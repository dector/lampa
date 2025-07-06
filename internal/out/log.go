package out

import (
	"fmt"
	"os"

	. "lampa/internal/globals"

	"github.com/fatih/color"
)

var red = color.New(color.FgRed).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()

func PrintlnErr(s string, a ...any) {
	msg := fmt.Sprintf("\nERROR: %s\n", fmt.Sprintf(s, a...))
	os.Stderr.WriteString(red(msg))
}

func PrintlnWarn(s string, a ...any) {
	msg := fmt.Sprintf("Warning: %s\n", fmt.Sprintf(s, a...))
	os.Stderr.WriteString(yellow(msg))
}

// --- Info

func Info(s string) {
	if G.Verbosity < VerbosityInfo {
		return
	}

	fmt.Println(s)
}
