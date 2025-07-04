package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

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
	version := fmt.Sprintf("%s+%s", G.Version, G.BuildCommitShort)
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
	fmt.Printf("%sv%s\n\n", spacer(header, version), version)
	fmt.Println()
}

func spacer(lines []string, text string) string {
	maxLength := 0
	for _, s := range lines {
		l := utf8.RuneCountInString(s)
		if l > maxLength {
			maxLength = l
		}
	}

	textLength := utf8.RuneCountInString(text)

	if textLength < maxLength {
		return strings.Repeat(" ", maxLength-textLength)
	} else {
		return ""
	}
}
