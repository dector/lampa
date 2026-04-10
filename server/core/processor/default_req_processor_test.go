package processor

import (
	"net/url"
	"testing"

	corehttp "github.com/dector/lampa/server/core/http"
	"github.com/dector/lampa/server/core/utils/optional"
)

func TestNewDefaultReqProcessor_IncludesRootQuickJSAndStaticFallback(t *testing.T) {
	p := NewDefaultReqProcessor()
	chain, ok := p.(DefaultReqProcessor)
	if !ok {
		t.Fatalf("unexpected processor type: %T", p)
	}

	if len(chain.processors) != 1 {
		t.Fatalf("unexpected processor count: got %d, want 1", len(chain.processors))
	}

	rootProcessor, ok := chain.processors["/"]
	if !ok {
		t.Fatal("expected processor for root endpoint '/'")
	}
	if _, ok := rootProcessor.(QuickJSReqProcessor); !ok {
		t.Fatalf("processor['/'] should be QuickJSReqProcessor, got %T", rootProcessor)
	}

	fallback, ok := chain.fallback.(StaticReqProcessor)
	if !ok {
		t.Fatalf("fallback should be StaticReqProcessor, got %T", chain.fallback)
	}
	if fallback.StatusCode != 404 {
		t.Fatalf("unexpected static status code: got %d, want 404", fallback.StatusCode)
	}
	if string(fallback.Body) != "Not Found" {
		t.Fatalf("unexpected static body: got %q, want %q", string(fallback.Body), "Not Found")
	}
}

func TestDefaultReqProcessor_FallbackBehavior(t *testing.T) {
	chain := DefaultReqProcessor{
		processors: EndpointProcessors{
			"/": noneReqProcessor{},
		},
		fallback: StaticReqProcessor{StatusCode: 404, Body: []byte("Not Found")},
	}

	response := chain.Process(corehttp.HttpRequest{Url: url.URL{Path: "/"}})
	if !response.IsPresent() {
		t.Fatal("expected Some response from fallback")
	}
	got := response.OrElse(corehttp.HttpResponse{})
	if got.StatusCode != 404 {
		t.Fatalf("unexpected status code: got %d, want 404", got.StatusCode)
	}
	if string(got.Body) != "Not Found" {
		t.Fatalf("unexpected body: got %q, want %q", string(got.Body), "Not Found")
	}
}

type noneReqProcessor struct{}

func (noneReqProcessor) Process(_ corehttp.HttpRequest) optional.Optional[corehttp.HttpResponse] {
	return optional.None[corehttp.HttpResponse]()
}
