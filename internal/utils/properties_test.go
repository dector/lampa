package utils

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
