package main

import (
	"context"

	"github.com/birabittoh/miri/internal/config"
	"github.com/birabittoh/miri/internal/deezer"
	"github.com/birabittoh/miri/internal/miri"
)

func Search(ctx context.Context, query string) []deezer.Song {
	// Mock implementation
	return []deezer.Song{
		{
			ID:     "3135556",
			Artist: "Daft Punk",
			Title:  "Harder, Better, Faster, Stronger",
		},
	}
}

func main() {
	ctx := context.Background()

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	query := "harder better faster stronger"
	res := Search(ctx, query)

	track := res[0]

	d := miri.New(cfg, "track")
	d.Logger.Infof("Downloading track: %s - %s", track.Artist, track.Title)

	err = d.Run(ctx, track.ID)
	if err != nil {
		panic(err)
	}
}
