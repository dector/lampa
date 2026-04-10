package processor

import (
	"testing"
)

func TestInMemoryReqProcessorStore_SettersAndGetters(t *testing.T) {
	store := NewInMemoryReqProcessorStore(nil, nil)

	if _, ok := store.EndpointProcessor("/"); ok {
		t.Fatal("did not expect endpoint processor for '/'")
	}
	if fallback := store.FallbackProcessor(); fallback != nil {
		t.Fatalf("expected nil fallback, got %T", fallback)
	}
	if got, want := store.ResponsesCount(), 0; got != want {
		t.Fatalf("unexpected response count: got %d, want %d", got, want)
	}

	store.SetEndpointProcessor("/", QuickJSReqProcessor{Script: okJsProcessor})
	store.SetFallbackProcessor(StaticReqProcessor{StatusCode: 404, Body: []byte("Not Found")})

	endpointProcessor, ok := store.EndpointProcessor("/")
	if !ok {
		t.Fatal("expected endpoint processor for '/'")
	}
	if _, ok := endpointProcessor.(QuickJSReqProcessor); !ok {
		t.Fatalf("processor['/'] should be QuickJSReqProcessor, got %T", endpointProcessor)
	}

	fallback, ok := store.FallbackProcessor().(StaticReqProcessor)
	if !ok {
		t.Fatalf("fallback should be StaticReqProcessor, got %T", store.FallbackProcessor())
	}
	if fallback.StatusCode != 404 {
		t.Fatalf("unexpected fallback status code: got %d, want 404", fallback.StatusCode)
	}
	if got, want := store.ResponsesCount(), 2; got != want {
		t.Fatalf("unexpected response count: got %d, want %d", got, want)
	}
}

func TestNewInMemoryReqProcessorStore_ClonesProcessorsMap(t *testing.T) {
	initial := EndpointProcessors{
		"/": QuickJSReqProcessor{Script: okJsProcessor},
	}
	store := NewInMemoryReqProcessorStore(initial, nil)

	initial["/health"] = StaticReqProcessor{StatusCode: 200, Body: []byte("ok")}

	if _, ok := store.EndpointProcessor("/health"); ok {
		t.Fatal("unexpected endpoint processor for '/health'; expected store to keep cloned snapshot")
	}
}
