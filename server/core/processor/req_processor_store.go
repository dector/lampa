package processor

import "sync"

// ReqProcessorStore stores request processors resolved by endpoint path.
type ReqProcessorStore interface {
	EndpointProcessor(path string) (ReqProcessor, bool)
	FallbackProcessor() ReqProcessor
	SetEndpointProcessor(path string, processor ReqProcessor)
	SetFallbackProcessor(processor ReqProcessor)
	ResponsesCount() int
}

// InMemoryReqProcessorStore stores request processors in memory.
type InMemoryReqProcessorStore struct {
	mu         sync.RWMutex
	processors EndpointProcessors
	fallback   ReqProcessor
}

// NewInMemoryReqProcessorStore creates in-memory processor store.
func NewInMemoryReqProcessorStore(processors EndpointProcessors, fallback ReqProcessor) *InMemoryReqProcessorStore {
	if processors == nil {
		processors = EndpointProcessors{}
	}

	return &InMemoryReqProcessorStore{
		processors: cloneEndpointProcessors(processors),
		fallback:   fallback,
	}
}

func (s *InMemoryReqProcessorStore) EndpointProcessor(path string) (ReqProcessor, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	processor, ok := s.processors[path]
	return processor, ok
}

func (s *InMemoryReqProcessorStore) FallbackProcessor() ReqProcessor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.fallback
}

func (s *InMemoryReqProcessorStore) SetEndpointProcessor(path string, processor ReqProcessor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.processors[path] = processor
}

func (s *InMemoryReqProcessorStore) SetFallbackProcessor(processor ReqProcessor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fallback = processor
}

func (s *InMemoryReqProcessorStore) ResponsesCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := len(s.processors)
	if s.fallback != nil {
		count++
	}
	return count
}

// NewDefaultReqProcessorStore creates default in-memory processor store.
func NewDefaultReqProcessorStore() ReqProcessorStore {
	return NewInMemoryReqProcessorStore(
		EndpointProcessors{
			"/": QuickJSReqProcessor{Script: okJsProcessor},
		},
		StaticReqProcessor{StatusCode: 404, Body: []byte("Not Found")},
	)
}

func cloneEndpointProcessors(src EndpointProcessors) EndpointProcessors {
	clone := make(EndpointProcessors, len(src))
	for endpoint, processor := range src {
		clone[endpoint] = processor
	}
	return clone
}
