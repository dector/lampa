package main

import (
	"flag"
	"fmt"
	"io"
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

type RuntimeOptions struct {
	WebUIEnabled bool
	WebUIPort    int
}

func LoadRuntimeConfig(getenv func(string) string) ServerConfig {
	return LoadRuntimeConfigWithOptions(getenv, RuntimeOptions{})
}

func LoadRuntimeConfigWithOptions(getenv func(string) string, options RuntimeOptions) ServerConfig {
	cfg := DefaultServerConfig()
	cfg.ProxyPort = parsePort(getenv("PORT"), cfg.ProxyPort)
	cfg.ControlPort = parsePort(getenv("PORT_CTRL"), cfg.ControlPort)
	cfg.ProxyBindHost = parseHost(getenv("BIND_HOST"), cfg.ProxyBindHost)
	cfg.ControlBindHost = parseHost(getenv("BIND_HOST_CTRL"), cfg.ControlBindHost)

	if options.WebUIEnabled {
		cfg.WebUI.Enabled = true
		cfg.CaptureTraffic = true
	}
	if options.WebUIPort != 0 {
		cfg.WebUI.Port = options.WebUIPort
	}

	cfg.ListenAddress = coreapp.ComposeListenAddress(cfg.ProxyBindHost, cfg.ProxyPort)
	cfg.ControlListenAddress = coreapp.ComposeListenAddress(cfg.ControlBindHost, cfg.ControlPort)
	cfg.WebUI.ListenAddress = coreapp.ComposeListenAddress(cfg.WebUI.BindHost, cfg.WebUI.Port)
	return cfg
}

func parseRuntimeOptions(args []string, defaults ServerConfig) (RuntimeOptions, error) {
	flags := flag.NewFlagSet("lampa-server", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	webUIEnabled := flags.Bool("webui", defaults.WebUI.Enabled, "enable the Web UI")
	webUIPort := flags.Int("webui-port", defaults.WebUI.Port, "Web UI port")

	if err := flags.Parse(args); err != nil {
		return RuntimeOptions{}, err
	}
	if flags.NArg() > 0 {
		return RuntimeOptions{}, fmt.Errorf("unexpected positional argument: %s", flags.Arg(0))
	}
	if *webUIPort <= 0 || *webUIPort > 65535 {
		return RuntimeOptions{}, fmt.Errorf("invalid --webui-port value: %d", *webUIPort)
	}

	return RuntimeOptions{
		WebUIEnabled: *webUIEnabled,
		WebUIPort:    *webUIPort,
	}, nil
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
	options, err := parseRuntimeOptions(os.Args[1:], DefaultServerConfig())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse arguments: %v\n", err)
		os.Exit(2)
	}
	cfg := LoadRuntimeConfigWithOptions(os.Getenv, options)
	if cfg.WebUI.Enabled {
		fmt.Printf("Web UI enabled on %s\n", cfg.WebUI.ListenAddress)
	}
	if err := RunWithControl(cfg, http.ListenAndServe); err != nil {
		panic(err)
	}
}
