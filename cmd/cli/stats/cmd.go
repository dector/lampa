package stats

import (
	"github.com/dector/lampa/cmd/cli/collect"
	"github.com/dector/lampa/cmd/cli/compare"
	"github.com/urfave/cli/v3"
)

func CreateCliCommand() *cli.Command {
	return &cli.Command{
		Name:  "stats",
		Usage: "collect and compare Android release statistics",
		Commands: []*cli.Command{
			collect.CreateCliCommand(),
			compare.CreateCliCommand(),
		},
	}
}
