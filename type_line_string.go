package gatig

import "errors"

// LineString is a curve with linear interpolation between points. Each
// consecutive pair of points defines a line segment.
type LineString struct {
	pts []Point
}

// NewLineString gives the LineString specified by its points. The number of
// points must be either zero or greater than 1, otherwise an error is
// returned.
func NewLineString(pts []Point) (LineString, error) {
	if len(pts) == 1 {
		return LineString{}, errors.New("line strings cannot have 1 point")
	}
	return LineString{pts}, nil
}

func (s LineString) IsSimple() bool {
	// TODO
	return false
}

func (s LineString) IsClosed() bool {
	// TODO
	return false
}
