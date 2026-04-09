package app

import (
	"net/http"
	"strconv"

	corehttp "github.com/dector/lampa/server/core/http"
	"github.com/dector/lampa/server/core/processor"
)

// NewRequestHandler returns HTTP handler that reads request and returns processed response.
func NewRequestHandler(_ ServerConfig) http.HandlerFunc {
	reqProcessor := processor.NewDefaultReqProcessor()

	return func(w http.ResponseWriter, r *http.Request) {
		request := corehttp.NewHttpRequest(r)
		response := reqProcessor.Process(request).OrElse(corehttp.HttpResponse{
			StatusCode: http.StatusNotFound,
			Headers:    make(http.Header),
			Body:       nil,
		})
		ToNetHTTP(w, response)
	}
}

func ToNetHTTP(w http.ResponseWriter, response corehttp.HttpResponse) {
	for name, values := range response.Headers {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	contentType := response.ContentType()
	if contentType == "" {
		contentType = "text/plain"
	}
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

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
