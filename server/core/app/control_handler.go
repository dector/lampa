package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dector/lampa/server/core/logstore"
	"github.com/dector/lampa/server/core/processor"
)

const (
	controlStatusOK    = "ok"
	controlStatusError = "error"

	controlProcKindStatic = "static"
	controlProcKindJS     = "js"
	controlProcKindPass   = "pass"
)

// ControlProcSetRequest defines payload for processor upsert.
type ControlProcSetRequest struct {
	Kind     string                   `json:"kind"`
	Endpoint string                   `json:"endpoint"`
	Response ControlStaticResponse    `json:"response"`
	JS       *ControlJSProcessorInput `json:"js,omitempty"`
}

// ControlStaticResponse defines static response returned by StaticReqProcessor.
type ControlStaticResponse struct {
	Status      int         `json:"status"`
	ContentType string      `json:"contentType"`
	Headers     http.Header `json:"headers"`
	Body        string      `json:"body"`
}

// ControlJSProcessorInput defines JS processor config for kind="js".
type ControlJSProcessorInput struct {
	Script string `json:"script"`
}

// ControlProcSetDefaultRequest defines payload for fallback/default processor update.
type ControlProcSetDefaultRequest struct {
	Kind   string `json:"kind"`
	Server string `json:"server"`
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

// NewControlProxyLogsHandler returns latest proxy request/response logs.
func NewControlProxyLogsHandler(logs logstore.Store) http.HandlerFunc {
	const (
		defaultN = 10
		maxN     = 1000
	)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			writeControlError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		if logs == nil {
			writeControlError(w, http.StatusInternalServerError, "log store is not configured")
			return
		}

		n := defaultN
		rawN := strings.TrimSpace(r.URL.Query().Get("n"))
		if rawN != "" {
			parsedN, err := strconv.Atoi(rawN)
			if err != nil || parsedN <= 0 {
				writeControlError(w, http.StatusBadRequest, "invalid n")
				return
			}
			n = parsedN
		}
		if n > maxN {
			n = maxN
		}

		entries := logs.Latest(n)
		writeJSON(w, http.StatusOK, map[string]any{
			"status":    controlStatusOK,
			"count":     len(entries),
			"total":     logs.Count(),
			"sizeBytes": logs.SizeBytes(),
			"entries":   entries,
		})
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

		switch strings.TrimSpace(payload.Kind) {
		case controlProcKindStatic:
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
		case controlProcKindJS:
			store.SetEndpointProcessor(payload.Endpoint, processor.QuickJSReqProcessor{
				Script: payload.JS.Script,
			})
		default:
			writeControlError(w, http.StatusBadRequest, "invalid kind")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"status":   controlStatusOK,
			"endpoint": payload.Endpoint,
			"kind":     payload.Kind,
		})
	}
}

// NewControlProcSetDefaultHandler sets fallback/default processor in runtime store.
func NewControlProcSetDefaultHandler(store processor.ReqProcessorStore) http.HandlerFunc {
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

		var payload ControlProcSetDefaultRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil {
			writeControlError(w, http.StatusBadRequest, "invalid request json")
			return
		}

		if err := validateControlProcSetDefaultRequest(payload); err != nil {
			writeControlError(w, http.StatusBadRequest, err.Error())
			return
		}

		serverURL := strings.TrimSpace(payload.Server)
		store.SetFallbackProcessor(processor.PassthroughReqProcessor{Server: serverURL})

		writeJSON(w, http.StatusOK, map[string]any{
			"status": controlStatusOK,
			"kind":   controlProcKindPass,
			"server": serverURL,
		})
	}
}

func validateControlProcSetRequest(payload ControlProcSetRequest) error {
	endpoint := strings.TrimSpace(payload.Endpoint)
	if endpoint == "" || !strings.HasPrefix(endpoint, "/") {
		return fmt.Errorf("invalid endpoint")
	}

	switch strings.TrimSpace(payload.Kind) {
	case controlProcKindStatic:
		status := payload.Response.Status
		if status < 100 || status > 599 {
			return fmt.Errorf("invalid status")
		}
		return nil
	case controlProcKindJS:
		if payload.JS == nil || strings.TrimSpace(payload.JS.Script) == "" {
			return fmt.Errorf("invalid js script")
		}
		return nil
	default:
		return fmt.Errorf("invalid kind")
	}
}

func validateControlProcSetDefaultRequest(payload ControlProcSetDefaultRequest) error {
	if strings.TrimSpace(payload.Kind) != controlProcKindPass {
		return fmt.Errorf("invalid kind")
	}

	parsedServer, err := url.Parse(strings.TrimSpace(payload.Server))
	if err != nil {
		return fmt.Errorf("invalid server")
	}
	if parsedServer.Scheme == "" || parsedServer.Host == "" {
		return fmt.Errorf("invalid server")
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
