package main

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"app/client"

	"github.com/gorilla/mux"
)

var ErrParseForm = errors.New("error parsing incoming request form")

func FormatGenre(genre string) string {
	if strings.ContainsAny(genre, " ") && !(strings.HasPrefix(genre, "\"") && strings.HasSuffix(genre, "\"")) {
		genre = "\"" + genre + "\""
	} else if !strings.ContainsAny(genre, " ") && (strings.HasPrefix(genre, "\"") && strings.HasSuffix(genre, "\"")) {
		genre = strings.Trim(genre, "\"")
	}
	genre = strings.ToLower(genre)
	genre = url.QueryEscape(genre)
	return genre
}
func GenreSearchHandler(w http.ResponseWriter, r *http.Request) {
	genreQueryStr := FormatGenre(r.URL.Query().Get("q"))
	genre, err := url.QueryUnescape(genreQueryStr)
	if err != nil {
		lg.Println("failed to unescape genre query string")
	}
	genre = strings.Trim(genre, "\"")
	/*
		partial := false
		if r.URL.Query().Get("partial") == "true" {
			partial = true
		}
	*/
	artists := make([]client.Artist, 0, 1000)

	req, err := clt.NewGenreRequest(genreQueryStr, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	res, err := clt.GenreSearch(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	total := res.Total
	// max offset will always be 950 due to spotify limitations
	if total > 1000 {
		total = 1000
	}
	offset := 50

	artists = append(artists, res.Artists...)
	queue := make([]*http.Request, 0, 19)

	for i := 0; i <= ((total/50)-1) && (i <= 18); i++ {
		req, err = clt.NewGenreRequest(genreQueryStr, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		queue = append(queue, req)
		offset += 50
	}

	wg := sync.WaitGroup{}
	for i, req := range queue {
		wg.Add(1)
		go func(i int, req *http.Request) {
			res, err := clt.GenreSearch(req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			artists = append(artists, res.Artists...)
			wg.Done()
		}(i, req)
	}
	wg.Wait()

	artists = client.ExactMatches(genre, artists)
	artists = client.SortArtists(artists)
	lg.Printf("sending %d/%d artists to client", len(artists), total)

	err = client.ToJSON(w, artists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var clt *client.Client
var lg *log.Logger

func init() {
	var err error
	clt, err = client.New()
	if err != nil {
		panic(err.Error() + "\n" + "panicking because we couldn't initialize client")
	}
	lg = log.New(os.Stdout, "", log.Ltime)
	clt.SetLogger(lg)
	err = clt.Authorize()
	if err != nil {
		panic(err.Error() + "\n" + "panicing because we couldn't authorize client")
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/artists/genre", GenreSearchHandler).Methods("GET").Headers("Content-Type", "application/json")

	s := &http.Server{
		Handler:      r,
		Addr:         "localhost:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	lg.Printf("listening on %s", s.Addr)
	lg.Fatal(s.ListenAndServe())
}
