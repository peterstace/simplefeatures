package simplefeatures

import "io"

// Line is a LineString with exactly two points.
type Line struct {
	ls LineString
}

// NewLine creates a line given the two points that define it.
func NewLine(p1, p2 Point) Line {
	ls, err := NewLineString([]Point{p1, p2})
	if err != nil {
		// Cannot panic because the size of the input is controlled.
		panic(err)
	}
	return Line{ls}
}

func (n Line) AsText(w io.Writer) error {
	return nil
}
