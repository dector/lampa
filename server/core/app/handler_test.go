package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dector/lampa/server/core/logstore"
	"github.com/dector/lampa/server/core/processor"
)

func TestNewRequestHandlerWithStoreAndLogs_RecordsRequestAndResponse(t *testing.T) {
	store := processor.NewInMemoryReqProcessorStore(processor.EndpointProcessors{
		"/logged": processor.StaticReqProcessor{
			StatusCode: 201,
			Headers: http.Header{
				"X-Resp": []string{"ok"},
			},
			Body: []byte("done"),
		},
	}, processor.StaticReqProcessor{StatusCode: 404, Body: []byte("Not Found")})

	logs := logstore.NewInMemoryStore(1024 * 1024)
	h := NewRequestHandlerWithStoreAndLogs(DefaultServerConfig(), store, logs)

	req := httptest.NewRequest(http.MethodPost, "http://example.local/logged?x=1", strings.NewReader("payload"))
	req.Header.Set("X-Test", "1")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if got, want := rr.Code, 201; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}

	entries := logs.Latest(1)
	if got, want := len(entries), 1; got != want {
		t.Fatalf("unexpected logs len: got %d, want %d", got, want)
	}

	entry := entries[0]
	if got, want := entry.ID, int64(1); got != want {
		t.Fatalf("unexpected entry id: got %d, want %d", got, want)
	}
	if entry.Timestamp.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}
	if got, want := entry.Request.Method, http.MethodPost; got != want {
		t.Fatalf("unexpected request method: got %q, want %q", got, want)
	}
	if got, want := entry.Request.Path, "/logged"; got != want {
		t.Fatalf("unexpected request path: got %q, want %q", got, want)
	}
	if got, want := string(entry.Request.Body), "payload"; got != want {
		t.Fatalf("unexpected request body: got %q, want %q", got, want)
	}
	if got, want := entry.Request.Headers.Get("X-Test"), "1"; got != want {
		t.Fatalf("unexpected request header: got %q, want %q", got, want)
	}
	if got, want := entry.Response.Status, 201; got != want {
		t.Fatalf("unexpected response status: got %d, want %d", got, want)
	}
	if got, want := string(entry.Response.Body), "done"; got != want {
		t.Fatalf("unexpected response body: got %q, want %q", got, want)
	}
	if got, want := entry.Response.Headers.Get("X-Resp"), "ok"; got != want {
		t.Fatalf("unexpected response header: got %q, want %q", got, want)
	}
}
