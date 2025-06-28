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
|    |    +--- jakarta.inject:jakarta.inject-api:2.0.1 -> 2.0.2
|    |    +--- javax.inject:javax.inject:1
|    |    \--- org.jspecify:jspecify:1.0.0 -> 2.0.0 (c)
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
									GroupID:          "jakarta.inject",
									ArtifactID:       "jakarta.inject-api",
									Version:          "2.0.2",
									RequestedVersion: "2.0.1",
								},
								{
									GroupID:    "javax.inject",
									ArtifactID: "javax.inject",
									Version:    "1",
								},
								{
									GroupID:          "org.jspecify",
									ArtifactID:       "jspecify",
									Version:          "2.0.0",
									RequestedVersion: "1.0.0",
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

func TestParseDependency_1(t *testing.T) {
	line := "|    |    +--- javax.inject:javax.inject:1.2.3"
	actual, err := parseDependencyLine(line)
	if err != nil {
		t.Fatalf("ParseDependency returned error: %v", err)
	}

	expected := ParsedDependency{
		Dependency: Dependency{
			GroupID:    "javax.inject",
			ArtifactID: "javax.inject",
			Version:    "1.2.3",
		},
		Level:      3,
		IsASummary: false,
	}
	if actual.IsEquals(expected) {
		t.Errorf("Parsed dependency does not match expected structure.\nGot: %#v\nWant: %#v", actual, expected)
	}
}

func TestParseDependency_2(t *testing.T) {
	line := "|    +--- javax.inject:javax.inject:1.2.3 -> 2.1.2"
	actual, err := parseDependencyLine(line)
	if err != nil {
		t.Fatalf("ParseDependency returned error: %v", err)
	}

	expected := ParsedDependency{
		Dependency: Dependency{
			GroupID:          "javax.inject",
			ArtifactID:       "javax.inject",
			Version:          "2.1.2",
			RequestedVersion: "1.2.3",
		},
		Level:      2,
		IsASummary: false,
	}
	if !actual.IsEquals(expected) {
		t.Errorf("Parsed dependency does not match expected structure.\nGot: %#v\nWant: %#v", actual, expected)
	}
}

func TestParseDependency_3(t *testing.T) {
	line := "|    +--- javax.inject:javax.inject:1.2.3 -> 2.1.2 (c)"
	actual, err := parseDependencyLine(line)
	if err != nil {
		t.Fatalf("ParseDependency returned error: %v", err)
	}

	expected := ParsedDependency{
		Dependency: Dependency{
			GroupID:          "javax.inject",
			ArtifactID:       "javax.inject",
			Version:          "2.1.2",
			RequestedVersion: "1.2.3",
		},
		Level:      2,
		IsASummary: false,
	}
	if !actual.IsEquals(expected) {
		t.Errorf("Parsed dependency does not match expected structure.\nGot: %#v\nWnt: %#v", expected, actual)
	}
}

func TestParseDependency_SummaryLine(t *testing.T) {
	line := "+--- org.jetbrains.kotlin:kotlin-stdlib:{strictly 1.0.10} -> 2.1.10 (c)"
	actual, err := parseDependencyLine(line)
	if err != nil {
		t.Fatalf("ParseDependency returned error: %v", err)
	}

	expected := ParsedDependency{
		Dependency: Dependency{
			GroupID:          "org.jetbrains.kotlin",
			ArtifactID:       "kotlin-stdlib",
			Version:          "2.1.10",
			RequestedVersion: "1.0.10",
		},
		Level:      1,
		IsASummary: true,
	}
	if !actual.IsEquals(expected) {
		t.Errorf("Parsed dependency does not match expected structure.\nGot: %#v\nWnt: %#v", expected, actual)
	}
}

func TestParseDependency_SummaryLine2(t *testing.T) {
	line := "+--- androidx.compose.ui:ui-tooling-preview -> 1.8.0-beta02"
	actual, err := parseDependencyLine(line)
	if err != nil {
		t.Fatalf("ParseDependency returned error: %v", err)
	}

	expected := ParsedDependency{
		Dependency: Dependency{
			GroupID:          "androidx.compose.ui",
			ArtifactID:       "ui-tooling-preview",
			Version:          "1.8.0-beta02",
			RequestedVersion: "",
		},
		Level:      1,
		IsASummary: true,
	}
	if !actual.IsEquals(expected) {
		t.Errorf("Parsed dependency does not match expected structure.\nGot: %#v\nWant: %#v", actual, expected)
	}
}

func TestParseDependency_Project(t *testing.T) {
	line := "+--- project :feature:interests"
	actual, err := parseDependencyLine(line)
	if err != nil {
		t.Fatalf("ParseDependency returned error: %v", err)
	}

	expected := ParsedDependency{
		Dependency: Dependency{
			GroupID:          "",
			ArtifactID:       ":feature:interests",
			Version:          "",
			RequestedVersion: "",
			IsAModule:        true,
		},
		Level:      1,
		IsASummary: false,
	}
	if !actual.IsEquals(expected) {
		t.Errorf("Parsed dependency does not match expected structure.\nGot: %#v\nWant: %#v", actual, expected)
	}
}
