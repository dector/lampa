package processor

import (
	"net/http"

	corehttp "github.com/dector/lampa/server/core/http"
	"github.com/dector/lampa/server/core/utils/optional"
)

// EmptyReqProcessor returns default successful response.
type EmptyReqProcessor struct{}

func (EmptyReqProcessor) Process(_ corehttp.HttpRequest) optional.Optional[corehttp.HttpResponse] {
	return optional.Some(corehttp.HttpResponse{
		StatusCode: http.StatusOK,
		Headers:    make(http.Header),
		Body:       nil,
	})
}
