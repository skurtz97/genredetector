package main

import (
	"genredetector/client"
	"net/http"
	"sync"
	"testing"
)

func TestGenreSearchHandler(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "correct formatting exact", query: "soft rock"},
		{name: "correct formatting partial", query: "soft rock"},
		{name: "leading space exact", query: "  soft rock"},
		{name: "leading space partial", query: "   soft rock"},
		{name: "leading/trailing/internal space", query: "  soft rock  "},
	}
	for _, tt := range tests {
		clt := client.New()
		clt.Authorize()

		t.Run(tt.name, func(t *testing.T) {
			tt.query = formatQueryString(tt.query)

			req, err := clt.NewGenreSearch(tt.query, 0)
			if err != nil {
				t.Errorf("failed to create new genre search request")
			}

			res, err := clt.GenreSearch(req)
			if err != nil {
				t.Errorf("error doing genre search")
			}

			total := getTotal(res.Total)
			nreqs := getNumRequests(res.Total)
			artists := make([]client.Artist, 0, total)
			artists = append(artists, res.Artists...)
			requests := make([]*http.Request, nreqs)

			for i, offset := 0, 50; i < nreqs; i++ {
				req, err = clt.NewGenreSearch(tt.query, offset)
				if err != nil {
					t.Errorf("failed to create new genre search request")
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
					res, err := clt.GenreSearch(req)
					if err != nil {
						t.Errorf("genre search failed")
					}
					// putting artist in explicit indexes is an alternative to using a mutex and appending
					// mutex.Lock(); artists = append(artists, res.Artists...); mutex.Unlock()
					m.Lock()
					artists = append(artists, res.Artists...)
					m.Unlock()

				}(i, req)
			}
			wg.Wait()

			artists = client.SortArtists(artists)
			lg.Printf("sending %d/%d artists to client", len(artists), total)

			for i, a := range artists {
				if a.Name == "" {
					t.Errorf("aritsts[%d] == \"\", something went wrong\n", i)
				}
			}

			if len(artists) != total {
				t.Errorf("len(artists) != total")
			}

		})

	}
}
