package processor

import (
	"net/http"
	"testing"

	corehttp "github.com/dector/lampa/server/core/http"
)

func TestStaticReqProcessor_Process(t *testing.T) {
	p := StaticReqProcessor{
		StatusCode: http.StatusCreated,
		Headers: http.Header{
			"Content-Type": []string{"text/plain"},
		},
		Body: []byte("created"),
	}

	response := p.Process(corehttp.HttpRequest{})
	if !response.IsPresent() {
		t.Fatal("expected Some response")
	}

	got := response.OrElse(corehttp.HttpResponse{})
	if got.StatusCode != http.StatusCreated {
		t.Fatalf("unexpected status: got %d, want %d", got.StatusCode, http.StatusCreated)
	}
	if got.Headers.Get("Content-Type") != "text/plain" {
		t.Fatalf("unexpected content-type: got %q", got.Headers.Get("Content-Type"))
	}
	if string(got.Body) != "created" {
		t.Fatalf("unexpected body: got %q, want %q", string(got.Body), "created")
	}
}
