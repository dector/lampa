# Gradle Wrapper

A Go package that provides a simple wrapper for executing Gradle commands using the Gradle Wrapper (gradlew).

## What is this package?

This package simplifies interaction with Gradle projects from Go code by providing a convenient API to execute Gradle commands. It automatically configures Gradle to run in non-daemon mode with plain console output and quiet logging, making it ideal for programmatic use.

## How to use it

### Basic Usage

```go
import "github.com/dector/lampa/pkg/gradle"

// Create a Gradle wrapper for a project directory
g := gradle.In("/path/to/gradle/project")

// Execute Gradle tasks
output, err := g.Execute("clean", "build")
if err != nil {
    log.Fatal(err)
}

fmt.Println(string(output))
```

### API Reference

#### `In(dir string) Gradle`

Creates a new Gradle instance for the specified project directory. The directory should contain a `gradlew` executable.

**Parameters:**
- `dir`: Path to the Gradle project root directory

**Returns:** A `Gradle` instance

#### `Execute(args ...string) (GradleOutput, error)`

Runs a Gradle command with the specified arguments. Automatically adds the following flags:
- `--no-daemon`: Prevents the Gradle daemon from starting
- `--console plain`: Uses plain console output (no ANSI colors)
- `-q`: Quiet mode (less verbose output)

**Parameters:**
- `args`: Gradle task names and arguments

**Returns:**
- `GradleOutput`: Combined stdout/stderr output
- `error`: Any error that occurred during execution

#### `GradlewPath() string`

Returns the complete file path to the gradlew executable in the project directory.

**Returns:** Full path to gradlew

#### `EnsureExistsAndIsAFile() error`

Validates that the gradlew wrapper exists in the project directory and is a regular file (not a directory). This is useful for validating the Gradle setup before attempting to execute commands.

**Returns:**
- `error`: An error if the wrapper doesn't exist, can't be accessed, or is a directory; `nil` if validation succeeds

### Example: Validating Gradle setup

```go
g := gradle.In("/path/to/project")

// Validate that gradlew exists and is a file
if err := g.EnsureExistsAndIsAFile(); err != nil {
    log.Fatalf("Gradle wrapper validation failed: %v", err)
}

// Now safe to execute Gradle commands
output, err := g.Execute("tasks")
```

### Example: Running multiple tasks

```go
g := gradle.In("/path/to/project")

// Run clean and build tasks
output, err := g.Execute("clean", "build", "--stacktrace")
if err != nil {
    fmt.Printf("Build failed: %s\n", output)
    return err
}

fmt.Printf("Build successful:\n%s\n", output)
```

## Requirements

- The target directory must contain a `gradlew` (Gradle Wrapper) executable.
- The gradlew executable must have execute permissions.

## Contributing

This package is part of the [Lampa](https://github.com/dector/lampa) project. See the main repository for contribution guidelines.

## License

MIT License - see the main Lampa project for details.
