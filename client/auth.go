package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

func (sc *SpotifyClient) newAuthRequest() *http.Request {
	req, err := http.NewRequest("POST", sc.AuthUrl, nil)
	if err != nil {
		panic("authorize: failed to build authorization request")
	}
	req.Header = map[string][]string{
		"Accept":       {"application/json"},
		"Content-Type": {"application/x-www-form-urlencoded"},
	}

	// if we don't have our environment variables then its time to panic
	cid, csec := os.Getenv("SPOTIFY_CLIENT_ID"), os.Getenv("SPOTIFY_CLIENT_SECRET")
	if cid == "" || csec == "" {
		panic("application not able to find environment varibles SPOTIFY_CLIENT_ID or SPOTIFY_CLIENT_SECRET")
	}

	req.SetBasicAuth(os.Getenv("SPOTIFY_CLIENT_ID"), os.Getenv("SPOTIFY_CLIENT_SECRET"))

	return req
}

// performs the SpotifyClient credentials authorization flow
// runs automatically at the end of NewSpotifyClient, and at the beginning
// of other request methods if we detect that the current
// authorization token is close to expiry or expired
func (sc *SpotifyClient) authorize() {

	req := sc.newAuthRequest()
	res, err := sc.Do(req)
	if err != nil {
		panic("authorize: authorization request failed")
	}
	defer res.Body.Close()
	// decode the response
	resJson := &AuthResponse{}
	resJson.FromJSON(res.Body)

	sc.AccessToken = resJson.AccessToken
	sc.AuthorizedAt = time.Now().Unix()
}

// returns true if we have less than two minutes left of time on our authorization
func (c *SpotifyClient) shouldRefresh() bool {
	return ((c.AuthorizedAt + (3600 - 120)) <= time.Now().Unix())
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
