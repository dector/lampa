package out

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

var red = color.New(color.FgRed).SprintFunc()

func PrintfErr(s string, a ...any) {
	msg := fmt.Sprintf("ERROR: %s", fmt.Sprintf(s, a...))
	os.Stderr.WriteString(red(msg))
}
