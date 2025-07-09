package report

import (
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

	r.PrevBuild.Dependencies.Compile = make([]MvnDependency, 0)
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

	d.Added = findAddedDeps(d1Deps, d2Deps)
	d.Removed = findRemovedDeps(d1Deps, d2Deps)
	d.Upgraded = findUpgradedDeps(d1Deps, d2Deps)
	d.Downgraded = findDowngradedDeps(d1Deps, d2Deps)
	d.Changed = findChangedDeps(d1Deps, d2Deps)
	d.Unchanged = findUnchangedDeps(d1Deps, d2Deps)

	return d
}

func findAddedDeps(deps1 []MvnDependency, deps2 []MvnDependency) []MvnDependency {
	added := make([]MvnDependency, 0)

	for _, d2 := range deps2 {
		found := false
		for _, d1 := range deps1 {
			if d1.HasSameCoordinates(d2) {
				found = true
				break
			}
		}
		if !found {
			added = append(added, d2)
		}
	}

	return added
}

func findRemovedDeps(deps1 []MvnDependency, deps2 []MvnDependency) []MvnDependency {
	removed := make([]MvnDependency, 0)

	for _, dep1 := range deps1 {
		found := false
		for _, dep2 := range deps2 {
			if dep1.HasSameCoordinates(dep2) {
				found = true
				break
			}
		}
		if !found {
			removed = append(removed, dep1)
		}
	}

	return removed
}

func findUpgradedDeps(deps1 []MvnDependency, deps2 []MvnDependency) []MvnDependencyDiff {
	upgraded := make([]MvnDependencyDiff, 0)

	for _, d1 := range deps1 {
		for _, d2 := range deps2 {
			if d1.HasSameCoordinates(d2) && d1.CompareVersion(d2) == Lower {
				upgraded = append(upgraded, MvnDependencyDiff{
					MvnDependency: d2,
					PrevVersion:   d1.Version,
				})
			}
		}
	}

	return upgraded
}

func findDowngradedDeps(deps1 []MvnDependency, deps2 []MvnDependency) []MvnDependencyDiff {
	downgraded := make([]MvnDependencyDiff, 0)

	for _, d1 := range deps1 {
		for _, d2 := range deps2 {
			if d1.HasSameCoordinates(d2) && d1.CompareVersion(d2) == Higher {
				downgraded = append(downgraded, MvnDependencyDiff{
					MvnDependency: d2,
					PrevVersion:   d1.Version,
				})
			}
		}
	}

	return downgraded
}

func findChangedDeps(deps1 []MvnDependency, deps2 []MvnDependency) []MvnDependencyDiff {
	changed := make([]MvnDependencyDiff, 0)

	for _, d1 := range deps1 {
		for _, d2 := range deps2 {
			if d1.HasSameCoordinates(d2) && d1.CompareVersion(d2) == Unknown {
				changed = append(changed, MvnDependencyDiff{
					MvnDependency: d2,
					PrevVersion:   d1.Version,
				})
			}
		}
	}

	return changed
}

func findUnchangedDeps(deps1 []MvnDependency, deps2 []MvnDependency) []MvnDependency {
	unchanged := make([]MvnDependency, 0)

	for _, d1 := range deps1 {
		for _, d2 := range deps2 {
			if d1.Equals(d2) {
				unchanged = append(unchanged, d2)
			}
		}
	}

	return unchanged
}
