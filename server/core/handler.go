package main

import (
	"encoding/json"
	"io"
	"net/http"
)

// NewRequestEchoHandler returns HTTP handler that echoes incoming request data as JSON.
func NewRequestEchoHandler(cfg ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()

		response := map[string]any{
			"request": map[string]any{
				"method":         r.Method,
				"url":            r.URL.String(),
				"path":           r.URL.Path,
				"raw_query":      r.URL.RawQuery,
				"query":          r.URL.Query(),
				"protocol":       r.Proto,
				"host":           r.Host,
				"remote_addr":    r.RemoteAddr,
				"headers":        r.Header,
				"content_length": r.ContentLength,
				"body":           string(body),
			},
		}

		w.Header().Set("Content-Type", cfg.ResponseContentType)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
