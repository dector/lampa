package proto

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

func ParseXml(data []byte) (*XmlNode, error) {
	result := &XmlNode{}

	err := proto.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func ToText(xml *XmlNode) string {
	var result string
	if elem := xml.GetElement(); elem != nil {
		result += fmt.Sprintf("<%s", elem.GetName())
		for _, attr := range elem.GetAttribute() {
			result += fmt.Sprintf(" %s=\"%s\"", attr.GetName(), attr.GetValue())
		}
		result += ">\n"
		for _, child := range elem.GetChild() {
			if childElem := child.GetElement(); childElem != nil {
				result += fmt.Sprintf("  <%s", childElem.GetName())
				for _, attr := range childElem.GetAttribute() {
					result += fmt.Sprintf(" %s=\"%s\"", attr.GetName(), attr.GetValue())
				}
				result += "/>\n"
			}
		}
		result += fmt.Sprintf("</%s>\n", elem.GetName())
	}
	return result
}
