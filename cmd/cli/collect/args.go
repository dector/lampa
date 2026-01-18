package collect

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/dector/lampa/internal/utils"
	"github.com/dector/lampa/pkg/gradle"

	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
)

type ExecArgs struct {
	ProjectDir string
	ReportsDir string

	JsonReportFile string
	HtmlReportFile string

	Module       string
	BuildVariant string

	OverwriteReport bool

	Formats FormatArgs
}

type FormatArgs struct {
	Json bool
	Html bool
}

func (self FormatArgs) Any() bool {
	return self.Json || self.Html
}

func parseExecArgs(c *cli.Command) ExecArgs {
	args := ExecArgs{}

	args.ProjectDir = c.String(OptProjectDir)
	args.ProjectDir = utils.TryResolveFsPath(args.ProjectDir)

	args.ReportsDir = c.String(OptReportsDir)

	args.Module = c.String(OptModule)
	args.Module = strings.TrimSpace(args.Module)

	args.BuildVariant = c.String(OptBuildVariant)
	args.BuildVariant = strings.TrimSpace(args.BuildVariant)

	args.OverwriteReport = c.Bool(OptOverwriteReport)

	formats := strings.Split(c.String(OptFormat), ",")
	args.Formats.Json = lo.Contains(formats, "json")
	args.Formats.Html = lo.Contains(formats, "html")

	reportName := c.String(OptFileName)
	args.JsonReportFile = path.Join(args.ReportsDir, reportName+".json")
	args.JsonReportFile = utils.TryResolveFsPath(args.JsonReportFile)
	args.HtmlReportFile = path.Join(args.ReportsDir, reportName+".html")
	args.HtmlReportFile = utils.TryResolveFsPath(args.HtmlReportFile)

	return args
}

func validateExecArgs(args *ExecArgs) error {
	// Build variant
	if args.BuildVariant == "" {
		return fmt.Errorf("'%s' cannot be empty", OptBuildVariant)

		// TODO Wrapping is not playing well with cli/v3 package
		// return exit.Wrap(
		// 	fmt.Errorf("build variant argument is missing"),
		// 	exit.UsageError,
		// )
	}

	// Project dir
	info, err := os.Stat(args.ProjectDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("project directory `%s` does not exist: %v", args.ProjectDir, err)
		}
		return fmt.Errorf("failed to stat project directory `%s`: %v", args.ProjectDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("`%s` is not a directory", args.ProjectDir)
	}

	// Reports
	if args.Formats.Json {
		if utils.FileExists(args.JsonReportFile) {
			if args.OverwriteReport {
				if utils.IsDir(args.JsonReportFile) {
					return fmt.Errorf("report file `%s` is a directory", args.JsonReportFile)
				}
			} else {
				return fmt.Errorf("report file `%s` already exists", args.JsonReportFile)
			}
		}
	}
	if args.Formats.Html {
		if utils.FileExists(args.HtmlReportFile) {
			if args.OverwriteReport {
				if utils.IsDir(args.HtmlReportFile) {
					return fmt.Errorf("HTML report file `%s` is a directory", args.HtmlReportFile)
				}
			} else {
				return fmt.Errorf("HTML report file `%s` already exists", args.HtmlReportFile)
			}
		}
	}
	if !args.Formats.Any() {
		return fmt.Errorf("No report formats selected. Choose at least one.")
	}

	// Java
	cmd := exec.Command("java", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("java not found or not executable: %v", err)
	}

	// Gradlew
	if err := gradle.In(args.ProjectDir).EnsureExistsAndIsAFile(); err != nil {
		return err
	}

	return nil
}
