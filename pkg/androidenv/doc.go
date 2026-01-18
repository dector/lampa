// Package androidenv provides utilities for working with Android/Gradle build environments.
//
// This package includes functions for parsing Java properties files, extracting
// Gradle versions from distribution URLs, and parsing Java version strings from
// various sources including version manager configuration files.
//
// Main use cases:
//   - Parse gradle-wrapper.properties files to extract Gradle distribution URLs
//   - Extract Gradle version numbers from distribution URLs
//   - Parse Java version from .tool-versions (asdf) configuration
//   - Parse Java version from .java-version (jenv) configuration
//   - Extract major version numbers from various Java version string formats
package androidenv
