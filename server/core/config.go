package main

// ServerConfig defines runtime HTTP server settings.
type ServerConfig struct {
	ListenAddress       string
	RoutePath           string
	ResponseContentType string
}

// DefaultServerConfig returns baseline server settings.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		ListenAddress:       ":8080",
		RoutePath:           "/",
		ResponseContentType: "application/json; charset=utf-8",
	}
}
