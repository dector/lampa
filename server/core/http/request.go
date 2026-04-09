package http

import (
	"bytes"
	"io"
	nethttp "net/http"
	"net/url"
)

// HttpRequest contains the most common and important request metadata.
type HttpRequest struct {
	Method  string         `json:"method"`
	Url     url.URL        `json:"url"`
	Headers nethttp.Header `json:"headers"`
	Body    []byte         `json:"body"`
}

func (r HttpRequest) ContentType() string {
	return r.Headers.Get("Content-Type")
}

func (r *HttpRequest) SetContentType(value string) {
	if r.Headers == nil {
		r.Headers = make(nethttp.Header)
	}
	r.Headers.Set("Content-Type", value)
}

// NewHttpRequest builds a concise request description from http.Request.
func NewHttpRequest(r *nethttp.Request) HttpRequest {
	var requestURL url.URL
	if r.URL != nil {
		requestURL = *r.URL
	}

	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(body))

	return HttpRequest{
		Method:  r.Method,
		Url:     requestURL,
		Headers: r.Header.Clone(),
		Body:    body,
	}
}
