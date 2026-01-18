# gradle-deps

A Go package for parsing Gradle dependency trees into structured data.

## Overview

`gradledeps` parses the text output from Gradle's `dependencies` task and converts it into structured Go types. It handles complex dependency scenarios including version conflicts, Gradle project modules, dependency constraints, and maintains both hierarchical tree structure and flattened summaries.

## Features

- Parse Gradle dependency tree output into structured Go types
- Support for Maven dependencies (group:artifact:version format)
- Handle version conflicts and resolution (marked with "->")
- Parse Gradle project modules (marked with "project")
- Process dependency constraints with "{strictly ...}"
- Maintain both tree hierarchy and flattened summary
- Extract specific configuration sections from full output

## Installation

```bash
go get github.com/dector/lampa/pkg/gradle-deps
```

## Usage

### Basic Example

Parse a complete Gradle dependencies output and extract a specific configuration:

```go
package main

import (
    "fmt"
    "log"
    "os"

    gradledeps "github.com/dector/lampa/pkg/gradle-deps"
)

func main() {
    // Read gradle dependencies output
    output, _ := os.ReadFile("gradle-dependencies.txt")

    // Parse specific configuration
    tree, err := gradledeps.ParseTreeFromOutput(string(output), "compileClasspath")
    if err != nil {
        log.Fatal(err)
    }

    // Access flattened summary
    for _, dep := range tree.Summary {
        fmt.Printf("%s:%s:%s\n", dep.GroupID, dep.ArtifactID, dep.Version)
    }
}
```

### Parse Pre-extracted Configuration

If you already have a specific configuration section extracted:

```go
tree, err := gradledeps.ParseTree(configurationOutput)
if err != nil {
    if gradledeps.IsEmptyInput(err) {
        fmt.Println("No dependencies found")
    } else {
        log.Fatal(err)
    }
}
```

### Navigate Tree Structure

Access the hierarchical dependency tree:

```go
// Traverse top-level dependencies
for _, child := range tree.Root.Children {
    fmt.Printf("Top-level: %s\n", child.String())

    // Access nested dependencies
    for _, nested := range child.Children {
        fmt.Printf("  Nested: %s\n", nested.String())
    }
}
```

### Distinguish Gradle Modules from Maven Dependencies

```go
for _, dep := range tree.Summary {
    if dep.IsAModule {
        fmt.Printf("Gradle module: %s\n", dep.ArtifactID)
    } else {
        fmt.Printf("Maven dependency: %s:%s:%s\n",
            dep.GroupID, dep.ArtifactID, dep.Version)
    }
}
```

### Handle Version Conflicts

```go
for _, dep := range tree.Summary {
    if dep.RequestedVersion != dep.Version {
        fmt.Printf("Version conflict: requested %s, resolved to %s\n",
            dep.RequestedVersion, dep.Version)
    }
}
```

## API Reference

### Types

#### `DependenciesTree`

Represents a parsed Gradle dependency tree with both hierarchical structure and flattened summary.

```go
type DependenciesTree struct {
    Root    Dependency   // Root node containing top-level dependencies as children
    Summary []Dependency // Flattened list of all dependencies
}
```

#### `Dependency`

Represents a single dependency (Maven or Gradle module).

```go
type Dependency struct {
    Children         []Dependency
    GroupID          string // Maven group (empty for Gradle modules)
    ArtifactID       string // Maven artifact or Gradle project path
    Version          string // Resolved version
    RequestedVersion string // Originally requested version
    IsAModule        bool   // true for Gradle project modules
}
```

### Functions

#### `ParseTreeFromOutput`

```go
func ParseTreeFromOutput(output string, configurationName string) (DependenciesTree, error)
```

Parses a Gradle dependency tree from the full output of the dependencies task, extracting the specified configuration section.

#### `ParseTree`

```go
func ParseTree(source string) (DependenciesTree, error)
```

Parses a Gradle dependency tree from a single configuration's text representation. Returns `ErrEmptyInput` if the input is empty.

#### `IsEmptyInput`

```go
func IsEmptyInput(err error) bool
```

Checks if an error is `ErrEmptyInput`.

### Methods

#### `Dependency.String()`

```go
func (d Dependency) String() string
```

Formats the dependency as "GroupID:ArtifactID:Version".

#### `Dependency.IsEquals()`

```go
func (d Dependency) IsEquals(other Dependency) bool
```

Performs deep equality check between dependencies, including nested children.

## Input Format

The parser expects output from Gradle's `dependencies` task, which looks like:

```
compileClasspath - Compile classpath for source set 'main'.
+--- org.jetbrains.kotlin:kotlin-stdlib:1.9.0
|    +--- org.jetbrains:annotations:13.0 -> 23.0.0
|    \--- org.jetbrains.kotlin:kotlin-stdlib-common:1.9.0
+--- com.google.dagger:hilt-android:2.48 -> 2.50
\--- project :feature:interests
```

## Error Handling

The package defines specific error types:

- `ErrEmptyInput`: Returned when input is empty or contains only whitespace
- Use `IsEmptyInput(err)` to check for this specific error

## Contributing

This package is part of the [Lampa](https://github.com/dector/lampa) project. See the main repository for contribution guidelines.

## License

MIT License - see the main Lampa project for details.
