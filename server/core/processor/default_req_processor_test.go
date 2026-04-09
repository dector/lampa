package processor

import (
	"net/url"
	"testing"

	corehttp "github.com/dector/lampa/server/core/http"
	"github.com/dector/lampa/server/core/utils/optional"
)

func TestNewDefaultReqProcessor_IncludesQuickJSAndFallback(t *testing.T) {
	p := NewDefaultReqProcessor()
	chain, ok := p.(DefaultReqProcessor)
	if !ok {
		t.Fatalf("unexpected processor type: %T", p)
	}

	if len(chain.processors) != 3 {
		t.Fatalf("unexpected processor count: got %d, want 3", len(chain.processors))
	}
	if _, ok := chain.processors[0].(QuickJSReqProcessor); !ok {
		t.Fatalf("processor[0] should be QuickJSReqProcessor, got %T", chain.processors[0])
	}
	if _, ok := chain.processors[1].(StarlarkReqProcessor); !ok {
		t.Fatalf("processor[1] should be StarlarkReqProcessor, got %T", chain.processors[1])
	}
	if _, ok := chain.processors[2].(EmptyReqProcessor); !ok {
		t.Fatalf("processor[2] should be EmptyReqProcessor, got %T", chain.processors[2])
	}
}

func TestDefaultReqProcessor_FallbackBehavior(t *testing.T) {
	chain := DefaultReqProcessor{
		processors: []ReqProcessor{
			noneReqProcessor{},
			EmptyReqProcessor{},
		},
	}

	response := chain.Process(corehttp.HttpRequest{Url: url.URL{Path: "/"}})
	if !response.IsPresent() {
		t.Fatal("expected Some response from fallback")
	}
	if got := response.OrElse(corehttp.HttpResponse{}).StatusCode; got != 200 {
		t.Fatalf("unexpected status code: got %d, want 200", got)
	}
}

type noneReqProcessor struct{}

func (noneReqProcessor) Process(_ corehttp.HttpRequest) optional.Optional[corehttp.HttpResponse] {
	return optional.None[corehttp.HttpResponse]()
}
