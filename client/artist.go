package client

import (
	"encoding/json"
	"errors"
	"io"
	"sort"
)

var ErrDecodeResponse = errors.New("failed to decode genre search response from json")
var ErrEncodeResponse = errors.New("failed to encode genre search response to json")
var ErrEncodeArtists = errors.New("failed to encode artists to json")

type Response struct {
	*Body `json:"artists"`
}

// deserializes a response struct from json
func (res *Response) FromJSON(r io.Reader) error {
	err := json.NewDecoder(r).Decode(res)
	if err != nil {
		return ErrDecodeResponse
	}
	return nil
}

// serializes a response struct to json
func (res *Response) ToJSON(w io.Writer) error {
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		return ErrEncodeResponse
	}
	return nil
}

type Body struct {
	Href    string   `json:"href"`
	Artists []Artist `json:"items"`
	Next    string   `json:"next"`
	Limit   int      `json:"limit"`
	Offset  int      `json:"offset"`
	Total   int      `json:"total"`
}

func ToJSON(w io.Writer, as []Artist) error {
	err := json.NewEncoder(w).Encode(as)
	if err != nil {
		return ErrEncodeArtists
	}
	return nil
}

type Artist struct {
	Name       string    `json:"name"`
	Followers  Followers `json:"followers"`
	Popularity int       `json:"popularity"`
	Genres     []string  `json:"genres"`
	Href       string    `json:"href"`
	Id         string    `json:"id"`
}

type Followers struct {
	Total int `json:"total"`
}

type By func(a1, a2 *Artist) bool

type ArtistsSorter struct {
	Artists []Artist
	By      func(a1, a2 *Artist) bool
}

func (as *ArtistsSorter) Len() int {
	return len(as.Artists)
}

func (as *ArtistsSorter) Swap(i, j int) {
	as.Artists[i], as.Artists[j] = as.Artists[j], as.Artists[i]
}

func (as *ArtistsSorter) Less(i, j int) bool {
	return as.By(&as.Artists[i], &as.Artists[j])
}

func (by By) Sort(artists []Artist) {
	as := &ArtistsSorter{
		Artists: artists,
		By:      by,
	}
	sort.Sort(as)
}

// returns true if the genres slice contains the specified genre
func GenresContains(genres []string, genre string) bool {
	for _, g := range genres {
		if g == genre {
			return true
		}
	}
	return false
}

// returns a new copy of artists but sorted on popularity in descending order
func SortArtists(artists []Artist) []Artist {
	popDesc := func(a1, a2 *Artist) bool {
		return a1.Popularity > a2.Popularity
	}
	By(popDesc).Sort(artists)
	return artists
}

// returns a new copy of artists that only includes artists that have a genre in Genres that exactly matches genre
func ExactMatches(genre string, artists []Artist) []Artist {
	exact := []Artist{}
	for i := range artists {
		if GenresContains(artists[i].Genres, genre) {
			exact = append(exact, artists[i])
		}
	}
	return exact
}
