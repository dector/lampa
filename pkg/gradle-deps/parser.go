package gradledeps

import (
	"errors"
	"strings"

	"github.com/dector/gx"
	"github.com/samber/lo"
)

// ErrEmptyInput is returned when the input string is empty or contains only whitespace.
var ErrEmptyInput = errors.New("input is empty")

// IsEmptyInput checks if an error is ErrEmptyInput.
func IsEmptyInput(err error) bool {
	return err == ErrEmptyInput
}

// ParseTreeFromOutput parses a Gradle dependency tree from the full output of the dependencies task.
// It searches for a section starting with the given configuration name and parses that section.
//
// Parameters:
//   - output: The complete output from gradle dependencies task
//   - configurationName: The configuration name to search for (e.g., "compileClasspath", "runtimeClasspath")
//
// Returns the parsed dependency tree or an error if the section cannot be found or parsed.
func ParseTreeFromOutput(output string, configurationName string) (DependenciesTree, error) {
	lines := strings.Split(output, "\n")
	_, startIdx, _ := lo.FindIndexOf(lines, func(it string) bool {
		return strings.HasPrefix(it, configurationName)
	})
	_, endIdx, _ := lo.FindIndexOf(lines[startIdx:], func(it string) bool {
		return strings.TrimSpace(it) == ""
	})

	return ParseTree(strings.Join(lines[startIdx+1:endIdx], "\n"))
}

// ParseTree parses a Gradle dependency tree from a text representation.
// The input should be the tree output from gradle's dependencies task for a single configuration.
//
// The parser handles:
//   - Tree structure markers (|, +---, \---)
//   - Version conflicts (marked with "->")
//   - Gradle project modules (marked with "project")
//   - Dependency constraints with "{strictly ...}"
//
// Returns ErrEmptyInput if the input is empty or contains only whitespace.
func ParseTree(source string) (DependenciesTree, error) {
	result := DependenciesTree{}

	source = strings.TrimSpace(source)
	if len(source) == 0 {
		return result, ErrEmptyInput
	}

	result.Root = Dependency{}

	for line := range strings.Lines(source) {
		line := strings.TrimSpace(line)
		dep := gx.Must(parseDependencyLine(line))

		if dep.IsASummary {
			result.Summary = append(result.Summary, dep.Dependency)
		} else {
			node := findLatestOnLevel(&result.Root, dep.Level-1)
			node.Children = append(node.Children, dep.Dependency)
		}
	}

	return result, nil
}

// isATreeMarker checks if a string is a Gradle dependency tree structure marker.
// Valid markers are: "|", "+---", "\---"
func isATreeMarker(it string) bool {
	return it == "|" || it == "+---" || it == "\\---"
}

// parseDependencyLine parses a single line from the dependency tree output.
// It extracts the dependency information, tree level, and whether it's a summary line.
func parseDependencyLine(line string) (ParsedDependency, error) {
	result := ParsedDependency{}

	line = strings.TrimSpace(line)

	// Handle constraint lines with {strictly ...}
	if strings.Contains(line, ":{strictly ") {
		line = strings.Replace(line, "{strictly ", "", 1)
		result.IsASummary = true
	}

	parts := strings.Fields(line)

	// Calculate tree depth by counting tree markers
	result.Level = lo.CountBy(parts, func(it string) bool {
		return isATreeMarker(it)
	})

	// Find the artifact coordinates (first non-marker field)
	artefact := ""
	for _, part := range parts {
		if !isATreeMarker(part) {
			artefact = part
			break
		}
	}

	// Handle Gradle project modules
	if artefact == "project" {
		projectMarkerIdx := lo.IndexOf(parts, "project")
		if projectMarkerIdx == -1 {
			panic("project marker found but no project name")
		}
		projectNameIdx := projectMarkerIdx + 1
		result.Dependency.ArtifactID = parts[projectNameIdx]
		result.Dependency.IsAModule = true
	} else {
		// Parse Maven coordinates (group:artifact:version)
		artefactParts := strings.Split(artefact, ":")
		result.Dependency.GroupID = artefactParts[0]
		result.Dependency.ArtifactID = artefactParts[1]

		if len(artefactParts) > 2 {
			result.Dependency.RequestedVersion = strings.TrimSuffix(artefactParts[2], "}")
		}

		// Check for resolved version (indicated by "->")
		resolvedVersionMarkerIdx := lo.IndexOf(parts, "->")
		resolvedVersionIdx := resolvedVersionMarkerIdx + 1
		if resolvedVersionMarkerIdx != -1 && resolvedVersionIdx < len(parts) {
			result.Dependency.Version = parts[resolvedVersionIdx]
		} else {
			result.Dependency.Version = result.Dependency.RequestedVersion
		}
	}

	return result, nil
}

// findLatestOnLevel traverses the dependency tree to find the most recently added
// dependency at the specified level. This is used to correctly parent new dependencies
// as they're parsed from the tree output.
func findLatestOnLevel(root *Dependency, level int) *Dependency {
	if level == 0 {
		return root
	}

	if len(root.Children) == 0 {
		return nil
	}
	latestChild := &root.Children[len(root.Children)-1]

	return findLatestOnLevel(latestChild, level-1)
}
