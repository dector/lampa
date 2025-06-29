package report

type Report struct {
	Version string `json:'v'`
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
	GitCommit string
}
