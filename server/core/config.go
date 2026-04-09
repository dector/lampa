package main

import coreapp "github.com/dector/lampa/server/core/app"

// ServerConfig defines runtime HTTP server settings.
type ServerConfig = coreapp.ServerConfig

// DefaultServerConfig returns baseline server settings.
func DefaultServerConfig() ServerConfig {
	return coreapp.DefaultServerConfig()
}
