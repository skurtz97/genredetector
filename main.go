package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"genredetector/client"

	"github.com/gorilla/mux"
)

var ErrParseForm = errors.New("error parsing incoming request form")

// package variables
var (
	clt *client.Client
	lg  *log.Logger
)

// panic if we can't initialize our client
// most likely Go can't find our id or secret, so check the environment variables
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

// we are good to go
func main() {
	var mode string
	if len(os.Args) == 2 {
		mode = os.Args[1]
		if mode != "-d" {
			mode = "-p"
		}
	} else {
		mode = "-p"
	}

	r := mux.NewRouter()

	if mode == "-d" {
		r.HandleFunc("/search/genre", GenreSearchHandler).Methods("GET", "OPTIONS")
		r.HandleFunc("/search/artist", ArtistSearchHandler).Methods("GET", "OPTIONS")
		r.HandleFunc("/search/artist/{id}", NewIdSearchHandler(ArtistId)).Methods("GET", "OPTIONS")
		r.HandleFunc("/search/track", TrackSearchHandler).Methods("GET", "OPTIONS")
		r.HandleFunc("/search/track/{id}", NewIdSearchHandler(TrackId)).Methods("GET", "OPTIONS")
		mux.CORSMethodMiddleware(r)
	} else {
		r.HandleFunc("/search/genre", GenreSearchHandler).Methods("GET")
		r.HandleFunc("/search/artist", ArtistSearchHandler).Methods("GET")
		r.HandleFunc("/search/artist/{id}", NewIdSearchHandler(ArtistId)).Methods("GET")
		r.HandleFunc("/search/track", TrackSearchHandler).Methods("GET")
		r.HandleFunc("/search/track/{id}", NewIdSearchHandler(TrackId)).Methods("GET")
	}

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
