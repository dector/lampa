package git

import (
	"os/exec"
)

type Git struct {
	ProjectDir string
}

func NewGit(projectDir string) Git {
	return Git{
		ProjectDir: projectDir,
	}
}

func (self Git) Run(args ...string) error {
	return self.asCommand(args...).Run()
}

func (self Git) RunWithOutput(args ...string) ([]byte, error) {
	return self.asCommand(args...).Output()
}

func (self Git) asCommand(args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = self.ProjectDir
	return cmd
}
