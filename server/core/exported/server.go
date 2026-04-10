package exported

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/dector/lampa/server/core/app"
	"github.com/dector/lampa/server/core/processor"
)

var (
	errServerAlreadyRunning = errors.New("server is already running")
	errConfigLocked         = errors.New("cannot change config while server is running")
)

// ServerConfig defines proxy/control server runtime settings.
type ServerConfig struct {
	ProxyListenAddress   string
	ProxyRoutePath       string
	ControlListenAddress string
	ControlRoutePath     string
}

// DefaultConfig returns baseline settings for both proxy and control servers.
func DefaultConfig() *ServerConfig {
	cfg := defaultConfig()
	return &cfg
}

func defaultConfig() ServerConfig {
	defaults := app.DefaultServerConfig()
	return ServerConfig{
		ProxyListenAddress:   defaults.ListenAddress,
		ProxyRoutePath:       defaults.RoutePath,
		ControlListenAddress: defaults.ControlListenAddress,
		ControlRoutePath:     "/",
	}
}

func mergeConfig(base ServerConfig, override *ServerConfig) ServerConfig {
	if override == nil {
		return base
	}

	cfg := base
	if override.ProxyListenAddress != "" {
		cfg.ProxyListenAddress = override.ProxyListenAddress
	}
	if override.ProxyRoutePath != "" {
		cfg.ProxyRoutePath = override.ProxyRoutePath
	}
	if override.ControlListenAddress != "" {
		cfg.ControlListenAddress = override.ControlListenAddress
	}
	if override.ControlRoutePath != "" {
		cfg.ControlRoutePath = override.ControlRoutePath
	}
	return cfg
}

// Server exposes embeddable HTTP server lifecycle for gomobile bindings.
type Server struct {
	mu                sync.Mutex
	config            ServerConfig
	proxyHTTPServer   *http.Server
	controlHTTPServer *http.Server
}

// NewServer creates server with default config.
func NewServer() *Server {
	return &Server{
		config: defaultConfig(),
	}
}

// NewServerWithConfig creates server with provided proxy/control config.
func NewServerWithConfig(config *ServerConfig) *Server {
	s := NewServer()
	_ = s.SetConfig(config)
	return s
}

// Config returns a copy of active runtime config.
func (s *Server) Config() *ServerConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	cfg := s.config
	return &cfg
}

// SetConfig updates proxy/control runtime config.
// Returns empty string on success or error message otherwise.
func (s *Server) SetConfig(config *ServerConfig) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunningLocked() {
		return errConfigLocked.Error()
	}

	s.config = mergeConfig(defaultConfig(), config)
	return ""
}

// ProxyListenAddress returns proxy listen address.
func (s *Server) ProxyListenAddress() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.config.ProxyListenAddress
}

// SetProxyListenAddress updates proxy listen address.
// Returns empty string on success or error message otherwise.
func (s *Server) SetProxyListenAddress(addr string) string {
	if addr == "" {
		return ""
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isRunningLocked() {
		return errConfigLocked.Error()
	}
	s.config.ProxyListenAddress = addr
	return ""
}

// ProxyRoutePath returns handled proxy route path.
func (s *Server) ProxyRoutePath() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.config.ProxyRoutePath
}

// SetProxyRoutePath updates handled proxy route path.
// Returns empty string on success or error message otherwise.
func (s *Server) SetProxyRoutePath(path string) string {
	if path == "" {
		return ""
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isRunningLocked() {
		return errConfigLocked.Error()
	}
	s.config.ProxyRoutePath = path
	return ""
}

// ControlListenAddress returns control listen address.
func (s *Server) ControlListenAddress() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.config.ControlListenAddress
}

// SetControlListenAddress updates control listen address.
// Returns empty string on success or error message otherwise.
func (s *Server) SetControlListenAddress(addr string) string {
	if addr == "" {
		return ""
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isRunningLocked() {
		return errConfigLocked.Error()
	}
	s.config.ControlListenAddress = addr
	return ""
}

// ControlRoutePath returns handled control route path.
func (s *Server) ControlRoutePath() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.config.ControlRoutePath
}

// SetControlRoutePath updates handled control route path.
// Returns empty string on success or error message otherwise.
func (s *Server) SetControlRoutePath(path string) string {
	if path == "" {
		return ""
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isRunningLocked() {
		return errConfigLocked.Error()
	}
	s.config.ControlRoutePath = path
	return ""
}

// IsRunning returns true when server has active instance.
func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isRunningLocked()
}

// Start starts proxy and control servers and blocks until both exit.
// Returns empty string on success or error message otherwise.
func (s *Server) Start() string {
	if err := s.startLocked(); err != nil {
		return err.Error()
	}
	if err := s.runServers(); err != nil {
		return err.Error()
	}
	return ""
}

// StartAsync starts proxy and control servers in a goroutine.
// Returns empty string on success or error message otherwise.
func (s *Server) StartAsync() string {
	if err := s.startLocked(); err != nil {
		return err.Error()
	}

	go func() {
		_ = s.runServers()
	}()

	return ""
}

// Stop shuts down running servers.
// Returns empty string on success or error message otherwise.
func (s *Server) Stop() string {
	s.mu.Lock()
	proxy := s.proxyHTTPServer
	control := s.controlHTTPServer
	s.mu.Unlock()

	if proxy == nil && control == nil {
		return ""
	}

	var errs []error
	if err := shutdownHTTPServer(proxy); err != nil {
		errs = append(errs, err)
	}
	if err := shutdownHTTPServer(control); err != nil {
		errs = append(errs, err)
	}

	s.clearRunningServers()

	if len(errs) > 0 {
		return errors.Join(errs...).Error()
	}
	return ""
}

func (s *Server) startLocked() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunningLocked() {
		return errServerAlreadyRunning
	}

	cfg := mergeConfig(defaultConfig(), &s.config)
	s.config = cfg

	sharedStore := processor.NewDefaultReqProcessorStore()

	proxyMux := app.BuildProxyMux(app.ServerConfig{
		ListenAddress: cfg.ProxyListenAddress,
		RoutePath:     cfg.ProxyRoutePath,
	}, sharedStore)

	controlMux := app.BuildControlMux(sharedStore, app.ControlMuxOptions{
		BasePath: cfg.ControlRoutePath,
	})

	s.proxyHTTPServer = &http.Server{
		Addr:    cfg.ProxyListenAddress,
		Handler: proxyMux,
	}
	s.controlHTTPServer = &http.Server{
		Addr:    cfg.ControlListenAddress,
		Handler: controlMux,
	}

	return nil
}

func (s *Server) runServers() error {
	s.mu.Lock()
	proxy := s.proxyHTTPServer
	control := s.controlHTTPServer
	s.mu.Unlock()

	errCh := make(chan error, 2)

	go func() {
		errCh <- normalizeServeError(proxy.ListenAndServe())
	}()
	go func() {
		errCh <- normalizeServeError(control.ListenAndServe())
	}()

	firstErr := <-errCh
	_ = shutdownHTTPServer(proxy)
	_ = shutdownHTTPServer(control)
	secondErr := <-errCh

	s.clearRunningServers()

	if firstErr != nil {
		return firstErr
	}
	if secondErr != nil {
		return secondErr
	}
	return nil
}

func normalizeServeError(err error) error {
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func shutdownHTTPServer(httpServer *http.Server) error {
	if httpServer == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := httpServer.Shutdown(ctx)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) isRunningLocked() bool {
	return s.proxyHTTPServer != nil || s.controlHTTPServer != nil
}

func (s *Server) clearRunningServers() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.proxyHTTPServer = nil
	s.controlHTTPServer = nil
}
