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
|    |    +--- com.google.code.findbugs:jsr305:3.0.2
|    |    \--- javax.inject:javax.inject:1
`
	result, err := ParseTree(input)
	if err != nil {
		t.Fatalf("ParseTree returned error: %v", err)
	}

	// TODO check result
	_ = result
}
