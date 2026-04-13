package proxy

import "github.com/urfave/cli/v3"

// CreateCliCommand creates top-level proxy command tree.
func CreateCliCommand() *cli.Command {
	return &cli.Command{
		Name:  "proxy",
		Usage: "proxy operations",
		Commands: []*cli.Command{
			createLogsCommand(),
		},
	}
}

func createLogsCommand() *cli.Command {
	return &cli.Command{
		Name:  "logs",
		Usage: "proxy logs operations",
		Commands: []*cli.Command{
			createGetAllCommand(),
		},
	}
}
