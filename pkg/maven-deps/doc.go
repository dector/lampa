// Package mavendeps provides utilities for working with Maven/Gradle dependencies.
//
// This package offers functionality to compare dependency trees, detect version
// changes (upgrades, downgrades), and identify added or removed dependencies.
// It supports semantic versioning for intelligent version comparison.
//
// Example usage:
//
//	// Compare two dependency lists
//	prev := []mavends.MvnDependency{
//		{Group: "com.example", Name: "library", Version: "1.0.0"},
//	}
//	curr := []mavends.MvnDependency{
//		{Group: "com.example", Name: "library", Version: "2.0.0"},
//	}
//
//	upgraded := mavends.FindUpgradedDeps(prev, curr)
//	for _, dep := range upgraded {
//		fmt.Printf("%s upgraded from %s to %s\n",
//			dep.MvnDependency.String(), dep.PrevVersion, dep.Version)
//	}
package mavendeps
