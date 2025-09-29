package miri

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type SongResult struct {
	ID       int    `json:"id"`
	Readable bool   `json:"readable"`
	Title    string `json:"title"`
	Artist   struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"artist"`
	Album struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	} `json:"album"`
	Duration       int    `json:"duration"`
	ExplicitLyrics bool   `json:"explicit_lyrics"`
	Preview        string `json:"preview"`
	Link           string `json:"link"`
	Type           string `json:"type"`
}

type SearchResults struct {
	Data  []SongResult `json:"data"`
	Total int          `json:"total"`
	Next  string       `json:"next,omitempty"`
}

const (
	deezerAPIBase  = "https://api.deezer.com/"
	endpointSearch = "search"
	endpointTrack  = "track"
)

func (c *Client) SearchTracks(ctx context.Context, query string) (results []SongResult, err error) {
	url := fmt.Sprintf("%s%s/%s?q=%s&output=json", deezerAPIBase, endpointSearch, endpointTrack, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var searchResults SearchResults
	err = json.Unmarshal(body, &searchResults)
	if err != nil {
		return nil, err
	}

	return searchResults.Data, nil
}
