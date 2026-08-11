package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status code: got %d, want %d", rr.Code, http.StatusNotFound)
		}
		if got := rr.Body.String(); got != "Not Found" {
			t.Fatalf("unexpected body: got %q, want %q", got, "Not Found")
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

func TestRunWithControl_RegistersPingHandler(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.ListenAddress = "localhost:18090"
	cfg.ControlListenAddress = "localhost:18091"

	controlChecked := make(chan struct{})

	err := RunWithControl(cfg, func(addr string, h http.Handler) error {
		switch addr {
		case cfg.ControlListenAddress:
			assertControlPingEndpoint(t, h, "/ping")
			close(controlChecked)
			return nil
		case cfg.ListenAddress:
			select {
			case <-controlChecked:
				return nil
			case <-time.After(250 * time.Millisecond):
				return errors.New("timeout waiting for control handler check")
			}
		default:
			return errors.New("unexpected listen address")
		}
	})

	if err != nil {
		t.Fatalf("RunWithControl returned unexpected error: %v", err)
	}
}

func TestRunWithControl_DoesNotRegisterRootAliasForPing(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.ListenAddress = "localhost:18094"
	cfg.ControlListenAddress = "localhost:18095"

	controlChecked := make(chan struct{})

	err := RunWithControl(cfg, func(addr string, h http.Handler) error {
		switch addr {
		case cfg.ControlListenAddress:
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != http.StatusNotFound {
				t.Fatalf("unexpected status code: got %d, want %d", rr.Code, http.StatusNotFound)
			}

			close(controlChecked)
			return nil
		case cfg.ListenAddress:
			select {
			case <-controlChecked:
				return nil
			case <-time.After(250 * time.Millisecond):
				return errors.New("timeout waiting for control handler check")
			}
		default:
			return errors.New("unexpected listen address")
		}
	})

	if err != nil {
		t.Fatalf("RunWithControl returned unexpected error: %v", err)
	}
}

func TestRunWithControl_RegistersProcCountHandler(t *testing.T) {
	cfg := DefaultServerConfig()
	cfg.ListenAddress = "localhost:18092"
	cfg.ControlListenAddress = "localhost:18093"

	controlChecked := make(chan struct{})

	err := RunWithControl(cfg, func(addr string, h http.Handler) error {
		switch addr {
		case cfg.ControlListenAddress:
			req := httptest.NewRequest(http.MethodGet, "/api/v0/proc_count", nil)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("unexpected status code: got %d, want %d", rr.Code, http.StatusOK)
			}
			var payload map[string]any
			if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
				t.Fatalf("failed to parse response body as JSON: %v", err)
			}
			if got, want := payload["count"], float64(2); got != want {
				t.Fatalf("unexpected count: got %v, want %v", got, want)
			}
			if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
				t.Fatalf("unexpected content type: got %q", got)
			}

			close(controlChecked)
			return nil
		case cfg.ListenAddress:
			select {
			case <-controlChecked:
				return nil
			case <-time.After(250 * time.Millisecond):
				return errors.New("timeout waiting for control handler check")
			}
		default:
			return errors.New("unexpected listen address")
		}
	})

	if err != nil {
		t.Fatalf("RunWithControl returned unexpected error: %v", err)
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

func assertControlPingEndpoint(t *testing.T, h http.Handler, requestPath string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, requestPath, nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want %d", rr.Code, http.StatusOK)
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
	if got, want := responses["count"], float64(2); got != want {
		t.Fatalf("unexpected responses.count: got %v, want %v", got, want)
	}
	if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("unexpected content type: got %q", got)
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

func TestLoadRuntimeConfig_WebUIDisabledByDefault(t *testing.T) {
	cfg := LoadRuntimeConfig(func(string) string { return "" })

	if cfg.WebUI.Enabled {
		t.Fatal("expected Web UI to be disabled by default")
	}
	if cfg.CaptureTraffic {
		t.Fatal("expected traffic capture to be disabled by default")
	}
	if got, want := cfg.WebUI.Port, 8880; got != want {
		t.Fatalf("unexpected Web UI port: got %d, want %d", got, want)
	}
	if got, want := cfg.WebUI.BindHost, "127.0.0.1"; got != want {
		t.Fatalf("unexpected Web UI host: got %q, want %q", got, want)
	}
	if got, want := cfg.WebUI.ListenAddress, "127.0.0.1:8880"; got != want {
		t.Fatalf("unexpected Web UI listen address: got %q, want %q", got, want)
	}
}

func TestLoadRuntimeConfigWithOptions_EnablesWebUIWithDefaultPort(t *testing.T) {
	cfg := LoadRuntimeConfigWithOptions(func(string) string { return "" }, RuntimeOptions{WebUIEnabled: true})

	if !cfg.WebUI.Enabled {
		t.Fatal("expected Web UI to be enabled")
	}
	if !cfg.CaptureTraffic {
		t.Fatal("expected traffic capture to be enabled")
	}
	if got, want := cfg.WebUI.Port, 8880; got != want {
		t.Fatalf("unexpected Web UI port: got %d, want %d", got, want)
	}
	if got, want := cfg.WebUI.ListenAddress, "127.0.0.1:8880"; got != want {
		t.Fatalf("unexpected Web UI listen address: got %q, want %q", got, want)
	}
}

func TestLoadRuntimeConfigWithOptions_EnablesWebUIWithCustomPort(t *testing.T) {
	cfg := LoadRuntimeConfigWithOptions(func(string) string { return "" }, RuntimeOptions{WebUIEnabled: true, WebUIPort: 8890})

	if !cfg.WebUI.Enabled {
		t.Fatal("expected Web UI to be enabled")
	}
	if !cfg.CaptureTraffic {
		t.Fatal("expected traffic capture to be enabled")
	}
	if got, want := cfg.WebUI.Port, 8890; got != want {
		t.Fatalf("unexpected Web UI port: got %d, want %d", got, want)
	}
	if got, want := cfg.WebUI.ListenAddress, "127.0.0.1:8890"; got != want {
		t.Fatalf("unexpected Web UI listen address: got %q, want %q", got, want)
	}
}

func TestParseRuntimeOptions_ParsesWebUIWithEqualsPort(t *testing.T) {
	options, err := parseRuntimeOptions([]string{"--webui", "--webui-port=8891"}, DefaultServerConfig())
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	cfg := LoadRuntimeConfigWithOptions(func(string) string { return "" }, options)

	if !cfg.WebUI.Enabled {
		t.Fatal("expected Web UI to be enabled")
	}
	if !cfg.CaptureTraffic {
		t.Fatal("expected traffic capture to be enabled")
	}
	if got, want := cfg.WebUI.Port, 8891; got != want {
		t.Fatalf("unexpected Web UI port: got %d, want %d", got, want)
	}
	if got, want := cfg.WebUI.ListenAddress, "127.0.0.1:8891"; got != want {
		t.Fatalf("unexpected Web UI listen address: got %q, want %q", got, want)
	}
}

func TestParseRuntimeOptions_RejectsInvalidWebUIPort(t *testing.T) {
	_, err := parseRuntimeOptions([]string{"--webui", "--webui-port", "70000"}, DefaultServerConfig())
	if err == nil {
		t.Fatal("expected invalid Web UI port error")
	}
	if !strings.Contains(err.Error(), "invalid --webui-port") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseRuntimeOptions_RejectsWebUIPositionalPort(t *testing.T) {
	_, err := parseRuntimeOptions([]string{"--webui", "8890"}, DefaultServerConfig())
	if err == nil {
		t.Fatal("expected positional argument error")
	}
	if !strings.Contains(err.Error(), "unexpected positional argument") {
		t.Fatalf("unexpected error: %v", err)
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
