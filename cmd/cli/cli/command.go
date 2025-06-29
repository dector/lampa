package cli

import (
	"context"
	"fmt"
	"lampa/cmd/cli/collect"
	"lampa/cmd/cli/compare"
	. "lampa/internal/globals"

	"github.com/urfave/cli/v3"
)

func CreateCliCommand() *cli.Command {
	cmd := &cli.Command{
		Name: "lampa",
		Commands: []*cli.Command{
			{
				Name: "collect",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "from",
						Usage: "specify project directory",
					},
					&cli.StringFlag{
						Name:  "to",
						Usage: "specify report directory",
					},
					&cli.StringFlag{
						Name:        "variant",
						Usage:       "build variant to use",
						DefaultText: "release",
					},
					&cli.StringFlag{
						Name:        "with-name",
						Usage:       "report file name (without extension)",
						DefaultText: "report",
					},

					&cli.BoolFlag{
						Name:  "rewrite-report",
						Usage: "allow to rewrite report file if it already exists",
					},
				},
				Action: collect.CmdActionCollect,
			},
			{
				Name:   "compare",
				Action: compare.ActionCmdCompare,
			},
			{
				Name: "version",
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Println(G.Version)
					return nil
				},
			},
		},
	}
	return cmd
}
