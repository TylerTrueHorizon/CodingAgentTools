package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"agent-tools-sandbox/internal/config"
	"agent-tools-sandbox/internal/handlers"
)

func main() {
	cfg := config.Load()
	files := &handlers.Files{}
	shell := handlers.NewShell(cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /files/read", files.Read)
	mux.HandleFunc("POST /files/write", files.Write)
	mux.HandleFunc("POST /files/edit", files.Edit)
	mux.HandleFunc("GET /files/list", files.List)
	mux.HandleFunc("POST /shell/run", shell.Run)

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("listening on %s", addr)
	handler := requireAPIKey(cfg)(limitBody(mux, cfg.MaxRequestBody))
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

// requireAPIKey returns middleware that enforces API_KEY when configured.
func requireAPIKey(cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.APIKey == "" {
				next.ServeHTTP(w, r)
				return
			}
			var key string
			if auth := r.Header.Get("Authorization"); strings.HasPrefix(strings.TrimSpace(auth), "Bearer ") {
				key = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(auth), "Bearer "))
			} else if k := r.Header.Get("X-Api-Key"); k != "" {
				key = k
			}
			if key != cfg.APIKey {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// limitBody wraps the handler to enforce max request body size.
func limitBody(next http.Handler, maxBytes int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		next.ServeHTTP(w, r)
	})
}
