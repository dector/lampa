package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dector/lampa/server/core/logstore"
)

func TestBuildWebUIMux_IndexShowsStatus(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.ListenAddress = "localhost:18080"
	cfg.ControlListenAddress = "localhost:18081"
	logs := logstore.NewInMemoryStore(logstore.UnlimitedMaxBytes)
	logs.Add(logstore.Entry{
		Request: logstore.Request{
			Method: "GET",
			URL:    "https://example.test/path",
			Path:   "/path",
		},
		Response: logstore.Response{Status: http.StatusOK},
	})

	mux := BuildWebUIMux(cfg, logs)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}
	body := rr.Body.String()
	for _, want := range []string{"Lampa Web UI", "Server: <span class=\"pill\">running", "Proxy port</dt><dd>8080", "Control port</dt><dd>8081", "Traffic count</dt><dd>1", "color-scheme: dark", "class=\"pill\"", "src=\"/assets/datastar-1.0.2.js\"", "data-on-interval__duration.1s=\"@get('/ds/stats')\""} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected body to contain %q, got %q", want, body)
		}
	}
	if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Fatalf("unexpected content type: got %q", got)
	}
}

func TestBuildWebUIMux_StatsReturnsDatastarStream(t *testing.T) {
	logs := logstore.NewInMemoryStore(logstore.UnlimitedMaxBytes)
	logs.Add(logstore.Entry{
		Request:  logstore.Request{Method: "POST", URL: "https://example.test/api", Path: "/api"},
		Response: logstore.Response{Status: http.StatusCreated},
	})

	mux := BuildWebUIMux(DefaultServerConfig(), logs)
	req := httptest.NewRequest(http.MethodGet, "/ds/stats", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}
	if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
		t.Fatalf("unexpected content type: got %q", got)
	}
	body := rr.Body.String()
	for _, want := range []string{"event: datastar-patch-elements", "id=\"stats\"", "Traffic count", "1", "Traffic size bytes"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected body to contain %q, got %q", want, body)
		}
	}
}

func TestBuildWebUIMux_DatastarAsset(t *testing.T) {
	mux := BuildWebUIMux(DefaultServerConfig(), nil)
	req := httptest.NewRequest(http.MethodGet, WebUIAssetsPath+"datastar-1.0.2.js", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}
	if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "javascript") {
		t.Fatalf("unexpected content type: got %q", got)
	}
	if got := rr.Header().Get("Cache-Control"); !strings.Contains(got, "immutable") {
		t.Fatalf("unexpected cache control: got %q", got)
	}
	if body := rr.Body.String(); !strings.Contains(body, "datastar") {
		t.Fatalf("expected datastar asset body, got %q", body)
	}
}

func TestBuildWebUIMux_MissingAssetNotFound(t *testing.T) {
	mux := BuildWebUIMux(DefaultServerConfig(), nil)
	req := httptest.NewRequest(http.MethodGet, WebUIAssetsPath+"missing.js", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusNotFound; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}
}

func TestBuildWebUIMux_Health(t *testing.T) {
	mux := BuildWebUIMux(DefaultServerConfig(), nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}
	if got, want := rr.Body.String(), "ok"; got != want {
		t.Fatalf("unexpected body: got %q, want %q", got, want)
	}
	if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/plain") {
		t.Fatalf("unexpected content type: got %q", got)
	}
}

func TestBuildWebUIMux_UnknownRouteNotFound(t *testing.T) {
	mux := BuildWebUIMux(DefaultServerConfig(), nil)
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusNotFound; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}
}
