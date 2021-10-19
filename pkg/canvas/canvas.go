package canvas

type Canvas struct {
	Name   string `json:"name,omitempty"`
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
	Data   string `json:"data,omitempty"`
}
