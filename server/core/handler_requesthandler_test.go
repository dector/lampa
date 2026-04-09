package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRequestHandler_Returns200ByDefault(t *testing.T) {
	cfg := DefaultServerConfig()
	h := NewRequestHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "http://dector.space/", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestNewRequestHandler_DefaultsContentTypeTextPlain(t *testing.T) {
	cfg := DefaultServerConfig()
	h := NewRequestHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "http://dector.space/", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("Content-Type"); got != "text/plain" {
		t.Fatalf("unexpected content type: got %q, want %q", got, "text/plain")
	}
}

func TestNewRequestHandler_AcceptsConfiguredRoute(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.RoutePath = "/health"

	mux := http.NewServeMux()
	mux.HandleFunc(cfg.RoutePath, NewRequestHandler(cfg))

	req := httptest.NewRequest(http.MethodGet, "http://dector.space/health", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code on configured route: got %d, want %d", rr.Code, http.StatusOK)
	}
}
