package http

import (
	"io"
	nethttp "net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHttpRequest_ContentType(t *testing.T) {
	r := HttpRequest{
		Headers: nethttp.Header{"Content-Type": {"application/json"}},
	}

	if diff := cmp.Diff("application/json", r.ContentType()); diff != "" {
		t.Fatalf("ContentType() mismatch (-want +got):\n%s", diff)
	}
}

func TestHttpRequest_SetContentType_InitializesHeaders(t *testing.T) {
	var r HttpRequest

	r.SetContentType("text/plain")

	if r.Headers == nil {
		t.Fatal("Headers is nil, expected initialized header map")
	}
	if diff := cmp.Diff("text/plain", r.ContentType()); diff != "" {
		t.Fatalf("ContentType() mismatch (-want +got):\n%s", diff)
	}
}

func TestNewHttpRequest_CopiesMethodURLHeadersBody(t *testing.T) {
	body := "hello world"
	r := &nethttp.Request{
		Method: "POST",
		URL: &url.URL{
			Path:     "/api/items",
			RawQuery: "a=1&b=2",
		},
		Header: nethttp.Header{"X-Test": {"v1", "v2"}},
		Body:   io.NopCloser(strings.NewReader(body)),
	}

	got := NewHttpRequest(r)

	want := HttpRequest{
		Method: "POST",
		Url: url.URL{
			Path:     "/api/items",
			RawQuery: "a=1&b=2",
		},
		Headers: nethttp.Header{"X-Test": {"v1", "v2"}},
		Body:    []byte(body),
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("NewHttpRequest() mismatch (-want +got):\n%s", diff)
	}

	if got.Headers == nil {
		t.Fatal("Headers is nil, want copied headers")
	}
	if diff := cmp.Diff("v1", got.Headers.Get("X-Test")); diff != "" {
		t.Fatalf("Headers.Get(X-Test) mismatch (-want +got):\n%s", diff)
	}
	if &got.Headers == &r.Header {
		t.Fatal("Headers should be a clone, but references are identical")
	}

	got.Headers.Set("X-Test", "changed")
	if r.Header.Get("X-Test") == "changed" {
		t.Fatal("mutating copied headers should not mutate original request headers")
	}
}

func TestNewHttpRequest_RestoresOriginalRequestBody(t *testing.T) {
	body := "payload"
	r := &nethttp.Request{
		Method: "PUT",
		URL:    &url.URL{Path: "/upload"},
		Body:   io.NopCloser(strings.NewReader(body)),
		Header: make(nethttp.Header),
	}

	_ = NewHttpRequest(r)

	restored, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("reading restored request body failed: %v", err)
	}
	if diff := cmp.Diff([]byte(body), restored); diff != "" {
		t.Fatalf("restored body mismatch (-want +got):\n%s", diff)
	}
}

func TestNewHttpRequest_WithNilURL(t *testing.T) {
	r := &nethttp.Request{
		Method: "GET",
		URL:    nil,
		Header: make(nethttp.Header),
		Body:   io.NopCloser(strings.NewReader("")),
	}

	got := NewHttpRequest(r)

	if diff := cmp.Diff(url.URL{}, got.Url); diff != "" {
		t.Fatalf("Url mismatch (-want +got):\n%s", diff)
	}
}
