package main

import (
	"context"
	"fmt"
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

	handleVersionArg()

	out.PrintHeader()

	cmd := CreateCliCommand()
	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		handleError(err)
	}
}

// handleVersionArg checks command line arguments for version commands.
// It supports "version" and "version --short" commands and exits after printing version info.
func handleVersionArg() {
	args := os.Args

	if len(args) == 2 {
		cmd := args[1]
		if cmd == "version" {
			printVersionAndExit("")
		}
	} else if len(args) == 3 {
		cmd := args[1]
		if cmd == "version" && args[2] == "--short" {
			printVersionAndExit("short")
		}
	}
}

// printVersionAndExit prints the version information in the specified format and exits.
// If format is "short", only the version is printed.
// Otherwise, version and build commit are printed.
func printVersionAndExit(format string) {
	switch format {
	case "short":
		fmt.Printf("%s\n", G.Version)
	default:
		fmt.Printf("%s+%s\n", G.Version, G.BuildCommit)
	}
	os.Exit(exit.OK)
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
