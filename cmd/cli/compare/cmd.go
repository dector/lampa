package compare

import (
	"context"
	"encoding/json"
	"fmt"
	"lampa/internal/report"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

func ActionCmdCompare(context context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 2 {
		return fmt.Errorf("expected exactly two files to compare, got %d", cmd.NArg())
	}

	file1 := cmd.Args().Get(0)
	file2 := cmd.Args().Get(1)

	info1, err := os.Stat(file1)
	if err != nil {
		return fmt.Errorf("could not stat %s: %v", file1, err)
	}
	if info1.IsDir() {
		return fmt.Errorf("%s exists but is a directory, not a file", file1)
	}

	info2, err := os.Stat(file2)
	if err != nil {
		return fmt.Errorf("could not stat %s: %v", file2, err)
	}
	if info2.IsDir() {
		return fmt.Errorf("%s exists but is a directory, not a file", file2)
	}

	data1, err := os.ReadFile(file1)
	if err != nil {
		return fmt.Errorf("could not read %s: %v", file1, err)
	}

	var report1 report.Report
	if err := json.Unmarshal(data1, &report1); err != nil {
		return fmt.Errorf("could not parse %s as report.Report: %v", file1, err)
	}

	data2, err := os.ReadFile(file2)
	if err != nil {
		return fmt.Errorf("could not read %s: %v", file2, err)
	}

	var report2 report.Report
	if err := json.Unmarshal(data2, &report2); err != nil {
		log.Fatalf("error: could not parse %s as report.Report: %v", file2, err)
	}

	fmt.Printf("Comparing releases %s...%s\n", report1.Context.Git.Commit, report2.Context.Git.Commit)

	dep1, err := parseDependencies(report1)
	if err != nil {
		return fmt.Errorf("could not parse %s as report.Report: %v", file1, err)
	}
	dep2, err := parseDependencies(report2)
	if err != nil {
		return fmt.Errorf("could not parse %s as report.Report: %v", file2, err)
	}

	newDeps := findNewDeps(dep1, dep2)
	removedDeps := findRemovedDeps(dep1, dep2)
	changedDeps := findChangedDeps(dep1, dep2)

	fmt.Println()
	fmt.Printf("Total dependencies before: %d\n", len(dep1))
	fmt.Printf("Total dependencies after: %d\n", len(dep2))

	fmt.Println()
	fmt.Printf("New dependencies:\n")
	for _, dep := range newDeps {
		fmt.Printf("- %s:%s: %s\n", dep.Group, dep.Name, dep.Version)
	}

	fmt.Println()
	fmt.Printf("Removed dependencies:\n")
	for _, dep := range removedDeps {
		fmt.Printf("- %s:%s: %s\n", dep.Group, dep.Name, dep.Version)
	}

	fmt.Println()
	fmt.Printf("Changed dependencies:\n")
	for _, dep := range changedDeps {
		fmt.Printf("- %s:%s: %s -> %s\n", dep.Dependency.Group, dep.Dependency.Name, dep.PrevVersion, dep.Dependency.Version)
	}

	fmt.Println()

	return nil
}

type Dependency struct {
	Group   string
	Name    string
	Version string
}

func parseDependencies(report report.Report) ([]Dependency, error) {
	result := make([]Dependency, 0, len(report.Dependencies.Compiled))

	for _, depStr := range report.Dependencies.Compiled {
		parts := strings.Split(depStr, ":")
		if len(parts) == 3 {
			result = append(result, Dependency{
				Group:   parts[0],
				Name:    parts[1],
				Version: parts[2],
			})
		} else {
			err := fmt.Errorf("dependency string %q does not have 3 parts (group:name:version)", depStr)
			return []Dependency{}, err
		}
	}

	return result, nil
}

func findNewDeps(oldDeps, newDeps []Dependency) []Dependency {
	result := []Dependency{}
	oldMap := make(map[string]Dependency)
	for _, dep := range oldDeps {
		key := dep.Group + ":" + dep.Name
		oldMap[key] = dep
	}
	for _, dep := range newDeps {
		key := dep.Group + ":" + dep.Name
		if _, exists := oldMap[key]; !exists {
			result = append(result, dep)
		}
	}
	return result
}

func findRemovedDeps(oldDeps, newDeps []Dependency) []Dependency {
	result := []Dependency{}
	newMap := make(map[string]Dependency)
	for _, dep := range newDeps {
		key := dep.Group + ":" + dep.Name
		newMap[key] = dep
	}
	for _, dep := range oldDeps {
		key := dep.Group + ":" + dep.Name
		if _, exists := newMap[key]; !exists {
			result = append(result, dep)
		}
	}
	return result
}

type DependencyChange struct {
	Dependency  Dependency
	PrevVersion string
}

func findChangedDeps(oldDeps, newDeps []Dependency) []DependencyChange {
	result := []DependencyChange{}
	oldMap := make(map[string]Dependency)
	for _, dep := range oldDeps {
		key := dep.Group + ":" + dep.Name
		oldMap[key] = dep
	}
	for _, dep := range newDeps {
		key := dep.Group + ":" + dep.Name
		if oldDep, exists := oldMap[key]; exists {
			if oldDep.Version != dep.Version {
				result = append(result, DependencyChange{
					Dependency:  dep,
					PrevVersion: oldDep.Version,
				})
			}
		}
	}
	return result
}
