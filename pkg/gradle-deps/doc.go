// Package gradledeps provides utilities for parsing Gradle dependency trees.
//
// This package can parse the text output from Gradle's dependencies task
// and convert it into structured data representing the dependency tree.
//
// The parser handles:
//   - Standard Maven dependencies (group:artifact:version)
//   - Version conflicts and resolution (marked with "->")
//   - Gradle project modules (marked with "project")
//   - Dependency constraints with "{strictly ...}"
//   - Tree structure markers (|, +---, \---)
//
// # Basic Usage
//
// To parse a dependency tree from Gradle output:
//
//	tree, err := gradledeps.ParseTreeFromOutput(gradleOutput, "compileClasspath")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Access the flattened summary
//	for _, dep := range tree.Summary {
//	    fmt.Printf("%s:%s:%s\n", dep.GroupID, dep.ArtifactID, dep.Version)
//	}
//
// # Direct Parsing
//
// If you already have a specific configuration section extracted:
//
//	tree, err := gradledeps.ParseTree(configurationOutput)
//	if err != nil {
//	    if gradledeps.IsEmptyInput(err) {
//	        fmt.Println("No dependencies found")
//	    } else {
//	        log.Fatal(err)
//	    }
//	}
//
// # Accessing the Tree Structure
//
// The parser builds a hierarchical tree structure as well as a flattened summary:
//
//	// Navigate the tree hierarchy
//	for _, child := range tree.Root.Children {
//	    fmt.Printf("Top-level: %s\n", child.String())
//	    for _, nested := range child.Children {
//	        fmt.Printf("  Nested: %s\n", nested.String())
//	    }
//	}
package gradledeps
