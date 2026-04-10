package app

import (
	"net/http"

	"github.com/dector/lampa/server/core/processor"
)

// BuildProxyMux constructs HTTP mux with proxy route registration.
func BuildProxyMux(cfg ServerConfig, store processor.ReqProcessorStore) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(cfg.RoutePath, NewRequestHandlerWithStore(cfg, store))
	return mux
}

// BuildControlMux constructs HTTP mux with control route registration.
func BuildControlMux(store processor.ReqProcessorStore, opts ControlMuxOptions) *http.ServeMux {
	mux := http.NewServeMux()

	pingHandler := NewControlPingHandler(store)
	mux.HandleFunc(ComposeControlPath(opts.BasePath, RoutePing), pingHandler)
	mux.HandleFunc(ComposeControlPath(opts.BasePath, RouteProcCount), NewControlProcCountHandler(store))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
		return
	})

	return mux
}
