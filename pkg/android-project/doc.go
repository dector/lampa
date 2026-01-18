// Package androidproject provides utilities for working with Android Gradle projects.
//
// This package helps locate build artifacts like AAB (Android App Bundle) files
// by following standard Gradle project conventions. It supports multi-module projects
// and different build variants.
//
// Example usage:
//
//	project := androidproject.NewAndroidProject("/path/to/project")
//	aabPath, err := project.FindAabFile("app", "release")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("AAB file found at:", aabPath)
package androidproject
