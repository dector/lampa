package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	coreapp "github.com/dector/lampa/server/core/app"
	"github.com/dector/lampa/server/core/processor"
)

func RunProxy(cfg ServerConfig, listen func(addr string, h http.Handler) error) error {
	return RunProxyWithStore(cfg, nil, listen)
}

func RunProxyWithStore(cfg ServerConfig, store processor.ReqProcessorStore, listen func(addr string, h http.Handler) error) error {
	mux := http.NewServeMux()
	mux.HandleFunc(cfg.RoutePath, NewRequestHandlerWithStore(cfg, store))

	fmt.Printf("Server listening on %s\n", cfg.ListenAddress)
	return listen(cfg.ListenAddress, mux)
}

func RunWithControl(cfg ServerConfig, listen func(addr string, h http.Handler) error) error {
	sharedStore := processor.NewDefaultReqProcessorStore()

	errCh := make(chan error, 2)

	go func() {
		errCh <- RunProxyWithStore(cfg, sharedStore, listen)
	}()

	go func() {
		controlMux := http.NewServeMux()
		pingHandler := coreapp.NewControlPingHandler(sharedStore)
		controlMux.HandleFunc("/ping", pingHandler)
		controlMux.HandleFunc("/api/v0/proc_count", coreapp.NewControlProcCountHandler(sharedStore))
		// Keep temporary backward-compatible alias.
		controlMux.HandleFunc("/", pingHandler)

		fmt.Printf("Control server listening on %s\n", cfg.ControlListenAddress)
		errCh <- listen(cfg.ControlListenAddress, controlMux)
	}()

	return <-errCh
}

func LoadRuntimeConfig(getenv func(string) string) ServerConfig {
	cfg := DefaultServerConfig()
	cfg.ProxyPort = parsePort(getenv("PORT"), cfg.ProxyPort)
	cfg.ControlPort = parsePort(getenv("PORT_CTRL"), cfg.ControlPort)
	cfg.ProxyBindHost = parseHost(getenv("BIND_HOST"), cfg.ProxyBindHost)
	cfg.ControlBindHost = parseHost(getenv("BIND_HOST_CTRL"), cfg.ControlBindHost)
	cfg.ListenAddress = coreapp.ComposeListenAddress(cfg.ProxyBindHost, cfg.ProxyPort)
	cfg.ControlListenAddress = coreapp.ComposeListenAddress(cfg.ControlBindHost, cfg.ControlPort)
	return cfg
}

func parsePort(raw string, fallback int) int {
	trimmed := strings.TrimSpace(strings.TrimPrefix(raw, ":"))
	if trimmed == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(trimmed)
	if err != nil || parsed <= 0 || parsed > 65535 {
		return fallback
	}
	return parsed
}

func parseHost(raw string, fallback string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func main() {
	cfg := LoadRuntimeConfig(os.Getenv)
	if err := RunWithControl(cfg, http.ListenAndServe); err != nil {
		panic(err)
	}
}
