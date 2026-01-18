package androidenv

import (
	"os"
	"testing"
)

func TestParsePropertiesFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected map[string]string
	}{
		{
			name: "valid properties file",
			content: `# Comment line
key1=value1
key2=value2
# Another comment
key3=value3`,
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "gradle wrapper properties",
			content: `distributionUrl=https\://services.gradle.org/distributions/gradle-7.5-bin.zip`,
			expected: map[string]string{
				"distributionUrl": "https://services.gradle.org/distributions/gradle-7.5-bin.zip",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpFile, err := os.CreateTemp("", "test-*.properties")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			// Write test content
			if err := os.WriteFile(tmpFile.Name(), []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Parse properties
			props, err := ParsePropertiesFile(tmpFile.Name())
			if err != nil {
				t.Fatalf("ParsePropertiesFile() error = %v", err)
			}

			// Verify expected properties
			for key, expectedValue := range tt.expected {
				if actualValue, ok := props[key]; !ok {
					t.Errorf("Missing expected key: %s", key)
				} else if actualValue != expectedValue {
					t.Errorf("For key %s: got %s, want %s", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestExtractGradleVersion(t *testing.T) {
	tests := []struct {
		name          string
		distributionUrl string
		want          string
		wantErr       bool
	}{
		{
			name:            "Standard bin distribution",
			distributionUrl: "https://services.gradle.org/distributions/gradle-7.5-bin.zip",
			want:            "7.5",
			wantErr:         false,
		},
		{
			name:            "All distribution",
			distributionUrl: "https://services.gradle.org/distributions/gradle-8.0.2-all.zip",
			want:            "8.0.2",
			wantErr:         false,
		},
		{
			name:            "Version with RC suffix",
			distributionUrl: "https://services.gradle.org/distributions/gradle-8.1-rc-1-bin.zip",
			want:            "8.1-rc-1",
			wantErr:         false,
		},
		{
			name:            "Invalid URL without gradle prefix",
			distributionUrl: "https://example.com/gradle.zip",
			wantErr:         true,
		},
		{
			name:            "Invalid URL without version",
			distributionUrl: "https://services.gradle.org/distributions/gradle-bin.zip",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := ExtractGradleVersion(tt.distributionUrl)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none for URL: %s", tt.distributionUrl)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if version != tt.want {
					t.Errorf("Expected version %s, got %s", tt.want, version)
				}
			}
		})
	}
}

func TestExtractJavaMajorVersion(t *testing.T) {
	tests := []struct {
		name       string
		versionStr string
		want       string
		wantErr    bool
	}{
		{
			name:       "Simple major version",
			versionStr: "17",
			want:       "17",
			wantErr:    false,
		},
		{
			name:       "Version with minor and patch",
			versionStr: "17.0.10",
			want:       "17",
			wantErr:    false,
		},
		{
			name:       "jenv style with distribution prefix",
			versionStr: "temurin64-17.0.10",
			want:       "17",
			wantErr:    false,
		},
		{
			name:       "jenv style with simple version",
			versionStr: "temurin64-17",
			want:       "17",
			wantErr:    false,
		},
		{
			name:       "Version with whitespace",
			versionStr: "  17.0.10  ",
			want:       "17",
			wantErr:    false,
		},
		{
			name:       "Empty string",
			versionStr: "",
			wantErr:    true,
		},
		{
			name:       "Only whitespace",
			versionStr: "   ",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := ExtractJavaMajorVersion(tt.versionStr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none for version: %s", tt.versionStr)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if version != tt.want {
					t.Errorf("Expected version %s, got %s", tt.want, version)
				}
			}
		})
	}
}

func TestParseJavaVersionFromToolVersions(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name:    "java@version format",
			content: "java@17",
			want:    "17",
			wantErr: false,
		},
		{
			name:    "java@version with full version",
			content: "java@17.0.10",
			want:    "17",
			wantErr: false,
		},
		{
			name:    "java space version format",
			content: "java 17",
			want:    "17",
			wantErr: false,
		},
		{
			name:    "java@latest should error",
			content: "java@latest",
			wantErr: true,
		},
		{
			name:    "empty file",
			content: "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "comment lines only",
			content: "# This is a comment\n# Another comment",
			want:    "",
			wantErr: false,
		},
		{
			name:    "multiple tools with java",
			content: "nodejs 18.0.0\njava@17",
			want:    "17",
			wantErr: false,
		},
		{
			name:    "java with comments",
			content: "# Comment\njava 17\nnodejs 18",
			want:    "17",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "test-project-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create .tool-versions file
			if tt.content != "" {
				toolVersionsPath := tmpDir + "/.tool-versions"
				if err := os.WriteFile(toolVersionsPath, []byte(tt.content), 0644); err != nil {
					t.Fatalf("Failed to write .tool-versions file: %v", err)
				}
			}

			// Parse Java version
			version, err := ParseJavaVersionFromToolVersions(tmpDir)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if version != tt.want {
					t.Errorf("Expected version %s, got %s", tt.want, version)
				}
			}
		})
	}
}

func TestParseJavaVersionFromJavaVersion(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name:    "simple major version",
			content: "17",
			want:    "17",
			wantErr: false,
		},
		{
			name:    "full version",
			content: "17.0.10",
			want:    "17",
			wantErr: false,
		},
		{
			name:    "jenv style",
			content: "temurin64-17.0.10",
			want:    "17",
			wantErr: false,
		},
		{
			name:    "version with whitespace",
			content: "  17.0.10  ",
			want:    "17",
			wantErr: false,
		},
		{
			name:    "empty file",
			content: "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "comment lines only",
			content: "# This is a comment",
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "test-project-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create .java-version file
			if tt.content != "" {
				javaVersionPath := tmpDir + "/.java-version"
				if err := os.WriteFile(javaVersionPath, []byte(tt.content), 0644); err != nil {
					t.Fatalf("Failed to write .java-version file: %v", err)
				}
			}

			// Parse Java version
			version, err := ParseJavaVersionFromJavaVersion(tmpDir)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if version != tt.want {
					t.Errorf("Expected version %s, got %s", tt.want, version)
				}
			}
		})
	}
}
