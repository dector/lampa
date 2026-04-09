package main

import (
	"fmt"
	"net/http"
)

func main() {
	cfg := DefaultServerConfig()

	http.HandleFunc(cfg.RoutePath, NewRequestEchoHandler(cfg))

	fmt.Printf("Server listening on %s\n", cfg.ListenAddress)
	if err := http.ListenAndServe(cfg.ListenAddress, nil); err != nil {
		panic(err)
	}
}
