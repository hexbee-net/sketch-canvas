package canvas

type Error string

func (s Error) Error() string {
	return string(s)
}

const (
	PointOutOfBound = Error("point out of bound")
	ObjectTooLarge  = Error("object too large")
	BadPattern      = Error("the drawing pattern is invalid")
)
