package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type echoResponse struct {
	Request struct {
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
	} `json:"request"`
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

	if got.Request.Method != http.MethodPost {
		t.Fatalf("method = %q, want %q", got.Request.Method, http.MethodPost)
	}
	if got.Request.URL != "/api/echo?tag=go&tag=test" {
		t.Fatalf("url = %q", got.Request.URL)
	}
	if got.Request.Path != "/api/echo" {
		t.Fatalf("path = %q", got.Request.Path)
	}
	if got.Request.RawQuery != "tag=go&tag=test" {
		t.Fatalf("raw_query = %q", got.Request.RawQuery)
	}
	if len(got.Request.Query["tag"]) != 2 || got.Request.Query["tag"][0] != "go" || got.Request.Query["tag"][1] != "test" {
		t.Fatalf("query[tag] = %#v, want [go test]", got.Request.Query["tag"])
	}
	if got.Request.Protocol != "HTTP/1.1" {
		t.Fatalf("protocol = %q, want HTTP/1.1", got.Request.Protocol)
	}
	if got.Request.Host != "dector.space" {
		t.Fatalf("host = %q, want dector.space", got.Request.Host)
	}
	if got.Request.RemoteAddr != "203.0.113.10:54321" {
		t.Fatalf("remote_addr = %q", got.Request.RemoteAddr)
	}
	if got.Request.Headers["Content-Type"][0] != "application/json" {
		t.Fatalf("headers[Content-Type] = %#v", got.Request.Headers["Content-Type"])
	}
	if len(got.Request.Headers["X-Trace-Id"]) != 2 || got.Request.Headers["X-Trace-Id"][0] != "abc123" || got.Request.Headers["X-Trace-Id"][1] != "def456" {
		t.Fatalf("headers[X-Trace-Id] = %#v", got.Request.Headers["X-Trace-Id"])
	}
	if got.Request.ContentLength != int64(len(reqBody)) {
		t.Fatalf("content_length = %d, want %d", got.Request.ContentLength, len(reqBody))
	}
	if got.Request.Body != reqBody {
		t.Fatalf("body = %q, want %q", got.Request.Body, reqBody)
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
