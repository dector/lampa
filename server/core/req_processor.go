package main

import (
	"net/http"

	corehttp "github.com/dector/lampa/server/core/http"
)

// ReqProcessor processes incoming requests and returns optional response.
type ReqProcessor interface {
	Process(request corehttp.HttpRequest) Optional[corehttp.HttpResponse]
}

// DefaultReqProcessor processes request with a chain of processors.
type DefaultReqProcessor struct {
	processors []ReqProcessor
}

// NewDefaultReqProcessor creates default request processor chain.
func NewDefaultReqProcessor() ReqProcessor {
	return DefaultReqProcessor{
		processors: []ReqProcessor{
			EmptyReqProcessor{},
		},
	}
}

func (p DefaultReqProcessor) Process(request corehttp.HttpRequest) Optional[corehttp.HttpResponse] {
	for _, processor := range p.processors {
		response := processor.Process(request)
		if response.IsPresent() {
			return response
		}
	}

	return None[corehttp.HttpResponse]()
}

// EmptyReqProcessor returns default successful response.
type EmptyReqProcessor struct{}

func (EmptyReqProcessor) Process(_ corehttp.HttpRequest) Optional[corehttp.HttpResponse] {
	return Some(corehttp.HttpResponse{
		StatusCode: http.StatusOK,
		Headers:    make(http.Header),
		Body:       nil,
	})
}
