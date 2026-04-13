package processor

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	corehttp "github.com/dector/lampa/server/core/http"
)

func TestPassthroughReqProcessor_Process_Success(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/api/example"; got != want {
			t.Fatalf("unexpected upstream path: got %q, want %q", got, want)
		}
		if got, want := r.URL.RawQuery, "a=1"; got != want {
			t.Fatalf("unexpected upstream query: got %q, want %q", got, want)
		}
		if got, want := r.Header.Get("X-Test"), "1"; got != want {
			t.Fatalf("unexpected upstream header X-Test: got %q, want %q", got, want)
		}
		raw, _ := io.ReadAll(r.Body)
		if got, want := string(raw), "request-body"; got != want {
			t.Fatalf("unexpected upstream body: got %q, want %q", got, want)
		}

		w.Header().Set("X-Upstream", "yes")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("proxied"))
	}))
	defer upstream.Close()

	proc := PassthroughReqProcessor{Server: upstream.URL, Client: upstream.Client()}

	response := proc.Process(corehttp.HttpRequest{
		Method: "POST",
		Url:    url.URL{Path: "/api/example", RawQuery: "a=1"},
		Headers: http.Header{
			"X-Test":       []string{"1"},
			"Content-Type": []string{"text/plain"},
		},
		Body: []byte("request-body"),
	})

	if !response.IsPresent() {
		t.Fatal("expected Some response")
	}
	got := response.OrElse(corehttp.HttpResponse{})
	if got.StatusCode != http.StatusCreated {
		t.Fatalf("unexpected status code: got %d, want %d", got.StatusCode, http.StatusCreated)
	}
	if gotBody := string(got.Body); gotBody != "proxied" {
		t.Fatalf("unexpected body: got %q, want %q", gotBody, "proxied")
	}
	if gotHeader := got.Headers.Get("X-Upstream"); gotHeader != "yes" {
		t.Fatalf("unexpected header X-Upstream: got %q, want %q", gotHeader, "yes")
	}
}

func TestPassthroughReqProcessor_Process_QueryPropagation(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.RawQuery, "one=1&two=2"; got != want {
			t.Fatalf("unexpected upstream query: got %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	proc := PassthroughReqProcessor{Server: upstream.URL, Client: upstream.Client()}
	response := proc.Process(corehttp.HttpRequest{
		Method: "GET",
		Url:    url.URL{Path: "/", RawQuery: "one=1&two=2"},
	})

	if !response.IsPresent() {
		t.Fatal("expected Some response")
	}
	if got, want := response.OrElse(corehttp.HttpResponse{}).StatusCode, http.StatusOK; got != want {
		t.Fatalf("unexpected status code: got %d, want %d", got, want)
	}
}

func TestPassthroughReqProcessor_Process_UpstreamFailure(t *testing.T) {
	proc := PassthroughReqProcessor{
		Server: "http://127.0.0.1:1",
		Client: &http.Client{Timeout: 150 * time.Millisecond},
	}

	response := proc.Process(corehttp.HttpRequest{Method: "GET", Url: url.URL{Path: "/x"}})
	if !response.IsPresent() {
		t.Fatal("expected Some response")
	}

	got := response.OrElse(corehttp.HttpResponse{})
	if got.StatusCode != http.StatusBadGateway {
		t.Fatalf("unexpected status code: got %d, want %d", got.StatusCode, http.StatusBadGateway)
	}
	if gotBody := strings.TrimSpace(string(got.Body)); gotBody != "Bad Gateway" {
		t.Fatalf("unexpected response body: got %q, want %q", gotBody, "Bad Gateway")
	}
}
