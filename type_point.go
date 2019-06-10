package simplefeatures

import (
	"errors"
	"math"
	"strconv"
)

// Point is a 0-dimensional geometry, and represents a single location in a
// coordinate space.
type Point struct {
	x, y  float64
	empty bool
}

// NewPoint creates a new point.
func NewPoint(x, y float64) (Point, error) {
	if math.IsNaN(x) || math.IsNaN(y) {
		return Point{}, errors.New("coordinate is NaN")
	}
	if math.IsInf(x, 0) || math.IsInf(y, 0) {
		return Point{}, errors.New("coordinate is Inf")
	}
	return Point{x, y, false}, nil
}

// NewEmptyPoint creates an empty point.
func NewEmptyPoint() Point {
	return Point{empty: true}
}

// NewPointFromCoords creates a new point gives its coordinates.
func NewPointFromCoords(c Coordinates) (Point, error) {
	return NewPoint(c.X, c.Y)
}

// NewPointFromOptionalCoords creates a new point given its coordinates (which
// may be empty).
func NewPointFromOptionalCoords(c OptionalCoordinates) (Point, error) {
	if c.Empty {
		return NewEmptyPoint(), nil
	}
	return NewPoint(c.Value.X, c.Value.Y)
}

func (p Point) AsText() []byte {
	return p.AppendWKT(nil)
}

func (p Point) AppendWKT(dst []byte) []byte {
	dst = append(dst, []byte("POINT")...)
	if p.empty {
		dst = append(dst, ' ')
	}
	return p.appendWKTBody(dst)
}

func (p Point) appendWKTBody(dst []byte) []byte {
	if p.empty {
		return append(dst, []byte("EMPTY")...)
	}
	dst = append(dst, '(')
	dst = strconv.AppendFloat(dst, p.x, 'f', -1, 64)
	dst = append(dst, ' ')
	dst = strconv.AppendFloat(dst, p.y, 'f', -1, 64)
	return append(dst, ')')
}

func (p Point) IsSimple() bool {
	panic("not implemented")
}
