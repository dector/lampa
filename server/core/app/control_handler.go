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

		payload, err := json.Marshal(map[string]any{
			"status": "ok",
			"responses": map[string]any{
				"count": responsesCount,
			},
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	}
}
