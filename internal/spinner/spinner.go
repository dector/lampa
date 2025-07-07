package spinner

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"

	. "lampa/internal/globals"
)

type SpinnerArgs struct {
	Msg             string
	MsgAfterSuccess string
	MsgAfterFail    string
}

func Create[T any](args SpinnerArgs, action func() (T, error)) (*T, error) {
	if G.UsePlainOutput {
		fmt.Println(args.Msg)
		data, err := action()
		if err != nil {
			fmt.Printf("✗ %s\n", args.MsgAfterFail)
			return nil, err
		}

		fmt.Printf("✔ %s\n", args.MsgAfterSuccess)
		return &data, err
	} else {
		blue := color.New(color.FgBlue).SprintfFunc()
		green := color.New(color.FgGreen).SprintfFunc()
		red := color.New(color.FgRed).SprintfFunc()
		cs := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		s := spinner.New(cs, 100*time.Millisecond)
		s.Color("blue")
		s.Suffix = blue(" " + args.Msg)
		s.FinalMSG = green("✔ " + args.MsgAfterSuccess + "\n")
		s.Start()
		defer s.Stop()

		data, err := action()
		if err != nil {
			s.FinalMSG = red("✗ " + args.MsgAfterFail + "\n")
			return nil, err
		}

		return &data, err
	}
}
