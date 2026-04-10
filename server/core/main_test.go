package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunProxy_PassesAddressAndRegistersHandler(t *testing.T) {
	cfg := ServerConfig{
		ListenAddress: ":9090",
		RoutePath:     "/health",
	}

	called := false
	err := RunProxy(cfg, func(addr string, h http.Handler) error {
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
		t.Fatalf("RunProxy returned unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected listen function to be called")
	}
}

func TestRunProxy_PropagatesListenError(t *testing.T) {
	cfg := DefaultServerConfig()
	wantErr := errors.New("listen failed")

	err := RunProxy(cfg, func(_ string, _ http.Handler) error {
		return wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("unexpected error: got %v, want %v", err, wantErr)
	}
}

func TestRunWithControl_StartsBothServers(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.ListenAddress = "localhost:18080"
	cfg.ControlListenAddress = "localhost:18081"

	ready := make(chan struct{})
	var calls atomic.Int32

	err := RunWithControl(cfg, func(_ string, _ http.Handler) error {
		if calls.Add(1) == 2 {
			close(ready)
		}
		<-ready
		return nil
	})

	if err != nil {
		t.Fatalf("RunWithControl returned unexpected error: %v", err)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("unexpected listen calls: got %d, want %d", got, 2)
	}
}

func TestRunWithControl_PropagatesError(t *testing.T) {
	cfg := DefaultServerConfig()
	wantErr := errors.New("proxy listen failed")

	err := RunWithControl(cfg, func(addr string, _ http.Handler) error {
		if addr == cfg.ListenAddress {
			return wantErr
		}
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("unexpected error: got %v, want %v", err, wantErr)
	}
}

func TestLoadRuntimeConfig_Defaults(t *testing.T) {
	cfg := LoadRuntimeConfig(func(string) string { return "" })

	if got, want := cfg.ProxyPort, 8080; got != want {
		t.Fatalf("unexpected proxy port: got %d, want %d", got, want)
	}
	if got, want := cfg.ControlPort, 8081; got != want {
		t.Fatalf("unexpected control port: got %d, want %d", got, want)
	}
	if got, want := cfg.ProxyBindHost, "localhost"; got != want {
		t.Fatalf("unexpected proxy host: got %q, want %q", got, want)
	}
	if got, want := cfg.ControlBindHost, "localhost"; got != want {
		t.Fatalf("unexpected control host: got %q, want %q", got, want)
	}
	if got, want := cfg.ListenAddress, "localhost:8080"; got != want {
		t.Fatalf("unexpected proxy listen address: got %q, want %q", got, want)
	}
	if got, want := cfg.ControlListenAddress, "localhost:8081"; got != want {
		t.Fatalf("unexpected control listen address: got %q, want %q", got, want)
	}
}

func TestLoadRuntimeConfig_UsesEnvValues(t *testing.T) {
	cfg := LoadRuntimeConfig(func(key string) string {
		switch key {
		case "PORT":
			return "19090"
		case "PORT_CTRL":
			return "19091"
		case "BIND_HOST":
			return "0.0.0.0"
		case "BIND_HOST_CTRL":
			return "127.0.0.1"
		default:
			return ""
		}
	})

	if got, want := cfg.ProxyPort, 19090; got != want {
		t.Fatalf("unexpected proxy port: got %d, want %d", got, want)
	}
	if got, want := cfg.ControlPort, 19091; got != want {
		t.Fatalf("unexpected control port: got %d, want %d", got, want)
	}
	if got, want := cfg.ProxyBindHost, "0.0.0.0"; got != want {
		t.Fatalf("unexpected proxy host: got %q, want %q", got, want)
	}
	if got, want := cfg.ControlBindHost, "127.0.0.1"; got != want {
		t.Fatalf("unexpected control host: got %q, want %q", got, want)
	}
	if got, want := cfg.ListenAddress, "0.0.0.0:19090"; got != want {
		t.Fatalf("unexpected proxy listen address: got %q, want %q", got, want)
	}
	if got, want := cfg.ControlListenAddress, "127.0.0.1:19091"; got != want {
		t.Fatalf("unexpected control listen address: got %q, want %q", got, want)
	}
}

func TestParsePort(t *testing.T) {
	if got, want := parsePort("", 8080), 8080; got != want {
		t.Fatalf("unexpected default port: got %d, want %d", got, want)
	}
	if got, want := parsePort("9090", 8080), 9090; got != want {
		t.Fatalf("unexpected parsed port: got %d, want %d", got, want)
	}
	if got, want := parsePort(":9090", 8080), 9090; got != want {
		t.Fatalf("unexpected parsed prefixed port: got %d, want %d", got, want)
	}
	if got, want := parsePort("invalid", 8080), 8080; got != want {
		t.Fatalf("unexpected fallback port: got %d, want %d", got, want)
	}
}

func TestParseHost(t *testing.T) {
	if got, want := parseHost("", "localhost"), "localhost"; got != want {
		t.Fatalf("unexpected fallback host: got %q, want %q", got, want)
	}
	if got, want := parseHost("0.0.0.0", "localhost"), "0.0.0.0"; got != want {
		t.Fatalf("unexpected parsed host: got %q, want %q", got, want)
	}
}
