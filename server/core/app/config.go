package app

// ServerConfig defines runtime HTTP server settings.
type ServerConfig struct {
	ListenAddress string
	RoutePath     string
}

// DefaultServerConfig returns baseline server settings.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		ListenAddress: ":8080",
		RoutePath:     "/",
	}
}
