package geom

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
func NewPointXY(xy XY, _ ...ConstructorOption) Point {
	return NewPointC(Coordinates{XY: xy})
}

// NewPointF creates a new point from float64 x and y values.
func NewPointF(x, y float64, _ ...ConstructorOption) Point {
	return NewPointXY(XY{x, y})
}

// NewPointC creates a new point gives its Coordinates.
func NewPointC(c Coordinates, _ ...ConstructorOption) Point {
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
	dst = appendFloat(dst, p.coords.X)
	dst = append(dst, ' ')
	dst = appendFloat(dst, p.coords.Y)
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
	return wkbAsBytes(p)
}

func (p Point) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypePoint)
	marsh.writeFloat64(p.coords.X)
	marsh.writeFloat64(p.coords.Y)
	return marsh.err
}

// ConvexHull returns the convex hull of this Point, which is always the same
// point.
func (p Point) ConvexHull() Geometry {
	return convexHull(p)
}

func (p Point) convexHullPointSet() []XY {
	return []XY{p.XY()}
}

func (p Point) MarshalJSON() ([]byte, error) {
	return marshalGeoJSON("Point", p.Coordinates())
}

// TransformXY transforms this Point into another Point according to fn.
func (p Point) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (Geometry, error) {
	coords := p.Coordinates()
	coords.XY = fn(coords.XY)
	return NewPointC(coords, opts...), nil
}

// EqualsExact checks if this Point is exactly equal to another Point.
func (p Point) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	o, ok := other.(Point)
	if !ok {
		return false
	}
	eq := newEqualsExactOptionSet(opts).eq
	return eq(p.XY(), o.XY())
}
