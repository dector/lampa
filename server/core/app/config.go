package app

import "strconv"

// WebUIConfig defines Web UI runtime settings.
type WebUIConfig struct {
	Enabled       bool
	Port          int
	BindHost      string
	ListenAddress string
}

// ServerConfig defines runtime HTTP server settings.
type ServerConfig struct {
	ProxyPort            int
	ControlPort          int
	ProxyBindHost        string
	ControlBindHost      string
	ListenAddress        string
	ControlListenAddress string
	RoutePath            string
	WebUI                WebUIConfig
	CaptureTraffic       bool
}

// DefaultServerConfig returns baseline server settings.
func DefaultServerConfig() ServerConfig {
	cfg := ServerConfig{
		ProxyPort:       8080,
		ControlPort:     8081,
		ProxyBindHost:   "localhost",
		ControlBindHost: "localhost",
		RoutePath:       "/",
		WebUI: WebUIConfig{
			Port:     8880,
			BindHost: "127.0.0.1",
		},
	}
	cfg.ListenAddress = ComposeListenAddress(cfg.ProxyBindHost, cfg.ProxyPort)
	cfg.ControlListenAddress = ComposeListenAddress(cfg.ControlBindHost, cfg.ControlPort)
	cfg.WebUI.ListenAddress = ComposeListenAddress(cfg.WebUI.BindHost, cfg.WebUI.Port)
	return cfg
}

// ComposeListenAddress builds host:port listen address.
func ComposeListenAddress(host string, port int) string {
	return host + ":" + strconv.Itoa(port)
}
