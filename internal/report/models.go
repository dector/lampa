package report

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

const StatsReportPrefix = "stats/"
const LatestStatsReportVersion = "0.0.1"

const DiffReportPrefix = "diff/"
const LatestDiffReportVersion = "0.0.1"

type Report_Version struct {
	Version string `json:"v"`
}

type CommonReport struct {
	Version string `json:"v"`

	Context ContextSegment
}

type StatsReport struct {
	CommonReport
	Build StatsBuildSegment
}

type DiffReport struct {
	CommonReport

	Build     DiffBuildSegment
	PrevBuild StatsBuildSegment
}

type CommonBuildSegment struct {
	AabName string
	AabSha1 string
	AabSize string

	AppName       string
	ApplicationId string
	VersionName   string
	VersionCode   string
	BuildVariant  string

	MinSdkVersion     string
	TargetSdkVersion  string
	CompileSdkVersion string

	// Locales []string
}

type StatsBuildSegment struct {
	CommonBuildSegment

	Dependencies StatsDependenciesSegment
}

type DiffBuildSegment struct {
	CommonBuildSegment

	Dependencies DiffDependenciesSegment
}

// --- Dependencies ---

type StatsDependenciesSegment struct {
	Compile []MvnDependency
}

type DiffDependenciesSegment struct {
	Added   []MvnDependency
	Removed []MvnDependency

	Upgraded   []MvnDependencyDiff
	Downgraded []MvnDependencyDiff

	Changed   []MvnDependencyDiff
	Unchanged []MvnDependency
}

type MvnDependencyDiff struct {
	MvnDependency

	PrevVersion string
}

type MvnDependency struct {
	Group   string
	Name    string
	Version string
}

func (self MvnDependency) String() string {
	return fmt.Sprintf("%s:%s:%s", self.Group, self.Name, self.Version)
}

func (a MvnDependency) Equals(b MvnDependency) bool {
	return a.Group == b.Group &&
		a.Name == b.Name &&
		a.Version == b.Version
}

func (a MvnDependency) HasSameCoordinates(b MvnDependency) bool {
	return a.Group == b.Group && a.Name == b.Name
}

func (self MvnDependency) HasSemanticVersion() bool {
	_, err := semver.NewVersion(self.Version)
	return err == nil
}

func (self MvnDependency) ToDiff() MvnDependencyDiff {
	return MvnDependencyDiff{
		MvnDependency: self,
	}
}

type Comparison int

const (
	Lower Comparison = iota
	Equals
	Higher
	Unknown
)

func (a MvnDependency) CompareVersion(b MvnDependency) Comparison {
	if a.Version == b.Version {
		return Equals
	}

	var err error
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

// --- /Dependencies ---

// --- Context ---

type ContextSegment struct {
	Tool ToolSegment
	Git  GitSegment

	GenerationTime string
}

type ToolSegment struct {
	Name        string
	Website     string
	Sources     string
	Version     string
	BuildCommit string
}

type GitSegment struct {
	Commit          string
	Branch          string
	Tag             string
	CommitsAfterTag uint
	IsDirty         bool
}

// --- /Context ---
