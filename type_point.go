package simplefeatures

import (
	"strconv"
)

// Point is a 0-dimensional geometry, and represents a single location in a
// coordinate space.
type Point struct {
	coords Coordinates
}

// NewPoint creates a new point.
func NewPoint(x, y float64) (Point, error) {
	return NewPointFromCoords(Coordinates{XY{x, y}})
}

// TODO: NewPointZ, NewPointM, and NewPointZM ctors.

// NewPointFromCoords creates a new point gives its coordinates.
func NewPointFromCoords(c Coordinates) (Point, error) {
	err := c.Validate()
	return Point{coords: c}, err
}

func (p Point) AsText() []byte {
	return p.AppendWKT(nil)
}

func (p Point) AppendWKT(dst []byte) []byte {
	dst = append(dst, []byte("POINT")...)
	return p.appendWKTBody(dst)
}

func (p Point) appendWKTBody(dst []byte) []byte {
	dst = append(dst, '(')
	dst = strconv.AppendFloat(dst, p.coords.X, 'f', -1, 64)
	dst = append(dst, ' ')
	dst = strconv.AppendFloat(dst, p.coords.Y, 'f', -1, 64)
	return append(dst, ')')
}

func (p Point) IsSimple() bool {
	panic("not implemented")
}

func (p Point) Intersection(g Geometry) Geometry {
	return intersection(p, g)
}

func (p Point) IsEmpty() bool {
	return false
}

func (p Point) Dimension() int {
	return 0
}

func (p Point) Equals(other Geometry) bool {
	return equals(p, other)
}
