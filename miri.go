package miri

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/birabittoh/miri/internal/config"
	"github.com/birabittoh/miri/internal/crypto"
	"github.com/birabittoh/miri/internal/deezer"
	"github.com/birabittoh/miri/internal/logger"
)

const chunkSize = 2048

type Client struct {
	appConfig    *config.Config
	deezerClient *deezer.Client
	Logger       *logger.Logger
}

func New(ctx context.Context, appConfig *config.Config) (c *Client, err error) {
	c = &Client{
		appConfig:    appConfig,
		deezerClient: nil,
		Logger:       logger.New(nil), // Initialize with a nil logger, can be set later
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

func (c *Client) downloadSong(ctx context.Context, song *deezer.Song) (content []byte, cover []byte, err error) {
	quality := c.appConfig.Quality

	media, err := c.deezerClient.FetchMedia(ctx, song, quality)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch media: %w", err)
	}

	stream, err := c.deezerClient.GetMediaStream(ctx, media, song.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get media stream: %w", err)
	}

	dlCtx, cancel := context.WithTimeout(ctx, c.appConfig.Timeout)
	defer cancel()

	mediaFormat := media.GetFormat()

	key := crypto.GetKey(c.appConfig.SecretKey, song.ID)

	var buffer bytes.Buffer
	if err := c.streamMedia(dlCtx, stream, key, &buffer); err != nil {
		return nil, nil, fmt.Errorf("failed to stream to file: %w", err)
	}

	if quality != strings.ToLower(mediaFormat) {
		log.Printf("requested quality '%s' not available, using '%s' instead", quality, strings.ToLower(mediaFormat))
	}

	cover, err = c.deezerClient.FetchCoverImage(ctx, song)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("requested quality '%s' not available, using '%s' instead", quality, strings.ToLower(mediaFormat))
	}

	return buffer.Bytes(), cover, nil
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
			buffer, err = crypto.Decrypt(buffer, key)
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

func (c *Client) DownloadTrackByID(ctx context.Context, trackID string) ([]byte, []byte, error) {
	resource := &deezer.Track{}
	if err := c.deezerClient.FetchResource(ctx, resource, trackID); err != nil {
		return nil, nil, fmt.Errorf("failed to fetch resource: %w", err)
	}

	songs := resource.GetSongs()
	if len(songs) == 0 {
		return nil, nil, fmt.Errorf("no songs found for track ID: %s", trackID)
	}

	return c.downloadSong(ctx, songs[0])
}
