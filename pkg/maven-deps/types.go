package mavendeps

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// Comparison represents the result of comparing two dependency versions.
type Comparison int

const (
	Lower   Comparison = iota // First version is lower than second
	Equals                    // Versions are equal
	Higher                    // First version is higher than second
	Unknown                   // Cannot compare (non-semantic versions)
)

// MvnDependency represents a Maven dependency with group, artifact, and version coordinates.
type MvnDependency struct {
	Group   string // Maven group ID (e.g., "com.example")
	Name    string // Maven artifact ID (e.g., "library")
	Version string // Version string (e.g., "1.0.0")
}

// String formats the dependency as a Maven coordinate: "group:artifact:version".
func (d MvnDependency) String() string {
	return fmt.Sprintf("%s:%s:%s", d.Group, d.Name, d.Version)
}

// Equals performs deep equality check, comparing group, artifact, and version.
func (a MvnDependency) Equals(b MvnDependency) bool {
	return a.Group == b.Group &&
		a.Name == b.Name &&
		a.Version == b.Version
}

// HasSameCoordinates checks if two dependencies have the same group and artifact (ignores version).
func (a MvnDependency) HasSameCoordinates(b MvnDependency) bool {
	return a.Group == b.Group && a.Name == b.Name
}

// HasSemanticVersion checks if the version string follows semantic versioning format.
func (d MvnDependency) HasSemanticVersion() bool {
	_, err := semver.NewVersion(d.Version)
	return err == nil
}

// CompareVersion compares versions semantically. Returns Unknown if either version is non-semantic.
func (a MvnDependency) CompareVersion(b MvnDependency) Comparison {
	if a.Version == b.Version {
		return Equals
	}

	aVer, err := semver.NewVersion(a.Version)
	if err != nil {
		return Unknown
	}
	bVer, err := semver.NewVersion(b.Version)
	if err != nil {
		return Unknown
	}

	if aVer.LessThan(bVer) {
		return Lower
	} else if aVer.GreaterThan(bVer) {
		return Higher
	} else {
		return Equals
	}
}

// ToDiff converts to MvnDependencyDiff with the specified previous version.
func (d MvnDependency) ToDiff(prevVersion string) MvnDependencyDiff {
	return MvnDependencyDiff{
		MvnDependency: d,
		PrevVersion:   prevVersion,
	}
}

// MvnDependencyDiff represents a dependency that has changed versions, including the previous version.
type MvnDependencyDiff struct {
	MvnDependency
	PrevVersion string // Previous version before the change
}
