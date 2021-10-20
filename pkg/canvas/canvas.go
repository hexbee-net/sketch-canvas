package canvas

import (
	"encoding/json"

	"github.com/apex/log"
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

func (c *Canvas) DrawRect(rect *Rectangle, fill string, outline string) error {
	log.
		WithField("rect", rect).
		WithField("fill", fill).
		WithField("outline", outline).
		Error("TODO - Canvas.DrawRect")

	return nil
}

func (c *Canvas) FloodFill(origin *Point, fill string) error {
	log.
		WithField("origin", origin).
		WithField("fill", fill).
		Error("TODO - Canvas.FloodFill")

	return nil
}
