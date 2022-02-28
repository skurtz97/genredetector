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

func (c *SpotifyClient) newAuthRequest() *http.Request {
	req, err := http.NewRequest("POST", c.ctx.AuthUrl, nil)
	if err != nil {
		panic("authorize: failed to build authorization request")
	}
	req.Header = map[string][]string{
		"Accept":       {"application/json"},
		"Content-Type": {"application/x-www-form-urlencoded"},
	}
	req.SetBasicAuth(os.Getenv("SPOTIFY_CLIENT_ID"), os.Getenv("SPOTIFY_CLIENT_SECRET"))
	return req
}

// performs the SpotifyClient credentials authorization flow
// runs automatically at the end of NewSpotifyClient, and at the beginning
// of other request methods if we detect that the current
// authorization token is close to expiry or expired
func (c *SpotifyClient) authorize() {

	req := c.newAuthRequest()
	res, err := c.hc.Do(req)
	if err != nil {
		panic("authorize: authorization request failed")
	}
	defer res.Body.Close()
	// decode the response
	resJson := &AuthResponse{}
	resJson.FromJSON(res.Body)

	c.ctx.AccessToken = resJson.AccessToken
	c.ctx.AuthorizedAt = time.Now().Unix()
}

// returns true if we have less than two minutes left of time on our authorization
func (c *SpotifyClient) shouldRefresh() bool {
	return ((c.ctx.AuthorizedAt + (3600 - 120)) <= time.Now().Unix())
}
