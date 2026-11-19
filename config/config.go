package config

import (
	"errors"
	"log"
	"os"
)

type Config struct {
	URL   string
	Token string
}

func Load() *Config {
	cfg, err := FromEnv()
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

// FromEnv reads the Metro API credentials without exiting, for callers that
// have a fallback when they are missing.
func FromEnv() (*Config, error) {
	url := os.Getenv("METRO_API_URL")
	token := os.Getenv("METRO_API_TOKEN")

	if url == "" || token == "" {
		return nil, errors.New("Env variable URL or TOKEN is empty")
	}

	return &Config{url, token}, nil
}
