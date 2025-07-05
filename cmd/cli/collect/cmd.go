package collect

import (
	"archive/zip"
	"context"
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"lampa/internal"
	"lampa/internal/out"
	"lampa/internal/report"
	pages "lampa/internal/templates/html"
	"lampa/internal/utils"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	. "lampa/internal/globals"
)

const (
	EnvAndroidSdkRoot = "ANDROID_SDK_ROOT"
	EnvBundletoolJar  = "BUNDLETOOL_JAR"
)

const (
	OptProjectDir   = "project"
	OptReportsDir   = "to-dir"
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

func parseExecArgs(c *cli.Command) ExecArgs {
	args := ExecArgs{}

	args.ProjectDir = c.String(OptProjectDir)
	args.ProjectDir = utils.TryResolveFsPath(args.ProjectDir)

	args.ReportsDir = c.String(OptReportsDir)

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

	args.GradlewPath = path.Join(args.ProjectDir, "gradlew")

	args.AndroidSdkPath = utils.TryResolveFsPath(os.Getenv(EnvAndroidSdkRoot))
	args.BundletoolPath = utils.TryResolveFsPath(os.Getenv(EnvBundletoolJar))

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

	// Bundletool
	if args.BundletoolPath == "" {
		return fmt.Errorf("%s environment variable is not set", EnvBundletoolJar)
	}
	if !utils.FileExists(args.BundletoolPath) {
		return fmt.Errorf("bundletool jar file `%s` does not exist", args.BundletoolPath)
	}
	if utils.IsDir(args.BundletoolPath) {
		return fmt.Errorf("bundletool jar file `%s` is a directory", args.BundletoolPath)
	}

	// Aapt
	if args.AndroidSdkPath == "" {
		return fmt.Errorf("%s environment variable is not set", EnvAndroidSdkRoot)
	}
	if !utils.FileExists(args.AndroidSdkPath) {
		return fmt.Errorf("Android SDK path `%s` does not exist", args.AndroidSdkPath)
	}
	if !utils.IsDir(args.AndroidSdkPath) {
		return fmt.Errorf("Android SDK path `%s` is not a directory", args.AndroidSdkPath)
	}
	args.AaptPath, err = findAaptExecutable(args.AndroidSdkPath)
	if err != nil {
		return err
	}

	// Gradlew
	info, err = os.Stat(args.GradlewPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist", args.GradlewPath)
		} else {
			return fmt.Errorf("could not stat %s: %v", args.GradlewPath, err)
		}
	}
	if info.IsDir() {
		return fmt.Errorf("%s exists but is a directory, not a file", args.GradlewPath)
	}

	return nil
}

type FormatArgs struct {
	Json bool
	Html bool
}

func (self FormatArgs) Any() bool {
	return self.Json || self.Html
}

