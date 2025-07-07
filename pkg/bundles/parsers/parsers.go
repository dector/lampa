package parsers

import (
	"lampa/pkg/bundles/model"

	"google.golang.org/protobuf/proto"
)

func ParseXml(data *[]byte) (*model.XmlNode, error) {
	result := &model.XmlNode{}

	err := proto.Unmarshal(*data, result)
	return result, err
}

func ParseResources(data *[]byte) (*model.ResourceTable, error) {
	result := &model.ResourceTable{}

	err := proto.Unmarshal(*data, result)
	return result, err
}
