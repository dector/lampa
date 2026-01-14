package generate

import (
	"github.com/urfave/cli/v3"
)

func CreateCliCommand() *cli.Command {
	return &cli.Command{
		Name:    "generate",
		Aliases: []string{"gen"},
		Usage:   "generate project files",
		Commands: []*cli.Command{
			CreateOciCommand(),
		},
	}
}
