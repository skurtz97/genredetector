package main

import (
	"net/url"
	"strings"
)

type SearchType int

const (
	Genre SearchType = iota
	Artist
	ArtistId
	Track
	TrackId
)

func getTotal(total int) int {
	if total > 1000 {
		return 1000
	} else {
		return total
	}
}
func getNumRequests(total int) int {
	if (total / 50) > 19 {
		return 19
	} else {
		return (total / 50)
	}
}

func formatQueryString(genre string) string {
	genre = strings.Trim(genre, " +%20")
	if strings.ContainsAny(genre, " ") && !(strings.HasPrefix(genre, "\"") && strings.HasSuffix(genre, "\"")) {
		genre = "\"" + genre + "\""
	} else if !strings.ContainsAny(genre, " ") && (strings.HasPrefix(genre, "\"") && strings.HasSuffix(genre, "\"")) {
		genre = strings.Trim(genre, "\"")
	}
	genre = strings.ToLower(genre)
	genre = url.QueryEscape(genre)
	return genre
}
