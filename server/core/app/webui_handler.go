package app

import (
	"net/http"

	"github.com/dector/lampa/server/core/logstore"
	"github.com/starfederation/datastar-go/datastar"
)

// NewWebUIIndexHandler returns the Web UI landing page handler.
func NewWebUIIndexHandler(cfg ServerConfig, logs logstore.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		trafficCount, trafficSize := webUIStatsSnapshot(logs)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := WebUIIndexPage(cfg, trafficCount, trafficSize).Render(r.Context(), w); err != nil {
			http.Error(w, "failed to render web ui", http.StatusInternalServerError)
			return
		}
	}
}

// NewWebUIStatsHandler returns the polling stats fragment handler.
func NewWebUIStatsHandler(cfg ServerConfig, logs logstore.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/ds/stats" {
			http.NotFound(w, r)
			return
		}

		trafficCount, trafficSize := webUIStatsSnapshot(logs)
		sse := datastar.NewSSE(w, r)
		if err := sse.PatchElementTempl(WebUIStats(cfg, trafficCount, trafficSize)); err != nil {
			http.Error(w, "failed to render stats", http.StatusInternalServerError)
			return
		}
	}
}

// NewWebUIHealthHandler returns the Web UI health-check handler.
func NewWebUIHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}
