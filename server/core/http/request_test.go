package http

import (
	nethttp "net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHttpRequest(t *testing.T) {
	r := httptest.NewRequest(nethttp.MethodPost, "https://dector.space/api/v1/items?tag=go&tag=http&sort=desc", nil)
	r.RemoteAddr = "203.0.113.10:54321"
	r.Header.Set("User-Agent", "lampa-test/1.0")
	r.Header.Set("Referer", "https://blog.dector.space/page")
	r.Header.Set("Content-Type", "application/json")
	r.Header["X-Multi"] = []string{"first", "second"}
	r.ContentLength = 123

	got := NewHttpRequest(r)

	if got.Method != nethttp.MethodPost {
		t.Fatalf("Method = %q, want %q", got.Method, nethttp.MethodPost)
	}
	if got.Scheme != "https" {
		t.Fatalf("Scheme = %q, want https", got.Scheme)
	}
	if got.Host != "dector.space" {
		t.Fatalf("Host = %q, want dector.space", got.Host)
	}
	if got.Path != "/api/v1/items" {
		t.Fatalf("Path = %q, want /api/v1/items", got.Path)
	}
	if got.RawQuery != "tag=go&tag=http&sort=desc" {
		t.Fatalf("RawQuery = %q", got.RawQuery)
	}
	if len(got.Query["tag"]) != 2 || got.Query["tag"][0] != "go" || got.Query["tag"][1] != "http" {
		t.Fatalf("Query[tag] = %#v, want [go http]", got.Query["tag"])
	}
	if got.Headers["X-Multi"] != "first" {
		t.Fatalf("Headers[X-Multi] = %q, want first", got.Headers["X-Multi"])
	}
	if got.RemoteAddr != "203.0.113.10:54321" {
		t.Fatalf("RemoteAddr = %q", got.RemoteAddr)
	}
	if got.UserAgent != "lampa-test/1.0" {
		t.Fatalf("UserAgent = %q", got.UserAgent)
	}
	if got.Referer != "https://blog.dector.space/page" {
		t.Fatalf("Referer = %q", got.Referer)
	}
	if got.ContentType != "application/json" {
		t.Fatalf("ContentType = %q", got.ContentType)
	}
	if got.ContentLength != 123 {
		t.Fatalf("ContentLength = %d, want 123", got.ContentLength)
	}
	if got.Protocol != "HTTP/1.1" {
		t.Fatalf("Protocol = %q, want HTTP/1.1", got.Protocol)
	}
}
