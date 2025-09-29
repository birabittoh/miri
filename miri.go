package miri

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/birabittoh/miri/deezer"
)

const chunkSize = 2048

type Client struct {
	appConfig    *deezer.Config
	deezerClient *deezer.Client
}

func New(ctx context.Context, appConfig *deezer.Config) (c *Client, err error) {
	c = &Client{
		appConfig:    appConfig,
		deezerClient: nil,
	}

	c.deezerClient, err = deezer.NewClient(ctx, c.appConfig)
	if err != nil {
		return
	}

	q := c.appConfig.Quality
	if !c.deezerClient.Session.Premium && (q == "mp3_320" || q == "flac") {
		return c, fmt.Errorf("premium account required for '%s' quality", q)
	}
	return
}

func (c *Client) getSongContent(ctx context.Context, song *deezer.Song, target io.Writer) error {
	quality := c.appConfig.Quality

	media, err := c.deezerClient.FetchMedia(ctx, song, quality)
	if err != nil {
		return fmt.Errorf("failed to fetch media: %w", err)
	}

	stream, err := c.deezerClient.GetMediaStream(ctx, media, song.ID)
	if err != nil {
		return fmt.Errorf("failed to get media stream: %w", err)
	}

	dlCtx, cancel := context.WithTimeout(ctx, c.appConfig.Timeout)
	defer cancel()

	mediaFormat := media.GetFormat()
	key := deezer.GetKey(c.appConfig.SecretKey, song.ID)
	if err := c.streamMedia(dlCtx, stream, key, target); err != nil {
		return fmt.Errorf("failed to stream to target: %w", err)
	}

	if quality != strings.ToLower(mediaFormat) {
		log.Printf("requested quality '%s' not available, using '%s' instead", quality, strings.ToLower(mediaFormat))
	}

	return nil
}

func (c *Client) streamMedia(ctx context.Context, stream io.ReadCloser, key []byte, target io.Writer) (err error) {
	defer stream.Close()

	buffer := make([]byte, chunkSize)
	for chunk := 0; ; chunk++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// continue
		}

		totalRead := 0
		for totalRead < chunkSize {
			n, err := stream.Read(buffer[totalRead:])
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}

			if n > 0 {
				totalRead += n
			}
		}

		if totalRead == 0 {
			break
		}

		if chunk%3 == 0 && totalRead == chunkSize {
			buffer, err = deezer.Decrypt(buffer, key)
			if err != nil {
				return err
			}
		}

		_, err = target.Write(buffer[:totalRead])
		if err != nil {
			return err
		}

		if totalRead < chunkSize {
			break
		}
	}

	return nil
}

func (c *Client) getSongsFromTrackID(ctx context.Context, trackID string) (songs []*deezer.Song, err error) {
	resource := &deezer.Track{}
	if err := c.deezerClient.FetchResource(ctx, resource, trackID); err != nil {
		return nil, fmt.Errorf("failed to fetch resource: %w", err)
	}

	songs = resource.GetSongs()
	if len(songs) == 0 {
		return nil, fmt.Errorf("no songs found for track ID: %s", trackID)
	}

	return songs, nil
}

func (c *Client) DownloadTrackByID(ctx context.Context, trackID string) ([]byte, error) {
	songs, err := c.getSongsFromTrackID(ctx, trackID)
	if err != nil {
		return nil, fmt.Errorf("failed to get songs from track ID: %w", err)
	}

	var buffer bytes.Buffer
	err = c.getSongContent(ctx, songs[0], &buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to get song content: %w", err)
	}

	return buffer.Bytes(), nil
}

func (c *Client) StreamTrackByID(ctx context.Context, trackID string, target io.Writer) error {
	songs, err := c.getSongsFromTrackID(ctx, trackID)
	if err != nil {
		return fmt.Errorf("failed to get songs from track ID: %w", err)
	}

	err = c.getSongContent(ctx, songs[0], target)
	if err != nil {
		return fmt.Errorf("failed to get song content: %w", err)
	}

	return nil
}
