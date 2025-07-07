package android

import (
	"fmt"
	"lampa/internal/out"
	"os"
	"path"
	"strings"
)

type AndroidProject struct {
	RootDir string
}

func NewAndroidProject(dir string) AndroidProject {
	return AndroidProject{
		RootDir: dir,
	}
}

func (self AndroidProject) FindAabFile(buildVariant string) (string, error) {
	out.Info("Searching for AAB file")

	relativeBundleDir := path.Join("app", "build", "outputs", "bundle", buildVariant)
	bundleDir := path.Join(self.RootDir, relativeBundleDir)

	info, err := os.Stat(bundleDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("AAB directory `%s` does not exist", bundleDir)
		}
		return "", fmt.Errorf("error accessing AAB directory `%s`: %v", bundleDir, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("AAB directory `%s` is not a directory", bundleDir)
	}

	files, err := os.ReadDir(bundleDir)
	if err != nil {
		return "", fmt.Errorf("could not read AAB directory `%s`: %v", bundleDir, err)
	}
	var aabFilePath string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".aab") {
			aabFilePath = path.Join(bundleDir, file.Name())
			break
		}
	}
	if aabFilePath == "" {
		return "", fmt.Errorf("no AAB file found in `%s`", bundleDir)
	}

	return aabFilePath, nil
}
