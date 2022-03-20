package client

import (
	"sort"
)

// response
type ArtistsResponse struct {
	*ArtistsBody `json:"artists"`
}

type ArtistsBody struct {
	//Href    string   `json:"href"`
	Artists []Artist `json:"items"`
	//Next    string   `json:"next"`
	//Limit   int      `json:"limit"`
	//Offset  int      `json:"offset"`
	Total  int `json:"total"`
	Length int `json:"length"`
}

type Artist struct {
	Name         string       `json:"name"`
	Followers    Followers    `json:"followers"`
	Popularity   int          `json:"popularity"`
	Genres       []string     `json:"genres"`
	ExternalUrls ExternalUrls `json:"external_urls"`
	Id           string       `json:"id"`
}

type ExternalUrls struct {
	Spotify string `json:"spotify"`
}

type Followers struct {
	Total int `json:"total"`
}

type ByArtist func(a1, a2 *Artist) bool

type ArtistsSorter struct {
	Artists []Artist
	By      ByArtist
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

func (by ByArtist) Sort(artists []Artist) {
	as := &ArtistsSorter{
		Artists: artists,
		By:      by,
	}
	sort.Sort(as)
}

// returns a new copy of artists but sorted on popularity in descending order
func SortArtists(artists []Artist) []Artist {
	popDesc := func(a1, a2 *Artist) bool {
		return a1.Popularity > a2.Popularity
	}
	ByArtist(popDesc).Sort(artists)
	return artists
}

// returns true if the genres slice contains the specified genre
func GenresContains(genres []string, genre string) bool {
	for i, g := range genres {
		if g == genre {
			temp := genres[0]
			genres[0] = genres[i]
			genres[i] = temp
			return true
		}
	}
	return false
}

// same as genres contains but we don't care about the boolean
// since this function always gets called by itself and not inside exact matches
func SortGenres(genre string, artists []Artist) {
	for i := range artists {
		for j := range artists[i].Genres {
			if artists[i].Genres[j] == genre {
				temp := artists[i].Genres[0]
				artists[i].Genres[0] = artists[i].Genres[j]
				artists[i].Genres[j] = temp
			}
		}
	}
}

// returns a new copy of artists that only includes artists that have a genre in Genres that exactly matches genre
func ExactMatches(genre string, artists []Artist) []Artist {
	exact := make([]Artist, 0, 1000)
	for i := range artists {
		if GenresContains(artists[i].Genres, genre) {
			exact = append(exact, artists[i])
		}
	}
	return exact
}
