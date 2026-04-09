package http

import (
	nethttp "net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseCaptureAndNewHttpResponse(t *testing.T) {
	rw := httptest.NewRecorder()
	capture := NewResponseCapture(rw)

	capture.Header().Set("Content-Type", "application/json")
	capture.Header().Add("X-Test", "one")
	capture.Header().Add("X-Test", "two")
	capture.WriteHeader(nethttp.StatusCreated)
	_, _ = capture.Write([]byte(`{"ok":true}`))

	got := NewHttpResponse(capture)

	if got.StatusCode != nethttp.StatusCreated {
		t.Fatalf("StatusCode = %d, want %d", got.StatusCode, nethttp.StatusCreated)
	}
	if got.Headers["X-Test"] != "one" {
		t.Fatalf("Headers[X-Test] = %q, want one", got.Headers["X-Test"])
	}
	if got.ContentType != "application/json" {
		t.Fatalf("ContentType = %q, want application/json", got.ContentType)
	}
	if got.ContentLength != int64(len(`{"ok":true}`)) {
		t.Fatalf("ContentLength = %d, want %d", got.ContentLength, len(`{"ok":true}`))
	}
}

func TestNewHttpResponse_UsesContentLengthHeaderWhenValid(t *testing.T) {
	rw := httptest.NewRecorder()
	capture := NewResponseCapture(rw)
	capture.Header().Set("Content-Length", "77")
	_, _ = capture.Write([]byte("abc"))

	got := NewHttpResponse(capture)
	if got.ContentLength != 77 {
		t.Fatalf("ContentLength = %d, want 77", got.ContentLength)
	}
}

func TestNewHttpResponse_FallsBackToBytesWrittenWhenContentLengthInvalid(t *testing.T) {
	rw := httptest.NewRecorder()
	capture := NewResponseCapture(rw)
	capture.Header().Set("Content-Length", "invalid")
	_, _ = capture.Write([]byte("abcd"))

	got := NewHttpResponse(capture)
	if got.ContentLength != 4 {
		t.Fatalf("ContentLength = %d, want 4", got.ContentLength)
	}
}
