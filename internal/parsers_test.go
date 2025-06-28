package internal

import (
	"testing"
)

func TestParseTree_EmptyInput(t *testing.T) {
	result, err := ParseTree("")
	if !IsEmptyInput(err) {
		t.Errorf("Expected empty DependenciesTree, got: %v", result)
	}
}

func TestParseTree_ShortInput(t *testing.T) {
	input := `
+--- com.google.dagger:hilt-android:2.56
|    +--- com.google.dagger:dagger:2.56
|    |    +--- jakarta.inject:jakarta.inject-api:2.0.1
|    |    +--- javax.inject:javax.inject:1
|    |    \--- org.jspecify:jspecify:1.0.0
|    +--- com.google.dagger:dagger-lint-aar:2.56
|    +--- com.google.dagger:hilt-core:2.56
|    |    +--- com.google.dagger:dagger:2.56 (*)
|    |    +--- com.google.code.findbugs:jsr305:3.0.2 (c)
|    |    \--- javax.inject:javax.inject:1
`
	result, err := ParseTree(input)
	if err != nil {
		t.Fatalf("ParseTree returned error: %v", err)
	}

	expected := DependenciesTree{
		Root: Dependency{
			GroupID:    "",
			ArtifactID: "",
			Version:    "",
			Children: []Dependency{
				{
					GroupID:    "com.google.dagger",
					ArtifactID: "hilt-android",
					Version:    "2.56",
					Children: []Dependency{
						{
							GroupID:    "com.google.dagger",
							ArtifactID: "dagger",
							Version:    "2.56",
							Children: []Dependency{
								{
									GroupID:    "jakarta.inject",
									ArtifactID: "jakarta.inject-api",
									Version:    "2.0.1",
								},
								{
									GroupID:    "javax.inject",
									ArtifactID: "javax.inject",
									Version:    "1",
								},
								{
									GroupID:    "org.jspecify",
									ArtifactID: "jspecify",
									Version:    "1.0.0",
								},
							},
						},
						{
							GroupID:    "com.google.dagger",
							ArtifactID: "dagger-lint-aar",
							Version:    "2.56",
						},
						{
							GroupID:    "com.google.dagger",
							ArtifactID: "hilt-core",
							Version:    "2.56",
							Children: []Dependency{
								{
									GroupID:    "com.google.dagger",
									ArtifactID: "dagger",
									Version:    "2.56",
								},
								{
									GroupID:    "com.google.code.findbugs",
									ArtifactID: "jsr305",
									Version:    "3.0.2",
								},
								{
									GroupID:    "javax.inject",
									ArtifactID: "javax.inject",
									Version:    "1",
								},
							},
						},
					},
				},
			},
		},
	}
	if !dependenciesTreeEqual(&result, &expected) {
		t.Errorf("ParseTree result does not match expected.\nGot: %#v\nExpected: %#v", result, expected)
	}
}

func dependenciesTreeEqual(a, b *DependenciesTree) bool {
	if a == nil || b == nil {
		return a == b
	}
	if len(a.Root.Children) != len(b.Root.Children) {
		return false
	}
	for i := range a.Root.Children {
		if !dependencyNodeEqual(a.Root.Children[i], b.Root.Children[i]) {
			return false
		}
	}
	return true
}

func dependencyNodeEqual(a, b Dependency) bool {
	if a.GroupID != b.GroupID {
		return false
	}
	if a.ArtifactID != b.ArtifactID {
		return false
	}
	if a.Version != b.Version {
		return false
	}
	if len(a.Children) != len(b.Children) {
		return false
	}
	for i := range a.Children {
		if !dependencyNodeEqual(a.Children[i], b.Children[i]) {
			return false
		}
	}
	return true
}

func TestParseDependency_FirstLevel(t *testing.T) {
	line := "|    |    +--- javax.inject:javax.inject:1.2.3"
	dep, err := parseDependencyLine(line)
	if err != nil {
		t.Fatalf("ParseDependency returned error: %v", err)
	}

	if dep.Dependency.GroupID != "javax.inject" {
		t.Errorf("Failed to parse dependency group, got: %s", dep.Dependency.GroupID)
	}
	if dep.Dependency.ArtifactID != "javax.inject" {
		t.Errorf("Failed to parse dependency name, got: %s", dep.Dependency.ArtifactID)
	}
	if dep.Dependency.Version != "1.2.3" {
		t.Errorf("Failed to parse dependency version, got: %s", dep.Dependency.Version)
	}
	if dep.Level != 3 {
		t.Errorf("Failed to parse depth, got: %d", dep.Level)
	}
}
