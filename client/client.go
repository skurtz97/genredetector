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

var ErrCreateRequest = errors.New("failed to create a new search request")
var ErrGenreSearch = errors.New("failed performing genre search")

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
		return ErrEncodeResponse
	}
	return nil
}

// deserializes a spotify error response struct from json
func (ser *SpotifyErrorResponse) FromJSON(r io.Reader) error {
	err := json.NewDecoder(r).Decode(ser)
	if err != nil {
		return ErrDecodeResponse
	}
	return nil
}

func (c *Client) Authorize() error {
	token, err := c.GetToken()
	if err != nil {
		return err
	}
	c.AccessToken = token
	c.AuthorizedAt = (time.Now())
	return nil
}
func (c *Client) GetToken() (string, error) {
	req, err := c.NewAuthRequest()
	if err != nil {
		return "", err
	}

	res, err := c.Do(req)
	if err != nil {
		return "", auth.ErrRequest
	}
	defer res.Body.Close()

	token := new(auth.AuthToken)
	err = token.FromJSON(res.Body)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func New() (*Client, error) {
	cid, csec := os.Getenv("SPOTIFY_CLIENT_ID"), os.Getenv("SPOTIFY_CLIENT_SECRET")
	if cid == "" || csec == "" {
		fmt.Println("cid: " + cid + " csec: " + csec)
		return nil, auth.ErrEnv
	}
	return &Client{
		Client: &http.Client{
			Timeout: time.Duration(10) * time.Second,
		},
		Auth: &auth.Auth{
			Id:           cid,
			Secret:       csec,
			AccessToken:  "",
			AuthorizedAt: time.Time{},
		},
		Lg: log.Default(),
	}, nil
}

func (c *Client) NewGenreSearch(genre string, offset int) (*http.Request, error) {
	url := "https://api.spotify.com/v1/search?q=genre:" + genre + "&type=artist&limit=50&offset=" + fmt.Sprint(offset)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ErrCreateRequest
	}
	req.Header = map[string][]string{
		"Accept":        {"application/json"},
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + c.AccessToken},
	}

	return req, nil
}
func (c *Client) NewArtistSearch(artist string, offset int) (*http.Request, error) {
	url := "https://api.spotify.com/v1/search?q=artist:" + artist + "&type=artist&limit=50&offset=" + fmt.Sprint(offset)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ErrCreateRequest
	}
	req.Header = map[string][]string{
		"Accept":        {"application/json"},
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + c.AccessToken},
	}

	return req, nil
}

func (c *Client) NewArtistIdSearch(id string) (*http.Request, error) {
	url := "https://api.spotify.com/v1/artists/" + id
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ErrCreateRequest
	}
	req.Header = map[string][]string{
		"Accept":        {"application/json"},
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + c.AccessToken},
	}
	return req, nil
}

func (c *Client) ArtistIdSearch(r *http.Request) (*Artist, error) {
	c.Lg.Printf("\033[32m%s: \033[33m%s \033[0m \n", r.Method, r.URL)
	res, err := c.Do(r)
	if err != nil {
		return nil, ErrGenreSearch
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		c.Lg.Println("STATUS: " + res.Status)
	}

	sr := new(Artist)
	err = sr.FromJSON(res.Body)

	if err != nil {
		return nil, err
	}
	return sr, nil

}
func (c *Client) ArtistSearch(r *http.Request) (*ArtistsResponse, error) {
	c.Lg.Printf("\033[32m%s: \033[33m%s \033[0m \n", r.Method, r.URL)
	res, err := c.Do(r)
	if err != nil {
		return nil, ErrGenreSearch
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		c.Lg.Println("STATUS: " + res.Status)
	}
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

	if res.StatusCode != 200 {
		c.Lg.Println("STATUS: " + res.Status)
	}
	sr := new(ArtistsResponse)
	err = sr.FromJSON(res.Body)

	if err != nil {
		return nil, err
	}
	return sr, nil
}
