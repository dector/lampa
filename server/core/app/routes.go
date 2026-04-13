package app

import (
	"path"
	"strings"
)

const (
	// RoutePing is control endpoint that returns basic server health payload.
	RoutePing = "/ping"
	// RouteProcCount is control endpoint that returns configured processor count.
	RouteProcCount = "/api/v0/proc_count"
	// RouteProcSet is control endpoint that sets endpoint processor runtime config.
	RouteProcSet = "/api/v0/proc/set"
	// RouteProcSetDefault is control endpoint that sets fallback/default processor runtime config.
	RouteProcSetDefault = "/api/v0/proc/default/set"
	// RouteProxyLogs is control endpoint that returns latest proxy request/response logs.
	RouteProxyLogs = "/api/v0/proxy/logs"
)

// ControlMuxOptions defines control routes mounting behavior.
type ControlMuxOptions struct {
	// BasePath is control API prefix, e.g. "/" or "/control".
	BasePath string
}

// ComposeControlPath joins control base path and route using consistent rules.
func ComposeControlPath(basePath string, route string) string {
	basePath = strings.TrimSpace(basePath)
	if basePath == "" {
		basePath = "/"
	}

	route = strings.TrimSpace(route)
	if route == "" {
		return path.Clean(basePath)
	}

	return path.Join(basePath, route)
}
