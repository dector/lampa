package components

import (
	mavendeps "github.com/dector/lampa/pkg/maven-deps"
)

// Alias for backward compatibility with templates
type MvnDependencyDiff = mavendeps.MvnDependencyDiff

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
	MvnDependencyDiff

	Kind DependencyDiffKind
}
