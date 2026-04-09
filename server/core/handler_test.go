package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type requestEcho struct {
	Method        string              `json:"method"`
	URL           string              `json:"url"`
	Path          string              `json:"path"`
	RawQuery      string              `json:"raw_query"`
	Query         map[string][]string `json:"query"`
	Protocol      string              `json:"protocol"`
	Host          string              `json:"host"`
	RemoteAddr    string              `json:"remote_addr"`
	Headers       map[string][]string `json:"headers"`
	ContentLength int64               `json:"content_length"`
	Body          string              `json:"body"`
}

type echoResponse struct {
	Request requestEcho `json:"request"`
}

func TestNewRequestEchoHandler_EchoesRequestData(t *testing.T) {
	cfg := ServerConfig{ResponseContentType: "application/json; charset=utf-8"}
	h := NewRequestEchoHandler(cfg)

	reqBody := `{"name":"lampa"}`
	r := httptest.NewRequest(http.MethodPost, "http://dector.space/api/echo?tag=go&tag=test", strings.NewReader(reqBody))
	r.RemoteAddr = "203.0.113.10:54321"
	r.Header.Set("Content-Type", "application/json")
	r.Header.Add("X-Trace-Id", "abc123")
	r.Header.Add("X-Trace-Id", "def456")

	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)

	if rw.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rw.Code, http.StatusOK)
	}

	var got echoResponse
	if err := json.Unmarshal(rw.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response JSON: %v", err)
	}

	want := requestEcho{
		Method:   http.MethodPost,
		URL:      "http://dector.space/api/echo?tag=go&tag=test",
		Path:     "/api/echo",
		RawQuery: "tag=go&tag=test",
		Query: map[string][]string{
			"tag": {"go", "test"},
		},
		Protocol:   "HTTP/1.1",
		Host:       "dector.space",
		RemoteAddr: "203.0.113.10:54321",
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
			"X-Trace-Id":   {"abc123", "def456"},
		},
		ContentLength: int64(len(reqBody)),
		Body:          reqBody,
	}

	if diff := cmp.Diff(want, got.Request); diff != "" {
		t.Fatalf("request mismatch (-want +got):\n%s", diff)
	}
}

func TestNewRequestEchoHandler_SetsConfiguredContentType(t *testing.T) {
	cfg := ServerConfig{ResponseContentType: "application/vnd.api+json"}
	h := NewRequestEchoHandler(cfg)

	r := httptest.NewRequest(http.MethodGet, "http://dector.space/", nil)
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, r)

	if got := rw.Header().Get("Content-Type"); got != "application/vnd.api+json" {
		t.Fatalf("content-type = %q, want %q", got, "application/vnd.api+json")
	}
}
