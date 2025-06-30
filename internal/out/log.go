package out

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

var red = color.New(color.FgRed).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()

func PrintlnErr(s string, a ...any) {
	msg := fmt.Sprintf("ERROR: %s\n", fmt.Sprintf(s, a...))
	os.Stderr.WriteString(red(msg))
}

func PrintlnWarn(s string, a ...any) {
	msg := fmt.Sprintf("Warning: %s\n", fmt.Sprintf(s, a...))
	os.Stderr.WriteString(yellow(msg))
}
