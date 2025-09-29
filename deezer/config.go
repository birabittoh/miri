package deezer

import (
	"fmt"
	"time"
)

const (
	defaultQuality = "mp3_128"
	defaultTimeout = 30 * time.Second
)

var validQualities = map[string]bool{
	"mp3_128": true,
	"mp3_320": true,
	"flac":    true,
}

type Config struct {
	ArlCookie string
	SecretKey string
	Quality   string
	Timeout   time.Duration
}

func NewConfig(arlCookie, secretKey string) (*Config, error) {
	config := &Config{
		ArlCookie: arlCookie,
		SecretKey: secretKey,
		Quality:   defaultQuality,
		Timeout:   defaultTimeout,
	}

	err := config.Validate()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) Validate() error {
	if c.ArlCookie == "" {
		return fmt.Errorf("arl_cookie is not set")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("secret_key is not set")
	}
	if len(c.SecretKey) != 16 {
		return fmt.Errorf("secret_key must be 16 bytes long")
	}

	if c.Timeout <= 0 {
		c.Timeout = defaultTimeout
	}

	_, ok := validQualities[c.Quality]
	if !ok {
		return fmt.Errorf("invalid quality: %s", c.Quality)
	}

	return nil
}
