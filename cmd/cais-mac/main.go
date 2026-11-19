//go:build darwin

// Command cais-mac is Cais for macOS: the Go server and the frontend in one
// process, shown in a native WebKit window.
//
// Where the data comes from, in order:
//  1. METRO_API_URL and METRO_API_TOKEN in the environment;
//  2. api_url and api_token in ~/Library/Application Support/Cais/config.json;
//  3. a deployed Cais server: backend_url in that file, else defaultBackend.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/HelderMartins1/metro-lis-go/config"
	"github.com/HelderMartins1/metro-lis-go/server"
	webview "github.com/webview/webview_go"
)

// A fixed port keeps the page's origin, and with it the saved favourites,
// the same from one launch to the next.
const addr = "127.0.0.1:47615"

// Set at build time with -ldflags "-X main.defaultBackend=https://…".
var defaultBackend = "https://api.caislx.tech"

type settings struct {
	APIURL     string `json:"api_url"`
	APIToken   string `json:"api_token"`
	BackendURL string `json:"backend_url"`
}

func init() {
	// Cocoa wants the window on the main thread.
	runtime.LockOSThread()
}

func main() {
	api, err := apiHandler()
	if err != nil {
		log.Fatal(err)
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("%s is busy (%v); using a random port, favourites will not carry over", addr, err)
		if ln, err = net.Listen("tcp", "127.0.0.1:0"); err != nil {
			log.Fatal(err)
		}
	}
	go func() {
		log.Fatal(http.Serve(ln, server.New(api, nil)))
	}()

	w := webview.New(false)
	defer w.Destroy()
	setupMenu()
	w.SetTitle("Cais")
	w.SetSize(1180, 780, webview.HintNone)
	w.SetSize(380, 560, webview.HintMin)
	w.Navigate("http://" + ln.Addr().String() + "/")
	w.Run()
}

func apiHandler() (http.Handler, error) {
	// The window has no Web Push, but the API keeps the same shape everywhere.
	opts := server.Options{PushContact: "https://caislx.tech"}
	if dir, err := os.UserConfigDir(); err == nil {
		opts.DataDir = filepath.Join(dir, "Cais")
	}
	if cfg, err := config.FromEnv(); err == nil {
		return server.API(cfg, opts), nil
	}

	s, err := loadSettings()
	if err != nil {
		return nil, err
	}
	if s.APIURL != "" && s.APIToken != "" {
		return server.API(&config.Config{URL: s.APIURL, Token: s.APIToken}, opts), nil
	}

	backend := s.BackendURL
	if backend == "" {
		backend = defaultBackend
	}
	u, err := url.Parse(backend)
	if err != nil || u.Host == "" {
		return nil, fmt.Errorf("backend_url is not a URL: %q", backend)
	}
	return server.Remote(u, opts), nil
}

func loadSettings() (settings, error) {
	var s settings
	dir, err := os.UserConfigDir()
	if err != nil {
		return s, nil
	}
	data, err := os.ReadFile(filepath.Join(dir, "Cais", "config.json"))
	if errors.Is(err, fs.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return s, fmt.Errorf("config.json: %w", err)
	}
	return s, nil
}
