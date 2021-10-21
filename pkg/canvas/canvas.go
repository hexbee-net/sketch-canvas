package canvas

import (
	"encoding/json"

	"github.com/apex/log"
	"golang.org/x/xerrors"
)

const backgroundChar = '-'

type Canvas struct {
	Name   string `json:"name,omitempty"`
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
	Data   []byte `json:"data,omitempty"`
}

func (c *Canvas) MarshalBinary() (data []byte, err error) {
	data, err = json.Marshal(c)
	if err != nil {
		return data, xerrors.Errorf("failed to marshal canvas to json: %w", err)
	}

	return data, nil
}

func (c *Canvas) DrawRect(rect *Rectangle, fill string, outline string) error {
	if rect.Origin.X > c.Width || rect.Origin.Y > c.Height {
		return PointOutOfBound
	}

	if rect.Origin.X+rect.Width > c.Width || rect.Origin.Y+rect.Height > c.Height {
		return ObjectTooLarge
	}

	if len(fill) > 1 || len(outline) > 1 {
		return BadPattern
	}

	if rect.Width == 0 || rect.Height == 0 {
		return nil
	}

	// If the data was never initialized, do it now.
	if len(c.Data) == 0 {
		c.initData(backgroundChar)
	}

	// We make a copy of the rectangle for the filling operation.
	// We will adjust the size of the filling rectangle if we draw an outline.
	fillOrigin := rect.Origin
	fillWidth := rect.Width
	fillHeight := rect.Height

	if outline != "" {
		outlineChar := outline[0]

		// Start with the horizontal lines
		upperOffset := rect.Origin.Y * c.Width
		lowerOffset := (rect.Origin.Y + rect.Height - 1) * c.Width

		for x := rect.Origin.X + rect.Width - 1; x >= rect.Origin.X; x-- {
			c.Data[upperOffset+x] = outlineChar
			c.Data[lowerOffset+x] = outlineChar
		}

		// Then draw the vertical lines.
		// We can skip the start and end chars since we just drew them with the horizontal lines.
		leftOffset := rect.Origin.X
		rightOffset := rect.Origin.X + rect.Width - 1

		for y := rect.Origin.Y + rect.Height - 1 - 1; y > rect.Origin.Y; y-- {
			yOffset := y * c.Width
			c.Data[yOffset+leftOffset] = outlineChar
			c.Data[yOffset+rightOffset] = outlineChar
		}

		// Shrink the fill by one char on each side to avoid overwriting the outline.
		fillOrigin.X++
		fillOrigin.Y++

		fillWidth -= 2
		fillHeight -= 2
	}

	if fill != "" {
		fillChar := fill[0]

		for y := fillOrigin.Y + fillHeight - 1; y >= fillOrigin.Y; y-- {
			yOffset := y * c.Width

			for x := fillOrigin.X + fillWidth - 1; x >= fillOrigin.X; x-- {
				c.Data[yOffset+x] = fillChar
			}
		}
	}

	return nil
}

func (c *Canvas) FloodFill(origin *Point, fill string) error {
	if origin.X > c.Width || origin.Y > c.Height {
		return PointOutOfBound
	}

	if len(fill) > 1 {
		return BadPattern
	}

	log.
		WithField("origin", origin).
		WithField("fill", fill).
		Error("TODO - Canvas.FloodFill")

	return nil
}
func (c *Canvas) initData(v byte) {
	c.Data = make([]byte, c.Width*c.Height)
	for i := range c.Data {
		c.Data[i] = v
	}
}

func (c *Canvas) Split() []string {
	if len(c.Data) == 0 {
		c.initData(backgroundChar)
	}

	data := make([]string, 0, c.Height)

	var y uint
	for y = 0; y < c.Height; y++ {
		start := y * c.Width
		line := c.Data[start : start+c.Width]
		data = append(data, string(line))
	}

	return data
}
