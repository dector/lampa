package gradle

import (
	"os/exec"
	"path/filepath"
)

type Gradle struct {
	ProjectDir string
}

type GradleOutput string

var gradlew = "gradlew"

func In(dir string) Gradle {
	g := Gradle{}
	g.ProjectDir = dir
	return g
}

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

func (self Gradle) FullPath() string {
	return filepath.Join(self.ProjectDir, gradlew)
}
