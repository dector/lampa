package http

import (
	"crypto/tls"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestScheme(t *testing.T) {
	httpReq := httptest.NewRequest(nethttp.MethodGet, "http://dector.space", nil)
	httpsReq := httptest.NewRequest(nethttp.MethodGet, "http://dector.space", nil)
	httpsReq.TLS = &tls.ConnectionState{}

	if got := requestScheme(httpReq); got != "http" {
		t.Fatalf("requestScheme(http) = %q, want http", got)
	}
	if got := requestScheme(httpsReq); got != "https" {
		t.Fatalf("requestScheme(https) = %q, want https", got)
	}
}

func TestFirstHeaderValues(t *testing.T) {
	h := nethttp.Header{}
	h["X-One"] = []string{"a", "b"}
	h["X-Empty"] = []string{}

	got := firstHeaderValues(h)

	if got["X-One"] != "a" {
		t.Fatalf("X-One = %q, want a", got["X-One"])
	}
	if _, exists := got["X-Empty"]; exists {
		t.Fatalf("X-Empty should not exist in result")
	}
}
