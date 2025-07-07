package bundles

import (
	"archive/zip"
	"fmt"
	"io"
)

func ReadFileFromAab(
	aabFile string,
	internalPath string,
) (*[]byte, error) {
	reader, err := zip.OpenReader(aabFile)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var data []byte
	for _, file := range reader.File {
		if file.Name == internalPath {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			data, err = io.ReadAll(rc)
			if err != nil {
				return nil, err
			}
			break
		}
	}

	if data == nil {
		return nil, fmt.Errorf("file %s not found", internalPath)
	}

	return &data, nil
}
