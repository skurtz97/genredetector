package client

import (
	"encoding/json"
	"errors"
	"io"
	"sort"
)

type TracksResponse struct {
	*TracksBody `json:"tracks"`
}

type TracksBody struct {
	Tracks []Track `json:"items"`
	Total  int     `json:"total"`
}

func (tb *TracksBody) ToJSON(w io.Writer) error {
	err := json.NewEncoder(w).Encode(tb)
	if err != nil {
		return ErrEncodeTracks
	}
	return nil
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

var ErrDecodeTrackResponse = errors.New("failed to decode track search response from json")
var ErrEncodeTrackResponse = errors.New("failed to encode track search response to json")
var ErrDecodeTracks = errors.New("failed to decode tracks from json")
var ErrEncodeTracks = errors.New("failed to encode tracks to json")
var ErrDecodeTrack = errors.New("failed to decode track from json")
var ErrEncodeTrack = errors.New("failed to encode track to json")

func (res *TracksResponse) FromJSON(r io.Reader) error {
	err := json.NewDecoder(r).Decode(res)
	if err != nil {
		return ErrDecodeTrackResponse
	}
	return nil
}
func (res *TracksResponse) ToJSON(w io.Writer) error {
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		return ErrEncodeTrackResponse
	}
	return nil
}

func (t *Track) FromJSON(r io.Reader) error {
	err := json.NewDecoder(r).Decode(t)
	if err != nil {
		return ErrDecodeTrack
	}
	return nil
}

func (t *Track) ToJSON(w io.Writer) error {
	err := json.NewEncoder(w).Encode(t)
	if err != nil {
		return ErrEncodeTrack
	}
	return nil
}

func TracksFromJSON(r io.Reader, ts []Track) error {
	err := json.NewDecoder(r).Decode(&ts)
	if err != nil {
		return ErrDecodeTracks
	}
	return nil
}

func TracksToJSON(w io.Writer, ts []Track) error {
	err := json.NewEncoder(w).Encode(ts)
	if err != nil {
		return ErrEncodeTracks
	}
	return nil
}
