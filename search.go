package miri

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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

// SearchOptions defines the parameters for searching tracks on Deezer.
type SearchOptions struct {
	Index  uint64 // Index of the first result to return (for pagination)
	Limit  uint64 // Maximum number of results to return
	Order  string // Order of results (RANKING, TRACK_ASC, TRACK_DESC, ARTIST_ASC, ARTIST_DESC, ALBUM_ASC, ALBUM_DESC, RATING_ASC, RATING_DESC, DURATION_ASC, DURATION_DESC)
	Strict bool   // Whether to perform a strict search ("on" or "off")
	Query  string // Search query (required)
}

const (
	deezerAPIBase  = "https://api.deezer.com/"
	endpointSearch = "search"
	endpointTrack  = "track"
	endpointAlbum  = "album"
)

var (
	defaultSearchOptions = SearchOptions{
		Index:  0,
		Limit:  25,
		Order:  "RANKING",
		Strict: false,
	}

	validSizes = map[string]bool{
		"small":  true,
		"medium": true,
		"big":    true,
		"xl":     true,
	}

	validOrders = map[string]bool{
		"RANKING":       true,
		"TRACK_ASC":     true,
		"TRACK_DESC":    true,
		"ARTIST_ASC":    true,
		"ARTIST_DESC":   true,
		"ALBUM_ASC":     true,
		"ALBUM_DESC":    true,
		"RATING_ASC":    true,
		"RATING_DESC":   true,
		"DURATION_ASC":  true,
		"DURATION_DESC": true,
	}
)

// Validate checks if the SearchOptions are valid and sets defaults where necessary.
func (opt *SearchOptions) Validate() error {
	if opt.Query == "" {
		return fmt.Errorf("search query cannot be empty")
	}

	if opt.Limit == 0 {
		opt.Limit = defaultSearchOptions.Limit
	}
	opt.Limit = min(opt.Limit, 100)

	if opt.Order == "" || !validOrders[opt.Order] {
		opt.Order = defaultSearchOptions.Order
	}

	return nil
}

// SearchTracks searches for tracks on Deezer matching the given query.
func (c *Client) SearchTracks(ctx context.Context, opt SearchOptions) (results []SongResult, err error) {
	err = opt.Validate()
	if err != nil {
		return nil, err
	}

	p := url.Values{}
	p.Set("output", "json")
	p.Set("q", opt.Query)
	p.Set("index", strconv.FormatUint(opt.Index, 10))
	p.Set("limit", strconv.FormatUint(opt.Limit, 10))
	p.Set("order", opt.Order)
	if opt.Strict {
		p.Set("strict", "on")
	} else {
		p.Set("strict", "off")
	}

	url := fmt.Sprintf("%s%s/%s?%s", deezerAPIBase, endpointSearch, endpointTrack, p.Encode())
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

// CoverURL returns the URL of the album cover image in the specified size.
// Valid sizes are "small", "medium", "big", and "xl".
func (s *SongResult) CoverURL(size string) string {
	if validSizes[size] {
		size = "?size=" + size
	} else {
		size = ""
	}
	return fmt.Sprintf("%s%s/%d/image%s", deezerAPIBase, endpointAlbum, s.Album.ID, size)
}
