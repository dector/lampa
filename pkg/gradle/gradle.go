// This pacakge provides a wrapper for executing Gradle build commands
// using the Gradle Wrapper (gradlew).
package gradle

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Gradle is a Gradle project wrapper and provides methods
// to execute Gradle commands within a specific project directory.
type Gradle struct {
	// ProjectDir is the root directory of the Gradle project
	ProjectDir string
}

// GradleOutput represents the output returned from a Gradle command execution.
type GradleOutput string

var gradlew = "gradlew"

// In creates a new Gradle instance for the specified project directory.
// The directory should contain a gradlew executable.
func In(dir string) Gradle {
	g := Gradle{}
	g.ProjectDir = dir
	return g
}

// Execute runs a Gradle command with the specified arguments.
// It automatically adds flags for non-daemon mode, plain console output,
// and quiet logging. Returns the combined stdout/stderr output and any error.
func (self Gradle) Execute(args ...string) (GradleOutput, error) {
	args = append(
		args,
		"--no-daemon", "--console",
		"plain", "-q",
	)

	cmd := exec.Command(filepath.Join(self.ProjectDir, gradlew), args...)
	cmd.Dir = self.ProjectDir

	output, err := cmd.CombinedOutput()
	return GradleOutput(output), err
}

// GradlewPath returns the complete file path to the gradlew executable
// in the project directory.
func (self Gradle) GradlewPath() string {
	return filepath.Join(self.ProjectDir, gradlew)
}

// EnsureExistsAndIsAFile checks that the gradlew wrapper exists in the project
// directory and is a regular file (not a directory). Returns an error if the
// wrapper doesn't exist, can't be accessed, or is a directory.
func (self Gradle) EnsureExistsAndIsAFile() error {
	gradlewPath := self.GradlewPath()
	info, err := os.Stat(gradlewPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist", gradlewPath)
		}
		return fmt.Errorf("could not stat %s: %v", gradlewPath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s exists but is a directory, not a file", gradlewPath)
	}
	return nil
}
