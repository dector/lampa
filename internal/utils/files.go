package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func TryResolveFsPath(s string) string {
	if s == "" {
		return ""
	}

	s = filepath.Clean(s)
	if strings.HasPrefix(s, "~") {
		s = filepath.Join(os.Getenv("HOME"), s[1:])
	}

	abs, err := filepath.Abs(s)
	if err == nil {
		s = abs
	}

	return s
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func EnsureParentDirExists(path string) error {
	parentDir := filepath.Dir(path)

	if FileExists(parentDir) {
		return nil
	}
	if IsDir(parentDir) {
		return nil
	}

	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("could not create parent directory for report file: %v", err)
	}

	return nil
}

// ReadFirstLine reads the first non-empty line from a file.
func ReadFirstLine(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			return line, nil
		}
	}

	return "", fmt.Errorf("no non-empty lines found in file")
}
