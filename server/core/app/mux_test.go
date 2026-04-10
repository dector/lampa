package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildProxyMux_RegistersConfiguredRoute(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.RoutePath = "/proxy"

	mux := BuildProxyMux(cfg, nil)

	t.Run("configured route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/proxy", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusNotFound; got != want {
			t.Fatalf("unexpected status code: got %d, want %d", got, want)
		}
		if got, want := rr.Body.String(), "Not Found"; got != want {
			t.Fatalf("unexpected body: got %q, want %q", got, want)
		}
	})

	t.Run("other routes are not handled", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if got, want := rr.Code, http.StatusNotFound; got != want {
			t.Fatalf("unexpected status code: got %d, want %d", got, want)
		}
		if got, want := rr.Body.String(), "404 page not found\n"; got != want {
			t.Fatalf("unexpected body: got %q, want %q", got, want)
		}
	})
}

func TestBuildControlMux_BasePathRoot_NoAlias(t *testing.T) {
	mux := BuildControlMux(nil, ControlMuxOptions{BasePath: "/", EnableRootPingAlias: false})

	assertPingEndpoint(t, mux, "/ping")
	assertProcCountEndpoint(t, mux, "/api/v0/proc_count")
	assertNotFound(t, mux, "/")
}

func TestBuildControlMux_BasePathRoot_WithAlias(t *testing.T) {
	mux := BuildControlMux(nil, ControlMuxOptions{BasePath: "/", EnableRootPingAlias: true})
	assertPingEndpoint(t, mux, "/")
}

func TestBuildControlMux_BasePathControl_NoAlias(t *testing.T) {
	mux := BuildControlMux(nil, ControlMuxOptions{BasePath: "/control", EnableRootPingAlias: false})

	assertPingEndpoint(t, mux, "/control/ping")
	assertProcCountEndpoint(t, mux, "/control/api/v0/proc_count")
	assertNotFound(t, mux, "/ping")
	assertNotFound(t, mux, "/")
}

func assertPingEndpoint(t *testing.T, h http.Handler, requestPath string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, requestPath, nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}

	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response body as JSON: %v", err)
	}

	if got, want := payload["status"], "ok"; got != want {
		t.Fatalf("unexpected status: got %v, want %v", got, want)
	}

	responses, ok := payload["responses"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected responses shape: %#v", payload["responses"])
	}
	if got, want := responses["count"], float64(0); got != want {
		t.Fatalf("unexpected responses.count: got %v, want %v", got, want)
	}

	if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("unexpected content type: got %q", got)
	}
}

func assertProcCountEndpoint(t *testing.T, h http.Handler, requestPath string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, requestPath, nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}

	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response body as JSON: %v", err)
	}
	if got, want := payload["count"], float64(0); got != want {
		t.Fatalf("unexpected count: got %v, want %v", got, want)
	}

	if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("unexpected content type: got %q", got)
	}
}

func assertNotFound(t *testing.T, h http.Handler, requestPath string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, requestPath, nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusNotFound; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}
}
