package client

import (
	"net/http"
	"time"
)

// wrapper for an http SpotifyClient that includes all the "global" state we will need
// stored in a Context that we pass in here and to all other instances of SpotifyClient.
// the Context is primarily for authorization purposes
type SpotifyClient struct {
	ctx *Context
	hc  *http.Client
}

// returns a new SpotifyClient that has already
// requested and received an auth token
// from spotify
func NewSpotifyClient() *SpotifyClient {
	ctx := newContext()
	SpotifyClient := &SpotifyClient{
		ctx: ctx,
		hc: &http.Client{
			Timeout: time.Duration(20) * time.Second,
		},
	}

	SpotifyClient.authorize()
	return SpotifyClient
}

// everything our SpotifyClient needs to know about can be contained in this Context instead
// making a bunch of globals
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
