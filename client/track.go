package client

import (
	"sort"
)

type TracksResponse struct {
	*TracksBody `json:"tracks"`
}

type TracksBody struct {
	Tracks []Track `json:"items"`
	Total  int     `json:"total"`
	Length int     `json:"length"`
}

type Track struct {
	Name         string        `json:"name"`
	Album        Album         `json:"album"`
	Artists      []TrackArtist `json:"artists"`
	ExternalUrls ExternalUrls  `json:"external_urls"`
	Popularity   int           `json:"popularity"`
	ReleaseDate  string        `json:"release_date"`
}

type Album struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	ReleaseDate string `json:"release_date"`
}

type TrackArtist struct {
	Name string `json:"name"`
}

type ByTrack func(t1, t2 *Track) bool

type TracksSorter struct {
	Tracks []Track
	By     ByTrack
}

func (ts *TracksSorter) Len() int {
	return len(ts.Tracks)
}

func (ts *TracksSorter) Swap(i, j int) {
	ts.Tracks[i], ts.Tracks[j] = ts.Tracks[j], ts.Tracks[i]
}

func (ts *TracksSorter) Less(i, j int) bool {
	return ts.By(&ts.Tracks[i], &ts.Tracks[j])
}

func (by ByTrack) Sort(tracks []Track) {
	ts := &TracksSorter{
		Tracks: tracks,
		By:     by,
	}
	sort.Sort(ts)
}

// returns a new copy of artists but sorted on popularity in descending order
func SortTracks(tracks []Track) []Track {
	popDesc := func(t1, t2 *Track) bool {
		return t1.Popularity > t2.Popularity
	}
	ByTrack(popDesc).Sort(tracks)
	return tracks
}
