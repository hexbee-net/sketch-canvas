package canvas

import (
	"encoding/json"

	"golang.org/x/xerrors"
)

type Point struct {
	X uint `json:"x"`
	Y uint `json:"y"`
}

func (c *Point) MarshalBinary() (data []byte, err error) {
	data, err = json.Marshal(c)
	if err != nil {
		return data, xerrors.Errorf("failed to marshal point to json: %w", err)
	}

	return data, nil
}
