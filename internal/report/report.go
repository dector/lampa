package report

type Report struct {
	Version string
	Tool    ToolSegment
}

type ToolSegment struct {
	Name       string
	Repository string
	Version    string
}
