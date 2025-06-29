package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"lampa/internal"
	. "lampa/internal/globals"
	"lampa/internal/report"

	"github.com/urfave/cli/v3"
)

func main() {
	log.Printf("os.Args: %v", os.Args)

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
				Action: func(ctx context.Context, c *cli.Command) error {
					execCollect(c)
					return nil
				},
			},
			{
				Name: "Compare",
				Action: func(ctx context.Context, c *cli.Command) error {
					return execCompare(c)
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

	fWithName := c.String("with-name")
	if fWithName == "" {
		fWithName = "lampa"
	}

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

	gradlewPath := path.Join(from, "gradlew")
	info, err := os.Stat(gradlewPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "error: %s does not exist\n", gradlewPath)
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stderr, "error: could not stat %s: %v\n", gradlewPath, err)
			os.Exit(1)
		}
	}
	if info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: %s exists but is a directory, not a file\n", gradlewPath)
		os.Exit(1)
	}

	reportFile := path.Join(to, fWithName+".report.json")
	if c.Bool("rewrite-report") {
		log.Printf("rewrite-report flag is enabled, existing report file (if any) will be overwritten")
	} else {
		if _, err := os.Stat(reportFile); err == nil {
			fmt.Fprintf(os.Stderr, "error: report file %s already exists\n", reportFile)
			os.Exit(1)
		}
	}

	report := collectReport(CollectReportArgs{
		ProjectDir: from,
		ReportDir:  to,
	})

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
}

type CollectReportArgs struct {
	ProjectDir string
	ReportDir  string
}

func collectReport(args CollectReportArgs) report.Report {
	result := report.Report{
		Version: "0.0.1-SNAPSHOT",
		Tool: report.ToolSegment{
			Name:       "lampa",
			Repository: "https://github.com/dector/lampa/",
			Version:    G.Version,
		},
	}

	configurationName := "prodReleaseCompileClasspath"

	gradlewPath := path.Join(args.ProjectDir, "gradlew")
	cmd := exec.Command(gradlewPath, "--no-daemon", "--console", "plain", "-q", "app:dependencies", "--configuration", configurationName)
	cmd.Dir = args.ProjectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("failed to execute gradlew: %v\nOutput:\n%s", err, string(output))
	}

	// fmt.Println(string(output))

	tree, err := internal.ParseTreeFromOutput(string(output), configurationName)
	if err != nil {
		log.Fatalf("failed to parse tree: %v", err)
	}

	result.Dependencies = report.DependenciesSegment{
		Compiled: make([]string, 0),
	}
	for _, info := range tree.Summary {
		d := info.String()
		result.Dependencies.Compiled = append(result.Dependencies.Compiled, d)
	}

	result.Context = parseContext(args)

	return result
}

func parseContext(args CollectReportArgs) report.ContextSegment {
	result := report.ContextSegment{}

	_, err := exec.LookPath("git")
	if err != nil {
		log.Fatalf("git not found in PATH: %v", err)
	}

	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = args.ProjectDir
	if err := cmd.Run(); err != nil {
		return result
	}

	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = args.ProjectDir
	out, err := cmd.Output()
	if err == nil {
		result.GitCommit = strings.TrimSpace(string(out))
	}

	return result
}

func execCompare(c *cli.Command) error {
	return nil
}
