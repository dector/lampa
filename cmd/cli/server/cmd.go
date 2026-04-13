package server

import "github.com/urfave/cli/v3"

func CreateCliCommand() *cli.Command {
	return &cli.Command{
		Name:  "server",
		Usage: "control server operations",
		Commands: []*cli.Command{
			createPingCommand(),
			createSetCommand(),
			createProxyCommand(),
		},
	}
}
