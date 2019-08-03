package simplefeatures

import (
	"database/sql/driver"
	"io"
)

// Point is a 0-dimensional geometry, and represents a single location in a
// coordinate space.
//
// There aren't any assertions about what constitutes a valid point, other than
// that it must be able to be represented by an XY.
type Point struct {
	coords Coordinates
}

// NewPointXY creates a new point from an XY.
func NewPointXY(xy XY) Point {
	return NewPointS(xy.X, xy.Y)
}

// NewPointS creates a new point from scalar values.
func NewPointS(x, y Scalar) Point {
	return NewPointC(Coordinates{XY{x, y}})
}

// NewPointC creates a new point gives its Coordinates.
func NewPointC(c Coordinates) Point {
	return Point{coords: c}
}

// XY gives the XY location of the point.
func (p Point) XY() XY {
	return p.coords.XY
}

// Coordinates returns the coordinates of the point.
func (p Point) Coordinates() Coordinates {
	return p.coords
}

func (p Point) AsText() string {
	return string(p.AppendWKT(nil))
}

func (p Point) AppendWKT(dst []byte) []byte {
	dst = append(dst, []byte("POINT")...)
	return p.appendWKTBody(dst)
}

func (p Point) appendWKTBody(dst []byte) []byte {
	dst = append(dst, '(')
	dst = p.coords.X.appendAsFloat(dst)
	dst = append(dst, ' ')
	dst = p.coords.Y.appendAsFloat(dst)
	return append(dst, ')')
}

func (p Point) IsSimple() bool {
	return true
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

func (p Point) Envelope() (Envelope, bool) {
	return NewEnvelope(p.coords.XY), true
}

func (p Point) Boundary() Geometry {
	return NewGeometryCollection(nil)
}

func (p Point) Value() (driver.Value, error) {
	return p.AsText(), nil
}

func (p Point) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypePoint)
	marsh.writeFloat64(p.coords.X.AsFloat())
	marsh.writeFloat64(p.coords.Y.AsFloat())
	return marsh.err
}

func (p Point) MarshalJSON() ([]byte, error) {
	return marshalGeoJSON("Point", p.Coordinates())
}
