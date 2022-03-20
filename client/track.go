package client

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
