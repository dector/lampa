# maven-deps

A Go package for comparing and analyzing Maven/Gradle dependencies.

## Overview

`mavendeps` provides utilities for comparing dependency trees, detecting version changes (upgrades, downgrades), and identifying added or removed dependencies. It supports semantic versioning for intelligent version comparison and handles non-semantic versions gracefully.

## Features

- Compare two dependency lists to find changes
- Detect added and removed dependencies
- Identify upgraded and downgraded dependencies
- Semantic version comparison using semver 2.0.0 spec
- Handle non-semantic versions (e.g., snapshot builds, custom versions)
- Find dependencies with unchanged versions
- Support for Maven coordinate format (group:artifact:version)

## Installation

```bash
go get github.com/dector/lampa/pkg/maven-deps
```

## Usage

### Basic Example

Compare two dependency lists to find what changed:

```go
package main

import (
    "fmt"

    mavendeps "github.com/dector/lampa/pkg/maven-deps"
)

func main() {
    // Previous dependencies
    prev := []mavendeps.MvnDependency{
        {Group: "com.example", Name: "library", Version: "1.0.0"},
        {Group: "com.example", Name: "old-lib", Version: "2.0.0"},
    }

    // Current dependencies
    curr := []mavendeps.MvnDependency{
        {Group: "com.example", Name: "library", Version: "2.0.0"},
        {Group: "com.example", Name: "new-lib", Version: "1.0.0"},
    }

    // Find upgraded dependencies
    upgraded := mavendeps.FindUpgradedDeps(prev, curr)
    for _, dep := range upgraded {
        fmt.Printf("Upgraded: %s from %s to %s\n",
            dep.MvnDependency.String(), dep.PrevVersion, dep.Version)
    }

    // Find added dependencies
    added := mavendeps.FindAddedDeps(prev, curr)
    for _, dep := range added {
        fmt.Printf("Added: %s\n", dep.String())
    }

    // Find removed dependencies
    removed := mavendeps.FindRemovedDeps(prev, curr)
    for _, dep := range removed {
        fmt.Printf("Removed: %s\n", dep.String())
    }
}
```

### Compare Versions

Compare dependency versions semantically:

```go
dep1 := mavendeps.MvnDependency{
    Group: "com.example",
    Name: "library",
    Version: "1.0.0",
}
dep2 := mavendeps.MvnDependency{
    Group: "com.example",
    Name: "library",
    Version: "2.0.0",
}

comparison := dep1.CompareVersion(&dep2)
switch comparison {
case mavendeps.Lower:
    fmt.Println("dep1 version is lower")
case mavendeps.Higher:
    fmt.Println("dep1 version is higher")
case mavendeps.Equals:
    fmt.Println("versions are equal")
case mavendeps.Unknown:
    fmt.Println("cannot compare (non-semantic versions)")
}
```

### Find All Changes

Get a complete picture of all dependency changes:

```go
upgraded := mavendeps.FindUpgradedDeps(prev, curr)
downgraded := mavendeps.FindDowngradedDeps(prev, curr)
changed := mavendeps.FindChangedDeps(prev, curr) // Non-semantic version changes
added := mavendeps.FindAddedDeps(prev, curr)
removed := mavendeps.FindRemovedDeps(prev, curr)
unchanged := mavendeps.FindUnchangedDeps(prev, curr)

fmt.Printf("Upgraded: %d, Downgraded: %d, Added: %d, Removed: %d, Unchanged: %d\n",
    len(upgraded), len(downgraded), len(added), len(removed), len(unchanged))
```

### Check Dependency Properties

```go
dep := mavendeps.MvnDependency{
    Group: "com.example",
    Name: "library",
    Version: "1.0.0",
}

// Check if version follows semantic versioning
if dep.HasSemanticVersion() {
    fmt.Println("Uses semantic versioning")
}

// Compare coordinates (ignoring version)
if dep.HasSameCoordinates(&other) {
    fmt.Println("Same group and artifact")
}

// Format as Maven coordinate
fmt.Println(dep.String()) // "com.example:library:1.0.0"
```

