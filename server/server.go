package server

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/HelderMartins1/metro-lis-go/config"
	"github.com/HelderMartins1/metro-lis-go/handlers"
	"github.com/HelderMartins1/metro-lis-go/push"
	"github.com/HelderMartins1/metro-lis-go/services"
	"github.com/HelderMartins1/metro-lis-go/web"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// New serves the frontend at / and hands everything under /api to api.
// allowedOrigins are the sites whose pages may call the API from a browser
// (the frontend's own domain when it is hosted apart); none means any.
func New(api http.Handler, allowedOrigins []string) http.Handler {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	// A bug in one request answers 500 instead of taking the connection down with it.
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT"},
		AllowedHeaders: []string{"Accept", "Content-Type"},
	}))

	r.Mount("/api", api)
	r.Handle("/*", web.Handler())
	return r
}

// Options are the settings both kinds of API share.
type Options struct {
	DataDir     string // where push keeps its keys and subscriptions
	PushContact string // how push services can reach whoever runs this server
}

// API talks to the Metro de Lisboa API with the given credentials.
func API(cfg *config.Config, opts Options) http.Handler {
	src := services.APISource{Cfg: cfg}
	r := chi.NewRouter()
	r.Get("/network", handlers.NetworkHandler())
	r.Get("/{line}/stations", handlers.StationsHandler())
	r.Get("/{line}/positions", handlers.PositionsHandler(services.NewPositions(src)))
	r.Get("/{code}/trains", handlers.TrainsHandler(cfg))
	r.Get("/destinations", handlers.DestinationsHandler(cfg))
	r.Get("/lines", handlers.LinesHandler(cfg))
	r.Mount("/push", pushHandler(src, opts))
	return r
}

// pushHandler starts push notifications, or answers 503 when they cannot run
// (the app keeps working, without notifications).
func pushHandler(src services.Source, opts Options) http.Handler {
	p, err := push.New(src, opts.DataDir, opts.PushContact)
	if err != nil {
		log.Printf("push notifications are off: %v", err)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, `{"error":"push notifications are not available"}`, http.StatusServiceUnavailable)
		})
	}
	go p.Run(context.Background())
	return p.Handler()
}

// Remote answers the static network data itself and forwards the live data
// to another instance of this server, so no Metro API token is needed here.
func Remote(backend *url.URL, opts Options) http.Handler {
	src := services.RemoteSource{Base: backend}
	proxy := httputil.NewSingleHostReverseProxy(backend)
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)
		req.Host = backend.Host
	}

	r := chi.NewRouter()
	r.Get("/network", handlers.NetworkHandler())
	r.Get("/{line}/stations", handlers.StationsHandler())
	r.Get("/{line}/positions", handlers.PositionsHandler(services.NewPositions(src)))
	r.Mount("/push", pushHandler(src, opts))
	r.Handle("/*", proxy)
	return r
}
