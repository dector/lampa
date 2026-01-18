# Android Project

A Go package that provides utilities for working with Android Gradle projects, helping locate build artifacts like AAB files.

## What is this package?

This package simplifies working with Android Gradle projects from Go code by providing a convenient API to locate build artifacts. It follows standard Gradle project conventions and supports multi-module projects with different build variants.

## How to use it

### Basic Usage

```go
import androidproject "github.com/dector/lampa/pkg/android-project"

// Create an AndroidProject instance for a project directory
project := androidproject.NewAndroidProject("/path/to/android/project")

// Find the AAB file for a module and build variant
aabPath, err := project.FindAabFile("app", "release")
if err != nil {
    log.Fatal(err)
}

fmt.Println("AAB file found at:", aabPath)
```

### API Reference

#### `NewAndroidProject(dir string) AndroidProject`

Creates a new AndroidProject instance for the specified directory.

**Parameters:**
- `dir`: Path to the Android project root directory

**Returns:** An `AndroidProject` instance

#### `FindAabFile(module string, buildVariant string) (string, error)`

Locates the AAB (Android App Bundle) file for a given module and build variant. It searches in the standard Gradle output directory: `module/build/outputs/bundle/buildVariant`.

**Parameters:**
- `module`: The Gradle module name (e.g., "app")
- `buildVariant`: The build variant name (e.g., "release", "debug")

**Returns:**
- `string`: Absolute path to the AAB file
- `error`: An error if the directory doesn't exist, is inaccessible, or contains no AAB files

### Example: Finding AAB for different variants

```go
project := androidproject.NewAndroidProject("/path/to/project")

// Find release AAB
releaseAab, err := project.FindAabFile("app", "release")
if err != nil {
    log.Fatalf("Release AAB not found: %v", err)
}

// Find debug AAB
debugAab, err := project.FindAabFile("app", "debug")
if err != nil {
    log.Fatalf("Debug AAB not found: %v", err)
}

fmt.Printf("Release: %s\nDebug: %s\n", releaseAab, debugAab)
```

### Example: Multi-module projects

```go
project := androidproject.NewAndroidProject("/path/to/project")

// Find AAB for a specific module
moduleAab, err := project.FindAabFile("feature-module", "release")
if err != nil {
    log.Fatalf("Module AAB not found: %v", err)
}

fmt.Println("Module AAB:", moduleAab)
```

## Requirements

- The target directory must be a valid Android Gradle project
- The AAB must have been built before attempting to locate it (e.g., by running `./gradlew bundleRelease`)
- The project must follow standard Gradle output conventions

## Directory Structure

This package expects the standard Gradle output structure:

```
<project-root>/
  <module>/
    build/
      outputs/
        bundle/
          <buildVariant>/
            *.aab
```

## Contributing

This package is part of the [Lampa](https://github.com/dector/lampa) project. See the main repository for contribution guidelines.

## License

MIT License - see the main Lampa project for details.
