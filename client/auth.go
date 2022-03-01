package client

import (
	"net/http"
	"os"
	"time"
)

func (c *SpotifyClient) newAuthRequest() *http.Request {
	req, err := http.NewRequest("POST", c.AuthUrl, nil)
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
