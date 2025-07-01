package compare

import (
	"context"
	"encoding/json"
	"fmt"
	"lampa/internal/report"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

func ActionCmdCompare(context context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 2 {
		return fmt.Errorf("expected exactly two files to compare, got %d", cmd.NArg())
	}

	file1, err := checkReportFile(cmd.Args().Get(0))
	if err != nil {
		return err
	}
	file2, err := checkReportFile(cmd.Args().Get(1))
	if err != nil {
		return err
	}

	dep1, release1, err := readReport(file1)
	if err != nil {
		return err
	}
	dep2, release2, err := readReport(file2)
	if err != nil {
		return err
	}

	fmt.Printf("Comparing releases %s...%s\n", release1, release2)

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

func checkReportFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("could not stat %s: %v", path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s exists but is a directory, not a file", path)
	}

	return path, nil
}

func readReport(file string) ([]Dependency, string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, "", fmt.Errorf("could not read %s: %v", file, err)
	}

	var report report.Report
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, "", fmt.Errorf("could not parse %s as report: %v", file, err)
	}

	dep, err := parseDependencies(report)
	if err != nil {
		return nil, "", fmt.Errorf("could not parse %s as report.Report: %v", file, err)
	}

	return dep, report.Context.Git.Commit, err
}
