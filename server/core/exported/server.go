package exported

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/dector/lampa/server/core/app"
)

var errServerAlreadyRunning = errors.New("server is already running")

// Server exposes embeddable HTTP server lifecycle for gomobile bindings.
type Server struct {
	mu            sync.Mutex
	listenAddress string
	routePath     string
	httpServer    *http.Server
}

// NewServer creates server with default config.
func NewServer() *Server {
	cfg := app.DefaultServerConfig()
	return &Server{
		listenAddress: cfg.ListenAddress,
		routePath:     cfg.RoutePath,
	}
}

// NewServerWithConfig creates server with provided config.
func NewServerWithConfig(listenAddress string, routePath string) *Server {
	s := NewServer()
	if listenAddress != "" {
		s.listenAddress = listenAddress
	}
	if routePath != "" {
		s.routePath = routePath
	}
	return s
}

// ListenAddress returns listen address.
func (s *Server) ListenAddress() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listenAddress
}

// SetListenAddress updates listen address.
func (s *Server) SetListenAddress(addr string) {
	if addr == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listenAddress = addr
}

// RoutePath returns handled route path.
func (s *Server) RoutePath() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.routePath
}

// SetRoutePath updates handled route path.
func (s *Server) SetRoutePath(path string) {
	if path == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routePath = path
}

// IsRunning returns true when server has active instance.
func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.httpServer != nil
}

// Start starts server and blocks until it exits.
// Returns empty string on success or error message otherwise.
func (s *Server) Start() string {
	_, _, err := s.startLocked()
	if err != nil {
		return err.Error()
	}

	err = s.httpServer.ListenAndServe()
	s.clearRunningServer()
	if err != nil && err != http.ErrServerClosed {
		return err.Error()
	}

	return ""
}

// StartAsync starts server in a goroutine.
// Returns empty string on success or error message otherwise.
func (s *Server) StartAsync() string {
	_, _, err := s.startLocked()
	if err != nil {
		return err.Error()
	}

	go func() {
		err := s.httpServer.ListenAndServe()
		s.clearRunningServer()
		if err != nil && err != http.ErrServerClosed {
			// no callback surface in gomobile-friendly API
		}
	}()

	return ""
}

// Stop shuts down running server.
// Returns empty string on success or error message otherwise.
func (s *Server) Stop() string {
	s.mu.Lock()
	httpServer := s.httpServer
	s.mu.Unlock()
	if httpServer == nil {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		return err.Error()
	}
	return ""
}

func (s *Server) startLocked() (string, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.httpServer != nil {
		return "", "", errServerAlreadyRunning
	}

	defaults := app.DefaultServerConfig()

	addr := s.listenAddress
	if addr == "" {
		addr = defaults.ListenAddress
	}

	routePath := s.routePath
	if routePath == "" {
		routePath = defaults.RoutePath
	}

	mux := http.NewServeMux()
	mux.HandleFunc(routePath, app.NewRequestHandler(app.ServerConfig{
		ListenAddress: addr,
		RoutePath:     routePath,
	}))
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return addr, routePath, nil
}

func (s *Server) clearRunningServer() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.httpServer = nil
}

