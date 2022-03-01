package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
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

	// make first request
	next, err := cl.Search(sreq.Query, sreq.Type, sreq.Limit, sreq.Offset)
	if err != nil {
		http.Error(w, "search failed", http.StatusBadRequest)
	}
	artists = append(artists, next.Items...)
	sreq.Offset += 50

	// continue making requests sequentially until we are done
	for next.Offset < (next.Total - 50) {
		next, err = cl.Search(sreq.Query, sreq.Type, sreq.Limit, sreq.Offset)
		if err != nil {
			http.Error(w, "search failed", http.StatusBadRequest)
		}
		artists = append(artists, next.Items...)
		sreq.Offset += 50
	}

	lg.Printf("sending %d items to client", len(artists))
	err = json.NewEncoder(w).Encode(artists)
	if err != nil {
		http.Error(w, "search failed", http.StatusBadRequest)
	}

}

// GLOBAL SPOTIFY CLIENT, REUSE THIS BUT MAYBE MAKE MORE FOR PARALLEL REQUESTS ?
var cl *client.SpotifyClient
var lg *log.Logger

const addr = "localhost:8080"

////////////////////////////////////////////////////////////////////////////

func init() {
	cl = client.New()
	lg = log.New(os.Stdout, "", log.Ltime)
}
func main() {
	rt := mux.NewRouter()

	rt.HandleFunc("/search", SearchHandler).Methods("GET").Headers("Content-Type", "application/json")
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
