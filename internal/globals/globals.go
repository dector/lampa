package globals

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
)

var G = Globals{
	Version: "0.3.3-snapshot",
}

type Globals struct {
	Version string

	BuildCommit      string
	BuildCommitShort string

	UsePlainOutput bool
	UseOnlyStdout  bool

	Verbosity VerbosityLevel
}

type VerbosityLevel int

const (
	VerbosityNormal VerbosityLevel = iota
	VerbosityInfo
	VerbosityDebug
	VerbosityTrace
)

func (self *Globals) Init() error {
	self.BuildCommit = commitHash
	self.BuildCommitShort = commitShortHash

	isCI, err := parseCIEnv(os.Getenv("CI"))
	if err != nil {
		return err
	}

	self.UsePlainOutput = isCI
	self.UseOnlyStdout = isCI
	return nil
}

func parseCIEnv(value string) (bool, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return false, nil
	}

	switch value {
	case "y", "yes", "1", "true":
		return true, nil
	case "n", "no", "0", "false":
		return false, nil
	default:
		return false, fmt.Errorf("invalid CI value %q: expected one of y, yes, 1, true, n, no, 0, false", value)
	}
}

//go:generate sh -c "printf %s $(git rev-parse HEAD) > gen/COMMIT.txt"
//go:embed gen/COMMIT.txt
var commitHash string

//go:generate sh -c "printf %s $(git rev-parse --short HEAD) > gen/COMMIT_SHORT.txt"
//go:embed gen/COMMIT_SHORT.txt
var commitShortHash string
