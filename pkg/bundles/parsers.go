package bundles

import "google.golang.org/protobuf/proto"

func ParseXml(data *[]byte) (*XmlNode, error) {
	result := &XmlNode{}

	err := proto.Unmarshal(*data, result)
	return result, err
}

func ParseResources(data *[]byte) (*ResourceTable, error) {
	result := &ResourceTable{}

	err := proto.Unmarshal(*data, result)
	return result, err
}
