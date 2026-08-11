package app

import (
	"net/http"

	"github.com/dector/lampa/server/core/logstore"
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

		trafficCount := 0
		trafficSize := int64(0)
		if logs != nil {
			trafficCount = logs.Count()
			trafficSize = logs.SizeBytes()
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := WebUIIndexPage(cfg, trafficCount, trafficSize).Render(r.Context(), w); err != nil {
			http.Error(w, "failed to render web ui", http.StatusInternalServerError)
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
