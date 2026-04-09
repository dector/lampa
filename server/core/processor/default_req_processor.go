package processor

import (
	corehttp "github.com/dector/lampa/server/core/http"
	"github.com/dector/lampa/server/core/utils/optional"
)

// DefaultReqProcessor processes request with a chain of processors.
type DefaultReqProcessor struct {
	processors []ReqProcessor
}

// NewDefaultReqProcessor creates default request processor chain.
func NewDefaultReqProcessor() ReqProcessor {
	return DefaultReqProcessor{
		processors: []ReqProcessor{
			QuickJSReqProcessor{Script: defaultQuickJSProgram},
			StarlarkReqProcessor{Script: defaultStarlarkProgram},
			EmptyReqProcessor{},
		},
	}
}

func (p DefaultReqProcessor) Process(request corehttp.HttpRequest) optional.Optional[corehttp.HttpResponse] {
	for _, processor := range p.processors {
		response := processor.Process(request)
		if response.IsPresent() {
			return response
		}
	}

	return optional.None[corehttp.HttpResponse]()
}
