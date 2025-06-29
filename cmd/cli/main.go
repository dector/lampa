package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
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
				Action: func(ctx context.Context, c *cli.Command) error {
					execCollect(c)
					return nil
				},
			},
			{
				Name: "compare",
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

	buildVariant := c.String("variant")
	if buildVariant == "" {
		buildVariant = "release"
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
		ProjectDir:   from,
		ReportDir:    to,
		BuildVariant: buildVariant,
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
	ProjectDir   string
	ReportDir    string
	BuildVariant string
}

func collectReport(args CollectReportArgs) report.Report {
	result := report.Report{
		Version: "0.0.1-SNAPSHOT",
		Type:    "CollectionReport",
		Tool: report.ToolSegment{
			Name:       "lampa",
			Repository: "https://github.com/dector/lampa/",
			Version:    G.Version,
		},
	}

	configurationName := args.BuildVariant + "CompileClasspath"

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
		result.Git.Commit = strings.TrimSpace(string(out))
	}

	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = args.ProjectDir
	out, err = cmd.Output()
	if err == nil {
		result.Git.IsDirty = len(strings.TrimSpace(string(out))) > 0
	}

	cmd = exec.Command("git", "describe", "--tags", "--long")
	cmd.Dir = args.ProjectDir
	out, err = cmd.Output()
	if err == nil {
		parts := strings.SplitN(strings.TrimSpace(string(out)), "-", 3)
		if len(parts) == 3 {
			result.Git.Tag = parts[0]
			commitsAfterTag, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				log.Printf("warning: could not parse commits after tag %q as uint: %v", parts[1], err)
			} else {
				result.Git.CommitsAfterTag = uint(commitsAfterTag)
			}
		} else {
			log.Printf("warning: unexpected format from git describe: %q", string(out))
		}
	} else {
		log.Printf("warning: git describe failed: %v", err)
	}

	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = args.ProjectDir
	out, err = cmd.Output()
	if err == nil {
		result.Git.Branch = strings.TrimSpace(string(out))
	}

	return result
}

func execCompare(c *cli.Command) error {
	if c.NArg() != 2 {
		log.Fatalf("error: expected exactly two files to compare, got %d", c.NArg())
	}

	file1 := c.Args().Get(0)
	file2 := c.Args().Get(1)

	info1, err := os.Stat(file1)
	if err != nil {
		log.Fatalf("error: could not stat %s: %v", file1, err)
	}
	if info1.IsDir() {
		log.Fatalf("error: %s exists but is a directory, not a file", file1)
	}

	info2, err := os.Stat(file2)
	if err != nil {
		log.Fatalf("error: could not stat %s: %v", file2, err)
	}
	if info2.IsDir() {
		log.Fatalf("error: %s exists but is a directory, not a file", file2)
	}

	data1, err := os.ReadFile(file1)
	if err != nil {
		log.Fatalf("error: could not read %s: %v", file1, err)
	}

	var report1 report.Report
	if err := json.Unmarshal(data1, &report1); err != nil {
		log.Fatalf("error: could not parse %s as report.Report: %v", file1, err)
	}

	data2, err := os.ReadFile(file2)
	if err != nil {
		log.Fatalf("error: could not read %s: %v", file2, err)
	}

	var report2 report.Report
	if err := json.Unmarshal(data2, &report2); err != nil {
		log.Fatalf("error: could not parse %s as report.Report: %v", file2, err)
	}

	fmt.Printf("Comparing releases %s...%s\n", report1.Context.Git.Commit, report2.Context.Git.Commit)

	dep1 := parseDependencies(report1)
	dep2 := parseDependencies(report2)

	newDeps := findNewDeps(dep1, dep2)
	removedDeps := findRemovedDeps(dep1, dep2)
	changedDeps := findChangedDeps(dep1, dep2)

	fmt.Println()
	fmt.Printf("Total dependencies before: %d\n", len(dep1))
	fmt.Printf("Total dependencies after: %d\n", len(dep2))

	fmt.Println()
	fmt.Printf("New dependencies:\n")
	for _, dep := range newDeps {
		fmt.Printf("- %s:%s: %s\n", dep.Group, dep.Name, dep.Version)
	}

	fmt.Println()
	fmt.Printf("Removed dependencies:\n")
	for _, dep := range removedDeps {
		fmt.Printf("- %s:%s: %s\n", dep.Group, dep.Name, dep.Version)
	}

	fmt.Println()
	fmt.Printf("Changed dependencies:\n")
	for _, dep := range changedDeps {
		fmt.Printf("- %s:%s: %s -> %s\n", dep.Dependency.Group, dep.Dependency.Name, dep.PrevVersion, dep.Dependency.Version)
	}

	fmt.Println()

	return nil
}

type Dependency struct {
	Group   string
	Name    string
	Version string
}

func parseDependencies(report report.Report) []Dependency {
	result := make([]Dependency, 0, len(report.Dependencies.Compiled))

	for _, depStr := range report.Dependencies.Compiled {
		parts := strings.Split(depStr, ":")
		if len(parts) == 3 {
			result = append(result, Dependency{
				Group:   parts[0],
				Name:    parts[1],
				Version: parts[2],
			})
		} else {
			log.Fatalf("error: dependency string %q does not have 3 parts (group:name:version)", depStr)
		}
	}

	return result
}

func findNewDeps(oldDeps, newDeps []Dependency) []Dependency {
	result := []Dependency{}
	oldMap := make(map[string]Dependency)
	for _, dep := range oldDeps {
		key := dep.Group + ":" + dep.Name
		oldMap[key] = dep
	}
	for _, dep := range newDeps {
		key := dep.Group + ":" + dep.Name
		if _, exists := oldMap[key]; !exists {
			result = append(result, dep)
		}
	}
	return result
}

func findRemovedDeps(oldDeps, newDeps []Dependency) []Dependency {
	result := []Dependency{}
	newMap := make(map[string]Dependency)
	for _, dep := range newDeps {
		key := dep.Group + ":" + dep.Name
		newMap[key] = dep
	}
	for _, dep := range oldDeps {
		key := dep.Group + ":" + dep.Name
		if _, exists := newMap[key]; !exists {
			result = append(result, dep)
		}
	}
	return result
}

type DependencyChange struct {
	Dependency  Dependency
	PrevVersion string
}

func findChangedDeps(oldDeps, newDeps []Dependency) []DependencyChange {
	result := []DependencyChange{}
	oldMap := make(map[string]Dependency)
	for _, dep := range oldDeps {
		key := dep.Group + ":" + dep.Name
		oldMap[key] = dep
	}
	for _, dep := range newDeps {
		key := dep.Group + ":" + dep.Name
		if oldDep, exists := oldMap[key]; exists {
			if oldDep.Version != dep.Version {
				result = append(result, DependencyChange{
					Dependency:  dep,
					PrevVersion: oldDep.Version,
				})
			}
		}
	}
	return result
}
