package processor

import (
	corehttp "github.com/dector/lampa/server/core/http"
	"github.com/dector/lampa/server/core/utils/optional"
)

// ReqProcessor processes incoming requests and returns optional response.
type ReqProcessor interface {
	Process(request corehttp.HttpRequest) optional.Optional[corehttp.HttpResponse]
}