type ExecArgs struct {
	ProjectDir string
	ReportsDir string

	JsonReportFile string
	HtmlReportFile string

	BuildVariant string

	OverwriteReport bool

	Formats FormatArgs

	BundletoolPath string
	AndroidSdkPath string
	AaptPath       string
	GradlewPath    string
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

	_, err := DynamicSpinner(SpinnerArgs{
		Msg:             "Building...",
		MsgAfterSuccess: "Building: Done.",
		MsgAfterFail:    "Building: Failed.",
	}, func() (string, error) {
		task := "bundle" + cases.Title(language.BritishEnglish).String(args.BuildVariant)
		output, err := executeGradleTask(args, task)
		if err != nil {
			return "", fmt.Errorf("failed to build app: %v\nOutput:\n%s", err, string(output))
		}

		return "", nil
	})
	if err != nil {
		return err
	}

	// pathToApk, err := DynamicSpinner(SpinnerArgs{
	// 	Msg:             "Building...",
	// 	MsgAfterSuccess: "Building: Done.",
	// 	MsgAfterFail:    "Building: Failed.",
	// }, func() (string, error) {
	// 	task := "assemble" + cases.Title(language.BritishEnglish).String(buildVariant)
	// 	cmd := exec.Command(gradlewPath, "--no-daemon", "--console", "plain", "-q", task)
	// 	cmd.Dir = from
	// 	output, err := cmd.CombinedOutput()
	// 	if err != nil {
	// 		return "", fmt.Errorf("failed to build app: %v\nOutput:\n%s", err, string(output))
	// 	}

	// 	var variantPaths []string
	// 	var wordStart int
	// 	for i, r := range buildVariant {
	// 		if i > 0 && unicode.IsUpper(r) {
	// 			variantPaths = append(variantPaths, strings.ToLower(buildVariant[wordStart:i]))
	// 			wordStart = i
	// 		}
	// 	}
	// 	variantPaths = append(variantPaths, strings.ToLower(buildVariant[wordStart:]))

	// 	apkDir := path.Join(append([]string{from, "app", "build", "outputs", "apk"}, variantPaths...)...)

	// 	info, err := os.Stat(apkDir)
	// 	if err != nil {
	// 		if os.IsNotExist(err) {
	// 			return "", fmt.Errorf("APK directory `%s` does not exist", apkDir)
	// 		}
	// 		return "", fmt.Errorf("error accessing APK directory `%s`: %v", apkDir, err)
	// 	}
	// 	if !info.IsDir() {
	// 		return "", fmt.Errorf("APK directory `%s` is not a directory", apkDir)
	// 	}

	// 	files, err := os.ReadDir(apkDir)
	// 	if err != nil {
	// 		return "", fmt.Errorf("could not read APK directory `%s`: %v", apkDir, err)
	// 	}
	// 	var apkFilePath string
	// 	for _, file := range files {
	// 		if !file.IsDir() && strings.HasSuffix(file.Name(), ".apk") {
	// 			apkFilePath = path.Join(apkDir, file.Name())
	// 			break
	// 		}
	// 	}
	// 	if apkFilePath == "" {
	// 		return "", fmt.Errorf("no APK file found in `%s`", apkDir)
	// 	}

	// 	return apkFilePath, nil
	// })
	// if err != nil {
	// 	return err
	// }

	err = StepReport(args)
	if err != nil {
		return err
	}

	return nil
}

