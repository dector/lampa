package exported

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig_ControlRoutePath(t *testing.T) {
	cfg := DefaultConfig()
	if got, want := cfg.ControlRoutePath, "/"; got != want {
		t.Fatalf("unexpected control route path: got %q, want %q", got, want)
	}
}

func TestServer_StartAsync_ControlPingEndpoint(t *testing.T) {
	proxyAddr := reserveAddress(t)
	controlAddr := reserveAddress(t)

	srv := NewServerWithConfig(&ServerConfig{
		ProxyListenAddress:   proxyAddr,
		ControlListenAddress: controlAddr,
	})

	if err := srv.StartAsync(); err != "" {
		t.Fatalf("StartAsync returned unexpected error: %s", err)
	}
	defer func() {
		if err := srv.Stop(); err != "" {
			t.Fatalf("Stop returned unexpected error: %s", err)
		}
	}()

	assertPingResponse(t, "http://"+controlAddr+"/ping")
}

func TestServer_StartAsync_ControlPingEndpoint_ResolvesAgainstControlRoutePath(t *testing.T) {
	proxyAddr := reserveAddress(t)
	controlAddr := reserveAddress(t)

	srv := NewServerWithConfig(&ServerConfig{
		ProxyListenAddress:   proxyAddr,
		ControlListenAddress: controlAddr,
		ControlRoutePath:     "/control",
	})

	if err := srv.StartAsync(); err != "" {
		t.Fatalf("StartAsync returned unexpected error: %s", err)
	}
	defer func() {
		if err := srv.Stop(); err != "" {
			t.Fatalf("Stop returned unexpected error: %s", err)
		}
	}()

	assertPingResponse(t, "http://"+controlAddr+"/control/ping")
}

func assertPingResponse(t *testing.T, url string) {
	t.Helper()

	resp, err := waitForGet(url)
	if err != nil {
		t.Fatalf("failed to call control ping endpoint: %v", err)
	}
	defer resp.Body.Close()

	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatalf("failed to parse response body as JSON: %v", err)
	}
	if got, want := body["status"], "ok"; got != want {
		t.Fatalf("unexpected status: got %v, want %v", got, want)
	}
	responses, ok := body["responses"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected responses shape: %#v", body["responses"])
	}
	if got, want := responses["count"], float64(2); got != want {
		t.Fatalf("unexpected responses.count: got %v, want %v", got, want)
	}
	if got := resp.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
		t.Fatalf("unexpected content type: got %q", got)
	}
}

func reserveAddress(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve address: %v", err)
	}
	defer ln.Close()
	return ln.Addr().String()
}

func waitForGet(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 200 * time.Millisecond}
	deadline := time.Now().Add(3 * time.Second)
	for {
		resp, err := client.Get(url)
		if err == nil {
			return resp, nil
		}
		if time.Now().After(deadline) {
			return nil, err
		}
		time.Sleep(50 * time.Millisecond)
	}
}
