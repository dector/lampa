package main

import (
	"context"
	"fmt"
	"os"

	. "lampa/internal/globals"
	"lampa/internal/out"

	"github.com/square/exit"
)

func main() {
	// log.Printf("os.Args: %v", os.Args)

	G.Init()

	printHeader()

	cmd := CreateCliCommand()
	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
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
}

func printHeader() {
	header := []string{
		"██╗      █████╗ ███╗   ███╗██████╗  █████╗",
		"██║     ██╔══██╗████╗ ████║██╔══██╗██╔══██╗",
		"██║     ███████║██╔████╔██║██████╔╝███████║",
		"██║     ██╔══██║██║╚██╔╝██║██╔═══╝ ██╔══██║",
		"███████╗██║  ██║██║ ╚═╝ ██║██║     ██║  ██║",
		"╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝     ╚═╝  ╚═╝",
	}

	for _, line := range header {
		fmt.Println(line)
	}
	fmt.Println()
}
