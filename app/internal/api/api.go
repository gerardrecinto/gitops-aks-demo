// Package api holds the HTTP handlers for the items service.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config carries the values injected at build or deploy time.
type Config struct {
	Version       string
	Environment   string
	APIKeyPresent bool
}

type item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var items = []item{
	{1, "keyboard"},
	{2, "monitor"},
	{3, "webcam"},
}

// NewHandler wires every route. It has no global state, so tests build their own.
func NewHandler(cfg Config) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"version":            cfg.Version,
			"environment":        cfg.Environment,
			"api_key_configured": cfg.APIKeyPresent,
		})
	})
	mux.HandleFunc("GET /api/items", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, items)
	})
	mux.Handle("GET /metrics", promhttp.Handler())
	return mux
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
