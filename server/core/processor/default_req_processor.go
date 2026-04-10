package processor

import (
	corehttp "github.com/dector/lampa/server/core/http"
	"github.com/dector/lampa/server/core/utils/optional"
)

type EndpointProcessors map[string]ReqProcessor

// DefaultReqProcessor processes request with a chain of processors.
type DefaultReqProcessor struct {
	store ReqProcessorStore
}

// NewDefaultReqProcessor creates default request processor chain.
func NewDefaultReqProcessor() ReqProcessor {
	return NewDefaultReqProcessorWithStore(nil)
}

// NewDefaultReqProcessorWithStore creates default request processor chain using provided store.
func NewDefaultReqProcessorWithStore(store ReqProcessorStore) ReqProcessor {
	if store == nil {
		store = NewDefaultReqProcessorStore()
	}

	return DefaultReqProcessor{store: store}
}

func (p DefaultReqProcessor) Process(request corehttp.HttpRequest) optional.Optional[corehttp.HttpResponse] {
	if p.store != nil {
		if endpointProcessor, ok := p.store.EndpointProcessor(request.Url.Path); ok {
			response := endpointProcessor.Process(request)
			if response.IsPresent() {
				return response
			}
		}

		fallback := p.store.FallbackProcessor()
		if fallback != nil {
			response := fallback.Process(request)
			if response.IsPresent() {
				return response
			}
		}
	}

	return optional.None[corehttp.HttpResponse]()
}