func StepReport(args ExecArgs) error {
	pathToAab, err := findAabFile(args)
	if err != nil {
		return err
	}
	report, err := DynamicSpinner(
		SpinnerArgs{
			Msg:             "Generating report...",
			MsgAfterSuccess: "Generating report: Done.",
			MsgAfterFail:    "Generating report: Failed.",
		}, func() (report.Report, error) {
			return collectReport(args, pathToAab)
			// return collectReport(CollectReportArgs{
			// 	ProjectDir:   args.ProjectDir,
			// 	ReportDir:    args.ReportsDir,
			// 	BuildVariant: args.BuildVariant,

			// 	PathToBundletool: args.BundletoolPath,
			// 	PathToAapt:       args.ApptPath,
			// 	PathToAab:        pathToAab,
			// 	// PathToApk:        *pathToApk,

			// })
		})
	if err != nil {
		return err
	}

	// Json Report
	if args.Formats.Json {
		err = WriteJsonReportToFile(report, args)
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

func WriteJsonReportToFile(report *report.Report, args ExecArgs) error {
	file, err := os.Create(args.JsonReportFile)
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

	return nil
}

func WriteHtmlReportToFile(report *report.Report, args ExecArgs) error {
	reportHtml, err := GenerateHtmlReport(report)
	if err != nil {
		return fmt.Errorf("could not generate HTML report: %v", err)
	}
	file, err := os.Create(args.HtmlReportFile)
	if err != nil {
		return fmt.Errorf("could not create HTML report file: %v", err)
	}
	defer file.Close()

	if _, err := file.Write([]byte(reportHtml)); err != nil {
		return fmt.Errorf("could not write HTML report: %v", err)
	}

	return nil
}

func GenerateHtmlReport(r *report.Report) (string, error) {
	w := &strings.Builder{}
	err := pages.CollectHtml(r).Render(context.Background(), w)
	if err != nil {
		return "", err
	}
	return w.String(), nil
}

func collectReport(args ExecArgs, pathToAab string) (report.Report, error) {
	result := report.Report{
		Version: "stats/0.0.1",
	}

	context, err := parseContext(args)
	if err != nil {
		return report.Report{}, err
	}
	result.Context = context

	configurationName := args.BuildVariant + "CompileClasspath"

	err = analyzeBuild(&result, args, pathToAab)
	if err != nil {
		return report.Report{}, err
	}

	output, err := executeGradleTask(args, "app:dependencies", "--configuration", configurationName)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to execute gradlew: %v\nOutput:\n%s", err, string(output))
	}

	// fmt.Println(string(output))

	tree, err := internal.ParseTreeFromOutput(string(output), configurationName)
	if err != nil {
		return report.Report{}, fmt.Errorf("failed to parse tree: %v", err)
	}

	for _, info := range tree.Summary {
		d := report.CoordinatedDependency{
			Group:   info.GroupID,
			Name:    info.ArtifactID,
			Version: info.Version,
		}
		result.Build.Dependencies.Compile = append(result.Build.Dependencies.Compile, d)
	}
	slices.SortFunc(result.Build.Dependencies.Compile, func(a, b report.CoordinatedDependency) int {
		if a.Group > b.Group {
			return 1
		} else if a.Group < b.Group {
			return -1
		} else if a.Name > b.Name {
			return 1
		} else if a.Name < b.Name {
			return -1
		} else if a.Version > b.Version {
			return 1
		} else if a.Version < b.Version {
			return -1
		} else {
			return 0
		}
	})

	return result, nil
}

func parseContext(args ExecArgs) (report.ContextSegment, error) {
	result := report.ContextSegment{
		Tool: report.ToolSegment{
			Name:        "Lampa",
			Website:     "https://github.com/dector/lampa",
			Sources:     "https://github.com/dector/lampa",
			Version:     G.Version,
			BuildCommit: G.BuildCommit,
		},
		GenerationTime: time.Now().UTC().Format(time.RFC3339),
	}

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

type SpinnerArgs struct {
	Msg             string
	MsgAfterSuccess string
	MsgAfterFail    string
}

func DynamicSpinner[T any](args SpinnerArgs, action func() (T, error)) (*T, error) {
	blue := color.New(color.FgBlue).SprintfFunc()
	green := color.New(color.FgGreen).SprintfFunc()
	red := color.New(color.FgRed).SprintfFunc()
	cs := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	s := spinner.New(cs, 100*time.Millisecond)
	s.Color("blue")
	s.Suffix = blue(" " + args.Msg)
	s.FinalMSG = green("✔ " + args.MsgAfterSuccess + "\n")
	s.Start()
	defer s.Stop()

	data, err := action()
	if err != nil {
		s.FinalMSG = red("✗ " + args.MsgAfterFail + "\n")
		return nil, err
	}

	return &data, err
}

func analyzeBuild(result *report.Report, args ExecArgs, pathToAab string) error {
	result.Build.BuildVariant = args.BuildVariant
	result.Build.AabName = filepath.Base(pathToAab)
	// result.Build.ApkName = filepath.Base(args.PathToApk)

	// file, err := os.Open(args.PathToApk)
	// if err == nil {
	// 	defer file.Close()
	// 	hasher := sha1.New()
	// 	if _, err := io.Copy(hasher, file); err == nil {
	// 		result.Build.ApkSha1 = fmt.Sprintf("%x", hasher.Sum(nil))
	// 	}
	// } else {
	// 	return err
	// }

	// info, err := os.Stat(args.PathToApk)
	// if err == nil {
	// 	result.Build.ApkSize = strconv.FormatInt(info.Size(), 10)
	// } else {
	// 	return fmt.Errorf("could not stat APK file: %v", err)
	// }

	fileAab, err := os.Open(pathToAab)
	if err == nil {
		defer fileAab.Close()
		hasher := sha1.New()
		if _, err := io.Copy(hasher, fileAab); err == nil {
			result.Build.AabSha1 = fmt.Sprintf("%x", hasher.Sum(nil))
		}
	} else {
		return err
	}

	infoAab, err := os.Stat(pathToAab)
	if err != nil {
		return fmt.Errorf("could not stat AAB file: %v", err)
	}
	result.Build.AabSize = strconv.FormatInt(infoAab.Size(), 10)

	// TODO analyze AAB manifest
	// Display AAB size range
	// Generate APK
	// Get other data from APK

	// Analyze AAB manifest using bundletool
	cmd := exec.Command("java", "-jar", args.BundletoolPath, "dump", "manifest", "--bundle", pathToAab)
	cmd.Dir = args.ProjectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to analyze AAB manifest with bundletool: %v.\nReason: %s", err, string(output))
	}
	manifest := string(output)

	// Parse manifest as XML
	type Manifest struct {
		XMLName          struct{} `xml:"manifest"`
		Package          string   `xml:"package,attr"`
		VersionCode      string   `xml:"versionCode,attr"`
		VersionName      string   `xml:"versionName,attr"`
		BuildVersionCode string   `xml:"platformBuildVersionCode,attr"`
		BuildVersionName string   `xml:"platformBuildVersionName,attr"`
		Application      struct {
			Label string `xml:"label,attr"`
		} `xml:"application"`
		UsesSdk struct {
			MinSdkVersion    string `xml:"minSdkVersion,attr"`
			TargetSdkVersion string `xml:"targetSdkVersion,attr"`
		} `xml:"uses-sdk"`
	}

	var manifestData Manifest
	if err := xml.Unmarshal([]byte(manifest), &manifestData); err != nil {
		return fmt.Errorf("could not parse manifest XML: %v", err)
	}
	result.Build.ApplicationId = manifestData.Package
	result.Build.VersionCode = manifestData.VersionCode
	result.Build.VersionName = manifestData.VersionName
	// result.Build.AppName = manifestData.Application.Label
	result.Build.MinSdkVersion = manifestData.UsesSdk.MinSdkVersion
	result.Build.TargetSdkVersion = manifestData.UsesSdk.TargetSdkVersion
	result.Build.CompileSdkVersion = manifestData.BuildVersionCode

	err = addDataFromApk(result, args, pathToAab)
	if err != nil {
		return err
	}
	// cmd := exec.Command(args.PathToAapt, "dump", "badging", args.PathToApk)
	// cmd.Dir = args.ProjectDir

	// output, err := cmd.CombinedOutput()
	// if err != nil {
	// 	return fmt.Errorf("failed to analyze build: %v.\nReason: %s", err, string(output))
	// }

	// lines := strings.Split(string(output), "\n")
	// props := make(map[string]string)
	// for _, l := range lines {
	// 	if idx := strings.Index(l, ":"); idx != -1 {
	// 		key := strings.TrimSpace(l[:idx])
	// 		val := strings.TrimSpace(l[idx+1:])
	// 		props[key] = val
	// 	}
	// }

	// for k, v := range props {
	// 	switch k {
	// 	case "minSdkVersion":
	// 		result.Build.MinSdkVersion = strings.Trim(v, "'")
	// 	case "targetSdkVersion":
	// 		result.Build.TargetSdkVersion = strings.Trim(v, "'")
	// 	case "application-label":
	// 		result.Build.AppName = strings.Trim(v, "'")
	// 	case "package":
	// 		{
	// 			parts := strings.Split(v, " ")
	// 			for _, part := range parts {
	// 				if kv := strings.SplitN(part, "=", 2); len(kv) == 2 {
	// 					key := strings.TrimSpace(kv[0])
	// 					val := strings.Trim(strings.TrimSpace(kv[1]), "'")
	// 					switch key {
	// 					case "name":
	// 						result.Build.ApplicationId = val
	// 					case "versionCode":
	// 						result.Build.VersionCode = val
	// 					case "versionName":
	// 						result.Build.VersionName = val
	// 					case "compileSdkVersion":
	// 						result.Build.CompileSdkVersion = val
	// 					}
	// 				}
	// 			}
	// 		}
	// 	case "locales":
	// 		{
	// 			result.Build.Locales = lo.Map(strings.Fields(v), func(locale string, _ int) string {
	// 				return strings.Trim(locale, "'")
	// 			})
	// 		}
	// 	}
	// }

	return nil
}

func addDataFromApk(result *report.Report, args ExecArgs, pathToAab string) error {
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("lampa-%x", sha1.Sum([]byte(args.ProjectDir))))
	if err != nil {
		return fmt.Errorf("failed to create temp dir for universal APK: %w", err)
	}
	defer os.RemoveAll(tempDir)

	universalApkPath := filepath.Join(tempDir, "universal.apk")

	// Use bundletool to build the universal APK from the AAB
	cmd := exec.Command(
		"java", "-jar", args.BundletoolPath, "build-apks",
		"--bundle", pathToAab,
		"--output", universalApkPath+".apks",
		"--mode", "universal",
		"--overwrite",
	)
	cmd.Dir = args.ProjectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to build universal APK with bundletool: %v\nOutput:\n%s", err, string(output))
	}

	// Extract universal.apk from the .apks file (which is a zip)
	apksFile, err := os.Open(universalApkPath + ".apks")
	if err != nil {
		return fmt.Errorf("failed to open .apks file: %w", err)
	}
	defer apksFile.Close()

	stat, err := apksFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat .apks file: %w", err)
	}

	zipReader, err := zip.NewReader(apksFile, stat.Size())
	if err != nil {
		return fmt.Errorf("failed to read .apks as zip: %w", err)
	}

	found := false
	for _, f := range zipReader.File {
		if f.Name == "universal.apk" {
			outFile, err := os.Create(universalApkPath)
			if err != nil {
				return fmt.Errorf("failed to create universal.apk: %w", err)
			}
			rc, err := f.Open()
			if err != nil {
				outFile.Close()
				return fmt.Errorf("failed to open universal.apk in zip: %w", err)
			}
			_, err = io.Copy(outFile, rc)
			rc.Close()
			outFile.Close()
			if err != nil {
				return fmt.Errorf("failed to extract universal.apk: %w", err)
			}
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("universal.apk not found in .apks file")
	}

	// Use aapt2 to extract the application label (app name) from the APK
	cmdAapt := exec.Command(args.AaptPath, "dump", "badging", universalApkPath)
	cmdAapt.Dir = args.ProjectDir
	outputAapt, err := cmdAapt.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run aapt2 on universal.apk: %v\nOutput:\n%s", err, string(outputAapt))
	}

	// Parse the output to find the application-label
	lines := strings.Split(string(outputAapt), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "application-label:") {
			// Format: application-label:'App Name'
			idx := strings.Index(line, ":")
			if idx != -1 {
				label := strings.Trim(line[idx+1:], "'")
				result.Build.AppName = label
				break
			}
		}
	}

	return nil
}

