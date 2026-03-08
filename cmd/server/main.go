package main

import (
	"fmt"
	"log"
	"net/http"

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
	if err := http.ListenAndServe(addr, limitBody(mux, cfg.MaxRequestBody)); err != nil {
		log.Fatal(err)
	}
}

// limitBody wraps the handler to enforce max request body size.
func limitBody(next http.Handler, maxBytes int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		next.ServeHTTP(w, r)
	})
}
