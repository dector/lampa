package report

import "encoding/json"

func (self Report) ToJsonBytes() ([]byte, error) {
	return json.MarshalIndent(self, "", "  ")
}

func (self Report) ToJsonString() (string, error) {
	jsonBytes, err := self.ToJsonBytes()
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
