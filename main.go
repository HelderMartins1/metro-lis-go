package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/HelderMartins1/metro-lis-go/config"
	"github.com/HelderMartins1/metro-lis-go/server"
)

func main() {
	// Settings come from the environment, filled in from .env in the working
	// directory (or CAIS_ENV_FILE) for whatever the environment leaves unset.
	if err := config.LoadEnvFile(envOr("CAIS_ENV_FILE", ".env")); err != nil {
		log.Fatal(err)
	}

	addr := ":" + envOr("PORT", "8080")
	// Comma-separated; empty allows any origin (fine on localhost).
	var origins []string
	if v := os.Getenv("CORS_ORIGINS"); v != "" {
		origins = strings.Split(v, ",")
	}

	err := http.ListenAndServe(addr, server.New(api(), origins))
	if err != nil {
		log.Fatal(err)
	}
}

// api uses the Metro API credentials when they are set; without them,
// METRO_BACKEND_URL points at a deployed instance that has them.
func api() http.Handler {
	opts := server.Options{
		DataDir:     envOr("CAIS_DATA_DIR", "data"),
		PushContact: envOr("CAIS_PUSH_CONTACT", "https://caislx.tech"),
	}
	if backend := os.Getenv("METRO_BACKEND_URL"); backend != "" {
		if _, err := config.FromEnv(); err != nil {
			u, err := url.Parse(backend)
			if err != nil || u.Host == "" {
				log.Fatalf("METRO_BACKEND_URL is not a URL: %q", backend)
			}
			log.Printf("no Metro API token: live data comes from %s", u)
			return server.Remote(u, opts)
		}
	}
	return server.API(config.Load(), opts)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
