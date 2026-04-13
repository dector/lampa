package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dector/lampa/server/core/logstore"
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

func TestControlProcSetHandler_Seq_Success(t *testing.T) {
	store := processor.NewInMemoryReqProcessorStore(nil, nil)
	handler := NewControlProcSetHandler(store)

	body := []byte(`{"kind":"seq","endpoint":"/seq","sequence":[{"response":{"status":500,"body":"fail","headers":{"X-Step":["1"]}}},{"response":{"status":200,"contentType":"application/json","body":"{\"ok\":true}"}}]}`)
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
	if got, want := payload["endpoint"], "/seq"; got != want {
		t.Fatalf("unexpected endpoint payload: got %v, want %v", got, want)
	}
	if got, want := payload["kind"], "seq"; got != want {
		t.Fatalf("unexpected kind payload: got %v, want %v", got, want)
	}

	proc, ok := store.EndpointProcessor("/seq")
	if !ok {
		t.Fatal("expected processor to be stored")
	}

	seqProc, ok := proc.(*processor.SequenceReqProcessor)
	if !ok {
		t.Fatalf("unexpected processor type: %T", proc)
	}
	if got, want := len(seqProc.Processors), 2; got != want {
		t.Fatalf("unexpected sequence processors count: got %d, want %d", got, want)
	}

	step1, ok := seqProc.Processors[0].(processor.StaticReqProcessor)
	if !ok {
		t.Fatalf("unexpected first step processor type: %T", seqProc.Processors[0])
	}
	if got, want := step1.StatusCode, 500; got != want {
		t.Fatalf("unexpected first step status: got %d, want %d", got, want)
	}
	if got, want := step1.Headers.Get("X-Step"), "1"; got != want {
		t.Fatalf("unexpected first step header: got %q, want %q", got, want)
	}

	step2, ok := seqProc.Processors[1].(processor.StaticReqProcessor)
	if !ok {
		t.Fatalf("unexpected second step processor type: %T", seqProc.Processors[1])
	}
	if got, want := step2.StatusCode, 200; got != want {
		t.Fatalf("unexpected second step status: got %d, want %d", got, want)
	}
	if got, want := step2.Headers.Get("Content-Type"), "application/json"; got != want {
		t.Fatalf("unexpected second step content type: got %q, want %q", got, want)
	}
}

func TestControlProcSetHandler_Seq_InvalidPayload(t *testing.T) {
	store := processor.NewInMemoryReqProcessorStore(nil, nil)
	handler := NewControlProcSetHandler(store)

	t.Run("empty sequence", func(t *testing.T) {
		body := []byte(`{"kind":"seq","endpoint":"/seq","sequence":[]}`)
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
		if got, want := payload["error"], "invalid sequence"; got != want {
			t.Fatalf("unexpected error payload: got %v, want %v", got, want)
		}
	})

	t.Run("invalid step status", func(t *testing.T) {
		body := []byte(`{"kind":"seq","endpoint":"/seq","sequence":[{"response":{"status":200,"body":"ok"}},{"response":{"status":99,"body":"bad"}}]}`)
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
		if got, want := payload["error"], "invalid sequence step 2: invalid status"; got != want {
			t.Fatalf("unexpected error payload: got %v, want %v", got, want)
		}
	})
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

func TestControlProcSetDefaultHandler_Success(t *testing.T) {
	store := processor.NewInMemoryReqProcessorStore(nil, nil)
	handler := NewControlProcSetDefaultHandler(store)

	body := []byte(`{"kind":"pass","server":"http://example.local:8080"}`)
	req := httptest.NewRequest(http.MethodPost, RouteProcSetDefault, bytes.NewReader(body))
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
	if got, want := payload["kind"], "pass"; got != want {
		t.Fatalf("unexpected kind payload: got %v, want %v", got, want)
	}

	fallback := store.FallbackProcessor()
	proc, ok := fallback.(processor.PassthroughReqProcessor)
	if !ok {
		t.Fatalf("unexpected fallback processor type: %T", fallback)
	}
	if got, want := proc.Server, "http://example.local:8080"; got != want {
		t.Fatalf("unexpected fallback server: got %q, want %q", got, want)
	}
}

func TestControlProcSetDefaultHandler_InvalidKind(t *testing.T) {
	store := processor.NewInMemoryReqProcessorStore(nil, nil)
	handler := NewControlProcSetDefaultHandler(store)

	body := []byte(`{"kind":"static","server":"http://example.local:8080"}`)
	req := httptest.NewRequest(http.MethodPost, RouteProcSetDefault, bytes.NewReader(body))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusBadRequest; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}
}

func TestControlProcSetDefaultHandler_InvalidServer(t *testing.T) {
	store := processor.NewInMemoryReqProcessorStore(nil, nil)
	handler := NewControlProcSetDefaultHandler(store)

	body := []byte(`{"kind":"pass","server":"example.local:8080"}`)
	req := httptest.NewRequest(http.MethodPost, RouteProcSetDefault, bytes.NewReader(body))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusBadRequest; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}

	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to parse response body as JSON: %v", err)
	}
	if got, want := payload["error"], "invalid server"; got != want {
		t.Fatalf("unexpected error payload: got %v, want %v", got, want)
	}
}

func TestControlProxyLogsHandler_Success(t *testing.T) {
	logs := logstore.NewInMemoryStore(1024 * 1024)
	logs.Add(logstore.Entry{Timestamp: time.Unix(1, 0).UTC(), Request: logstore.Request{Path: "/one"}})
	logs.Add(logstore.Entry{Timestamp: time.Unix(2, 0).UTC(), Request: logstore.Request{Path: "/two"}})
	logs.Add(logstore.Entry{Timestamp: time.Unix(3, 0).UTC(), Request: logstore.Request{Path: "/three"}})

	handler := NewControlProxyLogsHandler(logs)
	req := httptest.NewRequest(http.MethodGet, RouteProxyLogs+"?n=2", nil)
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
		t.Fatalf("unexpected status: got %v, want %v", got, want)
	}
	if got, want := payload["count"], float64(2); got != want {
		t.Fatalf("unexpected count: got %v, want %v", got, want)
	}
	if got, want := payload["total"], float64(3); got != want {
		t.Fatalf("unexpected total: got %v, want %v", got, want)
	}

	entries, ok := payload["entries"].([]any)
	if !ok {
		t.Fatalf("unexpected entries shape: %#v", payload["entries"])
	}
	if got, want := len(entries), 2; got != want {
		t.Fatalf("unexpected entries len: got %d, want %d", got, want)
	}

	entry0, ok := entries[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected first entry shape: %#v", entries[0])
	}
	request0, ok := entry0["request"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected first entry request shape: %#v", entry0["request"])
	}
	if got, want := request0["path"], "/three"; got != want {
		t.Fatalf("unexpected first entry path: got %v, want %v", got, want)
	}
}

func TestControlProxyLogsHandler_InvalidN(t *testing.T) {
	handler := NewControlProxyLogsHandler(logstore.NewInMemoryStore(1024))
	req := httptest.NewRequest(http.MethodGet, RouteProxyLogs+"?n=abc", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusBadRequest; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}
}

func TestControlProxyLogsHandler_NoStore(t *testing.T) {
	handler := NewControlProxyLogsHandler(nil)
	req := httptest.NewRequest(http.MethodGet, RouteProxyLogs, nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got, want := rr.Code, http.StatusInternalServerError; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}
}
