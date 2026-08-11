package app

import (
	"net/http"

	"github.com/dector/lampa/server/core/logstore"
	"github.com/dector/lampa/server/core/processor"
)

// BuildProxyMux constructs HTTP mux with proxy route registration.
func BuildProxyMux(cfg ServerConfig, store processor.ReqProcessorStore) *http.ServeMux {
	return BuildProxyMuxWithLogStore(cfg, store, nil)
}

// BuildProxyMuxWithLogStore constructs HTTP mux with proxy route registration and optional log storage.
func BuildProxyMuxWithLogStore(cfg ServerConfig, store processor.ReqProcessorStore, logs logstore.Store) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(cfg.RoutePath, NewRequestHandlerWithStoreAndLogs(cfg, store, logs))
	return mux
}

// BuildWebUIMux constructs HTTP mux with Web UI route registration.
func BuildWebUIMux(cfg ServerConfig, logs logstore.Store) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", NewWebUIHealthHandler())
	mux.HandleFunc(WebUIAssetsPath, NewWebUIAssetHandler())
	mux.HandleFunc("/ds/stats", NewWebUIStatsHandler(cfg, logs))
	mux.HandleFunc("/ds/requests", NewWebUIRequestsHandler(logs))
	mux.HandleFunc("/traffic", NewWebUITrafficPageHandler(logs))
	mux.HandleFunc("/", NewWebUIIndexHandler(cfg, logs))
	return mux
}

// BuildControlMux constructs HTTP mux with control route registration.
func BuildControlMux(store processor.ReqProcessorStore, opts ControlMuxOptions) *http.ServeMux {
	return BuildControlMuxWithLogStore(store, nil, opts)
}

// BuildControlMuxWithLogStore constructs HTTP mux with control route registration and optional log store endpoints.
func BuildControlMuxWithLogStore(store processor.ReqProcessorStore, logs logstore.Store, opts ControlMuxOptions) *http.ServeMux {
	mux := http.NewServeMux()

	pingHandler := NewControlPingHandler(store)
	mux.HandleFunc(ComposeControlPath(opts.BasePath, RoutePing), pingHandler)
	mux.HandleFunc(ComposeControlPath(opts.BasePath, RouteProcCount), NewControlProcCountHandler(store))
	mux.HandleFunc(ComposeControlPath(opts.BasePath, RouteProcSet), NewControlProcSetHandler(store))
	mux.HandleFunc(ComposeControlPath(opts.BasePath, RouteProcSetDefault), NewControlProcSetDefaultHandler(store))
	mux.HandleFunc(ComposeControlPath(opts.BasePath, RouteProcSequence), NewControlProcSequenceHandler(store))
	mux.HandleFunc(ComposeControlPath(opts.BasePath, RouteProcSequenceReset), NewControlProcSequenceResetHandler(store))
	mux.HandleFunc(ComposeControlPath(opts.BasePath, RouteProxyLogs), NewControlProxyLogsHandler(logs))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
		return
	})

	return mux
}
