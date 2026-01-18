// Package androidproject provides utilities for working with Android Gradle projects.
// It helps locate build artifacts like AAB files following standard Gradle project conventions.
package androidproject

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/dector/lampa/internal/out"
)

// AndroidProject represents an Android Gradle project and provides
// methods to locate build artifacts within the project structure.
type AndroidProject struct {
	// RootDir is the root directory of the Android project
	RootDir string
}

// NewAndroidProject creates a new AndroidProject instance for the specified directory.
// The directory should be the root of an Android Gradle project.
func NewAndroidProject(dir string) AndroidProject {
	return AndroidProject{
		RootDir: dir,
	}
}

// FindAabFile locates the AAB (Android App Bundle) file for a given module and build variant.
// It searches in the standard Gradle output directory: module/build/outputs/bundle/buildVariant
//
// Parameters:
//   - module: The Gradle module name (e.g., "app")
//   - buildVariant: The build variant name (e.g., "release", "debug")
//
// Returns the absolute path to the AAB file, or an error if not found.
func (self AndroidProject) FindAabFile(module string, buildVariant string) (string, error) {
	out.Info("Searching for AAB file")

	relativeBundleDir := path.Join(module, "build", "outputs", "bundle", buildVariant)
	bundleDir := path.Join(self.RootDir, relativeBundleDir)

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
