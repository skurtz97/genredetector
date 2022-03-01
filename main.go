package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/skurtz97/app/client"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lg.Printf("\033[36m%s: \033[33m%s \033[0m", r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

type IncSearchRequest struct {
	Query  string
	Type   client.SearchType
	Limit  int
	Offset int
	Exact  bool
}

func (isr *IncSearchRequest) LogJSON() {
	err := json.NewEncoder(lg.Writer()).Encode(isr)
	if err != nil {
		lg.Println("failed encoding incoming search request json")
	}
}

func ParseSearchRequest(r *http.Request) (*IncSearchRequest, error) {
	defer r.Body.Close()
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	// limit and offset aren't really necessary and we can just use defaults
	q, t := r.FormValue("q"), r.FormValue("type")
	ex := r.FormValue("exact")
	limit := 50
	offset := 0

	q = strings.Trim(q, " ")
	var st client.SearchType
	if t == "" {
		st = client.ArtistSearch
	} else {
		st = client.ToSearchType(t)
	}

	var exact bool
	if ex == "false" {
		exact = false
	} else {
		exact = true
	}

	return &IncSearchRequest{
		Query:  q,
		Type:   st,
		Limit:  limit,
		Offset: offset,
		Exact:  exact,
	}, nil

}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	sreq, err := ParseSearchRequest(r)
	if err != nil {
		http.Error(w, "search failed: parse failed", http.StatusBadRequest)
	}

	// todo: benchmark this big allocation
	artists := make([]client.Item, 0, 1000)
	next, err := cl.Search(sreq.Query, sreq.Type, sreq.Limit, sreq.Offset)
	if err != nil {
		http.Error(w, "search failed", http.StatusBadRequest)
	}

	artists = append(artists, next.Items...)
	sreq.Offset += 50

	// we are limited to an offset of 950, but for some genres total > 950 + 50 (limit),
	// so we have to keep track of both maximums and break from the loop if our offset
	// is going to exceed either total or the offset limit
	for sreq.Offset <= next.Total && sreq.Offset <= 950 {
		next, err = cl.Search(sreq.Query, sreq.Type, sreq.Limit, sreq.Offset)
		if err != nil {
			http.Error(w, "search failed", http.StatusBadRequest)
		}
		artists = append(artists, next.Items...)
		sreq.Offset += 50
	}

	lg.Printf("sending %d/%d items to client", len(artists), next.Total)
	err = json.NewEncoder(w).Encode(artists)
	if err != nil {
		http.Error(w, "search failed", http.StatusBadRequest)
	}

}
func SearchHandlerAsync(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	req, err := ParseSearchRequest(r)
	if err != nil {
		http.Error(w, "search failed: parse failed", http.StatusBadRequest)
	}

	artists := make([]client.Item, 0, 1000)
	next, err := cl.Search(req.Query, req.Type, req.Limit, req.Offset)
	if err != nil {
		http.Error(w, "search failed", http.StatusBadRequest)
	}

	total := next.Total
	artists = append(artists, next.Items...)
	req.Offset += 50
	queue := make([]*IncSearchRequest, 0, 19)

	// we started at offset 50, so we add -1 to calculation of num of reqs until we hit total,
	// and we subtract one from total maximum requests before we hit offset limit.
	// # NEW REQUESTS 		= 			(total/ offset) - 1
	// # MAX NEW REQUESTS = 			(max offset / offset) - 1
	for i := 0; i <= ((total/50)-1) && (i <= 18); i++ {
		nreq := &IncSearchRequest{
			Query:  req.Query,
			Type:   req.Type,
			Limit:  req.Limit,
			Offset: (req.Offset + (i * 50)),
		}

		queue = append(queue, nreq)
	}
	/* extra logs for debugging request slice building
	 	fmt.Println("total: " + fmt.Sprint(total))
		fmt.Println("len(reqs): " + fmt.Sprint(len(reqs)))
		fmt.Println("offsets: ")
		for i := 0; i < len(reqs); i++ {
			fmt.Println(fmt.Sprint(reqs[i].Offset))
		}
	*/

	// send the requests concurrently
	// a waitgroup is just a counter that blocks at wg.Wait() until it reaches zero
	// decrements on each wg.Done()
	wg := sync.WaitGroup{}
	for i, r := range queue {
		wg.Add(1)
		go func(i int, r *IncSearchRequest) {
			res, err := cl.Search(r.Query, r.Type, r.Limit, r.Offset)
			if err != nil {
				lg.Printf("error making concurrent request #%d", i)
			}
			artists = append(artists, res.Items...)
			fmt.Printf("%d/%d response received\n", i, len(queue))
			wg.Done()
		}(i, r)
	}
	wg.Wait()

	lg.Printf("sending %d/%d items to client", len(artists), total)
	err = json.NewEncoder(w).Encode(artists)
	if err != nil {
		http.Error(w, "search failed", http.StatusBadRequest)
	}

}

var cl *client.SpotifyClient
var lg *log.Logger

const addr = "localhost:8080"

func init() {
	cl = client.New()
	lg = log.New(os.Stdout, "", log.Ltime)
}
func main() {
	rt := mux.NewRouter()

	rt.HandleFunc("/search", SearchHandler).Methods("GET").Headers("Content-Type", "application/json")
	rt.HandleFunc("/searchasync", SearchHandlerAsync).Methods("GET").Headers("Content-Type", "application/json")
	rt.Use(loggingMiddleware)

	srv := &http.Server{
		Handler:      rt,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  100 * time.Second,
	}

	lg.Printf("listening on %s", addr)
	lg.Fatal(srv.ListenAndServe())
}
