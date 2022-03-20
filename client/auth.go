package client

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

var ErrCredentials = errors.New("client authorization does not have valid client id or client secret")
var ErrCreateAuthRequest = errors.New("failed to generate a new http request for authorization")
var ErrRequest = errors.New("authorization request to spotify failed")
var ErrDecode = errors.New("failed to decode authorization from json")
var ErrEncode = errors.New("failed to encode authorization to json")
var ErrTokenDecode = errors.New("failed to decode authorization token from json")
var ErrTokenEncode = errors.New("failed to encode authorization token to json")
var ErrTokenMissing = errors.New("authorization token is empty after decode")

type Auth struct {
	Id           string    `json:"id"`
	Secret       string    `json:"secret"`
	AccessToken  string    `json:"access_token"`
	AuthorizedAt time.Time `json:"authorized_at"`
}

type AuthToken struct {
	AccessToken string `json:"access_token"`
}

// deserializes an auth token from json
func (at *AuthToken) FromJSON(r io.Reader) error {
	err := json.NewDecoder(r).Decode(at)
	if err != nil {
		return ErrTokenDecode
	}
	if at.AccessToken == "" {
		return ErrTokenMissing
	}
	return nil
}

// serializes an auth token to json
func (at *AuthToken) ToJSON(w io.Writer) error {
	err := json.NewEncoder(w).Encode(at)
	if err != nil {
		return ErrTokenEncode
	}
	return nil
}

// deserializes an auth struct from json
func (a *Auth) FromJSON(r io.Reader) error {
	err := json.NewDecoder(r).Decode(a)
	if err != nil {
		return ErrDecode
	}
	if a.AccessToken == "" {
		return ErrTokenMissing
	}
	return nil
}

// serializes an auth struct to json
func (a *Auth) ToJSON(w io.Writer) error {
	err := json.NewEncoder(w).Encode(a)
	if err != nil {
		return ErrEncode
	}
	return nil
}

// returns true if it has been more than 3200 seconds since last authorization
func (a *Auth) ShouldRefresh() bool {
	return (a.AuthorizedAt.After(a.AuthorizedAt.Add(time.Duration(3200) * time.Second)))
}

func NewAuth(id string, secret string) *Auth {

	if id == "" || secret == "" {
		return &Auth{
			Id:           "",
			Secret:       "",
			AccessToken:  "",
			AuthorizedAt: time.Time{},
		}
	} else {
		return &Auth{
			Id:           id,
			Secret:       secret,
			AccessToken:  "",
			AuthorizedAt: time.Time{},
		}
	}
}

func (a *Auth) NewAuthRequest() *http.Request {
	if a.Id == "" || a.Secret == "" {
		return nil
	} else {
		req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token?grant_type=client_credentials", nil)
		if err != nil {
			return nil
		}
		req.Header = map[string][]string{
			"Accept":       {"application/json"},
			"Content-Type": {"application/x-www-form-urlencoded"},
		}
		req.SetBasicAuth(a.Id, a.Secret)
		return req
	}
}

func (a *Auth) Authorize() {
	token := a.getToken()
	a.AccessToken = token
	a.AuthorizedAt = (time.Now())
}

func (a *Auth) MaybeRefresh() {
	if time.Now().Unix() > a.AuthorizedAt.Unix()+3200 {
		a.Authorize()
	}
}

func (a *Auth) getToken() string {
	req := a.NewAuthRequest()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer res.Body.Close()

	token := AuthToken{}
	err = token.FromJSON(res.Body)
	if err != nil {
		return ""
	}

	return token.AccessToken
}
