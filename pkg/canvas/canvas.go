package canvas

import (
	"encoding/json"

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

// Split returns the content of the canvas split into lines.
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

	if len(c.Data) == 0 {
		c.initData(fill[0])

		return nil
	}

	fillChar := fill[0]
	orgChar := c.Data[origin.Y*c.Width+origin.X]

	if orgChar == fillChar {
		return nil
	}

	dir := []struct{ x, y int }{
		{-1, 0},
		{0, 1},
		{1, 0},
		{0, -1},
	}

	var recFill func(x, y uint)
	recFill = func(x, y uint) {
		if c.get(x, y) == fillChar {
			return
		}

		c.set(x, y, fillChar)

		for _, d := range dir {
			if int(x)+d.x < 0 || int(y)+d.y < 0 {
				return
			}

			dx := uint(int(x) + d.x)
			dy := uint(int(y) + d.y)

			if dx < c.Width && dy < c.Height && c.get(dx, dy) == orgChar {
				recFill(dx, dy)
			}
		}
	}

	recFill(origin.X, origin.Y)

	return nil
}

func (c *Canvas) initData(v byte) {
	c.Data = make([]byte, c.Width*c.Height)
	for i := range c.Data {
		c.Data[i] = v
	}
}

func (c *Canvas) set(x, y uint, v byte) {
	c.Data[y*c.Width+x] = v
}

func (c *Canvas) get(x, y uint) byte {
	return c.Data[y*c.Width+x]
}
