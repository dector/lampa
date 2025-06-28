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

	// parseDependencies(result, source, 1)

	return result, nil
}

type ParsedDependency struct {
	Dependency Dependency
	Level      int
}

func parseDependencyLine(line string) (ParsedDependency, error) {
	result := ParsedDependency{}

	line = strings.TrimSpace(line)
	parts := strings.Fields(line)

	// Parse level
	result.Level = lo.CountBy(parts, func(it string) bool {
		return it == "|" || it == "+---" || it == "\\---"
	})

	// Parse arttefact
	artefact := parts[len(parts)-1]
	artefactParts := strings.Split(artefact, ":")
	result.Dependency.GroupID = artefactParts[0]
	result.Dependency.ArtifactID = artefactParts[1]
	result.Dependency.Version = artefactParts[2]

	return result, nil
}

var ErrEmptyInput = errors.New("input is empty")

func IsEmptyInput(err error) bool {
	return err == ErrEmptyInput
}
