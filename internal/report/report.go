package report

import "fmt"

type Report struct {
	Version string `json:"v"`

	Context ContextSegment

	Build BuildSegment
}

type ToolSegment struct {
	Name        string
	Website     string
	Sources     string
	Version     string
	BuildCommit string
}

type BuildSegment struct {
	AabName string
	AabSha1 string
	AabSize string

	// ApkName string
	// ApkSha1 string
	// ApkSize string

	AppName       string
	ApplicationId string
	VersionName   string
	VersionCode   string
	BuildVariant  string

	MinSdkVersion     string
	TargetSdkVersion  string
	CompileSdkVersion string

	// Locales []string

	Dependencies DependenciesSegment
}

type DependenciesSegment struct {
	Compile []CoordinatedDependency
}

type CoordinatedDependency struct {
	Group   string
	Name    string
	Version string
}

type ContextSegment struct {
	Tool ToolSegment
	Git  GitSegment

	GenerationTime string
}

type GitSegment struct {
	Commit          string
	Branch          string
	Tag             string
	CommitsAfterTag uint
	IsDirty         bool
}

func (self CoordinatedDependency) String() string {
	return fmt.Sprintf("%s:%s:%s", self.Group, self.Name, self.Version)
}
