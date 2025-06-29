package main

import (
	"context"
	"log"
	"os"

	"lampa/cmd/cli/cli"
	. "lampa/internal/globals"
)

func main() {
	// log.Printf("os.Args: %v", os.Args)

	G.Init()

	cmd := cli.CreateCliCommand()
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.SetOutput(os.Stderr)
		log.Printf("ERROR: %+v", err)
		if errWithStack, ok := err.(interface{ StackTrace() any }); ok {
			log.Printf("%stacktrace: +v", errWithStack.StackTrace())
		}
		os.Exit(1)
	}
}
