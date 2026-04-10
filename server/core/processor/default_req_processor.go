package processor

import (
	corehttp "github.com/dector/lampa/server/core/http"
	"github.com/dector/lampa/server/core/utils/optional"
)

type EndpointProcessors map[string]ReqProcessor

// DefaultReqProcessor processes request with a chain of processors.
type DefaultReqProcessor struct {
	processors EndpointProcessors
	fallback   ReqProcessor
}

// NewDefaultReqProcessor creates default request processor chain.
func NewDefaultReqProcessor() ReqProcessor {
	return DefaultReqProcessor{
		processors: EndpointProcessors{
			"/": QuickJSReqProcessor{Script: okJsProcessor},
		},
		fallback: StaticReqProcessor{StatusCode: 404, Body: []byte("Not Found")},
	}
}

func (p DefaultReqProcessor) Process(request corehttp.HttpRequest) optional.Optional[corehttp.HttpResponse] {
	if endpointProcessor, ok := p.processors[request.Url.Path]; ok {
		response := endpointProcessor.Process(request)
		if response.IsPresent() {
			return response
		}
	}

	if p.fallback != nil {
		response := p.fallback.Process(request)
		if response.IsPresent() {
			return response
		}
	}

	return optional.None[corehttp.HttpResponse]()
}
