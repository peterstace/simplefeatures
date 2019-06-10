package simplefeatures

import (
	"errors"
	"strconv"
)

// Line is a LineString with exactly two distinct points.
type Line struct {
	a, b Coordinates
}

// NewLine creates a line segment given the coordinates of its two endpoints.
func NewLine(a, b Coordinates) (Line, error) {
	if err := a.Validate(); err != nil {
		return Line{}, err
	}
	if err := b.Validate(); err != nil {
		return Line{}, err
	}
	if a.XY == b.XY {
		return Line{}, errors.New("line endpoints must be distinct")
	}
	return Line{a, b}, nil
}

func (n Line) AsText() []byte {
	return n.AppendWKT(nil)
}

func (n Line) AppendWKT(dst []byte) []byte {
	dst = []byte("LINESTRING(")
	dst = strconv.AppendFloat(dst, n.a.X, 'f', -1, 64)
	dst = append(dst, ' ')
	dst = strconv.AppendFloat(dst, n.a.Y, 'f', -1, 64)
	dst = append(dst, ',')
	dst = strconv.AppendFloat(dst, n.b.X, 'f', -1, 64)
	dst = append(dst, ' ')
	dst = strconv.AppendFloat(dst, n.b.Y, 'f', -1, 64)
	return append(dst, ')')
}

func (n Line) IsSimple() bool {
	return true
}
