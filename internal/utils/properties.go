package utils

import (
	"fmt"
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
