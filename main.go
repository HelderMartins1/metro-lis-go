package main

import (
	"net/http"

	"github.com/HelderMartins1/metro-lis-go/config"
	"github.com/HelderMartins1/metro-lis-go/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg := config.Load()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Accept", "Content-Type"},
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/{line}/stations", handlers.StationsHandler())
		r.Get("/{code}/trains", handlers.TrainsHandler(cfg))
		r.Get("/destinations", handlers.DestinationsHandler(cfg))
		r.Get("/lines", handlers.LinesHandler(cfg))
	})

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		return
	}
}
