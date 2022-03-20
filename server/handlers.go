package server

import (
	"fmt"
	"genredetector/client"
	"genredetector/util"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
)

type SearchKind int

const (
	Genre SearchKind = iota
	Artist
	Track
)

func getTotal(total int) int {
	if total > 1000 {
		return 1000
	} else {
		return total
	}
}
func getNumRequests(total int) int {
	if (total / 50) > 19 {
		return 19
	} else {
		return (total / 50)
	}
}

func GenreSearchHandler(w http.ResponseWriter, r *http.Request) {
	c.MaybeRefresh()
	query := formatQueryString(r.URL.Query().Get("q"))
	genre, err := url.QueryUnescape(query)
	if err != nil {
		http.Error(w, "failed to unescape query string", http.StatusBadRequest)
	}
	genre, partial := strings.Trim(genre, "\""), r.URL.Query().Get("partial") == "true"

	req, err := c.NewSearch(query, client.GENRE, 0)
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
		req, err = c.NewSearch(query, client.GENRE, offset)
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

	sort.Slice(artists, func(i, j int) bool {
		return artists[i].Popularity > artists[j].Popularity
	})
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

func formatQueryString(genre string) string {
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

func ArtistSearchHandler(w http.ResponseWriter, r *http.Request) {
	c.MaybeRefresh()
	query := r.URL.Query().Get("q")
	query = formatQueryString(query)

	req, err := c.NewSearch(query, client.ARTIST, 0)
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
		req, err = c.NewSearch(query, client.ARTIST, offset)
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

	sort.Slice(artists, func(i, j int) bool {
		return artists[i].Popularity > artists[j].Popularity
	})

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

	req, err := c.NewSearch(query, client.TRACK, 0)
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
		req, err = c.NewSearch(query, client.TRACK, offset)
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
	err = util.ToJSON(w, &body)

}

func ArtistIdSearchHandler(w http.ResponseWriter, r *http.Request) {
	c.MaybeRefresh()
	id := mux.Vars(r)["id"]
	id = strings.Trim(id, " ")

	req, err := c.NewSearch(id, client.ARTIST_ID, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	res, err := c.ArtistIdSearch(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = util.ToJSON(w, res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func TrackIdSearchHandler(w http.ResponseWriter, r *http.Request) {
	c.MaybeRefresh()
	id := mux.Vars(r)["id"]
	id = strings.Trim(id, " ")

	req, err := c.NewSearch(id, client.TRACK_ID, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	res, err := c.TrackIdSearch(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	err = util.ToJSON(w, res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func NewIdSearchHandler(kind SearchKind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.MaybeRefresh()
		id := mux.Vars(r)["id"]
		fmt.Println(id)
		id = strings.Trim(id, " ")
		switch kind {
		case Artist:
			req, _ := c.NewSearch(id, client.ARTIST_ID, 0)
			res, _ := c.ArtistIdSearch(req)
			_ = util.ToJSON(w, &res)
		case Track:
			req, _ := c.NewSearch(id, client.TRACK_ID, 0)
			res, _ := c.TrackIdSearch(req)
			_ = util.ToJSON(w, &res)
		default:
			http.Error(w, "invalid search type", http.StatusBadRequest)
		}
	}
}
