package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	coreapp "github.com/dector/lampa/server/core/app"
	"github.com/dector/lampa/server/core/logstore"
	"github.com/dector/lampa/server/core/processor"
)

func RunProxy(cfg ServerConfig, listen func(addr string, h http.Handler) error) error {
	return RunProxyWithStore(cfg, nil, listen)
}

func RunProxyWithStore(cfg ServerConfig, store processor.ReqProcessorStore, listen func(addr string, h http.Handler) error) error {
	return RunProxyWithStoreAndLogs(cfg, store, nil, listen)
}

func RunProxyWithStoreAndLogs(cfg ServerConfig, store processor.ReqProcessorStore, logs logstore.Store, listen func(addr string, h http.Handler) error) error {
	mux := coreapp.BuildProxyMuxWithLogStore(cfg, store, logs)

	fmt.Printf("Server listening on %s\n", cfg.ListenAddress)
	return listen(cfg.ListenAddress, mux)
}

func RunWithControl(cfg ServerConfig, listen func(addr string, h http.Handler) error) error {
	sharedStore := processor.NewDefaultReqProcessorStore()
	sharedLogs := logstore.NewInMemoryStore(logstore.DefaultMaxBytes)

	errCh := make(chan error, 2)

	go func() {
		errCh <- RunProxyWithStoreAndLogs(cfg, sharedStore, sharedLogs, listen)
	}()

	go func() {
		controlMux := coreapp.BuildControlMuxWithLogStore(sharedStore, sharedLogs, coreapp.ControlMuxOptions{
			BasePath: "/",
		})

		fmt.Printf("Control server listening on %s\n", cfg.ControlListenAddress)
		errCh <- listen(cfg.ControlListenAddress, controlMux)
	}()

	return <-errCh
}

func LoadRuntimeConfig(getenv func(string) string) ServerConfig {
	return LoadRuntimeConfigWithArgs(getenv, nil)
}

func LoadRuntimeConfigWithArgs(getenv func(string) string, args []string) ServerConfig {
	cfg := DefaultServerConfig()
	cfg.ProxyPort = parsePort(getenv("PORT"), cfg.ProxyPort)
	cfg.ControlPort = parsePort(getenv("PORT_CTRL"), cfg.ControlPort)
	cfg.ProxyBindHost = parseHost(getenv("BIND_HOST"), cfg.ProxyBindHost)
	cfg.ControlBindHost = parseHost(getenv("BIND_HOST_CTRL"), cfg.ControlBindHost)
	cfg = applyCLIConfig(cfg, args)
	cfg.ListenAddress = coreapp.ComposeListenAddress(cfg.ProxyBindHost, cfg.ProxyPort)
	cfg.ControlListenAddress = coreapp.ComposeListenAddress(cfg.ControlBindHost, cfg.ControlPort)
	cfg.WebUI.ListenAddress = coreapp.ComposeListenAddress(cfg.WebUI.BindHost, cfg.WebUI.Port)
	return cfg
}

func applyCLIConfig(cfg ServerConfig, args []string) ServerConfig {
	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		switch {
		case arg == "--webui":
			cfg.WebUI.Enabled = true
			cfg.CaptureTraffic = true
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				cfg.WebUI.Port = parsePort(args[i+1], cfg.WebUI.Port)
				i++
			}
		case strings.HasPrefix(arg, "--webui="):
			cfg.WebUI.Enabled = true
			cfg.CaptureTraffic = true
			cfg.WebUI.Port = parsePort(strings.TrimPrefix(arg, "--webui="), cfg.WebUI.Port)
		}
	}
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
	cfg := LoadRuntimeConfigWithArgs(os.Getenv, os.Args[1:])
	if cfg.WebUI.Enabled {
		fmt.Printf("Web UI enabled on %s\n", cfg.WebUI.ListenAddress)
	}
	if err := RunWithControl(cfg, http.ListenAndServe); err != nil {
		panic(err)
	}
}
