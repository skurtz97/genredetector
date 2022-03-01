package client

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (ar *AuthResponse) FromJSON(r io.Reader) {
	err := json.NewDecoder(r).Decode(ar)
	if err != nil {
		fmt.Println("FromJSON: failed to encode json")
	}
}

func (ar *AuthResponse) ToJSON(w io.Writer) {
	err := json.NewEncoder(w).Encode(ar)
	if err != nil {
		fmt.Println("PrintJSON: failed to decode json")
	}
}

// context is just a bunch of helpful stuff that our client will need that could maybe also be global variables or hard coded constants,
// but we don't want to polute the namespace in case the app ends up using multiple custom clients for different apis
type Context struct {
	AccessToken  string `json:"access_token"`
	AuthorizedAt int64  `json:"authorized_at"`
	AuthUrl      string `json:"auth_url"`
	SearchUrl    string `json:"search_url"`
}

func newContext() *Context {
	return &Context{
		AccessToken:  "",
		AuthorizedAt: time.Now().Unix(),
		AuthUrl:      "https://accounts.spotify.com/api/token?grant_type=client_credentials",
		SearchUrl:    "https://api.spotify.com/v1/search?",
	}
}

type Artists struct {
	Href   string `json:"href"`
	Items  []Item `json:"items"`
	Next   string `json:"next"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Total  int    `json:"total"`
}

type Item struct {
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

type SearchResponse struct {
	*Artists `json:"artists"`
}

// writes the search response object to json
func (sr *SearchResponse) ToJSON(w io.Writer) {
	err := json.NewEncoder(w).Encode(sr)
	if err != nil {
		fmt.Println("PrintJSON: failed to decode json")
	}
}

// reads the search response object from json
func (sr *SearchResponse) FromJSON(r io.Reader) {
	err := json.NewDecoder(r).Decode(sr)
	if err != nil {
		fmt.Println("FromJSON: failed to encode json")
	}
}

type SearchType int

const (
	GenreSearch = iota
	ArtistSearch
	TrackSearch
)

func (st SearchType) String() string {
	return [...]string{"genre", "artist", "track"}[st]
}

func ToSearchType(st string) SearchType {
	switch st {
	case "genre":
		return GenreSearch
	case "artist":
		return ArtistSearch
	case "track":
		return TrackSearch
	default:
		return ArtistSearch
	}
}
