package deezer

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ArlCookie string
	SecretKey string
	DataDir   string
	Quality   string
	Limit     int
	Timeout   time.Duration
}

func ensureDir(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("file already exists at %s", path)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func NewConfig() (*Config, error) {
	limit, err := strconv.Atoi(getEnv("LIMIT", "100"))
	if err != nil || limit <= 0 {
		limit = 100
	}

	timeoutInt, err := strconv.Atoi(getEnv("TIMEOUT", "30"))
	if err != nil || timeoutInt <= 0 {
		timeoutInt = 30
	}
	timeout := time.Duration(timeoutInt) * time.Second

	config := &Config{
		ArlCookie: getEnv("ARL_COOKIE", ""),
		SecretKey: getEnv("SECRET_KEY", ""),
		DataDir:   getEnv("DATA_DIR", "data"),
		Quality:   getEnv("QUALITY", "mp3_128"),
		Limit:     limit,
		Timeout:   timeout,
	}

	if err := ensureDir(config.DataDir); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	if config.ArlCookie == "" {
		return nil, fmt.Errorf("ARL_COOKIE environment variable is not set")
	}

	if config.SecretKey == "" {
		return nil, fmt.Errorf("SECRET_KEY environment variable is not set")
	}

	err = config.Validate()
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

	return nil
}
