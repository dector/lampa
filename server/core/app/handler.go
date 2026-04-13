package app

import (
	"net/http"
	"strconv"
	"time"

	corehttp "github.com/dector/lampa/server/core/http"
	"github.com/dector/lampa/server/core/logstore"
	"github.com/dector/lampa/server/core/processor"
)

// NewRequestHandler returns HTTP handler that reads request and returns processed response.
func NewRequestHandler(cfg ServerConfig) http.HandlerFunc {
	return NewRequestHandlerWithStoreAndLogs(cfg, nil, nil)
}

// NewRequestHandlerWithStore returns HTTP handler that reads request and returns processed response.
func NewRequestHandlerWithStore(cfg ServerConfig, processorStore processor.ReqProcessorStore) http.HandlerFunc {
	return NewRequestHandlerWithStoreAndLogs(cfg, processorStore, nil)
}

// NewRequestHandlerWithStoreAndLogs returns HTTP handler with shared processor store and log store.
func NewRequestHandlerWithStoreAndLogs(_ ServerConfig, processorStore processor.ReqProcessorStore, logs logstore.Store) http.HandlerFunc {
	reqProcessor := processor.NewDefaultReqProcessorWithStore(processorStore)

	return func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		request := corehttp.NewHttpRequest(r)
		response := reqProcessor.Process(request).OrElse(corehttp.HttpResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    make(http.Header),
			Body:       nil,
		})
		ToNetHTTP(w, response)

		if logs != nil {
			responseStatus := response.StatusCode
			if responseStatus == 0 {
				responseStatus = http.StatusOK
			}

			logs.Add(logstore.Entry{
				Request: logstore.Request{
					Method:  request.Method,
					URL:     request.Url.String(),
					Path:    request.Url.Path,
					Headers: request.Headers.Clone(),
					Body:    append([]byte(nil), request.Body...),
				},
				Response: logstore.Response{
					Status:  responseStatus,
					Headers: response.Headers.Clone(),
					Body:    append([]byte(nil), response.Body...),
				},
				DurationMs: time.Since(startedAt).Milliseconds(),
			})
		}
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
