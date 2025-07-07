package bundles

import (
	"github.com/dector/lampa/pkg/bundles/model"
	"github.com/dector/lampa/pkg/bundles/parsers"
)

func LoadResources(aabFile string) (*model.ResourceTable, error) {
	content, err := ReadFileFromAab(
		aabFile,
		"base/resources.pb",
	)
	if err != nil {
		return nil, err
	}

	res, err := parsers.ParseResources(content)
	return res, err
}
