package server

import (
	"genredetector/client"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

var (
	c  *client.Client
	lg *log.Logger
)

func init() {
	c = client.New()
	if c == nil {
		panic("panicked because we couldn't initialize client\n")
	}
	lg = log.New(os.Stdout, "", log.Ltime)
	c.SetLogger(lg)
	c.Authorize()
	if c.AccessToken == "" {
		panic("panicked because we couldn't authorize client\n")
	}
}

func NewServer(addr string, dev bool) *http.Server {
	r := mux.NewRouter()
	r.Use()
	r.HandleFunc("/search/genre", GenreSearchHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/search/artist", ArtistSearchHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/search/artist/{id}", NewIdSearchHandler(Artist)).Methods("GET", "OPTIONS")
	r.HandleFunc("/search/track", TrackSearchHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/search/track/{id}", NewIdSearchHandler(Track)).Methods("GET", "OPTIONS")
	r.Use(middlewareCORS)

	s := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	return s
}
