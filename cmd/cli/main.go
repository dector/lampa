package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	. "lampa/internal/globals"
	"lampa/internal/report"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name: "lampa",
		Commands: []*cli.Command{
			{
				Name: "collect",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "from",
						Usage: "project directory",
					},
					&cli.StringFlag{
						Name:  "to",
						Usage: "report directory",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					execCollect(c)
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
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}

	// file := ""
	// configName := ""
	// if len(os.Args) > 2 {
	// 	file = os.Args[1]
	// 	configName = os.Args[2]
	// } else {
	// 	os.Stderr.WriteString("usage: lampa <file> <configuration name>\n")
	// 	os.Exit(1)
	// }
	// if _, err := os.Stat(file); os.IsNotExist(err) {
	// 	os.Stderr.WriteString("file does not exist\n")
	// 	os.Exit(1)
	// }

	// data, err := os.ReadFile(file)
	// if err != nil {
	// 	panic(err)
	// }
	// content := string(data)

	// tree, err := internal.ParseTreeFromOutput(content, configName)
	// if err != nil {
	// 	panic(err)
	// }

	// // jsonTree, err := json.MarshalIndent(tree.Summary, "", "  ")
	// // if err != nil {
	// // 	panic(err)
	// // }

	// for _, it := range tree.Summary {
	// 	fmt.Println(it.String())
	// }
}

func execCollect(c *cli.Command) {
	fFrom := c.String("from")
	fTo := c.String("to")

	from := "."
	if fFrom != "" {
		info, err := os.Stat(fFrom)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if !info.IsDir() {
			fmt.Fprintf(os.Stderr, "error: %s is not a directory\n", fFrom)
			os.Exit(1)
		}
		from = fFrom
	}

	to := from
	if fTo != "" {
		info, err := os.Stat(fTo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if !info.IsDir() {
			fmt.Fprintf(os.Stderr, "error: %s is not a directory\n", fTo)
			os.Exit(1)
		}
		to = fTo
	}

	log.Printf("collect: from=%s to=%s", from, to)

	reportFile := path.Join(to, "lampa.report.json")
	if _, err := os.Stat(reportFile); err == nil {
		fmt.Fprintf(os.Stderr, "error: report file %s already exists\n", reportFile)
		os.Exit(1)
	}

	report := report.Report{
		Version: "0.0.1",
		Tool: report.ToolSegment{
			Name:       "lampa",
			Repository: "https://github.com/dector/lampa/",
			Version:    G.Version,
		},
	}

	file, err := os.Create(reportFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not create report file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	reportJson, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not marshal report: %v\n", err)
		os.Exit(1)
	}

	if _, err := file.Write(reportJson); err != nil {
		fmt.Fprintf(os.Stderr, "error: could not write report: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Report written to %s\n", reportFile)

	fmt.Println("Not Implemented Yet")
	os.Exit(127)
}
