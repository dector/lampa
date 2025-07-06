package globals

import (
	_ "embed"
	"os"
	"strings"
)

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

var G = Globals{
	Version: "0.2.1-snapshot",
}

func (self *Globals) Init() {
	self.BuildCommit = commitHash
	self.BuildCommitShort = commitShortHash

	isCI := strings.TrimSpace(os.Getenv("CI")) != ""
	self.UsePlainOutput = isCI
	self.UseOnlyStdout = isCI
}

//go:generate sh -c "printf %s $(git rev-parse HEAD) > gen/COMMIT.txt"
//go:embed gen/COMMIT.txt
var commitHash string

//go:generate sh -c "printf %s $(git rev-parse --short HEAD) > gen/COMMIT_SHORT.txt"
//go:embed gen/COMMIT_SHORT.txt
var commitShortHash string
