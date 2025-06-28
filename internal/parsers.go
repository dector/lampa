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
}

func (d Dependency) String() string {
	return fmt.Sprintf("%s:%s:%s", d.GroupID, d.ArtifactID, d.Version)
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
	result.Dependency.Version = artefactParts[2]

	return result, nil
}

func IsATreeMarker(it string) bool {
	return it == "|" || it == "+---" || it == "\\---"
}

var ErrEmptyInput = errors.New("input is empty")

func IsEmptyInput(err error) bool {
	return err == ErrEmptyInput
}
