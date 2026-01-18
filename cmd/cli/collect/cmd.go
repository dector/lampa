package collect

import (
	"context"
	"fmt"

	"github.com/dector/lampa/internal/out"
	"github.com/dector/lampa/internal/report"
	"github.com/dector/lampa/internal/spinner"
	"github.com/dector/lampa/internal/utils"
	androidproject "github.com/dector/lampa/pkg/android-project"
	"github.com/dector/lampa/pkg/gradle"

	"github.com/urfave/cli/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	OptProjectDir   = "project"
	OptReportsDir   = "to-dir"
	OptModule       = "module"
	OptBuildVariant = "variant"
	OptFormat       = "format"

	OptOverwriteReport = "overwrite"
	OptFileName        = "file-name"
)

func CreateCliCommand() *cli.Command {
	return &cli.Command{
		Name:  "collect",
		Usage: "generate project report",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  OptProjectDir,
				Usage: "project directory root",
				Value: ".",
			},
			&cli.StringFlag{
				Name:  OptReportsDir,
				Usage: "directory where to put report",
				Value: ".",
			},
			&cli.StringFlag{
				Name:  OptModule,
				Usage: "gradle module to use for tasks",
				Value: "app",
			},
			&cli.StringFlag{
				Name:  OptBuildVariant,
				Usage: "build variant to use",
				Value: "release",
			},
			&cli.StringFlag{
				Name:  OptFileName,
				Usage: "report file name (without extension)",
				Value: "report.lampa",
			},
			&cli.StringFlag{
				Name:  OptFormat,
				Usage: "report formats to produce delimited with ',' (json,html)",
				Value: "json",
			},

			&cli.BoolFlag{
				Name:  OptOverwriteReport,
				Usage: "allow overwriting report file if it exists",
			},
		},
		Action: CmdActionCollect,
	}
}

func CmdActionCollect(ctx context.Context, cmd *cli.Command) error {
	args := parseExecArgs(cmd)
	err := validateExecArgs(&args)
	if err != nil {
		return err
	}

	return execute(args)
}

func execute(args ExecArgs) error {
	// Print run info
	fmt.Printf("Project directory: %s\n", args.ProjectDir)
	// fmt.Printf("Report directory: %s\n", to)
	fmt.Printf("Report file: %s\n", args.JsonReportFile)
	if args.Formats.Html {
		fmt.Printf("HTML report file: %s\n", args.HtmlReportFile)
	}
	fmt.Println()

	// Print warnings
	hasWarningSection := false
	if args.OverwriteReport {
		if args.Formats.Json {
			if utils.FileExists(args.JsonReportFile) {
				hasWarningSection = true
				out.PrintlnWarn("Existing report file will be overwritten")
			}
		}
		if args.Formats.Html {
			if utils.FileExists(args.HtmlReportFile) {
				hasWarningSection = true
				out.PrintlnWarn("Existing HTML report file will be overwritten")
			}
		}
	}
	if hasWarningSection {
		fmt.Println()
	}

	var err error

	err = StepBuild(args)
	if err != nil {
		return err
	}

	err = StepReport(args)
	if err != nil {
		return err
	}

	return nil
}

func StepBuild(args ExecArgs) error {
	_, err := spinner.Create(
		spinner.SpinnerArgs{
			Msg:             "Building...",
			MsgAfterSuccess: "Building: Done.",
			MsgAfterFail:    "Building: Failed.",
		}, func() (string, error) {
			out.Info("Building AAB")

			taskVariant := cases.Title(language.BritishEnglish).String(args.BuildVariant)
			task := ":" + args.Module + ":bundle" + taskVariant
			output, err := gradle.
				In(args.ProjectDir).
				Execute(task)
			if err != nil {
				return "", fmt.Errorf("failed to build app: %v\nOutput:\n%s", err, output)
			}

			return "", nil
		})
	if err != nil {
		return err
	}

	return nil
}

func StepReport(args ExecArgs) error {
	pathToAab, err := androidproject.
		NewAndroidProject(args.ProjectDir).
		FindAabFile(args.Module, args.BuildVariant)
	if err != nil {
		return err
	}

	report, err := spinner.Create(
		spinner.SpinnerArgs{
			Msg:             "Generating report...",
			MsgAfterSuccess: "Generating report: Done.",
			MsgAfterFail:    "Generating report: Failed.",
		}, func() (report.StatsReport, error) {
			return report.ParseFrom(report.ParseFromArgs{
				PathToAab:    pathToAab,
				Module:       args.Module,
				BuildVariant: args.BuildVariant,
				ProjectDir:   args.ProjectDir,
			})
		})
	if err != nil {
		return err
	}

	err = exportReports(report, args)
	if err != nil {
		return err
	}

	return nil
}

func exportReports(report *report.StatsReport, args ExecArgs) error {
	var err error

	// Json Report
	if args.Formats.Json {
		err = report.WriteToFile(args.JsonReportFile)
		if err != nil {
			return err
		}
		fmt.Printf("\nReport written to %s\n", args.JsonReportFile)
	}

	// Html Report
	if args.Formats.Html {
		err = WriteHtmlReportToFile(report, args)
		if err != nil {
			return err
		}
		fmt.Printf("Report written to %s\n", args.HtmlReportFile)
	}

	return nil
}