func findAaptExecutable(sdkRoot string) (string, error) {
	aaptPath := filepath.Join(sdkRoot, "build-tools")
	entries, err := os.ReadDir(aaptPath)

	if err == nil {
		// Find the latest build-tools version
		var latest string
		for _, entry := range entries {
			if entry.IsDir() {
				// TODO improve
				if latest == "" || entry.Name() > latest {
					latest = entry.Name()
				}
			}
		}
		if latest != "" {
			aaptFullPath := filepath.Join(aaptPath, latest, "aapt2")
			if _, err := os.Stat(aaptFullPath); err == nil {
				return aaptFullPath, nil
			}
		}
	}

	return "", fmt.Errorf("aapt executable not found in %s", sdkRoot)
}

func findAabFile(args ExecArgs) (string, error) {
	bundleDir := path.Join(args.ProjectDir, "app", "build", "outputs", "bundle", args.BuildVariant)

	info, err := os.Stat(bundleDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("AAB directory `%s` does not exist", bundleDir)
		}
		return "", fmt.Errorf("error accessing AAB directory `%s`: %v", bundleDir, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("AAB directory `%s` is not a directory", bundleDir)
	}

	files, err := os.ReadDir(bundleDir)
	if err != nil {
		return "", fmt.Errorf("could not read AAB directory `%s`: %v", bundleDir, err)
	}
	var aabFilePath string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".aab") {
			aabFilePath = path.Join(bundleDir, file.Name())
			break
		}
	}
	if aabFilePath == "" {
		return "", fmt.Errorf("no AAB file found in `%s`", bundleDir)
	}

	return aabFilePath, nil
}

func executeGradleTask(args ExecArgs, gradleArgs ...string) ([]byte, error) {
	cmd := exec.Command(
		args.GradlewPath,
		append(
			[]string{
				"--no-daemon", "--console",
				"plain", "-q",
			},
			gradleArgs...,
		)...,
	)
	cmd.Dir = args.ProjectDir
	return cmd.CombinedOutput()
}
