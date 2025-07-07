package out

import (
	"fmt"
	"os"

	. "github.com/dector/lampa/internal/globals"

	"github.com/fatih/color"
)

var red = color.New(color.FgRed).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()

func PrintlnErr(s string, a ...any) {
	msg := fmt.Sprintf("\nERROR: %s\n", fmt.Sprintf(s, a...))

	if G.UseOnlyStdout {
		os.Stdout.WriteString(red(msg))
	} else {
		os.Stderr.WriteString(red(msg))
	}
}

func PrintlnWarn(s string, a ...any) {
	msg := fmt.Sprintf("Warning: %s\n", fmt.Sprintf(s, a...))

	if G.UseOnlyStdout {
		os.Stdout.WriteString(yellow(msg))
	} else {
		os.Stderr.WriteString(yellow(msg))
	}
}

// --- Info

func Info(s string) {
	if G.Verbosity < VerbosityInfo {
		return
	}

	fmt.Println(s)
}
