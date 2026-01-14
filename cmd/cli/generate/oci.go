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

func parseVersionsFromProject(projectDir string) (containerfile.Versions, error) {
	// TODO: Implement version parsing from Gradle project files
	// This should:
	// 1. Read build.gradle or build.gradle.kts
	// 2. Parse compileSdkVersion/compileSdk for AndroidApiLevel
	// 3. Parse buildToolsVersion for AndroidBuildTools
	// 4. Read gradle/wrapper/gradle-wrapper.properties for Gradle version
	// 5. Parse sourceCompatibility/targetCompatibility for JDK version
	// 6. Use sensible defaults for AndroidCmdlineTools

	return containerfile.Versions{}, fmt.Errorf("version parsing not yet implemented")
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
	versions, err := parseVersionsFromProject(args.ProjectDir)
	if err != nil {
		return "", fmt.Errorf("failed to parse versions from project: %v", err)
	}

	// Create opts with parsed versions
	opts := containerfile.ContainerOpts{
		Image:    "ubuntu:24.04",
		Versions: versions,
	}

	return containerfile.GenerateContainerfileWithOpts(opts)
}
