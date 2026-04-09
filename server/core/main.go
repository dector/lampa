package main

import (
	"fmt"
	"net/http"
)

func Run(cfg ServerConfig, listen func(addr string, h http.Handler) error) error {
	mux := http.NewServeMux()
	mux.HandleFunc(cfg.RoutePath, NewRequestHandler(cfg))

	fmt.Printf("Server listening on %s\n", cfg.ListenAddress)
	return listen(cfg.ListenAddress, mux)
}

func main() {
	cfg := DefaultServerConfig()
	if err := Run(cfg, http.ListenAndServe); err != nil {
		panic(err)
	}
}
