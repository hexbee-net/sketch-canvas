package canvas

import (
	"encoding/json"

	"golang.org/x/xerrors"
)

type Rectangle struct {
	Origin Point `json:"origin"`
	Width  uint  `json:"width"`
	Height uint  `json:"height"`
}

func (c *Rectangle) MarshalBinary() (data []byte, err error) {
	data, err = json.Marshal(c)
	if err != nil {
		return data, xerrors.Errorf("failed to marshal point to json: %w", err)
	}

	return data, nil
}
