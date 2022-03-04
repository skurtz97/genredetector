package client

import (
	"encoding/json"
	"errors"
	"io"
)

type TracksResponse struct {
	*TracksBody `json:"tracks"`
}

type TracksBody struct {
	Href   string  `json:"href"`
	Tracks []Track `json:"items"`
	Next   string  `json:"next"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
	Total  int     `json:"total"`
}

type Track struct {
	Album   Album         `json:"album"`
	Artists []TrackArtist `json:"artists"`
}

type Album struct {
	Href        string `json:"href"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	ReleaseDate string `json:"release_date"`
}

type TrackArtist struct {
	Name string `json:"name"`
	Href string `json:"href"`
}

var ErrDecodeTrackResponse = errors.New("failed to decode track search response from json")
var ErrEncodeTrackResponse = errors.New("failed to encode track search response to json")
var ErrDecodeTracks = errors.New("failed to decode tracks from json")
var ErrEncodeTracks = errors.New("failed to encode tracks to json")

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
