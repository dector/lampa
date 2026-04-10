package app

import (
	"encoding/json"
	"net/http"

	"github.com/dector/lampa/server/core/processor"
)

// NewControlPingHandler returns control ping handler.
//
// store is shared with proxy request handler and reserved for control APIs that
// will update runtime proxy behavior.
func NewControlPingHandler(store processor.ReqProcessorStore) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		responsesCount := 0
		if store != nil {
			responsesCount = store.ResponsesCount()
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"status": "ok",
			"responses": map[string]any{
				"count": responsesCount,
			},
		})
	}
}

// NewControlProcCountHandler returns count of currently configured processors.
func NewControlProcCountHandler(store processor.ReqProcessorStore) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		count := 0
		if store != nil {
			count = store.ResponsesCount()
		}
		writeJSON(w, http.StatusOK, map[string]any{"count": count})
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(encoded)
}
