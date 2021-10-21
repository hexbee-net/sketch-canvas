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

// Split returns the content of the canvas split into lines
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

	ov := c.Data[origin.Y*c.Width+origin.X]
	fillChar := fill[0]

	if ov == fillChar {
		return nil
	}

	type stackEntry struct {
		xl, xr, y, dy int
	}

	stack := []stackEntry{
		{int(origin.X), int(origin.X), int(origin.Y), 1},      // needed in some cases
		{int(origin.X), int(origin.X), int(origin.Y) + 1, -1}, // seed segment (popped 1st)
	}

	const wx0 = 0
	const wy0 = 0
	wx1 := int(c.Width) - 1
	wy1 := int(c.Height) - 1

	push := func(xl, xr, y, dy int) {
		if y+dy >= wy0 && y+dy <= wy1 {
			stack = append(stack, stackEntry{xl: xl, xr: xr - 1, y: y, dy: dy})
		}
	}
	pop := func() (xl, xr, y, dy int) {
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		return v.xl, v.xr, v.y + v.dy, v.dy
	}
	set := func(x, y int) {
		c.Data[y*int(c.Width)+x] = fillChar
	}
	get := func(x, y int) byte {
		return c.Data[y*int(c.Width)+x]
	}

	dbgSplit("", c)

	for len(stack) > 0 {
		x1, x2, y, dy := pop() // pop segment off stack and fill a neighboring scan line

		// segment of scan line y-dy for x1<=x<=x2 was previously filled,
		// now explore adjacent pixels in scan line y

		x := x1

		for x >= wx0 && get(x, y) == ov {
			set(x, y)
			x--
		}

		var l int

		if x >= x1 {
			x++
			for x <= x2 && get(x, y) != ov {
				x++
			}

			l = x
		} else {
			l = x + 1
			if l < x1 { // leak on left?
				push(l, x1-1, y, -dy)
			}
			x = x1 + 1
		}

		for {
			for x <= wx1 && get(x, y) == ov {
				set(x, y)
				x++
			}
			push(l, x-1, y, dy)

			if x > x2+1 {
				push(x2+1, x-1, y, -dy) // leak on right?
			}

			x++
			for x <= x2 && get(x, y) != ov {
				x++
			}

			l = x

			if x > x2 {
				break
			}
		}
	}

	return nil
}

func (c *Canvas) initData(v byte) {
	c.Data = make([]byte, c.Width*c.Height)
	for i := range c.Data {
		c.Data[i] = v
	}
}
