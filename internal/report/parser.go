package report

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dector/lampa/internal"
	"github.com/dector/lampa/internal/git"
	"github.com/dector/lampa/internal/out"
	"github.com/dector/lampa/pkg/bundles"
	"github.com/dector/lampa/pkg/gradle"

	gogit "github.com/go-git/go-git/v6"

	. "github.com/dector/lampa/internal/globals"
)

type ParseFromArgs struct {
	PathToAab    string
	BuildVariant string
	ProjectDir   string
}

func ParseFrom(args ParseFromArgs) (Report, error) {
	result := Report{
		Version: StatsReportPrefix + LatestStatsReportVersion,
	}

	context, err := parseContext(args)
	if err != nil {
		return Report{}, err
	}
	result.Context = context

	configurationName := args.BuildVariant + "CompileClasspath"

	err = analyzeBuild(&result, args)
	if err != nil {
		return Report{}, err
	}

	out.Info("Fetching dependencies tree")

	output, err := gradle.
		In(args.ProjectDir).
		Execute("app:dependencies", "--configuration", configurationName)
	if err != nil {
		return Report{}, fmt.Errorf("failed to execute gradlew: %v\nOutput:\n%s", err, output)
	}

	// fmt.Println(string(output))

	tree, err := internal.ParseTreeFromOutput(string(output), configurationName)
	if err != nil {
		return Report{}, fmt.Errorf("failed to parse tree: %v", err)
	}

	for _, info := range tree.Summary {
		d := MvnDependency{
			Group:   info.GroupID,
			Name:    info.ArtifactID,
			Version: info.Version,
		}
		result.Build.Dependencies.Compile = append(result.Build.Dependencies.Compile, d)
	}
	slices.SortFunc(result.Build.Dependencies.Compile, func(a, b MvnDependency) int {
		if a.Group > b.Group {
			return 1
		} else if a.Group < b.Group {
			return -1
		} else if a.Name > b.Name {
			return 1
		} else if a.Name < b.Name {
			return -1
		} else if a.Version > b.Version {
			return 1
		} else if a.Version < b.Version {
			return -1
		} else {
			return 0
		}
	})

	return result, nil
}

func parseContext(args ParseFromArgs) (ContextSegment, error) {
	result := ContextSegment{
		Tool: ToolSegment{
			Name:        "Lampa",
			Website:     "https://github.com/dector/lampa",
			Sources:     "https://github.com/dector/lampa",
			Version:     G.Version,
			BuildCommit: G.BuildCommit,
		},
		GenerationTime: time.Now().UTC().Format(time.RFC3339),
	}

	out.Info("Parsing git repo information")

	_, err := exec.LookPath("git")
	if err != nil {
		return result, fmt.Errorf("git not found in PATH: %v", err)
	}

	git := git.NewGit(args.ProjectDir)

	repo, err := gogit.PlainOpen(args.ProjectDir)
	if err != nil {
		out.PrintlnWarn("git repository not found in %s", args.ProjectDir)
		return result, nil
	}

	headRef, err := repo.Head()
	if err == nil {
		result.Git.Commit = headRef.Hash().String()

		refName := headRef.Name()
		if refName.IsBranch() {
			result.Git.Branch = refName.Short()
		}
	}

	worktree, err := repo.Worktree()
	if err == nil {
		status, err := worktree.Status()
		if err == nil {
			result.Git.IsDirty = !status.IsClean()
		}
	}

	output, err := git.RunWithOutput("describe", "--tags", "--long")
	if err == nil {
		parts := strings.SplitN(strings.TrimSpace(string(output)), "-", 3)
		if len(parts) == 3 {
			result.Git.Tag = parts[0]
			commitsAfterTag, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				out.PrintlnWarn("could not parse commits after tag from %q: %v", parts[1], err)
			} else {
				result.Git.CommitsAfterTag = uint(commitsAfterTag)
			}
		} else {
			log.Printf("warning: unexpected format from git describe: %q", string(output))
		}
	} else {
		log.Printf("warning: git describe failed: %v", err)
	}

	return result, nil
}

func analyzeBuild(result *Report, args ParseFromArgs) error {
	out.Info("Analyzing AAB file")

	result.Build.BuildVariant = args.BuildVariant
	result.Build.AabName = filepath.Base(args.PathToAab)

	fileAab, err := os.Open(args.PathToAab)
	if err == nil {
		defer fileAab.Close()
		hasher := sha1.New()
		if _, err := io.Copy(hasher, fileAab); err == nil {
			result.Build.AabSha1 = fmt.Sprintf("%x", hasher.Sum(nil))
		}
	} else {
		return err
	}

	infoAab, err := os.Stat(args.PathToAab)
	if err != nil {
		return fmt.Errorf("could not stat AAB file: %v", err)
	}
	result.Build.AabSize = strconv.FormatInt(infoAab.Size(), 10)

	manifestData, err := bundles.LoadManifest(args.PathToAab)
	if err != nil {
		return fmt.Errorf("failed to load AAB manifest: %v", err)
	}

	result.Build.ApplicationId = manifestData.Package
	result.Build.VersionCode = manifestData.VersionCode
	result.Build.VersionName = manifestData.VersionName
	result.Build.MinSdkVersion = manifestData.MinSdkVersion
	result.Build.TargetSdkVersion = manifestData.TargetSdkVersion
	result.Build.CompileSdkVersion = manifestData.CompileSdkVersion

	appLabel := manifestData.Label
	if strings.HasPrefix(appLabel, "@string/") {
		str, err := findString(args.PathToAab, strings.TrimPrefix(appLabel, "@string/"))
		if err != nil {
			return err
		}
		if str != "" {
			result.Build.AppName = str
		}
	} else {
		result.Build.AppName = manifestData.Label
	}

	// 	case "locales":
	// 		{
	// 			result.Build.Locales = lo.Map(strings.Fields(v), func(locale string, _ int) string {
	// 				return strings.Trim(locale, "'")
	// 			})
	// 		}
	// 	}
	// }

	return nil
}

func findString(pathToAab string, stringName string) (string, error) {
	res, err := bundles.LoadResources(pathToAab)
	if err != nil {
		return "", fmt.Errorf("could not load resources from AAB file `%s`: %v", pathToAab, err)
	}

	for _, p := range res.GetPackage() {
		for _, t := range p.GetType() {
			if t.GetName() != "string" {
				continue
			}

			for _, e := range t.GetEntry() {
				if e.GetName() != stringName {
					continue
				}

				for _, cv := range e.GetConfigValue() {
					text := cv.GetValue().GetItem().GetStr().GetValue()
					return text, nil
				}
			}
		}
	}

	return "", fmt.Errorf("string `%s` not found in AAB file `%s`", stringName, pathToAab)
}
