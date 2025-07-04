package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func TryResolveFsPath(s string) string {
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
