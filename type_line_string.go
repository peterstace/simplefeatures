package simplefeatures

import (
	"errors"
	"strconv"
)

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
	// TODO: check empties are as appropriate
	return LineString{pts}, nil
}

// NewLineStringFromCoords creates a line string from the coordinates defining
// its points.
func NewLineStringFromCoords(coords []Coordinates) (LineString, error) {
	var pts []Point
	for _, c := range coords {
		pt, err := NewPointFromCoords(c)
		if err != nil {
			return LineString{}, err
		}
		pts = append(pts, pt)
	}
	return NewLineString(pts)
}

func (s LineString) AsText() []byte {
	return s.AppendWKT(nil)
}

func (s LineString) AppendWKT(dst []byte) []byte {
	dst = append(dst, []byte("LINESTRING")...)
	if len(s.pts) == 0 {
		dst = append(dst, ' ')
	}
	return s.appendWKTBody(dst)
}

func (s LineString) appendWKTBody(dst []byte) []byte {
	if len(s.pts) == 0 {
		return append(dst, []byte("EMPTY")...)
	}
	dst = append(dst, '(')
	for i, pt := range s.pts {
		dst = strconv.AppendFloat(dst, pt.x, 'f', -1, 64)
		dst = append(dst, ' ')
		dst = strconv.AppendFloat(dst, pt.y, 'f', -1, 64)
		if i != len(s.pts)-1 {
			dst = append(dst, ',')
		}
	}
	return append(dst, ')')
}

// IsSimple returns true iff the curve defined by the LineString doesn't pass
// through the same point twice (with the possible exception of the two
// endpoints).
func (s LineString) IsSimple() bool {
	// TODO
	return true
}

func (s LineString) IsClosed() bool {
	if len(s.pts) == 0 {
		return false
	}
	return s.pts[0] == s.pts[len(s.pts)-1]
}