## API Reference

### Types

#### `MvnDependency`

Represents a Maven dependency with group, artifact, and version coordinates.

```go
type MvnDependency struct {
    Group   string // Maven group ID (e.g., "com.example")
    Name    string // Maven artifact ID (e.g., "library")
    Version string // Version string (e.g., "1.0.0")
}
```

#### `MvnDependencyDiff`

Represents a dependency that has changed versions, including the previous version.

```go
type MvnDependencyDiff struct {
    MvnDependency
    PrevVersion string // Previous version before the change
}
```

#### `Comparison`

Enum representing version comparison results.

```go
type Comparison int

const (
    Lower   Comparison = -1 // First version is lower
    Equals  Comparison = 0  // Versions are equal
    Higher  Comparison = 1  // First version is higher
    Unknown Comparison = 2  // Cannot compare (non-semantic versions)
)
```

### Functions

#### `FindAddedDeps`

```go
func FindAddedDeps(prev, curr []MvnDependency) []MvnDependency
```

Finds dependencies that exist in `curr` but not in `prev`.

#### `FindRemovedDeps`

```go
func FindRemovedDeps(prev, curr []MvnDependency) []MvnDependency
```

Finds dependencies that exist in `prev` but not in `curr`.

#### `FindUpgradedDeps`

```go
func FindUpgradedDeps(prev, curr []MvnDependency) []MvnDependencyDiff
```

Finds dependencies where the version increased (semantic versioning).

#### `FindDowngradedDeps`

```go
func FindDowngradedDeps(prev, curr []MvnDependency) []MvnDependencyDiff
```

Finds dependencies where the version decreased (semantic versioning).

#### `FindChangedDeps`

```go
func FindChangedDeps(prev, curr []MvnDependency) []MvnDependencyDiff
```

Finds dependencies with different versions that don't follow semantic versioning or have non-comparable versions.

#### `FindUnchangedDeps`

```go
func FindUnchangedDeps(prev, curr []MvnDependency) []MvnDependency
```

Finds dependencies that exist in both lists with identical versions.

### Methods

#### `MvnDependency.String()`

```go
func (d MvnDependency) String() string
```

Formats the dependency as a Maven coordinate: "group:artifact:version".

#### `MvnDependency.Equals()`

```go
func (d *MvnDependency) Equals(other *MvnDependency) bool
```

Performs deep equality check, comparing group, artifact, and version.

#### `MvnDependency.HasSameCoordinates()`

```go
func (d *MvnDependency) HasSameCoordinates(other *MvnDependency) bool
```

Checks if two dependencies have the same group and artifact (ignores version).

#### `MvnDependency.HasSemanticVersion()`

```go
func (d *MvnDependency) HasSemanticVersion() bool
```

Checks if the version string follows semantic versioning format.

#### `MvnDependency.CompareVersion()`

```go
func (d *MvnDependency) CompareVersion(other *MvnDependency) Comparison
```

Compares versions semantically. Returns `Unknown` if either version is non-semantic.

#### `MvnDependency.ToDiff()`

```go
func (d *MvnDependency) ToDiff(prevVersion string) MvnDependencyDiff
```

Converts to `MvnDependencyDiff` with the specified previous version.

## Version Comparison

The package uses semantic versioning (semver 2.0.0) for version comparison when possible:

- Versions like "1.0.0", "2.1.3", "1.0.0-alpha" are compared semantically
- Non-semantic versions (e.g., "SNAPSHOT", "local", custom strings) return `Unknown`
- Version comparison is case-sensitive

## Contributing

This package is part of the [Lampa](https://github.com/dector/lampa) project. See the main repository for contribution guidelines.

## License

MIT License - see the main Lampa project for details.
