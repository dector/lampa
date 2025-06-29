package report

type Report struct {
	Version string
	Tool    ToolSegment

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
