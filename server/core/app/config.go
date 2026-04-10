package app

import "strconv"

// ServerConfig defines runtime HTTP server settings.
type ServerConfig struct {
	ProxyPort            int
	ControlPort          int
	ProxyBindHost        string
	ControlBindHost      string
	ListenAddress        string
	ControlListenAddress string
	RoutePath            string
}

// DefaultServerConfig returns baseline server settings.
func DefaultServerConfig() ServerConfig {
	cfg := ServerConfig{
		ProxyPort:       8080,
		ControlPort:     8081,
		ProxyBindHost:   "localhost",
		ControlBindHost: "localhost",
		RoutePath:       "/",
	}
	cfg.ListenAddress = ComposeListenAddress(cfg.ProxyBindHost, cfg.ProxyPort)
	cfg.ControlListenAddress = ComposeListenAddress(cfg.ControlBindHost, cfg.ControlPort)
	return cfg
}

// ComposeListenAddress builds host:port listen address.
func ComposeListenAddress(host string, port int) string {
	return host + ":" + strconv.Itoa(port)
}
