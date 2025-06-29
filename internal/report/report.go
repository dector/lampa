package report

type Report struct {
	Version string `json:"v"`
	Type    string
	Tool    ToolSegment
	Context ContextSegment

	Dependencies DependenciesSegment
}

type ToolSegment struct {
	Name       string
	Repository string
	Version    string
}

type DependenciesSegment struct {
	Compiled []string
}

type ContextSegment struct {
	Git GitSegment
}

type GitSegment struct {
	Commit          string
	Branch          string
	Tag             string
	CommitsAfterTag uint
	IsDirty         bool
}
