package globals

import (
	_ "embed"
)

type Globals struct {
	Version     string
	BuildCommit string
}

var G = Globals{
	Version: "0.1.0-SNAPSHOT",
}

func (self *Globals) Init() {
	self.BuildCommit = commitHash
}

//go:generate sh -c "printf %s $(git rev-parse HEAD) > gen/COMMIT.txt"
//go:embed gen/COMMIT.txt
var commitHash string
