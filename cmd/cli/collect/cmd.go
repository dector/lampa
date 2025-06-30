package collect

import (
	"context"
	"encoding/json"
	"fmt"
	"lampa/internal"
	"lampa/internal/out"
	"lampa/internal/report"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"

	. "lampa/internal/globals"
)

func CmdActionCollect(ctx context.Context, cmd *cli.Command) error {
	buildVariant := cmd.String("variant")
	if buildVariant == "" {
		buildVariant = "release"
	}

	from, err := decodeProjectPath(cmd)
	if err != nil {
		return err
	}

	to, err := decodeTargetPath(cmd)
	if err != nil {
		return err
	}

	fWithName := cmd.String("with-name")
	if fWithName == "" {
		fWithName = "lampa"
	}
	reportFile := path.Join(to, fWithName+".report.json")

	fmt.Printf("Project directory: %s\n", from)
	// fmt.Printf("Report directory: %s\n", to)
	fmt.Printf("Report file: %s\n", reportFile)

	if cmd.Bool("rewrite-report") {
		out.PrintlnWarn("Existing report file will be overwritten (if it exists)")
	} else {
		if _, err := os.Stat(reportFile); err == nil {
			return fmt.Errorf("report file `%s` already exists", reportFile)
		}
	}

	gradlewPath := path.Join(from, "gradlew")
	info, err := os.Stat(gradlewPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist", gradlewPath)
		} else {
			return fmt.Errorf("could not stat %s: %v", gradlewPath, err)
		}
	}
	if info.IsDir() {
		return fmt.Errorf("%s exists but is a directory, not a file", gradlewPath)
	}

	blue := color.New(color.FgBlue).SprintfFunc()
	green := color.New(color.FgGreen).SprintfFunc()
	red := color.New(color.FgRed).SprintfFunc()
	cs := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	s := spinner.New(cs, 100*time.Millisecond)
	s.Color("blue")
	s.Suffix = blue(" Capturing report...")
	s.FinalMSG = green("✔ Capturing report: Done.\n")
	s.Start()
	defer s.Stop()

	report, err := collectReport(CollectReportArgs{
		ProjectDir:   from,
		ReportDir:    to,
		BuildVariant: buildVariant,
	})
	if err != nil {
		s.FinalMSG = red("✗ Capturing report: Failed.\n")
		return err
	}

	file, err := os.Create(reportFile)
	if err != nil {
		return fmt.Errorf("could not create report file: %v", err)
	}
	defer file.Close()

	reportJson, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal report: %v", err)
	}

	if _, err := file.Write(reportJson); err != nil {
		return fmt.Errorf("could not write report: %v", err)
	}
	fmt.Printf("Report written to %s\n", reportFile)

	return nil
}

func collectReport(args CollectReportArgs) (report.Report, error) {
	result := report.Report{
		Version: "0.0.1-SNAPSHOT",
		Type:    "CollectionReport",
		Tool: report.ToolSegment{
			Name:        "lampa",
			Repository:  "https://github.com/dector/lampa/",
			Version:     G.Version,
			BuildCommit: G.BuildCommit,
		},
	}

	context, err := parseContext(args)
	if err != nil {
		return report.Report{}, err
	}
	result.Context = context

	configurationName := args.BuildVariant + "CompileClasspath"

	gradlewPath := path.Join(args.ProjectDir, "gradlew")
	cmd := exec.Command(gradlewPath, "--no-daemon", "--console", "plain", "-q", "app:dependencies", "--configuration", configurationName)
	cmd.Dir = args.ProjectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to execute gradlew: %v\nOutput:\n%s", err, string(output))
	}

	// fmt.Println(string(output))

	tree, err := internal.ParseTreeFromOutput(string(output), configurationName)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to parse tree: %v", err)
	}

	result.Dependencies = report.DependenciesSegment{
		Compiled: make([]string, 0),
	}
	for _, info := range tree.Summary {
		d := info.String()
		result.Dependencies.Compiled = append(result.Dependencies.Compiled, d)
	}

	return result, nil
}

func parseContext(args CollectReportArgs) (report.ContextSegment, error) {
	result := report.ContextSegment{}

	_, err := exec.LookPath("git")
	if err != nil {
		return result, fmt.Errorf("git not found in PATH: %v", err)
	}

	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = args.ProjectDir
	if err := cmd.Run(); err != nil {
		return result, nil
	}

	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = args.ProjectDir
	output, err := cmd.Output()
	if err == nil {
		result.Git.Commit = strings.TrimSpace(string(output))
	}

	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = args.ProjectDir
	output, err = cmd.Output()
	if err == nil {
		result.Git.IsDirty = len(strings.TrimSpace(string(output))) > 0
	}

	cmd = exec.Command("git", "describe", "--tags", "--long")
	cmd.Dir = args.ProjectDir
	output, err = cmd.Output()
	if err == nil {
		parts := strings.SplitN(strings.TrimSpace(string(output)), "-", 3)
		if len(parts) == 3 {
			result.Git.Tag = parts[0]
			commitsAfterTag, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				out.PrintlnWarn("could not parse commits after tag from %q: %v", parts[1], err)
			} else {
				result.Git.CommitsAfterTag = uint(commitsAfterTag)
			}
		} else {
			log.Printf("warning: unexpected format from git describe: %q", string(output))
		}
	} else {
		log.Printf("warning: git describe failed: %v", err)
	}

	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = args.ProjectDir
	output, err = cmd.Output()
	if err == nil {
		result.Git.Branch = strings.TrimSpace(string(output))
	}

	return result, nil
}

type CollectReportArgs struct {
	ProjectDir   string
	ReportDir    string
	BuildVariant string
}

func decodeProjectPath(cmd *cli.Command) (string, error) {
	path := cmd.String("from")
	if path == "" {
		return ".", nil
	}

	inf, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("project directory `%s` does not exist", path)
		} else {
			return "", fmt.Errorf("internal error: %v", err)
		}
	}
	if !inf.IsDir() {
		return "", fmt.Errorf("error: `%s` is not a directory", path)
	}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return path, nil
	}
	return absolutePath, nil
}

func decodeTargetPath(cmd *cli.Command) (string, error) {
	path := cmd.String("to")
	if path == "" {
		return ".", nil
	}

	inf, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("target directory `%s` does not exist", path)
		} else {
			return "", fmt.Errorf("internal error: %v", err)
		}
	}
	if !inf.IsDir() {
		return "", fmt.Errorf("`%s` is not a directory", path)
	}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return path, nil
	}
	return absolutePath, nil
}
