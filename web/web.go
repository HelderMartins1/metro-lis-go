package web

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"strings"
)

func init() {
	// Go does not know this one and would serve the PWA manifest as text/plain.
	_ = mime.AddExtensionType(".webmanifest", "application/manifest+json")
}

// dist is filled by `npm run build` in frontend/ (see the Makefile).
//
//go:embed all:dist
var files embed.FS

// Handler serves the frontend bundled into the binary.
func Handler() http.Handler {
	dist, _ := fs.Sub(files, "dist")
	if _, err := fs.Stat(dist, "index.html"); err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "frontend not built: run `make frontend`", http.StatusNotFound)
		})
	}

	static := http.FileServerFS(dist)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/assets/"):
			// Vite puts a content hash in every file name here.
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		case r.URL.Path == "/sw.js", r.URL.Path == "/", r.URL.Path == "/index.html":
			w.Header().Set("Cache-Control", "no-cache")
		}
		static.ServeHTTP(w, r)
	})
}
