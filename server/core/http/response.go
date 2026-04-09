package http

import (
	nethttp "net/http"
)

// HttpResponse contains common and important response metadata.
type HttpResponse struct {
	StatusCode int            `json:"statusCode"`
	Headers    nethttp.Header `json:"headers"`
	Body       []byte         `json:"body"`
}

func (r HttpResponse) ContentType() string {
	return r.Headers.Get("Content-Type")
}

func (r *HttpResponse) SetContentType(value string) {
	if r.Headers == nil {
		r.Headers = make(nethttp.Header)
	}
	r.Headers.Set("Content-Type", value)
}
