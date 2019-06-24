package simplefeatures

import (
	"fmt"
	"strconv"
)

// Line is a single line segment between two points.
//
// Its assertions are:
//
// 1. The two points must be distinct.
type Line struct {
	a, b Coordinates
}

// NewLine creates a line segment given the coordinates of its two endpoints.
func NewLine(a, b Coordinates) (Line, error) {
	if a.XY.Equals(b.XY) {
		return Line{}, fmt.Errorf("line endpoints must be distinct: %v", a.XY)
	}
	return Line{a, b}, nil
}

func (n Line) AsText() []byte {
	return n.AppendWKT(nil)
}

func (n Line) AppendWKT(dst []byte) []byte {
	dst = append(dst, []byte("LINESTRING(")...)
	dst = strconv.AppendFloat(dst, n.a.X.AsFloat(), 'f', -1, 64)
	dst = append(dst, ' ')
	dst = strconv.AppendFloat(dst, n.a.Y.AsFloat(), 'f', -1, 64)
	dst = append(dst, ',')
	dst = strconv.AppendFloat(dst, n.b.X.AsFloat(), 'f', -1, 64)
	dst = append(dst, ' ')
	dst = strconv.AppendFloat(dst, n.b.Y.AsFloat(), 'f', -1, 64)
	return append(dst, ')')
}

func (n Line) IsSimple() bool {
	return true
}

func (n Line) Intersection(g Geometry) Geometry {
	return intersection(n, g)
}

func (n Line) IsEmpty() bool {
	return false
}

func (n Line) Dimension() int {
	return 1
}

func (n Line) Equals(other Geometry) bool {
	return equals(n, other)
}

func (n Line) Envelope() (Envelope, bool) {
	return NewEnvelope(n.a.XY, n.b.XY), true
}
