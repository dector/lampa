package logstore

import (
	"net/http"
	"sync"
	"time"
)

const (
	// DefaultMaxBytes caps retained logs in memory at 30MB.
	DefaultMaxBytes int64 = 30 * 1024 * 1024

	// UnlimitedMaxBytes disables log retention size limits.
	UnlimitedMaxBytes int64 = -1
)

// Entry describes one proxied request/response interaction.
type Entry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`

	Request  Request  `json:"request"`
	Response Response `json:"response"`

	DurationMs int64  `json:"durationMs,omitempty"`
	Error      string `json:"error,omitempty"`

	SizeBytes         int64 `json:"sizeBytes"`
	RequestTruncated  bool  `json:"requestTruncated,omitempty"`
	ResponseTruncated bool  `json:"responseTruncated,omitempty"`
}

// Request is request payload persisted in logs.
type Request struct {
	Method  string      `json:"method"`
	URL     string      `json:"url"`
	Path    string      `json:"path"`
	Headers http.Header `json:"headers"`
	Body    []byte      `json:"body"`
}

// Response is response payload persisted in logs.
type Response struct {
	Status  int         `json:"status"`
	Headers http.Header `json:"headers"`
	Body    []byte      `json:"body"`
}

// Store is in-memory request/response log store.
type Store interface {
	Add(entry Entry) Entry
	Latest(n int) []Entry
	Get(id int64) (Entry, bool)
	Clear()
	Count() int
	SizeBytes() int64
}

// InMemoryStore is thread-safe bounded log store.
type InMemoryStore struct {
	mu        sync.Mutex
	entries   []Entry
	totalSize int64
	maxBytes  int64
	nextID    int64
}

// NewInMemoryStore creates bounded in-memory log store.
func NewInMemoryStore(maxBytes int64) *InMemoryStore {
	if maxBytes == 0 || maxBytes < UnlimitedMaxBytes {
		maxBytes = DefaultMaxBytes
	}
	return &InMemoryStore{maxBytes: maxBytes}
}

// Add stores entry, assigning sequence ID and timestamp if they are empty.
func (s *InMemoryStore) Add(entry Entry) Entry {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	entry.ID = s.nextID
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	if s.maxBytes != UnlimitedMaxBytes {
		entry = fitEntryToCap(entry, s.maxBytes)
	}
	entry.SizeBytes = EstimateSizeBytes(entry)

	s.entries = append(s.entries, cloneEntry(entry))
	s.totalSize += entry.SizeBytes

	if s.maxBytes != UnlimitedMaxBytes {
		for s.totalSize > s.maxBytes && len(s.entries) > 0 {
			oldest := s.entries[0]
			s.entries = s.entries[1:]
			s.totalSize -= oldest.SizeBytes
		}
	}

	return cloneEntry(entry)
}

// Latest returns up to n latest entries ordered latest-first.
// If n <= 0, it returns all entries ordered latest-first.
func (s *InMemoryStore) Latest(n int) []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()

	if n <= 0 || n > len(s.entries) {
		n = len(s.entries)
	}

	out := make([]Entry, 0, n)
	for i := len(s.entries) - 1; i >= len(s.entries)-n; i-- {
		out = append(out, cloneEntry(s.entries[i]))
	}
	return out
}

// Get returns entry by ID.
func (s *InMemoryStore) Get(id int64) (Entry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, entry := range s.entries {
		if entry.ID == id {
			return cloneEntry(entry), true
		}
	}
	return Entry{}, false
}

// Clear removes all retained entries. Sequence IDs keep increasing.
func (s *InMemoryStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries = nil
	s.totalSize = 0
}

// Count returns current entries count.
func (s *InMemoryStore) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}

// SizeBytes returns approximate current in-memory size.
func (s *InMemoryStore) SizeBytes() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.totalSize
}

func fitEntryToCap(entry Entry, maxBytes int64) Entry {
	if maxBytes <= 0 {
		return entry
	}

	estimated := EstimateSizeBytes(entry)
	if estimated <= maxBytes {
		return entry
	}

	reqLen := len(entry.Request.Body)
	resLen := len(entry.Response.Body)
	baseSize := estimated - int64(reqLen+resLen)

	if baseSize >= maxBytes {
		if reqLen > 0 {
			entry.Request.Body = nil
			entry.RequestTruncated = true
		}
		if resLen > 0 {
			entry.Response.Body = nil
			entry.ResponseTruncated = true
		}
		return entry
	}

	allowedBodies := int(maxBytes - baseSize)
	if allowedBodies < 0 {
		allowedBodies = 0
	}

	totalBodies := reqLen + resLen
	if totalBodies == 0 {
		return entry
	}

	reqKeep := allowedBodies * reqLen / totalBodies
	resKeep := allowedBodies - reqKeep

	if reqKeep < reqLen {
		entry.Request.Body = append([]byte(nil), entry.Request.Body[:reqKeep]...)
		entry.RequestTruncated = true
	}
	if resKeep < resLen {
		entry.Response.Body = append([]byte(nil), entry.Response.Body[:resKeep]...)
		entry.ResponseTruncated = true
	}

	return entry
}

// EstimateSizeBytes returns deterministic approximation of entry memory usage.
func EstimateSizeBytes(entry Entry) int64 {
	const fixedOverhead = 128

	size := int64(fixedOverhead)
	size += int64(len(entry.Request.Method))
	size += int64(len(entry.Request.URL))
	size += int64(len(entry.Request.Path))
	size += estimateHeaderSize(entry.Request.Headers)
	size += int64(len(entry.Request.Body))

	size += int64(8) // status int
	size += estimateHeaderSize(entry.Response.Headers)
	size += int64(len(entry.Response.Body))

	size += int64(len(entry.Error))
	return size
}

func estimateHeaderSize(headers http.Header) int64 {
	if headers == nil {
		return 0
	}

	size := int64(0)
	for key, values := range headers {
		size += int64(len(key))
		for _, value := range values {
			size += int64(len(value))
		}
	}
	return size
}

func cloneEntry(entry Entry) Entry {
	entry.Request.Headers = entry.Request.Headers.Clone()
	entry.Response.Headers = entry.Response.Headers.Clone()
	entry.Request.Body = append([]byte(nil), entry.Request.Body...)
	entry.Response.Body = append([]byte(nil), entry.Response.Body...)
	return entry
}
