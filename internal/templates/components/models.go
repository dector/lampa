package components

import "github.com/dector/lampa/internal/report"

type DependencyDiffKind int

const (
	DiffKindNone DependencyDiffKind = iota
	DiffKindAdded
	DiffKindRemoved
	DiffKindUpgraded
	DiffKindDowngraded
	DiffKindUnchanged
	DiffKindChanged
)

type MvnDependencyVM struct {
	report.MvnDependencyDiff

	Kind DependencyDiffKind
}
