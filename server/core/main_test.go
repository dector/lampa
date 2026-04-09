package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRun_PassesAddressAndRegistersHandler(t *testing.T) {
	cfg := ServerConfig{
		ListenAddress: ":9090",
		RoutePath:     "/health",
	}

	called := false
	err := Run(cfg, func(addr string, h http.Handler) error {
		called = true

		if addr != cfg.ListenAddress {
			t.Fatalf("unexpected listen address: got %q, want %q", addr, cfg.ListenAddress)
		}
		if h == nil {
			t.Fatal("expected non-nil handler")
		}

		req := httptest.NewRequest(http.MethodGet, cfg.RoutePath, nil)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status code: got %d, want %d", rr.Code, http.StatusOK)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Run returned unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected listen function to be called")
	}
}

func TestRun_PropagatesListenError(t *testing.T) {
	cfg := DefaultServerConfig()
	wantErr := errors.New("listen failed")

	err := Run(cfg, func(_ string, _ http.Handler) error {
		return wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("unexpected error: got %v, want %v", err, wantErr)
	}
}
