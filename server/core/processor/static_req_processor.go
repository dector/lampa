package processor

import (
	"net/http"

	corehttp "github.com/dector/lampa/server/core/http"
	"github.com/dector/lampa/server/core/utils/optional"
)

// StaticReqProcessor always returns preconfigured response.
type StaticReqProcessor struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func (p StaticReqProcessor) Process(_ corehttp.HttpRequest) optional.Optional[corehttp.HttpResponse] {
	return optional.Some(corehttp.HttpResponse{
		StatusCode: p.StatusCode,
		Headers:    p.Headers,
		Body:       p.Body,
	})
}
