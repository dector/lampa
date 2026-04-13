package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dector/lampa/server/core/processor"
)

const (
	controlStatusOK    = "ok"
	controlStatusError = "error"
)

// ControlProcSetRequest defines payload for processor upsert.
type ControlProcSetRequest struct {
	Kind     string                `json:"kind"`
	Endpoint string                `json:"endpoint"`
	Response ControlStaticResponse `json:"response"`
}

// ControlStaticResponse defines static response returned by StaticReqProcessor.
type ControlStaticResponse struct {
	Status      int         `json:"status"`
	ContentType string      `json:"contentType"`
	Headers     http.Header `json:"headers"`
	Body        string      `json:"body"`
}

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
			"status": controlStatusOK,
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

// NewControlProcSetHandler sets endpoint processor in runtime store.
func NewControlProcSetHandler(store processor.ReqProcessorStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			writeControlError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		if store == nil {
			writeControlError(w, http.StatusInternalServerError, "processor store is not configured")
			return
		}

		var payload ControlProcSetRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil {
			writeControlError(w, http.StatusBadRequest, "invalid request json")
			return
		}

		if err := validateControlProcSetRequest(payload); err != nil {
			writeControlError(w, http.StatusBadRequest, err.Error())
			return
		}

		staticResponse := payload.Response
		headers := staticResponse.Headers.Clone()
		if headers == nil {
			headers = make(http.Header)
		}
		if strings.TrimSpace(staticResponse.ContentType) != "" {
			headers.Set("Content-Type", strings.TrimSpace(staticResponse.ContentType))
		}

		store.SetEndpointProcessor(payload.Endpoint, processor.StaticReqProcessor{
			StatusCode: staticResponse.Status,
			Headers:    headers,
			Body:       []byte(staticResponse.Body),
		})

		writeJSON(w, http.StatusOK, map[string]any{
			"status":   controlStatusOK,
			"endpoint": payload.Endpoint,
			"kind":     payload.Kind,
		})
	}
}

func validateControlProcSetRequest(payload ControlProcSetRequest) error {
	endpoint := strings.TrimSpace(payload.Endpoint)
	if endpoint == "" || !strings.HasPrefix(endpoint, "/") {
		return fmt.Errorf("invalid endpoint")
	}

	if strings.TrimSpace(payload.Kind) != "static" {
		return fmt.Errorf("invalid kind")
	}

	status := payload.Response.Status
	if status < 100 || status > 599 {
		return fmt.Errorf("invalid status")
	}

	return nil
}

func writeControlError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]any{
		"status": controlStatusError,
		"error":  message,
	})
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
