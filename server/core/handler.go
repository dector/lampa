package main

import (
	"net/http"

	coreapp "github.com/dector/lampa/server/core/app"
	corehttp "github.com/dector/lampa/server/core/http"
)

// NewRequestHandler returns HTTP handler that reads request and returns processed response.
func NewRequestHandler(cfg ServerConfig) http.HandlerFunc {
	return coreapp.NewRequestHandler(cfg)
}

func ToNetHTTP(w http.ResponseWriter, response corehttp.HttpResponse) {
	coreapp.ToNetHTTP(w, response)
}
