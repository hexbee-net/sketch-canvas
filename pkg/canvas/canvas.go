package canvas

import (
	"encoding/json"

	"golang.org/x/xerrors"
)

type Canvas struct {
	Name   string `json:"name,omitempty"`
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
	Data   string `json:"data,omitempty"`
}

func (c *Canvas) MarshalBinary() (data []byte, err error) {
	data, err = json.Marshal(c)
	if err != nil {
		return data, xerrors.Errorf("failed to marshal canvas to json: %w", err)
	}

	return data, nil
}
