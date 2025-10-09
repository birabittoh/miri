package miri

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const lyricsURLFormat = "https://lyrics.lewdhutao.my.eu.org/v2/musixmatch/lyrics?title=%s&artist=%s"

type LyricsResponse struct {
	Data struct {
		ArtistName   string `json:"artistName"`
		TrackName    string `json:"trackName"`
		TrackID      string `json:"trackId"`
		SearchEngine string `json:"searchEngine"`
		ArtworkURL   string `json:"artworkUrl"`
		Lyrics       string `json:"lyrics"`
	} `json:"data"`
	Metadata struct {
		ApiVersion string `json:"apiVersion"`
	} `json:"metadata"`
}

func (s *SongResult) Lyrics(ctx context.Context) (string, error) {
	artist := url.QueryEscape(s.Artist.Name)
	title := url.QueryEscape(s.Title)
	lyricsURL := fmt.Sprintf(lyricsURLFormat, title, artist)

	req, err := http.NewRequestWithContext(ctx, "GET", lyricsURL, nil)
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

	if lyricsResp.Data.Lyrics == "" {
		return "", fmt.Errorf("no lyrics found for %s - %s", s.Artist.Name, s.Title)
	}

	return lyricsResp.Data.Lyrics, nil
}
