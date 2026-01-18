# androidenv

A Go package for working with Android and Gradle build environments. This package provides utilities for parsing Java properties files, extracting version information, and reading configuration from various version manager files.

## Features

- Parse Java `.properties` files (including gradle-wrapper.properties)
- Extract Gradle version from distribution URLs
- Parse Java version strings in multiple formats
- Read Java version from `.tool-versions` (asdf/mise)
- Read Java version from `.java-version` (jenv)

## Installation

```bash
go get github.com/dector/lampa/pkg/androidenv
```

## Usage

### Parse Properties Files

```go
import "github.com/dector/lampa/pkg/androidenv"

// Parse gradle-wrapper.properties
props, err := androidenv.ParsePropertiesFile("gradle/wrapper/gradle-wrapper.properties")
if err != nil {
    log.Fatal(err)
}
distributionUrl := props["distributionUrl"]
```

### Extract Gradle Version

```go
url := "https://services.gradle.org/distributions/gradle-8.0.2-all.zip"
version, err := androidenv.ExtractGradleVersion(url)
if err != nil {
    log.Fatal(err)
}
fmt.Println(version) // Output: 8.0.2
```

### Parse Java Version Strings

The package supports various Java version string formats:

```go
// Simple version
version, _ := androidenv.ExtractJavaMajorVersion("17")
// Output: "17"

// Full version with minor and patch
version, _ := androidenv.ExtractJavaMajorVersion("17.0.10")
// Output: "17"

// jenv style with distribution prefix
version, _ := androidenv.ExtractJavaMajorVersion("temurin64-17.0.10")
// Output: "17"
```

### Read from Version Manager Files

```go
// Read from .tool-versions (asdf/mise)
version, err := androidenv.ParseJavaVersionFromToolVersions("/path/to/project")
if err != nil {
    log.Fatal(err)
}

// Read from .java-version (jenv)
version, err := androidenv.ParseJavaVersionFromJavaVersion("/path/to/project")
if err != nil {
    log.Fatal(err)
}
```

## Supported Formats

### Gradle Distribution URLs
- `https://services.gradle.org/distributions/gradle-X.Y.Z-bin.zip`
- `https://services.gradle.org/distributions/gradle-X.Y.Z-all.zip`
- Supports version suffixes like `-rc-1`

### Java Version Strings
- Simple major version: `17`
- Full version: `17.0.10`
- jenv style: `temurin64-17.0.10`
- With whitespace trimming

### .tool-versions Format
- `java@17` or `java 17`
- Returns error for `java@latest`
- Ignores comment lines starting with `#`

### .java-version Format
- Single line with version string
- Supports all Java version string formats above
- Ignores comment lines starting with `#`

## Dependencies

- [github.com/magiconair/properties](https://github.com/magiconair/properties) - For parsing Java properties files

## Contributing

This package is part of the [Lampa](https://github.com/dector/lampa) project. See the main repository for contribution guidelines.

## License

MIT License - see the main Lampa project for details.
