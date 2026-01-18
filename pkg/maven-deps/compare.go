package mavendeps

import "slices"

// FindAddedDeps finds dependencies that exist in curr but not in prev.
// A dependency is considered "added" if its group:artifact coordinates don't exist in prev.
func FindAddedDeps(prev, curr []MvnDependency) []MvnDependency {
	added := make([]MvnDependency, 0)

	for _, d2 := range curr {
		found := false
		for _, d1 := range prev {
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

// FindRemovedDeps finds dependencies that exist in prev but not in curr.
// A dependency is considered "removed" if its group:artifact coordinates don't exist in curr.
func FindRemovedDeps(prev, curr []MvnDependency) []MvnDependency {
	removed := make([]MvnDependency, 0)

	for _, dep1 := range prev {
		found := slices.ContainsFunc(curr, dep1.HasSameCoordinates)
		if !found {
			removed = append(removed, dep1)
		}
	}

	return removed
}

// FindUpgradedDeps finds dependencies where the version increased.
// Uses semantic version comparison. Only returns dependencies with valid semantic versions
// where curr version > prev version.
func FindUpgradedDeps(prev, curr []MvnDependency) []MvnDependencyDiff {
	upgraded := make([]MvnDependencyDiff, 0)

	for _, d1 := range prev {
		for _, d2 := range curr {
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

// FindDowngradedDeps finds dependencies where the version decreased.
// Uses semantic version comparison. Only returns dependencies with valid semantic versions
// where curr version < prev version.
func FindDowngradedDeps(prev, curr []MvnDependency) []MvnDependencyDiff {
	downgraded := make([]MvnDependencyDiff, 0)

	for _, d1 := range prev {
		for _, d2 := range curr {
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

// FindChangedDeps finds dependencies with different versions that cannot be compared semantically.
// Returns dependencies where versions changed but don't follow semantic versioning
// (e.g., SNAPSHOT, custom version strings, etc.).
func FindChangedDeps(prev, curr []MvnDependency) []MvnDependencyDiff {
	changed := make([]MvnDependencyDiff, 0)

	for _, d1 := range prev {
		for _, d2 := range curr {
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

// FindUnchangedDeps finds dependencies that exist in both lists with identical versions.
// Returns dependencies where group, artifact, and version are all equal.
func FindUnchangedDeps(prev, curr []MvnDependency) []MvnDependency {
	unchanged := make([]MvnDependency, 0)

	for _, d1 := range prev {
		for _, d2 := range curr {
			if d1.Equals(d2) {
				unchanged = append(unchanged, d2)
			}
		}
	}

	return unchanged
}
