package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	corehttp "github.com/dector/lampa/server/core/http"
)

func TestToNetHTTP(t *testing.T) {
	tests := []struct {
		name               string
		response           corehttp.HttpResponse
		wantStatusCode     int
		wantContentType    string
		wantContentLength  string
		wantBody           string
		wantLengthHeaderOk bool
	}{
		{
			name: "defaults content type and status code",
			response: corehttp.HttpResponse{
				Headers: nil,
				Body:    nil,
			},
			wantStatusCode:     http.StatusOK,
			wantContentType:    "text/plain",
			wantContentLength:  "0",
			wantLengthHeaderOk: true,
			wantBody:           "",
		},
		{
			name: "respects explicit content type and status code",
			response: corehttp.HttpResponse{
				StatusCode: http.StatusCreated,
				Headers:    http.Header{"Content-Type": {"application/json"}},
				Body:       nil,
			},
			wantStatusCode:     http.StatusCreated,
			wantContentType:    "application/json",
			wantContentLength:  "0",
			wantLengthHeaderOk: true,
			wantBody:           "",
		},
		{
			name: "sets content length and writes body for non-empty body",
			response: corehttp.HttpResponse{
				StatusCode: http.StatusOK,
				Headers:    make(http.Header),
				Body:       []byte("hello"),
			},
			wantStatusCode:     http.StatusOK,
			wantContentType:    "text/plain",
			wantContentLength:  "5",
			wantLengthHeaderOk: true,
			wantBody:           "hello",
		},
		{
			name: "sets content length for empty body",
			response: corehttp.HttpResponse{
				StatusCode: http.StatusOK,
				Headers:    make(http.Header),
				Body:       []byte{},
			},
			wantStatusCode:     http.StatusOK,
			wantContentType:    "text/plain",
			wantContentLength:  "0",
			wantLengthHeaderOk: true,
			wantBody:           "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			ToNetHTTP(rr, tt.response)

			if rr.Code != tt.wantStatusCode {
				t.Fatalf("unexpected status code: got %d, want %d", rr.Code, tt.wantStatusCode)
			}

			if got := rr.Header().Get("Content-Type"); got != tt.wantContentType {
				t.Fatalf("unexpected content type: got %q, want %q", got, tt.wantContentType)
			}

			gotContentLength, hasContentLength := rr.Header()["Content-Length"]
			if hasContentLength != tt.wantLengthHeaderOk {
				t.Fatalf("content-length header presence mismatch: got %v, want %v", hasContentLength, tt.wantLengthHeaderOk)
			}
			if tt.wantLengthHeaderOk {
				if len(gotContentLength) != 1 || gotContentLength[0] != tt.wantContentLength {
					t.Fatalf("unexpected content-length: got %v, want [%q]", gotContentLength, tt.wantContentLength)
				}
			}

			if got := rr.Body.String(); got != tt.wantBody {
				t.Fatalf("unexpected body: got %q, want %q", got, tt.wantBody)
			}
		})
	}
}

func TestToNetHTTP_CopiesMultiValueHeaders(t *testing.T) {
	rr := httptest.NewRecorder()

	response := corehttp.HttpResponse{
		StatusCode: http.StatusOK,
		Headers: http.Header{
			"Set-Cookie": {"a=1", "b=2"},
		},
	}

	ToNetHTTP(rr, response)

	cookies := rr.Header().Values("Set-Cookie")
	if len(cookies) != 2 {
		t.Fatalf("unexpected number of set-cookie headers: got %d, want 2", len(cookies))
	}
	if cookies[0] != "a=1" || cookies[1] != "b=2" {
		t.Fatalf("unexpected set-cookie values: got %v, want [a=1 b=2]", cookies)
	}
}
