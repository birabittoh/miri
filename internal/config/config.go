package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/birabittoh/miri/internal/fileutil"
	"github.com/joho/godotenv"
)

type Config struct {
	ArlCookie string
	SecretKey string
	DataDir   string
	Quality   string
	Limit     int
	Timeout   time.Duration
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func New() (*Config, error) {
	godotenv.Load()

	limit := getEnv("LIMIT", "100")
	timeout := getEnv("TIMEOUT", "30")

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		limitInt = 100
	}

	timeoutInt, err := strconv.Atoi(timeout)
	if err != nil || timeoutInt <= 0 {
		timeoutInt = 30
	}

	config := &Config{
		ArlCookie: getEnv("ARL_COOKIE", ""),
		SecretKey: getEnv("SECRET_KEY", ""),
		DataDir:   getEnv("DATA_DIR", "data"),
		Quality:   getEnv("QUALITY", "mp3_128"),
		Limit:     limitInt,
		Timeout:   time.Duration(timeoutInt) * time.Second,
	}

	if err := fileutil.EnsureDir(config.DataDir); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	if config.ArlCookie == "" {
		return nil, fmt.Errorf("ARL_COOKIE environment variable is not set")
	}

	if config.SecretKey == "" {
		return nil, fmt.Errorf("SECRET_KEY environment variable is not set")
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

	return nil
}
