package quickjsexec

import (
	"errors"
	nethttp "net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"modernc.org/quickjs"
)

func TestFromResponseValue_Map(t *testing.T) {
	value := map[string]any{
		"statusCode": 201,
		"headers": map[string]any{
			"Content-Type": "application/json",
			"Set-Cookie":   []any{"a=1", "b=2"},
		},
		"body": "ok",
	}

	got, err := FromResponseValue(value)
	if err != nil {
		t.Fatalf("FromResponseValue() unexpected error: %v", err)
	}

	want := struct {
		StatusCode int
		Headers    nethttp.Header
		Body       []byte
	}{
		StatusCode: 201,
		Headers: nethttp.Header{
			"Content-Type": {"application/json"},
			"Set-Cookie":   {"a=1", "b=2"},
		},
		Body: []byte("ok"),
	}
	if diff := cmp.Diff(want.StatusCode, got.StatusCode); diff != "" {
		t.Fatalf("StatusCode mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(want.Headers, got.Headers); diff != "" {
		t.Fatalf("Headers mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(want.Body, got.Body); diff != "" {
		t.Fatalf("Body mismatch (-want +got):\n%s", diff)
	}
}

func TestFromResponseValue_ObjectWithUint8ArrayBody(t *testing.T) {
	vm, err := quickjs.NewVM()
	if err != nil {
		t.Fatalf("quickjs.NewVM() failed: %v", err)
	}
	defer vm.Close()

	if _, err := vm.Eval(`
		function makeResp() {
			return {
				statusCode: 202,
				headers: {"Content-Type": "application/octet-stream"},
				body: new Uint8Array([65, 66, 67]),
			};
		}
	`, quickjs.EvalGlobal); err != nil {
		t.Fatalf("vm.Eval() failed: %v", err)
	}

	value, err := vm.Call("makeResp")
	if err != nil {
		t.Fatalf("vm.Call() failed: %v", err)
	}

	got, err := FromResponseValue(value)
	if err != nil {
		t.Fatalf("FromResponseValue() unexpected error: %v", err)
	}

	if diff := cmp.Diff(202, got.StatusCode); diff != "" {
		t.Fatalf("StatusCode mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]byte("ABC"), got.Body); diff != "" {
		t.Fatalf("Body mismatch (-want +got):\n%s", diff)
	}
}

func TestFromResponseValue_InvalidShape(t *testing.T) {
	_, err := FromResponseValue(42)
	if err == nil {
		t.Fatal("expected error")
	}
	var target InvalidResponseTypeError
	if !errors.As(err, &target) {
		t.Fatalf("unexpected error type: %T", err)
	}
}

func TestFromResponseValue_InvalidHeadersShape(t *testing.T) {
	_, err := FromResponseValue(map[string]any{"headers": 42})
	if err == nil {
		t.Fatal("expected error")
	}
	var target InvalidHeadersTypeError
	if !errors.As(err, &target) {
		t.Fatalf("unexpected error type: %T", err)
	}
}

func TestFromResponseValue_InvalidBodyShape(t *testing.T) {
	_, err := FromResponseValue(map[string]any{"body": true})
	if err == nil {
		t.Fatal("expected error")
	}
	var target InvalidBodyTypeError
	if !errors.As(err, &target) {
		t.Fatalf("unexpected error type: %T", err)
	}
}
