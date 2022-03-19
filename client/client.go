package client

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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
	lg *log.Logger
	*Auth
	config map[string]string
}

type RequestKind int

const (
	GENRE RequestKind = iota
	ARTIST
	ARTIST_ID
	TRACK
	TRACK_ID
)

func getRequestHeader(token string) map[string][]string {
	return map[string][]string{
		"Accept":        {"application/json"},
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + token},
	}
}

func (c *Client) NewSearch(query string, kind RequestKind, offset int) (*http.Request, error) {
	header := getRequestHeader(c.AccessToken)
	var req *http.Request
	var err error
	switch kind {
	case GENRE:
		req, err = http.NewRequest("GET", "https://api.spotify.com/v1/search?q=genre:"+query+"&type=artist&limit=50&offset="+fmt.Sprint(offset), nil)
		req.Header = header
	case ARTIST:
		req, err = http.NewRequest("GET", "https://api.spotify.com/v1/search?q=artist:"+query+"&type=artist&limit=50&offset="+fmt.Sprint(offset), nil)
		req.Header = header
	case ARTIST_ID:
		req, err = http.NewRequest("GET", "https://api.spotify.com/v1/artists/"+query, nil)
		req.Header = header
	case TRACK:
		req, err = http.NewRequest("GET", "https://api.spotify.com/v1/search?q=track:"+query+"&type=track&limit=50&offset="+fmt.Sprint(offset), nil)
		req.Header = header
	case TRACK_ID:
		req, err = http.NewRequest("GET", "https://api.spotify.com/v1/tracks/"+query, nil)
		req.Header = header
	}
	if err != nil {
		return nil, err
	}
	return req, nil
}

/*
func (c *Client) NewGenreSearch(genre string, offset int) (*http.Request, error) {
	url := "https://api.spotify.com/v1/search?q=genre:" + genre + "&type=artist&limit=50&offset=" + fmt.Sprint(offset)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ErrCreateAuthRequest
	}
	req.Header = getRequestHeader(c.AccessToken)
	return req, nil
}
func (c *Client) NewArtistSearch(artist string, offset int) (*http.Request, error) {
	url := "https://api.spotify.com/v1/search?q=artist:" + artist + "&type=artist&limit=50&offset=" + fmt.Sprint(offset)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ErrCreateAuthRequest
	}
	req.Header = getRequestHeader(c.AccessToken)
	return req, nil
}

func (c *Client) NewArtistIdSearch(id string) (*http.Request, error) {
	url := "https://api.spotify.com/v1/artists/" + id
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ErrCreateAuthRequest
	}
	req.Header = getRequestHeader(c.AccessToken)
	return req, nil
}

func (c *Client) NewTrackIdSearch(id string) (*http.Request, error) {
	url := "https://api.spotify.com/v1/tracks/" + id
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ErrCreateAuthRequest
	}
	req.Header = getRequestHeader(c.AccessToken)
	return req, nil
}

func (c *Client) NewTrackSearch(track string, offset int) (*http.Request, error) {
	url := "https://api.spotify.com/v1/search?q=track:" + track + "&type=track&limit=50&offset=" + fmt.Sprint(offset)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, ErrCreateAuthRequest
	}
	req.Header = getRequestHeader(c.AccessToken)
	return req, nil
}
*/

func (c *Client) ArtistIdSearch(r *http.Request) (*Artist, error) {
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

func (c *Client) ArtistSearch(r *http.Request) (*ArtistsResponse, error) {
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

func readConfig() map[string]string {
	file, err := os.Open(".env")
	if err != nil {
		panic("failed to open config")
	}
	config := make(map[string]string)

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	for i := range lines {
		splitStr := strings.Split(lines[i], "=")
		key, value := splitStr[0], splitStr[1]
		config[key] = value
	}

	return config
}

func New() *Client {
	config := readConfig()
	return &Client{
		Client: &http.Client{
			Timeout: time.Duration(10) * time.Second,
		},
		Auth:   NewAuth(config["CLIENT_ID"], config["CLIENT_SECRET"]),
		lg:     log.Default(),
		config: config,
	}
}

func (c *Client) SetLogger(l *log.Logger) {
	c.lg = l
}
