package compare

import (
	"context"
	"encoding/json"
	"fmt"
	"lampa/internal/report"
	"lampa/internal/templates/html/compare"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

func CreateCliCommand() *cli.Command {
	return &cli.Command{
		Name:   "compare",
		Action: ActionCmdCompare,
	}
}

func ActionCmdCompare(context context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 3 {
		return fmt.Errorf("usage: lampa report1.json report2.json out.html")
	}

	file1, err := checkReportFile(cmd.Args().Get(0))
	if err != nil {
		return err
	}
	file2, err := checkReportFile(cmd.Args().Get(1))
	if err != nil {
		return err
	}

	r1, err := ReadReportFromFile(file1)
	if err != nil {
		return err
	}
	r2, err := ReadReportFromFile(file2)
	if err != nil {
		return err
	}

	fmt.Printf("Comparing releases %s...%s\n", r1.Build.VersionName, r2.Build.VersionName)

	html, err := GenerateComparingHtmlReport(r1, r2)
	if err != nil {
		return err
	}

	outFile := cmd.Args().Get(2)
	outF, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer outF.Close()
	_, err = outF.Write([]byte(html))
	if err != nil {
		return err
	}

	// newDeps := findNewDeps(dep1, dep2)
	// removedDeps := findRemovedDeps(dep1, dep2)
	// changedDeps := findChangedDeps(dep1, dep2)

	// fmt.Println()
	// fmt.Printf("Total dependencies before: %d\n", len(dep1))
	// fmt.Printf("Total dependencies after: %d\n", len(dep2))

	// fmt.Println()
	// fmt.Printf("New dependencies:\n")
	// for _, dep := range newDeps {
	// 	fmt.Printf("- %s:%s: %s\n", dep.Group, dep.Name, dep.Version)
	// }

	// fmt.Println()
	// fmt.Printf("Removed dependencies:\n")
	// for _, dep := range removedDeps {
	// 	fmt.Printf("- %s:%s: %s\n", dep.Group, dep.Name, dep.Version)
	// }

	// fmt.Println()
	// fmt.Printf("Changed dependencies:\n")
	// for _, dep := range changedDeps {
	// 	fmt.Printf("- %s:%s: %s -> %s\n", dep.Dependency.Group, dep.Dependency.Name, dep.PrevVersion, dep.Dependency.Version)
	// }

	// fmt.Println()

	return nil
}

type Dependency struct {
	Group   string
	Name    string
	Version string
}

func parseDependencies(report report.Report) ([]Dependency, error) {
	result := make([]Dependency, 0, len(report.Build.CompileDependencies))

	for _, depStr := range report.Build.CompileDependencies {
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

func ReadReportFromFile(file string) (*report.Report, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %v", file, err)
	}

	var report report.Report
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("could not parse %s as report: %v", file, err)
	}

	return &report, nil
}

// func readReport(file string) ([]Dependency, string, error) {
// 	report, err := ReadReportFromFile(file)
// 	if err != nil {
// 		return nil, "", err
// 	}

// 	dep, err := parseDependencies(*report)
// 	if err != nil {
// 		return nil, "", fmt.Errorf("could not parse %s as report.Report: %v", file, err)
// 	}

// 	return dep, report.Context.Git.Commit, err
// }

func GenerateComparingHtmlReport(r1 *report.Report, r2 *report.Report) (string, error) {
	w := &strings.Builder{}
	err := compare.CompareHtml(r1, r2).Render(context.Background(), w)
	if err != nil {
		return "", err
	}
	return w.String(), nil
}
