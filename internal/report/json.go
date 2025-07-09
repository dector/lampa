package report

import "encoding/json"

func (self StatsReport) ToJsonBytes() ([]byte, error) {
	return json.MarshalIndent(self, "", "  ")
}

func (self StatsReport) ToJsonString() (string, error) {
	jsonBytes, err := self.ToJsonBytes()
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (self DiffReport) ToJsonBytes() ([]byte, error) {
	return json.MarshalIndent(self, "", "  ")
}

func (self DiffReport) ToJsonString() (string, error) {
	jsonBytes, err := self.ToJsonBytes()
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
