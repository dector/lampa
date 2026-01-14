package generate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dector/lampa/internal/templates/containerfile"
	"github.com/dector/lampa/internal/utils"
	"github.com/urfave/cli/v3"
)

const (
	OptToDir      = "to-dir"
	OptProjectDir = "project-dir"
)

type OciArgs struct {
	ToDir      string
	ProjectDir string
}

func CreateOciCommand() *cli.Command {
	return &cli.Command{
		Name:  "oci",
		Usage: "generate Containerfile",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  OptToDir,
				Usage: "directory where to put Containerfile",
				Value: ".",
			},
			&cli.StringFlag{
				Name:  OptProjectDir,
				Usage: "Gradle project directory to extract versions from",
			},
		},
		Action: ActionCmdOci,
	}
}

func parseOciArgs(cmd *cli.Command) OciArgs {
	args := OciArgs{}

	args.ToDir = cmd.String(OptToDir)

	args.ProjectDir = cmd.String(OptProjectDir)
	if args.ProjectDir != "" {
		args.ProjectDir = utils.TryResolveFsPath(args.ProjectDir)
	}

	return args
}

func validateOciArgs(args *OciArgs) error {
	// Only validate project-dir if provided
	if args.ProjectDir == "" {
		return nil
	}

	// Check if directory exists and is a directory
	info, err := os.Stat(args.ProjectDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("project directory `%s` does not exist", args.ProjectDir)
		}
		return fmt.Errorf("failed to access project directory `%s`: %v", args.ProjectDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("`%s` is not a directory", args.ProjectDir)
	}

	// Check if it's a Gradle project root
	hasSettingsGradle := utils.FileExists(filepath.Join(args.ProjectDir, "settings.gradle"))
	hasSettingsGradleKts := utils.FileExists(filepath.Join(args.ProjectDir, "settings.gradle.kts"))

	if !hasSettingsGradle && !hasSettingsGradleKts {
		return fmt.Errorf(
			"directory `%s` is not a Gradle project root (missing settings.gradle or settings.gradle.kts)",
			args.ProjectDir,
		)
	}

	return nil
}

func parseGradleVersion(projectDir string) (string, error) {
	wrapperPropsPath := filepath.Join(projectDir, "gradle", "wrapper", "gradle-wrapper.properties")
	if !utils.FileExists(wrapperPropsPath) {
		return "", nil
	}

	props, err := utils.ParsePropertiesFile(wrapperPropsPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse gradle-wrapper.properties: %v", err)
	}

	distributionUrl, ok := props["distributionUrl"]
	if !ok {
		return "", nil
	}

	gradleVersion, err := utils.ExtractGradleVersion(distributionUrl)
	if err != nil {
		return "", fmt.Errorf("failed to extract Gradle version: %v", err)
	}

	return gradleVersion, nil
}

func parseJavaVersion(projectDir string) (string, error) {
	// Try .tool-versions first
	version, err := utils.ParseJavaVersionFromToolVersions(projectDir)
	if err != nil {
		return "", err
	}
	if version != "" {
		return version, nil
	}

	// Try .java-version
	version, err = utils.ParseJavaVersionFromJavaVersion(projectDir)
	if err != nil {
		return "", err
	}
	return version, nil
}

func parseVersionsFromProject(projectDir string) (containerfile.Versions, error) {
	versions := containerfile.Versions{}

	// Parse Gradle version
	gradleVersion, err := parseGradleVersion(projectDir)
	if err != nil {
		return versions, err
	}
	versions.Gradle = gradleVersion

	// Parse Java version
	javaVersion, err := parseJavaVersion(projectDir)
	if err != nil {
		return versions, err
	}
	versions.Jdk = javaVersion

	// TODO: Implement remaining version parsing
	// 1. Read build.gradle or build.gradle.kts
	// 2. Parse compileSdkVersion/compileSdk for AndroidApiLevel
	// 3. Parse buildToolsVersion for AndroidBuildTools

	return versions, nil
}

func ActionCmdOci(ctx context.Context, cmd *cli.Command) error {
	// Parse arguments
	args := parseOciArgs(cmd)

	// Validate arguments
	if err := validateOciArgs(&args); err != nil {
		return err
	}

	// Ensure output directory exists
	if err := os.MkdirAll(args.ToDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", args.ToDir, err)
	}

	// Generate Containerfile content
	content, err := generateContainerfileContent(args)
	if err != nil {
		return fmt.Errorf("failed to generate Containerfile content: %v", err)
	}

	// Write to file
	filePath := filepath.Join(args.ToDir, "Containerfile")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write Containerfile: %v", err)
	}

	fmt.Printf("Containerfile created at %s\n", filePath)
	return nil
}

func generateContainerfileContent(args OciArgs) (string, error) {
	// If no project dir specified, use default behavior
	if args.ProjectDir == "" {
		return containerfile.GenerateContainerfile()
	}

	// Parse versions from the project
	parsedVersions, err := parseVersionsFromProject(args.ProjectDir)
	if err != nil {
		return "", fmt.Errorf("failed to parse versions from project: %v", err)
	}

	// Start with defaults
	opts := containerfile.NewContainerOpts()

	// Override with parsed versions (only non-empty values)
	if parsedVersions.Gradle != "" {
		opts.Versions.Gradle = parsedVersions.Gradle
	}
	if parsedVersions.Jdk != "" {
		opts.Versions.Jdk = parsedVersions.Jdk
	}
	if parsedVersions.AndroidApiLevel != "" {
		opts.Versions.AndroidApiLevel = parsedVersions.AndroidApiLevel
	}
	if parsedVersions.AndroidBuildTools != "" {
		opts.Versions.AndroidBuildTools = parsedVersions.AndroidBuildTools
	}
	if parsedVersions.AndroidCmdlineTools != "" {
		opts.Versions.AndroidCmdlineTools = parsedVersions.AndroidCmdlineTools
	}

	return containerfile.GenerateContainerfileWithOpts(opts)
}
