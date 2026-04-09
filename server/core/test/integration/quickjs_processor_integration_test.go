package integration_test

// Integration coverage diagram (request -> response pipeline under test):
//
//   net/http client
//        |
//        v
//   httptest.Server
//        |
//        v
//   http.HandlerFunc (newQuickJSTestServer)
//        |
//        v
//   core/http.NewHttpRequest
//        |
//        v
//   processor.QuickJSReqProcessor.Process
//        |
//        +--> quickjs.NewVM
//        +--> quickjsexec.LoadHandler (Response DSL prelude + user script)
//        +--> quickjsexec.ToRequestValue (Go request -> JS req)
//        +--> vm.Call("handle", req)
//        +--> quickjsexec.FromResponseValue (JS response -> Go response)
//        |
//        v
//   fallback (when JS returns null/undefined): Go 404 response
//        |
//        v
//   writeResponse (headers/body/status)
//        |
//        v
//   net/http client assertions
//
// These tests validate the full integration across transport, mapping, JS runtime,
// Response DSL, and fallback behavior.

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	corehttp "github.com/dector/lampa/server/core/http"
	"github.com/dector/lampa/server/core/processor"
)

const routingScript = `
function handle(req) {
  if (req.url.indexOf("/ping") === 0) {
    return Response.text("pong").status(302);
  }

  if (req.url.indexOf("/echo") === 0) {
    return Response.json({ req: req });
  }

  if (req.url.indexOf("/404") === 0) {
    return Response.text("not found").status(404);
  }

  if (req.url.indexOf("/skip") === 0) {
    return null;
  }

  return null;
}
`

func TestQuickJSProcessorIntegration_PingReturnsPongWithFound(t *testing.T) {
	ts := newQuickJSTestServer(routingScript)
	defer ts.Close()

	client := &http.Client{
		Transport: &http.Transport{DisableCompression: true},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(ts.URL + "/ping")
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("unexpected status: got %d, want %d", resp.StatusCode, http.StatusFound)
	}
	if got := resp.Header.Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Fatalf("unexpected content type: got %q, want %q", got, "text/plain; charset=utf-8")
	}
	if got := readBody(t, resp); got != "pong" {
		t.Fatalf("unexpected body: got %q, want %q", got, "pong")
	}
}

func TestQuickJSProcessorIntegration_EchoesGETRequestAsJSON(t *testing.T) {
	ts := newQuickJSTestServer(routingScript)
	defer ts.Close()

	client := &http.Client{Transport: &http.Transport{DisableCompression: true}}

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/echo?foo=bar", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("User-Agent", "integration-test")
	req.Header.Set("X-Test", "alpha")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got := resp.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content type: got %q, want %q", got, "application/json")
	}

	got := parseEchoedResponse(t, readBody(t, resp))
	want := echoedResponse{
		Req: echoedRequest{
			Method: "GET",
			URL:    "/echo?foo=bar",
			Headers: map[string][]string{
				"User-Agent": {"integration-test"},
				"X-Test":     {"alpha"},
			},
			Body: "",
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("response mismatch (-want +got):\n%s", diff)
	}
}

func TestQuickJSProcessorIntegration_EchoesPOSTBodyAndMultiHeaders(t *testing.T) {
	ts := newQuickJSTestServer(routingScript)
	defer ts.Close()

	client := &http.Client{Transport: &http.Transport{DisableCompression: true}}

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/echo", strings.NewReader(`{"name":"pi"}`))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("User-Agent", "integration-test")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("X-Tag", "a")
	req.Header.Add("X-Tag", "b")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	got := parseEchoedResponse(t, readBody(t, resp))
	want := echoedResponse{
		Req: echoedRequest{
			Method: "POST",
			URL:    "/echo",
			Headers: map[string][]string{
				"Content-Length": {"13"},
				"Content-Type":   {"application/json"},
				"User-Agent":     {"integration-test"},
				"X-Tag":          {"a", "b"},
			},
			Body: `{"name":"pi"}`,
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("response mismatch (-want +got):\n%s", diff)
	}
}

func TestQuickJSProcessorIntegration_ReturnsNotFoundFromScript(t *testing.T) {
	ts := newQuickJSTestServer(routingScript)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/404")
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected status: got %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
	if got := resp.Header.Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Fatalf("unexpected content type: got %q, want %q", got, "text/plain; charset=utf-8")
	}
	if got := readBody(t, resp); got != "not found" {
		t.Fatalf("unexpected body: got %q, want %q", got, "not found")
	}
}

func TestQuickJSProcessorIntegration_FallsBackWhenScriptReturnsNull(t *testing.T) {
	ts := newQuickJSTestServer(routingScript)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/skip")
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected status: got %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
	if got := resp.Header.Get("Content-Type"); got != "text/plain" {
		t.Fatalf("unexpected content type: got %q, want %q", got, "text/plain")
	}
}

func newQuickJSTestServer(script string) *httptest.Server {
	reqProcessor := processor.QuickJSReqProcessor{Script: script}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := corehttp.NewHttpRequest(r)
		response := reqProcessor.Process(request).OrElse(corehttp.HttpResponse{
			StatusCode: http.StatusNotFound,
			Headers:    make(http.Header),
			Body:       nil,
		})
		writeResponse(w, response)
	})

	return httptest.NewServer(h)
}

func writeResponse(w http.ResponseWriter, response corehttp.HttpResponse) {
	for name, values := range response.Headers {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	contentType := response.ContentType()
	if contentType == "" {
		contentType = "text/plain"
	}
	w.Header().Set("Content-Type", contentType)

	contentLength := int64(len(response.Body))
	w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))

	statusCode := response.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	w.WriteHeader(statusCode)
	if len(response.Body) > 0 {
		_, _ = w.Write(response.Body)
	}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return string(body)
}

type echoedResponse struct {
	Req echoedRequest `json:"req"`
}

type echoedRequest struct {
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

func parseEchoedResponse(t *testing.T, body string) echoedResponse {
	t.Helper()

	var decoded echoedResponse
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		t.Fatalf("decode JSON body: %v\nbody: %s", err, body)
	}

	return decoded
}
