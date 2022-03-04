package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"genredetector/auth"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var ErrEnv = errors.New("client id and client secret environment variables missing")
var ErrCreateRequest = errors.New("failed to create a new search request")
var ErrGenreSearch = errors.New("failed doing genre search")
var ErrArtistSearch = errors.New("failed doing artist search")
var ErrArtistIdSearch = errors.New("failed doing artist id search")
var ErrTrackSearch = errors.New("failed doing track search")
var ErrTrackIdSearch = errors.New("failed doing track id search")

type Client struct {
	*http.Client
	*auth.Auth
	Lg *log.Logger
}

type SpotifyError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type SpotifyErrorResponse struct {
	*SpotifyError `json:"error"`
}

// serializes a response struct to json
func (ser *SpotifyErrorResponse) ToJSON(w io.Writer) error {
	err := json.NewEncoder(w).Encode(ser)
	if err != nil {
		return ErrEncodeArtistsResponse
	}
	return nil
}

// deserializes a spotify error response struct from json
func (ser *SpotifyErrorResponse) FromJSON(r io.Reader) error {
	err := json.NewDecoder(r).Decode(ser)
	if err != nil {
		return ErrDecodeArtistsResponse
	}
	return nil
}

func getRequestHeader(token string) map[string][]string {
	return map[string][]string{
		"Accept":        {"application/json"},
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + token},
	}
}

func (c *Client) NewGenreSearch(genre string, offset int) (*http.Request, error) {
	url := "https://api.spotify.com/v1/search?q=genre:" + genre + "&type=artist&limit=50&offset=" + fmt.Sprint(offset)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ErrCreateRequest
	}
	req.Header = getRequestHeader(c.AccessToken)
	return req, nil
}
func (c *Client) NewArtistSearch(artist string, offset int) (*http.Request, error) {
	url := "https://api.spotify.com/v1/search?q=artist:" + artist + "&type=artist&limit=50&offset=" + fmt.Sprint(offset)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ErrCreateRequest
	}
	req.Header = getRequestHeader(c.AccessToken)
	return req, nil
}

func (c *Client) NewArtistIdSearch(id string) (*http.Request, error) {
	url := "https://api.spotify.com/v1/artists/" + id
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ErrCreateRequest
	}
	req.Header = getRequestHeader(c.AccessToken)
	return req, nil
}

func (c *Client) NewTrackIdSearch(id string) (*http.Request, error) {
	url := "https://api.spotify.com/v1/tracks/" + id
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ErrCreateRequest
	}
	req.Header = getRequestHeader(c.AccessToken)
	return req, nil
}

func (c *Client) NewTrackSearch(track string, offset int) (*http.Request, error) {
	url := "https://api.spotify.com/v1/search?q=track:" + track + "&type=track&limit=50&offset=" + fmt.Sprint(offset)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ErrCreateRequest
	}
	req.Header = getRequestHeader(c.AccessToken)
	return req, nil
}

func (c *Client) ArtistIdSearch(r *http.Request) (*Artist, error) {
	c.Lg.Printf("\033[32m%s: \033[33m%s \033[0m \n", r.Method, r.URL)
	res, err := c.Do(r)
	if err != nil {
		return nil, ErrArtistIdSearch
	}
	defer res.Body.Close()

	sr := new(Artist)
	err = sr.FromJSON(res.Body)

	if err != nil {
		return nil, err
	}
	return sr, nil

}

// when Go adds generics in 1.18, these are all going to become a single function
func (c *Client) ArtistSearch(r *http.Request) (*ArtistsResponse, error) {
	c.Lg.Printf("\033[32m%s: \033[33m%s \033[0m \n", r.Method, r.URL)
	res, err := c.Do(r)
	if err != nil {
		return nil, ErrArtistSearch
	}
	defer res.Body.Close()
	sr := new(ArtistsResponse)
	err = sr.FromJSON(res.Body)
	if err != nil {
		return nil, err
	}
	return sr, nil
}

func (c *Client) GenreSearch(r *http.Request) (*ArtistsResponse, error) {
	c.Lg.Printf("\033[32m%s: \033[33m%s \033[0m \n", r.Method, r.URL)
	res, err := c.Do(r)
	if err != nil {
		return nil, ErrGenreSearch
	}
	defer res.Body.Close()

	sr := new(ArtistsResponse)
	err = sr.FromJSON(res.Body)

	if err != nil {
		return nil, err
	}
	return sr, nil
}

func (c *Client) TrackSearch(r *http.Request) (*TracksResponse, error) {
	c.Lg.Printf("\033[32m%s: \033[33m%s \033[0m \n", r.Method, r.URL)
	res, err := c.Do(r)
	if err != nil {
		return nil, ErrTrackSearch
	}

	defer res.Body.Close()
	sr := new(TracksResponse)
	err = sr.FromJSON(res.Body)
	if err != nil {
		return nil, ErrDecodeTrackResponse
	}

	return sr, nil
}

func (c *Client) TrackIdSearch(r *http.Request) (*Track, error) {
	c.Lg.Printf("\033[32m%s: \033[33m%s \033[0m \n", r.Method, r.URL)
	res, err := c.Do(r)
	if err != nil {
		return nil, ErrTrackSearch
	}
	defer res.Body.Close()
	sr := new(Track)
	err = sr.FromJSON(res.Body)
	if err != nil {
		return nil, ErrDecodeTrack
	}
	return sr, nil
}

func New() *Client {
	cid, csec := os.Getenv("SPOTIFY_CLIENT_ID"), os.Getenv("SPOTIFY_CLIENT_SECRET")
	if cid == "" || csec == "" {
		return nil
	}
	return &Client{
		Client: &http.Client{
			Timeout: time.Duration(10) * time.Second,
		},
		Auth: auth.New(),
		Lg:   log.Default(),
	}
}
