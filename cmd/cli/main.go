package main

import (
	"context"
	"os"

	. "github.com/dector/lampa/internal/globals"
	"github.com/dector/lampa/internal/out"

	"github.com/square/exit"
)

// main is the entry point of the app.
// It initializes globals, handles version arguments and runs the CLI command.
func main() {
	// log.Printf("os.Args: %v", os.Args)

	G.Init()

	cmd := CreateCliCommand()
	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		handleError(err)
	}
}

// handleError prints the error message and stack trace if available,
// then exits with a non-zero status.
func handleError(err error) {
	// if e, ok := err.(exit.Error); ok {
	// 	err = e.Cause
	// }

	out.PrintlnErr("%+v", err)

	errWithStack, ok := err.(interface{ StackTrace() any })
	if ok {
		out.PrintlnErr("%+v", errWithStack.StackTrace())
	}

	os.Exit(exit.NotOK)
}
