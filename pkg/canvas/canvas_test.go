package canvas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanvas_DrawRect(t *testing.T) {
	c := Canvas{
		Width:  10,
		Height: 10,
	}

	rect := &Rectangle{
		Origin: Point{
			X: 2,
			Y: 3,
		},
		Width:  4,
		Height: 5,
	}

	err := c.DrawRect(rect, "#", "*")

	expected :=
		"" +
			"----------" +
			"----------" +
			"----------" +
			"--****----" +
			"--*##*----" +
			"--*##*----" +
			"--*##*----" +
			"--****----" +
			"----------" +
			"----------" +
			""

	for _, l := range c.Split() {
		println(l)
	}

	assert.NoError(t, err)
	assert.Equal(t, expected, string(c.Data))
}

func TestCanvas_Split(t *testing.T) {
	c := Canvas{
		Width:  4,
		Height: 4,
		Data:   []byte("12345678abcdefgh"),
	}

	lines := c.Split()
	assert.Equal(t, []string{"1234", "5678", "abcd", "efgh"}, lines)
}
