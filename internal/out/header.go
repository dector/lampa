package out

import (
	"fmt"
	"strings"
	"unicode/utf8"

	. "lampa/internal/globals"
)

func PrintHeader() {
	version := fmt.Sprintf("%s+%s", G.Version, G.BuildCommitShort)
	header := []string{
		"██╗      █████╗ ███╗   ███╗██████╗  █████╗",
		"██║     ██╔══██╗████╗ ████║██╔══██╗██╔══██╗",
		"██║     ███████║██╔████╔██║██████╔╝███████║",
		"██║     ██╔══██║██║╚██╔╝██║██╔═══╝ ██╔══██║",
		"███████╗██║  ██║██║ ╚═╝ ██║██║     ██║  ██║",
		"╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝     ╚═╝  ╚═╝",
	}

	fmt.Println()
	for _, line := range header {
		fmt.Println(line)
	}
	fmt.Printf("%sv%s\n", spacer(header, version), version)
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
