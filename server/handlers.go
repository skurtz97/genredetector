package server

import (
	"fmt"
	"genredetector/client"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
)

func GenreSearchHandler(w http.ResponseWriter, r *http.Request) {
	c.MaybeRefresh()
	query := formatQueryString(r.URL.Query().Get("q"))
	genre, err := url.QueryUnescape(query)
	if err != nil {
		http.Error(w, "failed to unescape query string", http.StatusBadRequest)
	}
	genre, partial := strings.Trim(genre, "\""), r.URL.Query().Get("partial") == "true"

	req, err := c.NewGenreSearch(query, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	res, err := c.GenreSearch(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	total := getTotal(res.Total)
	nreqs := getNumRequests(res.Total)
	artists := make([]client.Artist, 0, total)
	artists = append(artists, res.Artists...)
	requests := make([]*http.Request, nreqs)

	for i, offset := 0, 50; i < nreqs; i++ {
		req, err = c.NewGenreSearch(query, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		requests[i] = req
		offset += 50
	}

	wg := sync.WaitGroup{}
	var m sync.Mutex
	for i, req := range requests {
		wg.Add(1)
		go func(i int, req *http.Request) {
			defer wg.Done()
			res, err := c.GenreSearch(req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			m.Lock()
			artists = append(artists, res.Artists...)
			m.Unlock()
		}(i, req)
	}
	wg.Wait()

	if !partial {
		artists = client.ExactMatches(genre, artists)
	} else {
		client.SortGenres(genre, artists)
	}

	artists = client.SortArtists(artists)
	lg.Printf("sending %d/%d artists to client", len(artists), total)

	err = jsoniter.NewEncoder(w).Encode(client.ArtistsBody{
		Total:   total,
		Length:  len(artists),
		Artists: artists,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ArtistSearchHandler(w http.ResponseWriter, r *http.Request) {
	c.MaybeRefresh()
	query := r.URL.Query().Get("q")
	query = formatQueryString(query)

	req, err := c.NewArtistSearch(query, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	res, err := c.ArtistSearch(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	total := getTotal(res.Total)
	nreqs := getNumRequests(total)
	artists := make([]client.Artist, 0, total)
	artists = append(artists, res.Artists...)
	requests := make([]*http.Request, nreqs)

	for i, offset := 0, 50; i < nreqs; i++ {
		req, err = c.NewArtistSearch(query, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		requests[i] = req
		offset += 50
	}

	lg.Printf("total: %d\t nreqs: %d\t len(artists): %d\t len(requests): %d\n", total, nreqs, len(artists), len(requests))
	wg := sync.WaitGroup{}
	var m sync.Mutex
	for i, req := range requests {
		wg.Add(1)
		go func(i int, req *http.Request) {
			defer wg.Done()
			res, err := c.ArtistSearch(req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			m.Lock()
			artists = append(artists, res.Artists...)
			m.Unlock()
		}(i, req)
	}
	wg.Wait()

	artists = client.SortArtists(artists)
	lg.Printf("sending %d/%d artists to client", len(artists), total)

	err = jsoniter.NewEncoder(w).Encode(client.ArtistsBody{
		Total:   total,
		Length:  len(artists),
		Artists: artists,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func TrackSearchHandler(w http.ResponseWriter, r *http.Request) {
	c.MaybeRefresh()
	query := formatQueryString(r.URL.Query().Get("q"))

	req, err := c.NewTrackSearch(query, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	res, err := c.TrackSearch(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	total := getTotal(res.Total)
	lg.Printf("total: %d", total)
	nreqs := getNumRequests(res.Total)
	tracks := make([]client.Track, 0, total)
	tracks = append(tracks, res.Tracks...)
	requests := make([]*http.Request, nreqs)

	for i, offset := 0, 50; i < nreqs; i++ {
		req, err = c.NewTrackSearch(query, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		requests[i] = req
		offset += 50
	}

	lg.Printf("\ntotal: %d\t nreqs: %d\t len(tracks): %d\t len(requests): %d", total, nreqs, len(tracks), len(requests))
	wg := sync.WaitGroup{}
	var m sync.Mutex
	for i, req := range requests {
		wg.Add(1)
		go func(i int, req *http.Request) {
			defer wg.Done()
			res, err := c.TrackSearch(req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			m.Lock()
			tracks = append(tracks, res.Tracks...)
			m.Unlock()

		}(i, req)
	}
	wg.Wait()

	tracks = client.SortTracks(tracks)
	lg.Printf("sending %d/%d artists to client", len(tracks), total)
	body := client.TracksBody{
		Total:  total,
		Length: len(tracks),
		Tracks: tracks,
	}
	err = body.ToJSON(w)

}

func ArtistIdSearchHandler(w http.ResponseWriter, r *http.Request) {
	c.MaybeRefresh()
	id := mux.Vars(r)["id"]
	id = strings.Trim(id, " ")

	req, err := c.NewArtistIdSearch(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	res, err := c.ArtistIdSearch(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = res.ToJSON(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func TrackIdSearchHandler(w http.ResponseWriter, r *http.Request) {
	c.MaybeRefresh()
	id := mux.Vars(r)["id"]
	id = strings.Trim(id, " ")

	req, err := c.NewTrackIdSearch(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	res, err := c.TrackIdSearch(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	err = res.ToJSON(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func NewIdSearchHandler(t SearchType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.MaybeRefresh()
		id := mux.Vars(r)["id"]
		fmt.Println(id)
		id = strings.Trim(id, " ")
		switch t {
		case Artist:
			req, _ := c.NewArtistIdSearch(id)
			res, _ := c.ArtistIdSearch(req)
			_ = res.ToJSON(w)
		case Track:
			req, _ := c.NewTrackIdSearch(id)
			res, _ := c.TrackIdSearch(req)
			_ = res.ToJSON(w)
		default:
			http.Error(w, "invalid search type", http.StatusBadRequest)
		}
	}
}
