package config

import (
	"log"
	"os"
)

type Config struct {
	URL   string
	Token string
}

func Load() *Config {
	url := os.Getenv("METRO_API_URL")
	token := os.Getenv("METRO_API_TOKEN")

	if url == "" || token == "" {
		log.Fatal("Env variable URL or TOKEN is empty")
	}

	return &Config{url, token}
}
