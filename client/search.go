package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type SearchType int

const (
	GenreSearch = iota
	ArtistSearch
	TrackSearch
)

func (st SearchType) String() string {
	return [...]string{"genre", "artist", "track"}[st]
}

type SearchResponse struct {
	Response *Artists `json:"artists"`
}

func (sr *SearchResponse) ToJSON(w io.Writer) {
	err := json.NewEncoder(w).Encode(sr)
	if err != nil {
		fmt.Println("PrintJSON: failed to decode json")
	}
}

func (sr *SearchResponse) FromJSON(r io.Reader) {
	err := json.NewDecoder(r).Decode(sr)
	if err != nil {
		fmt.Println("FromJSON: failed to encode json")
	}
}

func (sr *SearchResponse) PrintJSON(w io.Writer) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err := enc.Encode(sr)
	if err != nil {
		fmt.Println("FromJSON: failed to print json")
	}
}

// returns the user input query string in a sanitized and spotify friendly format
// this escapes all special characters, trims leading and trailing whitespace, adds
// quotations in the case of internal whitespace, and in the case of a genre search
// adds the genre field filter
func encodeQueryStr(query string, st SearchType) string {
	// if query has any leading or trailing whitespace, remove it
	query = strings.Trim(query, " ")

	// if query has internal whitespace, enclose in quotations
	if strings.ContainsAny(query, " ") {
		query = fmt.Sprintf("\"%s\"", query)
	}

	// if we are searching by genre, convert to lowercase and add genre field filter
	if st == GenreSearch {
		return url.QueryEscape("genre:" + strings.ToLower(query))
	} else {
		return url.QueryEscape(query)
	}
}

func (c *SpotifyClient) newSearchRequest(query string, st SearchType, limit int, offset int) *http.Request {
	queryStr := encodeQueryStr(query, st)
	if st == GenreSearch {
		st = ArtistSearch
	}

	searchUrl, err := url.Parse(c.ctx.SearchUrl + "q=" + queryStr + "&type=" + st.String() + "&limit=" + fmt.Sprint(limit) + "&offset=" + fmt.Sprint(offset))
	if err != nil {
		panic("NewSearchRequest: failed to parse url")
	}

	req, err := http.NewRequest("GET", searchUrl.String(), nil)
	if err != nil {
		panic("authorize: failed to build search request")
	}
	req.Header = map[string][]string{
		"Accept":        {"application/json"},
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + c.ctx.AccessToken},
	}
	return req
}

func (sc *SpotifyClient) Search(query string, st SearchType, limit int, offset int) error {
	if sc.shouldRefresh() {
		sc.authorize()
	}

	req := sc.newSearchRequest(query, st, limit, offset)
	res, err := sc.hc.Do(req)
	if err != nil {
		fmt.Println(errors.New("search: search request failed"))
	}
	defer res.Body.Close()

	sr := new(SearchResponse)
	sr.FromJSON(res.Body)
	sr.PrintJSON(log.Writer())

	return nil
}

type Artists struct {
	Href   string  `json:"href"`
	Items  []*Item `json:"items"`
	Next   string  `json:"next"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
	Total  int     `json:"total"`
}

type Item struct {
	Name       string    `json:"name"`
	Followers  Followers `json:"followers"`
	Popularity int       `json:"popularity"`
	Genres     []string  `json:"genres"`
	Href       string    `json:"href"`
	Id         string    `json:"id"`
}

type Followers struct {
	Total int `json:"total"`
}
