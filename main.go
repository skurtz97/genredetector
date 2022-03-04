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

	"genredetector/client"

	"github.com/gorilla/mux"
)

var ErrParseForm = errors.New("error parsing incoming request form")

func FormatQueryString(genre string) string {
	genre = strings.Trim(genre, " +%20")
	if strings.ContainsAny(genre, " ") && !(strings.HasPrefix(genre, "\"") && strings.HasSuffix(genre, "\"")) {
		genre = "\"" + genre + "\""
	} else if !strings.ContainsAny(genre, " ") && (strings.HasPrefix(genre, "\"") && strings.HasSuffix(genre, "\"")) {
		genre = strings.Trim(genre, "\"")
	}
	genre = strings.ToLower(genre)
	genre = url.QueryEscape(genre)
	return genre
}

func ArtistIdSearchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	artistId := mux.Vars(r)["id"]
	artistId = strings.Trim(artistId, " ")

	req, err := clt.NewArtistIdSearch(artistId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	res, err := clt.ArtistIdSearch(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = res.ToJSON(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func ArtistSearchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	artistQueryStr := r.URL.Query().Get("q")
	artistQueryStr = FormatQueryString(artistQueryStr)

	artists := make([]client.Artist, 0, 300)
	req, err := clt.NewArtistSearch(artistQueryStr, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	res, err := clt.ArtistSearch(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	total := getTotal(res.Total)
	nreqs := getNumRequests(total)

	artists = append(artists, res.Artists...)
	requests := make([]*http.Request, nreqs)

	for i, offset := 0, 50; i <= nreqs; i++ {
		req, err = clt.NewArtistSearch(artistQueryStr, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		requests = append(requests, req)
		offset += 50
	}

	wg := sync.WaitGroup{}
	for i, req := range requests {
		wg.Add(1)
		go func(i int, req *http.Request) {
			res, err := clt.ArtistSearch(req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			for j, artist := range res.Artists {
				artists[50+(50*i)+j] = artist
			}
			wg.Done()
		}(i, req)
	}
	wg.Wait()

	artists = client.SortArtists(artists)
	lg.Printf("sending %d/%d artists to client", len(artists), total)

	err = client.ToJSON(w, artists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func getTotal(total int) int {
	if total > 1000 {
		return 1000
	} else {
		return total
	}
}
func getNumRequests(total int) int {
	if ((total / 50) - 1) > 19 {
		return 19
	} else {
		return ((total / 50) - 1)
	}
}

func GenreSearchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	query := FormatQueryString(r.URL.Query().Get("q"))
	genre, err := url.QueryUnescape(query)
	if err != nil {
		http.Error(w, "failed to unescape query string", http.StatusBadRequest)
	}
	genre, partial := strings.Trim(genre, "\""), r.URL.Query().Get("partial") == "true"

	req, err := clt.NewGenreSearch(query, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	res, err := clt.GenreSearch(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	total := getTotal(res.Total)
	nreqs := getNumRequests(res.Total)
	artists := make([]client.Artist, total)
	for i, artist := range res.Artists {
		artists[i] = artist
	}
	requests := make([]*http.Request, nreqs)

	for i, offset := 0, 50; i < nreqs; i++ {
		req, err = clt.NewGenreSearch(query, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		requests[i] = req
		offset += 50
	}

	wg := sync.WaitGroup{}
	for i, req := range requests {
		wg.Add(1)
		go func(i int, req *http.Request) {
			defer wg.Done()
			res, err := clt.GenreSearch(req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			// putting artist in explicit indexes is an alternative to using a mutex and appending
			// mutex.Lock(); artists = append(artists, res.Artists...); mutex.Unlock()
			for j, artist := range res.Artists {
				artists[50+(50*i)+j] = artist
			}
		}(i, req)
	}
	wg.Wait()

	if !partial {
		artists = client.ExactMatches(genre, artists)
	}
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
	clt = client.New()
	if clt == nil {
		panic("panicking because we couldn't initialize client\n")
	}
	lg = log.New(os.Stdout, "", log.Ltime)
	clt.Lg = lg
	clt.Authorize()
	if clt.AccessToken == "" {
		panic("panicing because we couldn't authorize client\n")
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/search/genre", GenreSearchHandler)
	r.HandleFunc("/search/artist", ArtistSearchHandler)
	r.HandleFunc("/search/artist/{id}", ArtistIdSearchHandler)
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
