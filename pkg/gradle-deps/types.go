package gradledeps

import "fmt"

// DependenciesTree represents a parsed Gradle dependency tree.
// It contains both the hierarchical tree structure and a flattened summary.
type DependenciesTree struct {
	// Root is the root node of the dependency tree.
	// The root itself has no dependency information but contains all top-level dependencies as children.
	Root Dependency

	// Summary is a flattened list of all dependencies from constraint lines.
	// This is typically what's used for analysis and reporting.
	Summary []Dependency
}

// Dependency represents a single dependency in the tree.
// It can be either a Maven dependency (GroupID:ArtifactID:Version) or a Gradle project module.
type Dependency struct {
	// Children contains nested dependencies for this dependency.
	Children []Dependency

	// GroupID is the Maven group identifier (e.g., "com.google.dagger").
	// Empty for Gradle project modules.
	GroupID string

	// ArtifactID is the Maven artifact identifier (e.g., "hilt-android").
	// For Gradle project modules, this contains the project path (e.g., ":feature:interests").
	ArtifactID string

	// Version is the resolved version of the dependency.
	// Empty for Gradle project modules.
	Version string

	// RequestedVersion is the originally requested version before resolution.
	// May differ from Version when Gradle resolves version conflicts.
	// Empty if no specific version was requested.
	RequestedVersion string

	// IsAModule indicates whether this is a Gradle project module (true)
	// or a Maven dependency (false).
	IsAModule bool
}

// String formats the dependency as "GroupID:ArtifactID:Version".
func (d Dependency) String() string {
	return fmt.Sprintf("%s:%s:%s", d.GroupID, d.ArtifactID, d.Version)
}

// IsEquals performs a deep equality check between two Dependencies.
// It compares all fields including nested children recursively.
func (d Dependency) IsEquals(other Dependency) bool {
	if len(d.Children) != len(other.Children) {
		return false
	}
	for i := range d.Children {
		if !d.Children[i].IsEquals(other.Children[i]) {
			return false
		}
	}
	return d.GroupID == other.GroupID &&
		d.ArtifactID == other.ArtifactID &&
		d.Version == other.Version &&
		d.RequestedVersion == other.RequestedVersion
}

// ParsedDependency represents a dependency with parsing metadata.
// This is primarily used internally during parsing and in tests.
type ParsedDependency struct {
	// Dependency is the actual dependency information.
	Dependency Dependency

	// Level is the depth in the dependency tree (0 = root).
	Level int

	// IsASummary indicates if this was parsed from a summary/constraint line.
	IsASummary bool
}

// IsEquals performs an equality check between two ParsedDependencies.
// It compares both the dependency data and the parsing metadata.
func (d ParsedDependency) IsEquals(other ParsedDependency) bool {
	return d.Dependency.IsEquals(other.Dependency) &&
		d.Level == other.Level
}
