package logstore

import (
	"bytes"
	"net/http"
	"testing"
)

func TestInMemoryStore_LatestReturnsLatestFirst(t *testing.T) {
	store := NewInMemoryStore(1024 * 1024)

	store.Add(Entry{Request: Request{Method: http.MethodGet, Path: "/1"}})
	store.Add(Entry{Request: Request{Method: http.MethodGet, Path: "/2"}})
	store.Add(Entry{Request: Request{Method: http.MethodGet, Path: "/3"}})

	latest := store.Latest(2)
	if got, want := len(latest), 2; got != want {
		t.Fatalf("unexpected latest len: got %d, want %d", got, want)
	}
	if got, want := latest[0].Request.Path, "/3"; got != want {
		t.Fatalf("unexpected latest[0] path: got %q, want %q", got, want)
	}
	if got, want := latest[1].Request.Path, "/2"; got != want {
		t.Fatalf("unexpected latest[1] path: got %q, want %q", got, want)
	}
}

func TestInMemoryStore_EvictsOldestWhenExceedsCap(t *testing.T) {
	e1 := Entry{Request: Request{Path: "/1", Body: bytes.Repeat([]byte("a"), 256)}}
	e2 := Entry{Request: Request{Path: "/2", Body: bytes.Repeat([]byte("b"), 256)}}
	e3 := Entry{Request: Request{Path: "/3", Body: bytes.Repeat([]byte("c"), 256)}}

	capBytes := EstimateSizeBytes(e1) + EstimateSizeBytes(e2) + 1
	store := NewInMemoryStore(capBytes)

	store.Add(e1)
	store.Add(e2)
	store.Add(e3)

	if got, want := store.Count(), 2; got != want {
		t.Fatalf("unexpected entry count after eviction: got %d, want %d", got, want)
	}
	if got, want := store.SizeBytes() <= capBytes, true; got != want {
		t.Fatalf("size exceeds cap: got %d, cap %d", store.SizeBytes(), capBytes)
	}

	latest := store.Latest(2)
	if got, want := latest[0].Request.Path, "/3"; got != want {
		t.Fatalf("unexpected latest[0] path: got %q, want %q", got, want)
	}
	if got, want := latest[1].Request.Path, "/2"; got != want {
		t.Fatalf("unexpected latest[1] path: got %q, want %q", got, want)
	}
}

func TestInMemoryStore_TruncatesOversizedEntry(t *testing.T) {
	store := NewInMemoryStore(300)

	entry := store.Add(Entry{
		Request: Request{
			Method: http.MethodPost,
			URL:    "http://example.local/proxy",
			Path:   "/proxy",
			Body:   bytes.Repeat([]byte("r"), 2000),
		},
		Response: Response{
			Status: 200,
			Body:   bytes.Repeat([]byte("s"), 2000),
		},
	})

	if got, want := store.Count(), 1; got != want {
		t.Fatalf("unexpected entry count: got %d, want %d", got, want)
	}
	if got, want := store.SizeBytes() <= 300, true; got != want {
		t.Fatalf("size exceeds cap: got %d", store.SizeBytes())
	}
	if !entry.RequestTruncated && !entry.ResponseTruncated {
		t.Fatal("expected at least one truncation flag to be set")
	}
	if got, want := len(entry.Request.Body) < 2000 || len(entry.Response.Body) < 2000, true; got != want {
		t.Fatal("expected request or response body to be truncated")
	}
}
