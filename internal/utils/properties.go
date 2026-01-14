package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magiconair/properties"
)

// ParsePropertiesFile parses a Java properties file and returns a map of key-value pairs.
func ParsePropertiesFile(path string) (map[string]string, error) {
	props, err := properties.LoadFile(path, properties.UTF8)
	if err != nil {
		return nil, fmt.Errorf("failed to load properties file: %v", err)
	}

	return props.Map(), nil
}

// ExtractGradleVersion extracts the Gradle version from a distribution URL.
// Expected format: https://services.gradle.org/distributions/gradle-X.Y.Z-bin.zip
// or: https://services.gradle.org/distributions/gradle-X.Y.Z-all.zip
func ExtractGradleVersion(distributionUrl string) (string, error) {
	const prefix = "gradle-"

	// Find the prefix
	prefixIdx := strings.Index(distributionUrl, prefix)
	if prefixIdx == -1 {
		return "", fmt.Errorf("could not find 'gradle-' prefix in URL: %s", distributionUrl)
	}

	// Start after the prefix
	versionStart := prefixIdx + len(prefix)
	remaining := distributionUrl[versionStart:]

	// Find the suffix (-bin.zip or -all.zip)
	var suffixIdx int
	if idx := strings.Index(remaining, "-bin.zip"); idx != -1 {
		suffixIdx = idx
	} else if idx := strings.Index(remaining, "-all.zip"); idx != -1 {
		suffixIdx = idx
	} else {
		return "", fmt.Errorf("could not find '-bin.zip' or '-all.zip' suffix in URL: %s", distributionUrl)
	}

	// Extract and clean up the version
	version := remaining[:suffixIdx]
	version = strings.TrimSpace(version)

	if version == "" {
		return "", fmt.Errorf("extracted empty version from URL: %s", distributionUrl)
	}

	return version, nil
}

// ExtractJavaMajorVersion extracts the major version number from various Java version formats.
// Supports formats:
// - "17" -> "17"
// - "17.0.10" -> "17"
// - "temurin64-17.0.10" -> "17" (jenv style)
func ExtractJavaMajorVersion(versionStr string) (string, error) {
	versionStr = strings.TrimSpace(versionStr)
	if versionStr == "" {
		return "", fmt.Errorf("empty version string")
	}

	// Handle jenv style: "temurin64-17.0.10"
	if dashIdx := strings.Index(versionStr, "-"); dashIdx != -1 {
		versionStr = versionStr[dashIdx+1:]
	}

	// Extract major version (first part before '.')
	if dotIdx := strings.Index(versionStr, "."); dotIdx != -1 {
		versionStr = versionStr[:dotIdx]
	}

	// Validate it's a number
	versionStr = strings.TrimSpace(versionStr)
	if versionStr == "" {
		return "", fmt.Errorf("could not extract version number")
	}

	return versionStr, nil
}

// ParseJavaVersionFromToolVersions parses Java version from .tool-versions file.
// Expected format: "java@17" or "java 17"
// Returns error for "java@latest"
func ParseJavaVersionFromToolVersions(projectDir string) (string, error) {
	toolVersionsPath := filepath.Join(projectDir, ".tool-versions")
	if !FileExists(toolVersionsPath) {
		return "", nil
	}

	content, err := os.ReadFile(toolVersionsPath)
	if err != nil {
		return "", nil
	}

	// Search through all lines to find java entry
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse "java@17" or "java 17" format
		var versionStr string
		if strings.Contains(line, "@") {
			parts := strings.SplitN(line, "@", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[0]) == "java" {
				versionStr = strings.TrimSpace(parts[1])
			}
		} else {
			parts := strings.Fields(line)
			if len(parts) >= 2 && parts[0] == "java" {
				versionStr = parts[1]
			}
		}

		if versionStr == "" {
			continue
		}

		// Error on "latest"
		if versionStr == "latest" {
			return "", fmt.Errorf("'latest' is not supported for Java version in .tool-versions")
		}

		return ExtractJavaMajorVersion(versionStr)
	}

	return "", nil
}

// ParseJavaVersionFromJavaVersion parses Java version from .java-version file.
// Supports formats:
// - "17"
// - "17.0.10"
// - "temurin64-17.0.10" (jenv style)
func ParseJavaVersionFromJavaVersion(projectDir string) (string, error) {
	javaVersionPath := filepath.Join(projectDir, ".java-version")
	if !FileExists(javaVersionPath) {
		return "", nil
	}

	versionStr, err := ReadFirstLine(javaVersionPath)
	if err != nil {
		return "", nil
	}

	return ExtractJavaMajorVersion(versionStr)
}
