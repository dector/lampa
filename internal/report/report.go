package report

type Report struct {
	Version string `json:"v"`
	Type    string

	Context ContextSegment

	Build BuildSegment
}

type ToolSegment struct {
	Name        string
	Repository  string
	Version     string
	BuildCommit string
}

type BuildSegment struct {
	ApkName string
	ApkSha1 string

	ApplicationId string
	VersionName   string
	VersionCode   string
	BuildVariant  string

	MinSdkVersion     string
	TargetSdkVersion  string
	CompileSdkVersion string

	Locales []string

	CompileDependencies []string
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
