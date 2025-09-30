package miri

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const lyricsURLFormat = "https://api.lyrics.ovh/v1/%s/%s"

type LyricsResponse struct {
	Lyrics string `json:"lyrics"`
}

func (s *SongResult) Lyrics() (string, error) {
	artist := url.QueryEscape(s.Artist.Name)
	title := url.QueryEscape(s.Title)
	lyricsURL := fmt.Sprintf(lyricsURLFormat, artist, title)

	req, err := http.NewRequest("GET", lyricsURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch lyrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch lyrics: status code %d", resp.StatusCode)
	}

	var lyricsResp LyricsResponse
	if err := json.NewDecoder(resp.Body).Decode(&lyricsResp); err != nil {
		return "", fmt.Errorf("failed to decode lyrics response: %w", err)
	}

	if lyricsResp.Lyrics == "" {
		return "", fmt.Errorf("no lyrics found for %s - %s", s.Artist.Name, s.Title)
	}

	return lyricsResp.Lyrics, nil
}
