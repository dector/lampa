package processor

import (
	nethttp "net/http"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"

	corehttp "github.com/dector/lampa/server/core/http"
)

func TestQuickJSReqProcessor_ReturnsNoneOnScriptLoadError(t *testing.T) {
	p := QuickJSReqProcessor{
		Script: `function handle(`,
	}

	response := p.Process(testRequest())
	if response.IsPresent() {
		t.Fatal("expected None response")
	}
}

func TestQuickJSReqProcessor_ReturnsNoneWhenHandlerMissing(t *testing.T) {
	p := QuickJSReqProcessor{
		Script: `function ping(req) { return {}; }`,
	}

	response := p.Process(testRequest())
	if response.IsPresent() {
		t.Fatal("expected None response")
	}
}

func TestQuickJSReqProcessor_ReturnsNoneOnNullOrUndefined(t *testing.T) {
	t.Run("null", func(t *testing.T) {
		p := QuickJSReqProcessor{
			Script: `function handle(req) { return null; }`,
		}

		response := p.Process(testRequest())
		if response.IsPresent() {
			t.Fatal("expected None response")
		}
	})

	t.Run("undefined", func(t *testing.T) {
		p := QuickJSReqProcessor{
			Script: `function handle(req) { return undefined; }`,
		}

		response := p.Process(testRequest())
		if response.IsPresent() {
			t.Fatal("expected None response")
		}
	})
}

func TestQuickJSReqProcessor_MapsValidResponse(t *testing.T) {
	p := QuickJSReqProcessor{
		Script: `
			function handle(req) {
				return {
					statusCode: 201,
					headers: {
						"Content-Type": "application/json",
						"Set-Cookie": ["a=1", "b=2"],
					},
					body: "{\"ok\":true}",
				};
			}
		`,
	}

	response := p.Process(testRequest())
	if !response.IsPresent() {
		t.Fatal("expected Some response")
	}
	got := response.OrElse(corehttp.HttpResponse{})

	if diff := cmp.Diff(201, got.StatusCode); diff != "" {
		t.Fatalf("StatusCode mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(nethttp.Header{
		"Content-Type": {"application/json"},
		"Set-Cookie":   {"a=1", "b=2"},
	}, got.Headers); diff != "" {
		t.Fatalf("Headers mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]byte(`{"ok":true}`), got.Body); diff != "" {
		t.Fatalf("Body mismatch (-want +got):\n%s", diff)
	}
}

func TestQuickJSReqProcessor_ReturnsNoneOnInvalidHeaderOrBodyShape(t *testing.T) {
	t.Run("invalid headers", func(t *testing.T) {
		p := QuickJSReqProcessor{
			Script: `function handle(req) { return { headers: 42 }; }`,
		}

		response := p.Process(testRequest())
		if response.IsPresent() {
			t.Fatal("expected None response")
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		p := QuickJSReqProcessor{
			Script: `function handle(req) { return { body: true }; }`,
		}

		response := p.Process(testRequest())
		if response.IsPresent() {
			t.Fatal("expected None response")
		}
	})
}

func TestQuickJSReqProcessor_MapsResponseDSLBuilder(t *testing.T) {
	p := QuickJSReqProcessor{
		Script: `
			function handle(req) {
				return Response
					.json({ ok: true })
					.status(202)
					.header("X-Test", "ok");
			}
		`,
	}

	response := p.Process(testRequest())
	if !response.IsPresent() {
		t.Fatal("expected Some response")
	}
	got := response.OrElse(corehttp.HttpResponse{})

	if diff := cmp.Diff(202, got.StatusCode); diff != "" {
		t.Fatalf("StatusCode mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(nethttp.Header{
		"Content-Type": {"application/json"},
		"X-Test":       {"ok"},
	}, got.Headers); diff != "" {
		t.Fatalf("Headers mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]byte(`{"ok":true}`), got.Body); diff != "" {
		t.Fatalf("Body mismatch (-want +got):\n%s", diff)
	}
}

func testRequest() corehttp.HttpRequest {
	return corehttp.HttpRequest{
		Method:  "GET",
		Url:     url.URL{Scheme: "https", Host: "example.com", Path: "/"},
		Headers: nethttp.Header{"X-Test": {"ok"}},
		Body:    []byte("hello"),
	}
}
