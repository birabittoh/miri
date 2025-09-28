package miri

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/birabittoh/miri/internal/config"
	"github.com/birabittoh/miri/internal/crypto"
	"github.com/birabittoh/miri/internal/deezer"
	"github.com/birabittoh/miri/internal/fileutil"
	"github.com/birabittoh/miri/internal/logger"
)

const chunkSize = 2048

type Client struct {
	appConfig    *config.Config
	resourceType string
	deezerClient *deezer.Client
	Logger       *logger.Logger
}

func New(appConfig *config.Config, resourceType string) *Client {
	return &Client{
		appConfig:    appConfig,
		resourceType: resourceType,
		deezerClient: nil,
		Logger:       logger.New(nil), // Initialize with a nil logger, can be set later
	}
}

func (c *Client) Run(ctx context.Context, id string) error {
	if err := c.initDeezerClient(ctx); err != nil {
		return err
	}

	resource, err := c.prepareResource(ctx, id)
	if err != nil {
		return err
	}

	return c.downloadAllSongs(ctx, resource)
}

func (c *Client) initDeezerClient(ctx context.Context) error {
	var err error
	c.deezerClient, err = deezer.NewClient(ctx, c.appConfig)
	if err != nil {
		return err
	}

	q := c.appConfig.Quality

	if !c.deezerClient.Session.Premium && (q == "mp3_320" || q == "flac") {
		return fmt.Errorf("premium account required for '%s' quality", q)
	}

	return nil
}

func (c *Client) prepareResource(ctx context.Context, id string) (deezer.Resource, error) {
	resource, err := c.createResource()
	if err != nil {
		return nil, err
	}

	if err := c.deezerClient.FetchResource(ctx, resource, id); err != nil {
		return nil, fmt.Errorf("failed to fetch resource: %w", err)
	}

	songs := resource.GetSongs()
	if len(songs) == 0 {
		if c.resourceType == "track" {
			return nil, fmt.Errorf("track with ID %s not found", id)
		}
		return nil, fmt.Errorf("%s has no songs", c.resourceType)
	}

	if c.resourceType == "artist" && len(songs) > c.appConfig.Limit {
		songs = songs[:c.appConfig.Limit]
		resource.SetSongs(songs)
	}

	return resource, nil
}

func (c *Client) createResource() (deezer.Resource, error) {
	switch c.resourceType {
	case "album":
		return &deezer.Album{}, nil
	case "playlist":
		return &deezer.Playlist{}, nil
	case "artist":
		return &deezer.Artist{}, nil
	case "track":
		return &deezer.Track{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", c.resourceType)
	}
}

func (c *Client) downloadAllSongs(ctx context.Context, resource deezer.Resource) error {
	songs := resource.GetSongs()

	if c.resourceType != "track" {
		fmt.Printf("%s\n\nStarting download...\n\n", resource)
	}

	for _, song := range songs {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		data, _, err := c.downloadSong(ctx, song)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
		}

		file, err := os.Create("test.mp3")
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		_, err = file.Write(data)
		if err != nil {
			return fmt.Errorf("failed to write to output file: %w", err)
		}
		file.Close()
	}

	return nil
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

	fileName := song.GetFileName(c.resourceType, mediaFormat, song)
	outputPath := path.Join(c.appConfig.DataDir, fileName)

	key := crypto.GetKey(c.appConfig.SecretKey, song.ID)

	var buffer bytes.Buffer
	if err := c.streamMedia(dlCtx, stream, key, &buffer); err != nil {
		fileutil.DeleteFile(outputPath)
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
