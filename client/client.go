package client

import (
	"net/http"
	"time"
)

// wrapper for an http SpotifyClient that includes all the "global" state we will need
// stored in a Context that we pass in here and to all other instances of SpotifyClient.
// the Context is primarily for authorization purposes
type SpotifyClient struct {
	*http.Client
	*Context
}

// initializes, authorizes, and returns a new spotify client
func New() *SpotifyClient {
	ctx := newContext()
	SpotifyClient := &SpotifyClient{
		Client: &http.Client{
			Timeout: time.Duration(10) * time.Second,
		},
		Context: ctx,
	}

	SpotifyClient.authorize()
	return SpotifyClient
}
