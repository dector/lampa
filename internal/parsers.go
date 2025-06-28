package internal

import (
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"
)

type DependenciesTree struct {
	Root Dependency
}

type Dependency struct {
	Children []Dependency

	GroupID    string
	ArtifactID string
	Version    string

	RequestedVersion string
}

func (d Dependency) String() string {
	return fmt.Sprintf("%s:%s:%s", d.GroupID, d.ArtifactID, d.Version)
}

func (d ParsedDependency) IsEquals(other ParsedDependency) bool {
	return d.Dependency.IsEquals(other.Dependency) &&
		d.Level == other.Level
}

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

func ParseTree(source string) (DependenciesTree, error) {
	result := DependenciesTree{}

	source = strings.TrimSpace(source)
	if len(source) == 0 {
		return result, ErrEmptyInput
	}

	result.Root = Dependency{}

	for line := range strings.Lines(source) {
		line := strings.TrimSpace(line)
		dep := lo.Must(parseDependencyLine(line))

		node := findLatestOnLevel(&result.Root, dep.Level-1)
		node.Children = append(node.Children, dep.Dependency)
	}

	return result, nil
}

type ParsedDependency struct {
	Dependency Dependency
	Level      int
}

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

func parseDependencyLine(line string) (ParsedDependency, error) {
	result := ParsedDependency{}

	line = strings.TrimSpace(line)
	parts := strings.Fields(line)

	// Parse level
	result.Level = lo.CountBy(parts, func(it string) bool {
		return IsATreeMarker(it)
	})

	// Parse arttefact
	artefact := ""
	for _, part := range parts {
		if !IsATreeMarker(part) {
			artefact = part
			break
		}
	}
	artefactParts := strings.Split(artefact, ":")
	result.Dependency.GroupID = artefactParts[0]
	result.Dependency.ArtifactID = artefactParts[1]
	result.Dependency.RequestedVersion = artefactParts[2]

	// Parse resolved version
	resolvedVersionMarkerIdx := lo.IndexOf(parts, "->")
	resolvedVersionIdx := resolvedVersionMarkerIdx + 1
	if resolvedVersionMarkerIdx != -1 && resolvedVersionIdx < len(parts) {
		result.Dependency.Version = parts[resolvedVersionIdx]
	} else {
		result.Dependency.Version = result.Dependency.RequestedVersion
	}

	return result, nil
}

func IsATreeMarker(it string) bool {
	return it == "|" || it == "+---" || it == "\\---"
}

var ErrEmptyInput = errors.New("input is empty")

func IsEmptyInput(err error) bool {
	return err == ErrEmptyInput
}
