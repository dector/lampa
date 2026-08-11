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

func TestInMemoryStore_LatestZeroReturnsAllLatestFirst(t *testing.T) {
	store := NewInMemoryStore(1024 * 1024)

	store.Add(Entry{Request: Request{Method: http.MethodGet, Path: "/1"}})
	store.Add(Entry{Request: Request{Method: http.MethodGet, Path: "/2"}})
	store.Add(Entry{Request: Request{Method: http.MethodGet, Path: "/3"}})

	latest := store.Latest(0)
	if got, want := len(latest), 3; got != want {
		t.Fatalf("unexpected latest len: got %d, want %d", got, want)
	}
	if got, want := latest[0].Request.Path, "/3"; got != want {
		t.Fatalf("unexpected latest[0] path: got %q, want %q", got, want)
	}
	if got, want := latest[2].Request.Path, "/1"; got != want {
		t.Fatalf("unexpected latest[2] path: got %q, want %q", got, want)
	}
}

func TestInMemoryStore_GetReturnsEntryByID(t *testing.T) {
	store := NewInMemoryStore(1024 * 1024)

	first := store.Add(Entry{Request: Request{Method: http.MethodGet, Path: "/1"}})
	second := store.Add(Entry{Request: Request{Method: http.MethodPost, Path: "/2", Body: []byte("body")}})

	entry, ok := store.Get(second.ID)
	if !ok {
		t.Fatal("expected entry to be found")
	}
	if got, want := entry.ID, second.ID; got != want {
		t.Fatalf("unexpected entry ID: got %d, want %d", got, want)
	}
	if got, want := entry.Request.Path, "/2"; got != want {
		t.Fatalf("unexpected entry path: got %q, want %q", got, want)
	}

	entry.Request.Body[0] = 'X'
	entryAgain, ok := store.Get(second.ID)
	if !ok {
		t.Fatal("expected entry to be found again")
	}
	if got, want := string(entryAgain.Request.Body), "body"; got != want {
		t.Fatalf("expected cloned body: got %q, want %q", got, want)
	}

	if _, ok := store.Get(first.ID + second.ID + 100); ok {
		t.Fatal("expected missing entry")
	}
}

func TestInMemoryStore_ClearRemovesEntriesAndKeepsIDsIncreasing(t *testing.T) {
	store := NewInMemoryStore(1024 * 1024)

	first := store.Add(Entry{Request: Request{Path: "/1", Body: []byte("body")}})
	store.Clear()

	if got, want := store.Count(), 0; got != want {
		t.Fatalf("unexpected entry count: got %d, want %d", got, want)
	}
	if got, want := store.SizeBytes(), int64(0); got != want {
		t.Fatalf("unexpected size bytes: got %d, want %d", got, want)
	}
	if latest := store.Latest(0); len(latest) != 0 {
		t.Fatalf("expected empty latest after clear, got %d", len(latest))
	}

	second := store.Add(Entry{Request: Request{Path: "/2"}})
	if second.ID <= first.ID {
		t.Fatalf("expected IDs to keep increasing: first %d, second %d", first.ID, second.ID)
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

func TestInMemoryStore_UnlimitedDoesNotEvictOrTruncate(t *testing.T) {
	entry := Entry{
		Request:  Request{Path: "/large", Body: bytes.Repeat([]byte("r"), 2000)},
		Response: Response{Status: 200, Body: bytes.Repeat([]byte("s"), 2000)},
	}
	store := NewInMemoryStore(UnlimitedMaxBytes)

	store.Add(entry)
	store.Add(entry)
	store.Add(entry)

	if got, want := store.Count(), 3; got != want {
		t.Fatalf("unexpected entry count: got %d, want %d", got, want)
	}
	latest := store.Latest(1)
	if got, want := len(latest[0].Request.Body), 2000; got != want {
		t.Fatalf("unexpected request body len: got %d, want %d", got, want)
	}
	if got, want := len(latest[0].Response.Body), 2000; got != want {
		t.Fatalf("unexpected response body len: got %d, want %d", got, want)
	}
	if latest[0].RequestTruncated || latest[0].ResponseTruncated {
		t.Fatal("did not expect truncation flags")
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
