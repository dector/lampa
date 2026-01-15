package generate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dector/kdly"
	"github.com/dector/lampa/internal/templates/containerfile"
	"github.com/dector/lampa/internal/utils"
	"github.com/dector/lampa/pkg/gradle"
	"github.com/urfave/cli/v3"
)

const (
	OptToDir       = "to-dir"
	OptProjectDir  = "project-dir"
	OptFromConfig  = "from-config"
)

const (
	androidCompileSdkInitScriptPath = "/tmp/lampa-printCompileSdk.init.gradle"
	androidCompileSdkTaskName       = "lampaCompileSdk"
)

const androidCompileSdkInitScript = `allprojects {
    plugins.withId("com.android.application") {
        afterEvaluate {
            def android = extensions.findByName("android")
            if (android != null && android.compileSdk != null) {
                tasks.register("lampaCompileSdk") {
                    doLast {
                        println("compileSdk=${android.compileSdk}")
                    }
                }
            }
        }
    }
}
`

type OciArgs struct {
	ToDir      string
	ProjectDir string
	FromConfig string
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
			&cli.StringFlag{
				Name:  OptFromConfig,
				Usage: "path to KDL config file with version information",
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

	args.FromConfig = cmd.String(OptFromConfig)
	if args.FromConfig != "" {
		args.FromConfig = utils.TryResolveFsPath(args.FromConfig)
	}

	return args
}

func validateOciArgs(args *OciArgs) error {
	// Validate project-dir if provided
	if args.ProjectDir != "" {
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
	}

	// Validate from-config if provided
	if args.FromConfig != "" {
		// Check if file exists
		if !utils.FileExists(args.FromConfig) {
			return fmt.Errorf("config file `%s` does not exist", args.FromConfig)
		}

		// Check if it's a file (not a directory)
		info, err := os.Stat(args.FromConfig)
		if err != nil {
			return fmt.Errorf("failed to access config file `%s`: %v", args.FromConfig, err)
		}
		if info.IsDir() {
			return fmt.Errorf("`%s` is a directory, not a file", args.FromConfig)
		}
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
	fmt.Println("Looking for Java version...")

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

func ensureAndroidCompileSdkInitScript() error {
	// Check if file already exists
	if utils.FileExists(androidCompileSdkInitScriptPath) {
		return nil
	}

	// Create the init script file
	err := os.WriteFile(androidCompileSdkInitScriptPath, []byte(androidCompileSdkInitScript), 0644)
	if err != nil {
		return fmt.Errorf("failed to create Android compile SDK init script: %v", err)
	}

	return nil
}

func extractCompileSdkFromOutput(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "compileSdk=") {
			version := strings.TrimPrefix(line, "compileSdk=")
			return strings.TrimSpace(version)
		}
	}
	return ""
}

func parseAndroidCompileSdk(projectDir string) (string, error) {
	fmt.Println("Looking for compileSdk version...")

	// Check if gradlew exists
	gradlewPath := filepath.Join(projectDir, "gradlew")
	if !utils.FileExists(gradlewPath) {
		return "", nil // Not an error, just no gradlew
	}

	// Ensure init script exists
	if err := ensureAndroidCompileSdkInitScript(); err != nil {
		// Log warning but continue - not critical
		return "", nil
	}

	// Execute Gradle command
	g := gradle.In(projectDir)
	output, err := g.Execute(androidCompileSdkTaskName, "-q", "-I", androidCompileSdkInitScriptPath)
	if err != nil {
		// Gradle execution failed - likely not an Android project
		return "", nil
	}

	// Parse output
	version := extractCompileSdkFromOutput(string(output))
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

	// Parse Android compile SDK
	androidCompileSdk, err := parseAndroidCompileSdk(projectDir)
	if err != nil {
		return versions, err
	}
	versions.AndroidApiLevel = androidCompileSdk

	// TODO: Implement remaining version parsing
	// 1. Parse buildToolsVersion for AndroidBuildTools

	return versions, nil
}

// parseVersionsFromConfig parses version information from a KDL config file.
// Expected structure:
//
//	generator {
//	  oci {
//	    version jdk="21" androidApi="36" androidBuildTools="36.1.0" gradle="9.2.1" androidCmdlineTools="13114758"
//	  }
//	}
func parseVersionsFromConfig(configPath string) (containerfile.Versions, error) {
	versions := containerfile.Versions{}

	// Read config file
	content, err := os.ReadFile(configPath)
	if err != nil {
		return versions, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse KDL document
	doc, err := kdly.Parse(string(content))
	if err != nil {
		return versions, fmt.Errorf("failed to parse KDL config: %v", err)
	}

	// Navigate to generator.oci.version node
	versionNode, err := findVersionNode(doc)
	if err != nil {
		// Return empty versions if structure not found (all properties optional)
		// Don't treat as error - just means no versions specified
		return versions, nil
	}

	// Extract version properties
	versions = extractVersionProperties(versionNode)

	return versions, nil
}

// findVersionNode navigates the KDL document tree to find the version node.
// Path: generator -> oci -> version
func findVersionNode(doc *kdly.Document) (*kdly.Node, error) {
	// Find "generator" node
	var generatorNode *kdly.Node
	for i := range doc.Nodes {
		if doc.Nodes[i].Name == "generator" {
			generatorNode = &doc.Nodes[i]
			break
		}
	}
	if generatorNode == nil {
		return nil, fmt.Errorf("'generator' node not found")
	}

	// Find "oci" child node
	var ociNode *kdly.Node
	for i := range generatorNode.Children {
		if generatorNode.Children[i].Name == "oci" {
			ociNode = &generatorNode.Children[i]
			break
		}
	}
	if ociNode == nil {
		return nil, fmt.Errorf("'oci' node not found under 'generator'")
	}

	// Find "version" child node
	var versionNode *kdly.Node
	for i := range ociNode.Children {
		if ociNode.Children[i].Name == "version" {
			versionNode = &ociNode.Children[i]
			break
		}
	}
	if versionNode == nil {
		return nil, fmt.Errorf("'version' node not found under 'generator.oci'")
	}

	return versionNode, nil
}

// extractVersionProperties extracts version information from the version node properties.
func extractVersionProperties(node *kdly.Node) containerfile.Versions {
	versions := containerfile.Versions{}

	// Iterate through properties
	for _, prop := range node.Properties {
		// Extract string value from property (Value.Value is the actual string)
		strValue := prop.Value.Value

		switch prop.Key {
		case "jdk":
			versions.Jdk = strValue
		case "androidApi":
			versions.AndroidApiLevel = strValue
		case "androidBuildTools":
			versions.AndroidBuildTools = strValue
		case "gradle":
			versions.Gradle = strValue
		case "androidCmdlineTools":
			versions.AndroidCmdlineTools = strValue
		}
	}

	return versions
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

// mergeVersions merges source versions into target, only overriding non-empty values.
func mergeVersions(target *containerfile.Versions, source containerfile.Versions) {
	if source.Gradle != "" {
		target.Gradle = source.Gradle
	}
	if source.Jdk != "" {
		target.Jdk = source.Jdk
	}
	if source.AndroidApiLevel != "" {
		target.AndroidApiLevel = source.AndroidApiLevel
	}
	if source.AndroidBuildTools != "" {
		target.AndroidBuildTools = source.AndroidBuildTools
	}
	if source.AndroidCmdlineTools != "" {
		target.AndroidCmdlineTools = source.AndroidCmdlineTools
	}
}

func generateContainerfileContent(args OciArgs) (string, error) {
	// If no project dir and no config specified, use default behavior
	if args.ProjectDir == "" && args.FromConfig == "" {
		return containerfile.GenerateContainerfile()
	}

	// Start with defaults (lowest priority)
	opts := containerfile.NewContainerOpts()

	// Parse and merge project-detected versions (middle priority)
	if args.ProjectDir != "" {
		projectVersions, err := parseVersionsFromProject(args.ProjectDir)
		if err != nil {
			return "", fmt.Errorf("failed to parse versions from project: %v", err)
		}
		mergeVersions(&opts.Versions, projectVersions)
	}

	// Parse and merge config file versions (highest priority)
	if args.FromConfig != "" {
		configVersions, err := parseVersionsFromConfig(args.FromConfig)
		if err != nil {
			return "", fmt.Errorf("failed to parse versions from config: %v", err)
		}
		mergeVersions(&opts.Versions, configVersions)
	}

	return containerfile.GenerateContainerfileWithOpts(opts)
}
