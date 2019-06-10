package simplefeatures

import (
	"strconv"
)

// Point is a 0-dimensional geometry, and can represent a single location in a
// coordinate space. It can be empty, and not represent any point.
type Point struct {
	empty  bool
	coords Coordinates
}

// NewPoint creates a new point.
func NewPoint(x, y float64) (Point, error) {
	return NewPointFromCoords(Coordinates{XY{x, y}})
}

// TODO: NewPointZ, NewPointM, and NewPointZM ctors.

// NewEmptyPoint creates an empty point.
func NewEmptyPoint() Point {
	return Point{empty: true}
}

// NewPointFromCoords creates a new point gives its coordinates.
func NewPointFromCoords(c Coordinates) (Point, error) {
	err := c.Validate()
	return Point{coords: c}, err
}

// NewPointFromOptionalCoords creates a new point given its coordinates (which
// may be empty).
func NewPointFromOptionalCoords(c OptionalCoordinates) (Point, error) {
	if c.Empty {
		return NewEmptyPoint(), nil
	}
	return NewPointFromCoords(c.Value)
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
	dst = strconv.AppendFloat(dst, p.coords.X, 'f', -1, 64)
	dst = append(dst, ' ')
	dst = strconv.AppendFloat(dst, p.coords.Y, 'f', -1, 64)
	return append(dst, ')')
}

func (p Point) IsSimple() bool {
	panic("not implemented")
}
