package globals

import (
	_ "embed"
)

type Globals struct {
	Version string

	BuildCommit      string
	BuildCommitShort string
}

var G = Globals{
	Version: "0.1.0-1.dev",
}

func (self *Globals) Init() {
	self.BuildCommit = commitHash
	self.BuildCommitShort = commitShortHash
}

//go:generate sh -c "printf %s $(git rev-parse HEAD) > gen/COMMIT.txt"
//go:embed gen/COMMIT.txt
var commitHash string

//go:generate sh -c "printf %s $(git rev-parse --short HEAD) > gen/COMMIT_SHORT.txt"
//go:embed gen/COMMIT_SHORT.txt
var commitShortHash string
