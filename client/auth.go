package client

import (
	"genredetector/util"
	"net/http"
	"time"
)

type Auth struct {
	Id           string    `json:"id"`
	Secret       string    `json:"secret"`
	AccessToken  string    `json:"access_token"`
	AuthorizedAt time.Time `json:"authorized_at"`
}

type AuthToken struct {
	AccessToken string `json:"access_token"`
}

// returns true if it has been more than 3200 seconds since last authorization
func (a *Auth) ShouldRefresh() bool {
	return (a.AuthorizedAt.After(a.AuthorizedAt.Add(time.Duration(3200) * time.Second)))
}

func NewAuth(id string, secret string) *Auth {
	if id == "" || secret == "" {
		panic("id or secret missing")
	}
	return &Auth{
		Id:           id,
		Secret:       secret,
		AccessToken:  "",
		AuthorizedAt: time.Time{},
	}

}

func (a *Auth) NewAuthRequest() *http.Request {
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
	err = util.FromJSON(res.Body, &token)
	if err != nil {
		return ""
	}

	return token.AccessToken
}
