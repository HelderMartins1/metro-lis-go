package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/HelderMartins1/metro-lis-go/config"
	"github.com/HelderMartins1/metro-lis-go/services"
	"github.com/go-chi/chi/v5"
)

func StationsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		line := chi.URLParam(r, "line")
		writeJSON(w, http.StatusOK, services.GetStations(line))
	}
}

func TrainsHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := chi.URLParam(r, "code")
		data, err := services.GetTrains(cfg, code)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, data)
	}
}

func LinesHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := services.GetLines(cfg)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, data)
	}
}

func DestinationsHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := services.GetDestinations(cfg)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, data)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Backend", "go")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		return
	}
}
