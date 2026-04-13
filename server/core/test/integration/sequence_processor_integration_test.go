package integration_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dector/lampa/server/core/app"
	"github.com/dector/lampa/server/core/processor"
)

func TestSequenceProcessorIntegration_RotatesInOrderAndWrapsAround(t *testing.T) {
	ts := newSequenceTestServer([]processor.ReqProcessor{
		processor.StaticReqProcessor{StatusCode: 200, Body: []byte("first")},
		processor.StaticReqProcessor{StatusCode: 200, Body: []byte("second")},
		processor.StaticReqProcessor{StatusCode: 200, Body: []byte("third")},
	})
	defer ts.Close()

	client := &http.Client{Transport: &http.Transport{DisableCompression: true}}

	wantBodies := []string{"first", "second", "third", "first", "second"}
	for i, want := range wantBodies {
		resp, err := client.Get(ts.URL + "/flaky")
		if err != nil {
			t.Fatalf("request %d failed: %v", i+1, err)
		}

		gotBody := readSequenceBody(t, resp)
		_ = resp.Body.Close()

		if gotBody != want {
			t.Fatalf("request %d: unexpected body: got %q, want %q", i+1, gotBody, want)
		}
	}
}

func TestSequenceProcessorIntegration_HonorsPerStepStatusHeadersContentTypeAndBody(t *testing.T) {
	ts := newSequenceTestServer([]processor.ReqProcessor{
		processor.StaticReqProcessor{
			StatusCode: 503,
			Headers: http.Header{
				"Content-Type": {"text/plain"},
				"X-Step":       {"1"},
			},
			Body: []byte("temporary error"),
		},
		processor.StaticReqProcessor{
			StatusCode: 201,
			Headers: http.Header{
				"Content-Type":  {"application/json"},
				"X-Step":        {"2"},
				"Cache-Control": {"no-store"},
			},
			Body: []byte(`{"ok":true}`),
		},
	})
	defer ts.Close()

	client := &http.Client{Transport: &http.Transport{DisableCompression: true}}

	tests := []struct {
		name            string
		wantStatus      int
		wantContentType string
		wantStep        string
		wantBody        string
		wantCache       string
	}{
		{
			name:            "step 1",
			wantStatus:      503,
			wantContentType: "text/plain",
			wantStep:        "1",
			wantBody:        "temporary error",
			wantCache:       "",
		},
		{
			name:            "step 2",
			wantStatus:      201,
			wantContentType: "application/json",
			wantStep:        "2",
			wantBody:        `{"ok":true}`,
			wantCache:       "no-store",
		},
		{
			name:            "step 1 again after wrap-around",
			wantStatus:      503,
			wantContentType: "text/plain",
			wantStep:        "1",
			wantBody:        "temporary error",
			wantCache:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Get(ts.URL + "/flaky")
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if got, want := resp.StatusCode, tt.wantStatus; got != want {
				t.Fatalf("unexpected status: got %d, want %d", got, want)
			}
			if got, want := resp.Header.Get("Content-Type"), tt.wantContentType; got != want {
				t.Fatalf("unexpected content type: got %q, want %q", got, want)
			}
			if got, want := resp.Header.Get("X-Step"), tt.wantStep; got != want {
				t.Fatalf("unexpected X-Step header: got %q, want %q", got, want)
			}
			if got, want := resp.Header.Get("Cache-Control"), tt.wantCache; got != want {
				t.Fatalf("unexpected Cache-Control header: got %q, want %q", got, want)
			}
			if got, want := readSequenceBody(t, resp), tt.wantBody; got != want {
				t.Fatalf("unexpected body: got %q, want %q", got, want)
			}
		})
	}
}

func newSequenceTestServer(steps []processor.ReqProcessor) *httptest.Server {
	store := processor.NewInMemoryReqProcessorStore(processor.EndpointProcessors{
		"/flaky": processor.NewSequenceReqProcessor(steps),
	}, processor.StaticReqProcessor{StatusCode: http.StatusNotFound, Body: []byte("Not Found")})

	h := app.NewRequestHandlerWithStore(app.DefaultServerConfig(), store)
	return httptest.NewServer(h)
}

func readSequenceBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return string(body)
}
