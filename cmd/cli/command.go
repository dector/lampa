package main

import (
	"context"
	"fmt"
	"lampa/cmd/cli/collect"
	"lampa/cmd/cli/compare"
	"lampa/internal/out"
	"net/http"

	"github.com/samber/lo"
	"github.com/square/exit"
	"github.com/urfave/cli/v3"
)

func CreateCliCommand() *cli.Command {
	cmd := &cli.Command{
		Name: "lampa",
		// Version: G.Version,
		Usage: "Android releases analyzer",
		Commands: []*cli.Command{
			collect.CreateCliCommand(),
			compare.CreateCliCommand(),
			CreateVersionCommand(),
			// devReportCommand(),
		},
		CommandNotFound: handleCommandNotFound,
	}
	return cmd
}

func CreateVersionCommand() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"--version"},
		Usage:   "show version and exit",
	}
}

func handleCommandNotFound(ctx context.Context, c *cli.Command, s string) {
	out.PrintlnErr("Command '%s' not found\n", s)

	cli.ShowAppHelpAndExit(c, exit.UnknownSubcommand)
}

func devReportCommand() *cli.Command {
	return &cli.Command{
		Name: "testhtml",
		Action: func(ctx context.Context, c *cli.Command) error {
			srv := &http.Server{Addr: ":8080"}
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")

				r1 := lo.Must(compare.ReadReportFromFile("out/libretube-prev.lampa.json"))
				// d := lo.Must(collect.GenerateHtmlReport(r1))

				r2 := lo.Must(compare.ReadReportFromFile("out/libretube-next.lampa.json"))
				// d := lo.Must(collect.GenerateHtmlReport(r2))

				d := lo.Must(compare.GenerateComparingHtmlReport(r1, r2))

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
