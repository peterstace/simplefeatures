package simplefeatures

import (
	"errors"
	"strconv"
)

// LineString is a curve defined by linear interpolation between a finite set
// of points. Each consecutive pair of points defines a line segment. It must
// contain at least 2 distinct points.
type LineString struct {
	pts []Point
}

// NewLineString gives the LineString specified by its points.
func NewLineString(pts []Point) (LineString, error) {
	// Empty line string.
	if len(pts) == 0 {
		return LineString{}, nil
	}

	// Must have at least 2 distinct points.
	type xy struct{ x, y float64 }
	pointSet := make(map[xy]struct{})
	for _, pt := range pts {
		pointSet[xy{pt.x, pt.y}] = struct{}{}
		if len(pointSet) == 2 {
			break
		}
	}
	if len(pointSet) < 2 {
		return LineString{}, errors.New("LineString must contain either zero or at least two distinct points")
	}

	// TODO: cannot have empty points as part of a line string
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
