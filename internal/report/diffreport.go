package report

import (
	mavendeps "github.com/dector/lampa/pkg/maven-deps"
	mavends "github.com/dector/lampa/pkg/maven-deps"
	"github.com/tiendc/go-deepcopy"
)

func BuildDiffReport(r1 *StatsReport, r2 *StatsReport) DiffReport {
	r := DiffReport{
		CommonReport: CommonReport{
			Version: DiffReportPrefix + LatestDiffReportVersion,
			Context: r2.Context,
		},
	}
	deepcopy.Copy(&r.PrevBuild, r1.Build)
	deepcopy.Copy(&r.Build, r2.Build)

	r.PrevBuild.Dependencies.Compile = make([]mavends.MvnDependency, 0)
	r.Build.Dependencies = buildDependenciesSegment(
		r1.Build.Dependencies,
		r2.Build.Dependencies,
	)

	return r
}

func buildDependenciesSegment(
	deps1 StatsDependenciesSegment,
	deps2 StatsDependenciesSegment,
) DiffDependenciesSegment {
	d := DiffDependenciesSegment{}

	d1Deps := deps1.Compile
	d2Deps := deps2.Compile

	d.Added = mavendeps.FindAddedDeps(d1Deps, d2Deps)
	d.Removed = mavendeps.FindRemovedDeps(d1Deps, d2Deps)
	d.Upgraded = mavendeps.FindUpgradedDeps(d1Deps, d2Deps)
	d.Downgraded = mavendeps.FindDowngradedDeps(d1Deps, d2Deps)
	d.Changed = mavendeps.FindChangedDeps(d1Deps, d2Deps)
	d.Unchanged = mavendeps.FindUnchangedDeps(d1Deps, d2Deps)

	return d
}
