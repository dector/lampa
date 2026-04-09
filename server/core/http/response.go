package http

import (
	nethttp "net/http"
	"strconv"
)

// HttpResponse contains common and important response metadata.
type HttpResponse struct {
	StatusCode    int               `json:"statusCode"`
	Headers       map[string]string `json:"headers"`
	ContentType   string            `json:"contentType"`
	ContentLength int64             `json:"contentLength"`
}

// ResponseCapture wraps http.ResponseWriter and records status/size for HttpResponse.
type ResponseCapture struct {
	nethttp.ResponseWriter
	StatusCode   int
	BytesWritten int64
}

func NewResponseCapture(w nethttp.ResponseWriter) *ResponseCapture {
	return &ResponseCapture{ResponseWriter: w, StatusCode: nethttp.StatusOK}
}

func (r *ResponseCapture) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseCapture) Write(p []byte) (int, error) {
	if r.StatusCode == 0 {
		r.StatusCode = nethttp.StatusOK
	}
	n, err := r.ResponseWriter.Write(p)
	r.BytesWritten += int64(n)
	return n, err
}

// NewHttpResponse builds a concise response description from ResponseCapture.
func NewHttpResponse(r *ResponseCapture) HttpResponse {
	contentLength := r.BytesWritten
	if cl := r.Header().Get("Content-Length"); cl != "" {
		if parsed, err := strconv.ParseInt(cl, 10, 64); err == nil {
			contentLength = parsed
		}
	}

	return HttpResponse{
		StatusCode:    r.StatusCode,
		Headers:       firstHeaderValues(r.Header()),
		ContentType:   r.Header().Get("Content-Type"),
		ContentLength: contentLength,
	}
}
