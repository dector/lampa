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
)

// ControlMuxOptions defines control routes mounting behavior.
type ControlMuxOptions struct {
	// BasePath is control API prefix, e.g. "/" or "/control".
	BasePath string
	// EnableRootPingAlias exposes "/" as an alias for ping endpoint.
	EnableRootPingAlias bool
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
