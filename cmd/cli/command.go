package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/dector/lampa/cmd/cli/collect"
	"github.com/dector/lampa/cmd/cli/compare"
	"github.com/dector/lampa/cmd/cli/generate"
	"github.com/dector/lampa/cmd/cli/server"
	"github.com/dector/lampa/internal/out"

	. "github.com/dector/lampa/internal/globals"

	"github.com/samber/lo"
	"github.com/square/exit"
	"github.com/urfave/cli/v3"
)

func CreateCliCommand() *cli.Command {
	cmd := &cli.Command{
		Name: "lampa",
		// Version: G.Version,
		Usage: "Android releases analyzer",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "v",
				Aliases: []string{"vv", "vvv"},
				Usage:   "Increase verbosity (-v, -vv, -vvv)",
				Value:   false,
			},
		},
		Commands: []*cli.Command{
			collect.CreateCliCommand(),
			compare.CreateCliCommand(),
			generate.CreateCliCommand(),
			server.CreateCliCommand(),
			CreateVersionCommand(),
			devReportCommand(),
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			// Set verbosity
			verbosity := c.Count("v")
			switch verbosity {
			case 0:
				G.Verbosity = VerbosityNormal
			case 1:
				G.Verbosity = VerbosityInfo
			case 2:
				G.Verbosity = VerbosityDebug
			default:
				if verbosity >= 3 {
					G.Verbosity = VerbosityTrace
				}
			}

			return ctx, nil
		},
		CommandNotFound: handleCommandNotFound,
	}
	return cmd
}

func CreateVersionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "print long build version and exit",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "short",
				Usage: "display version without build info",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			// Empty command because it's handled on the top
			return nil
		},
	}
}

func handleCommandNotFound(ctx context.Context, c *cli.Command, s string) {
	out.PrintlnErr("Command '%s' not found\n", s)

	cli.ShowAppHelpAndExit(c, exit.UnknownSubcommand)
}

func devReportCommand() *cli.Command {
	if os.Getenv("DEV") != "1" {
		return &cli.Command{}
	}

	return &cli.Command{
		Name: "testhtml",
		Action: func(ctx context.Context, c *cli.Command) error {
			srv := &http.Server{Addr: ":8080"}
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")

				// r1 := lo.Must(compare.ReadReportFromFile("out/libretube-prev.lampa.json"))
				// d := lo.Must(collect.GenerateHtmlReport(r1))

				// r2 := lo.Must(compare.ReadReportFromFile("out/libretube-next.lampa.json"))
				// d := lo.Must(collect.GenerateHtmlReport(r2))

				// d := lo.Must(compare.GenerateComparingHtmlReport(r1, r2))

				rp := lo.Must(compare.ReadReportFromFile("out/report.lampa.json"))
				d := lo.Must(collect.GenerateHtmlReport(rp))

				w.Write([]byte(d))
			})
			fmt.Println("HTTP server started on :8080")
			err := srv.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				fmt.Printf("HTTP server error: %v\n", err)
				return err
			}
			return nil
		},
	}
}
