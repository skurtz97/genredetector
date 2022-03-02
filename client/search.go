package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Body struct {
	Href    string   `json:"href"`
	Artists []Artist `json:"items"`
	Next    string   `json:"next"`
	Limit   int      `json:"limit"`
	Offset  int      `json:"offset"`
	Total   int      `json:"total"`
}

type Artist struct {
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

type SearchResponse struct {
	*Body `json:"artists"`
}

// writes the search response object to json
func (sr *SearchResponse) ToJSON(w io.Writer) {
	err := json.NewEncoder(w).Encode(sr)
	if err != nil {
		fmt.Println("PrintJSON: failed to decode json")
	}
}

// reads the search response object from json
func (sr *SearchResponse) FromJSON(r io.Reader) {
	err := json.NewDecoder(r).Decode(sr)
	if err != nil {
		fmt.Println("FromJSON: failed to encode json")
	}
}

func ArtistsToJSON(w io.Writer, ar []*Artist) {
	err := json.NewEncoder(w).Encode(&ar)
	if err != nil {
		fmt.Println("PrintJSON: failed to decode json")
	}
}
func ArtistsFromJSON(r io.Reader, ar []*Artist) {
	err := json.NewDecoder(r).Decode(&ar)
	if err != nil {
		fmt.Println("PrintJSON: failed to decode json")
	}
}

type SearchType int

const (
	GenreSearch = iota
	ArtistSearch
	TrackSearch
)

func (st SearchType) String() string {
	return [...]string{"genre", "artist", "track"}[st]
}

func ToSearchType(st string) SearchType {
	switch st {
	case "genre":
		return GenreSearch
	case "artist":
		return ArtistSearch
	case "track":
		return TrackSearch
	default:
		return ArtistSearch
	}
}

var lg *log.Logger = log.New(os.Stdout, "", log.Ltime)

// returns the user input query string in a sanitized and spotify friendly format
// escapes all special characters, trims leading and trailing whitespace, adds
// quotations in the case of internal whitespace, and in the case of a genre search
// adds the genre field filter
func encodeQueryStr(query string, st SearchType) string {
	query = strings.Trim(query, " ")
	// if query has internal whitespace, enclose in quotations
	/*
		if strings.ContainsAny(query, " ") {
			query = fmt.Sprintf("\"%s\"", query)
		}
	*/

	// if we are searching by genre, convert to lowercase and add genre field filter
	if st == GenreSearch {
		return url.QueryEscape("genre:" + strings.ToLower(query))
	} else {
		return url.QueryEscape(query)
	}
}

// sends a search request to the spotify search endpoint
func (sc *SpotifyClient) Search(query string, st SearchType, limit int, offset int) (*SearchResponse, error) {
	if sc.shouldRefresh() {
		sc.authorize()
	}

	req, err := sc.newSearchRequest(query, st, limit, offset)
	if err != nil {
		panic("search: failed to build new search request")
	}

	lg.Printf("\033[32m%s: \033[33m%s \033[0m \n", req.Method, req.URL)
	res, err := sc.Do(req)
	if err != nil {
		return nil, err
	}
	//defer res.Body.Close()

	sr := new(SearchResponse)
	sr.FromJSON(res.Body)

	return sr, nil
}

// builds and then returns a pointer to a new search request
func (sc *SpotifyClient) newSearchRequest(query string, st SearchType, limit int, offset int) (*http.Request, error) {
	queryStr := encodeQueryStr(query, st)
	if st == GenreSearch {
		st = ArtistSearch
	}

	searchUrl, err := url.Parse(sc.SearchUrl + "q=" + queryStr + "&type=" + st.String() + "&limit=" + fmt.Sprint(limit) + "&offset=" + fmt.Sprint(offset))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", searchUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header = map[string][]string{
		"Accept":        {"application/json"},
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + sc.AccessToken},
	}
	return req, nil
}
