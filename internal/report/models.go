package report

import (
	mavendeps "github.com/dector/lampa/pkg/maven-deps"
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
	Compile []mavendeps.MvnDependency
}

type DiffDependenciesSegment struct {
	Added   []mavendeps.MvnDependency
	Removed []mavendeps.MvnDependency

	Upgraded   []mavendeps.MvnDependencyDiff
	Downgraded []mavendeps.MvnDependencyDiff

	Changed   []mavendeps.MvnDependencyDiff
	Unchanged []mavendeps.MvnDependency
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
