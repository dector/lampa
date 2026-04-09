package http

import (
	nethttp "net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHttpResponse_ContentType(t *testing.T) {
	r := HttpResponse{
		Headers: nethttp.Header{"Content-Type": {"application/json"}},
	}

	if diff := cmp.Diff("application/json", r.ContentType()); diff != "" {
		t.Fatalf("ContentType() mismatch (-want +got):\n%s", diff)
	}
}

func TestHttpResponse_SetContentType_InitializesHeaders(t *testing.T) {
	var r HttpResponse

	r.SetContentType("text/plain")

	if r.Headers == nil {
		t.Fatal("Headers is nil, expected initialized header map")
	}
	if diff := cmp.Diff("text/plain", r.ContentType()); diff != "" {
		t.Fatalf("ContentType() mismatch (-want +got):\n%s", diff)
	}
}
