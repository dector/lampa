package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dector/lampa/server/core/processor"
)

func TestControlProcSetHandler_Success(t *testing.T) {
	store := processor.NewInMemoryReqProcessorStore(nil, nil)
	handler := NewControlProcSetHandler(store)

	body := []byte(`{"kind":"static","endpoint":"/example","response":{"status":200,"contentType":"application/json","headers":{"X-Test":["1"]},"body":"{\"ok\":true}"}}`)
	req := httptest.NewRequest(http.MethodPost, RouteProcSet, bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}

	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response body as JSON: %v", err)
	}
	if got, want := payload["status"], "ok"; got != want {
		t.Fatalf("unexpected status payload: got %v, want %v", got, want)
	}
	if got, want := payload["endpoint"], "/example"; got != want {
		t.Fatalf("unexpected endpoint payload: got %v, want %v", got, want)
	}

	proc, ok := store.EndpointProcessor("/example")
	if !ok {
		t.Fatal("expected processor to be stored")
	}

	staticProc, ok := proc.(processor.StaticReqProcessor)
	if !ok {
		t.Fatalf("unexpected processor type: %T", proc)
	}
	if got, want := staticProc.StatusCode, 200; got != want {
		t.Fatalf("unexpected status code in processor: got %d, want %d", got, want)
	}
	if got, want := staticProc.Headers.Get("Content-Type"), "application/json"; got != want {
		t.Fatalf("unexpected content type in processor: got %q, want %q", got, want)
	}
	xTestValues := staticProc.Headers.Values("X-Test")
	if len(xTestValues) != 1 {
		t.Fatalf("unexpected X-Test values: %#v", xTestValues)
	}
	if got, want := xTestValues[0], "1"; got != want {
		t.Fatalf("unexpected X-Test header value: got %q, want %q", got, want)
	}
	if got, want := string(staticProc.Body), `{"ok":true}`; got != want {
		t.Fatalf("unexpected processor body: got %q, want %q", got, want)
	}
}

func TestControlProcSetHandler_JS_Success(t *testing.T) {
	store := processor.NewInMemoryReqProcessorStore(nil, nil)
	handler := NewControlProcSetHandler(store)

	body := []byte(`{"kind":"js","endpoint":"/js","js":{"script":"function handle(req){ return Response.json({ok:true}); }"}}`)
	req := httptest.NewRequest(http.MethodPost, RouteProcSet, bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusOK; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}

	proc, ok := store.EndpointProcessor("/js")
	if !ok {
		t.Fatal("expected processor to be stored")
	}

	jsProc, ok := proc.(processor.QuickJSReqProcessor)
	if !ok {
		t.Fatalf("unexpected processor type: %T", proc)
	}
	if jsProc.Script == "" {
		t.Fatal("expected non-empty script")
	}
}

func TestControlProcSetHandler_JS_MissingConfig(t *testing.T) {
	store := processor.NewInMemoryReqProcessorStore(nil, nil)
	handler := NewControlProcSetHandler(store)

	body := []byte(`{"kind":"js","endpoint":"/js"}`)
	req := httptest.NewRequest(http.MethodPost, RouteProcSet, bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusBadRequest; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}
}

func TestControlProcSetHandler_JS_InvalidScript(t *testing.T) {
	store := processor.NewInMemoryReqProcessorStore(nil, nil)
	handler := NewControlProcSetHandler(store)

	body := []byte(`{"kind":"js","endpoint":"/js","js":{"script":"  "}}`)
	req := httptest.NewRequest(http.MethodPost, RouteProcSet, bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusBadRequest; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}

	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response body as JSON: %v", err)
	}
	if got, want := payload["error"], "invalid js script"; got != want {
		t.Fatalf("unexpected error payload: got %v, want %v", got, want)
	}
}

func TestControlProcSetHandler_BadJSON(t *testing.T) {
	store := processor.NewInMemoryReqProcessorStore(nil, nil)
	handler := NewControlProcSetHandler(store)

	req := httptest.NewRequest(http.MethodPost, RouteProcSet, bytes.NewReader([]byte(`{"kind":`)))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusBadRequest; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}
}

func TestControlProcSetHandler_InvalidEndpoint(t *testing.T) {
	store := processor.NewInMemoryReqProcessorStore(nil, nil)
	handler := NewControlProcSetHandler(store)

	body := []byte(`{"kind":"static","endpoint":"example","response":{"status":200,"body":"ok"}}`)
	req := httptest.NewRequest(http.MethodPost, RouteProcSet, bytes.NewReader(body))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusBadRequest; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}

	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response body as JSON: %v", err)
	}
	if got, want := payload["error"], "invalid endpoint"; got != want {
		t.Fatalf("unexpected error payload: got %v, want %v", got, want)
	}
}

func TestControlProcSetHandler_InvalidStatus(t *testing.T) {
	store := processor.NewInMemoryReqProcessorStore(nil, nil)
	handler := NewControlProcSetHandler(store)

	body := []byte(`{"kind":"static","endpoint":"/example","response":{"status":99,"body":"ok"}}`)
	req := httptest.NewRequest(http.MethodPost, RouteProcSet, bytes.NewReader(body))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusBadRequest; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}

	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response body as JSON: %v", err)
	}
	if got, want := payload["error"], "invalid status"; got != want {
		t.Fatalf("unexpected error payload: got %v, want %v", got, want)
	}
}
