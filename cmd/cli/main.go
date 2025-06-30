package main

import (
	"context"
	"os"

	"lampa/cmd/cli/cli"
	. "lampa/internal/globals"
	"lampa/internal/out"
)

func main() {
	// log.Printf("os.Args: %v", os.Args)

	G.Init()

	cmd := cli.CreateCliCommand()
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		out.PrintlnErr("%+v", err)
		if errWithStack, ok := err.(interface{ StackTrace() any }); ok {
			out.PrintlnErr("%+v", errWithStack.StackTrace())
		}
		os.Exit(1)
	}
}
