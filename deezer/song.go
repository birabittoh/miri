package deezer

import (
	"encoding/json"
	"fmt"
)

type Contributors struct {
	MainArtists []string `json:"main_artist"`
	Composers   []string `json:"composer"`
	Authors     []string `json:"author"`
}

func (c *Contributors) UnmarshalJSON(data []byte) error {
	if string(data) == "[]" {
		*c = Contributors{}
		return nil
	}

	type Alias Contributors
	aux := (*Alias)(c)

	return json.Unmarshal(data, aux)
}

type Song struct {
	ID           string       `json:"SNG_ID"`
	Artist       string       `json:"ART_NAME"`
	Title        string       `json:"SNG_TITLE"`
	Version      string       `json:"VERSION"`
	Cover        string       `json:"ALB_PICTURE"`
	Contributors Contributors `json:"SNG_CONTRIBUTORS"`
	Duration     string       `json:"DURATION"`
	Gain         string       `json:"GAIN"`
	ISRC         string       `json:"ISRC"`
	TrackNumber  string       `json:"TRACK_NUMBER"`
	TrackToken   string       `json:"TRACK_TOKEN"`
}

func (s *Song) GetTitle() string {
	songTitle := s.Title
	if s.Version != "" {
		songTitle = fmt.Sprintf("%s %s", s.Title, s.Version)
	}

	return songTitle
}
