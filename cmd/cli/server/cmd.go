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

// CreatePingCliCommand exports ping command constructor for reuse by other namespaces.
func CreatePingCliCommand() *cli.Command {
	return createPingCommand()
}

// CreateSetCliCommand exports set command constructor for reuse by other namespaces.
func CreateSetCliCommand() *cli.Command {
	return createSetCommand()
}

// CreateSetDefaultCliCommand exports set-default command constructor for reuse by other namespaces.
func CreateSetDefaultCliCommand() *cli.Command {
	return createSetDefaultCommand()
}
