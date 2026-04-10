package main

import (
	"net/http"

	coreapp "github.com/dector/lampa/server/core/app"
	corehttp "github.com/dector/lampa/server/core/http"
	"github.com/dector/lampa/server/core/processor"
)

// NewRequestHandler returns HTTP handler that reads request and returns processed response.
func NewRequestHandler(cfg ServerConfig) http.HandlerFunc {
	return NewRequestHandlerWithStore(cfg, nil)
}

// NewRequestHandlerWithStore returns HTTP handler with shared processor store.
func NewRequestHandlerWithStore(cfg ServerConfig, store processor.ReqProcessorStore) http.HandlerFunc {
	return coreapp.NewRequestHandlerWithStore(cfg, store)
}

func ToNetHTTP(w http.ResponseWriter, response corehttp.HttpResponse) {
	coreapp.ToNetHTTP(w, response)
}
