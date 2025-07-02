package cli

import (
	"context"
	"fmt"
	"lampa/cmd/cli/collect"
	"lampa/cmd/cli/compare"
	. "lampa/internal/globals"
	"lampa/internal/out"
	"net/http"

	"github.com/samber/lo"
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
						Name:  "variant",
						Usage: "build variant to use",
						Value: "release",
					},
					&cli.StringFlag{
						Name:  "with-name",
						Usage: "report file name (without extension)",
						Value: "report",
					},
					&cli.BoolFlag{
						Name:  "with-html",
						Usage: "generate HTML report as well",
						Value: false,
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
			// TODO hide in production
			{
				Name: "testhtml",
				Action: func(ctx context.Context, c *cli.Command) error {
					srv := &http.Server{Addr: ":8080"}
					http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "text/html")
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
			},
			{
				Name: "version",
				Action: func(ctx context.Context, c *cli.Command) error {
					fmt.Println(G.Version)
					return nil
				},
			},
		},
		CommandNotFound: func(ctx context.Context, c *cli.Command, s string) {
			out.PrintlnErr("Command '%s' not found\n", s)

			cli.ShowAppHelpAndExit(c, 127)
		},
	}
	return cmd
}
