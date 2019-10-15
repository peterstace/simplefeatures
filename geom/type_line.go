package geom

import (
	"database/sql/driver"
	"fmt"
	"io"
	"math"
)

// Line is a single line segment between two points.
//
// Its assertions are:
//
// 1. The two points must be distinct.
type Line struct {
	a, b Coordinates
}

// NewLineC creates a line segment given the Coordinates of its two endpoints.
func NewLineC(a, b Coordinates, opts ...ConstructorOption) (Line, error) {
	if doCheapValidations(opts) && a.XY.Equals(b.XY) {
		return Line{}, fmt.Errorf("line endpoints must be distinct: %v", a.XY)
	}
	return Line{a, b}, nil
}

// StartPoint gives the first point of the line.
func (n Line) StartPoint() Point {
	return NewPointC(n.a)
}

// EndPoint gives the second (last) point of the line.
func (n Line) EndPoint() Point {
	return NewPointC(n.b)
}

// NumPoints always returns 2.
func (Line) NumPoints() int {
	return 2
}

// PointN returns the first point when n is 0, and the second point when n is
// 1. It panics if n is any other value.
func (ln Line) PointN(n int) Point {
	switch n {
	case 0:
		return ln.StartPoint()
	case 1:
		return ln.EndPoint()
	default:
		panic("n must be 0 or 1")
	}
}

func (n Line) AsText() string {
	return string(n.AppendWKT(nil))
}

func (n Line) AppendWKT(dst []byte) []byte {
	dst = append(dst, []byte("LINESTRING(")...)
	dst = appendFloat(dst, n.a.X)
	dst = append(dst, ' ')
	dst = appendFloat(dst, n.a.Y)
	dst = append(dst, ',')
	dst = appendFloat(dst, n.b.X)
	dst = append(dst, ' ')
	dst = appendFloat(dst, n.b.Y)
	return append(dst, ')')
}

func (n Line) IsSimple() bool {
	return true
}

func (n Line) Intersection(g Geometry) (Geometry, error) {
	return intersection(n, g)
}

func (n Line) Intersects(g Geometry) (bool, error) {
	has, _, err := hasIntersection(n, g)
	return has, err
}

func (n Line) IsEmpty() bool {
	return false
}

func (n Line) Dimension() int {
	return 1
}

func (n Line) Equals(other Geometry) (bool, error) {
	return equals(n, other)
}

func (n Line) Envelope() (Envelope, bool) {
	return NewEnvelope(n.a.XY, n.b.XY), true
}

func (n Line) Boundary() Geometry {
	return NewMultiPoint([]Point{
		NewPointXY(n.a.XY),
		NewPointXY(n.b.XY),
	})
}

func (n Line) Value() (driver.Value, error) {
	return wkbAsBytes(n)
}

func (n Line) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeLineString)
	marsh.writeCount(2)
	marsh.writeFloat64(n.StartPoint().XY().X)
	marsh.writeFloat64(n.StartPoint().XY().Y)
	marsh.writeFloat64(n.EndPoint().XY().X)
	marsh.writeFloat64(n.EndPoint().XY().Y)
	return marsh.err
}

func (n Line) ConvexHull() Geometry {
	return convexHull(n)
}

func (n Line) convexHullPointSet() []XY {
	return []XY{
		n.StartPoint().XY(),
		n.EndPoint().XY(),
	}
}

func (n Line) MarshalJSON() ([]byte, error) {
	return marshalGeoJSON("LineString", n.Coordinates())
}

// Coordinates returns the coordinates of the start and end point of the Line.
func (n Line) Coordinates() []Coordinates {
	return []Coordinates{n.a, n.b}
}

// TransformXY transforms this Line into another Line according to fn.
func (n Line) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (Geometry, error) {
	coords := n.Coordinates()
	transform1dCoords(coords, fn)
	return NewLineC(coords[0], coords[1], opts...)
}

// EqualsExact checks if this Line is exactly equal to another curve.
func (n Line) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	c, ok := other.(curve)
	return ok && other.Dimension() == 1 && curvesExactEqual(n, c, opts)
}

// IsValid checks if this Line is valid
func (n Line) IsValid() bool {
	_, err := NewLineC(n.a, n.b)
	return err == nil
}

// IsRing always returns false for Lines because they are never rings. In
// particular, they are never closed because they only contain two points.
func (n Line) IsRing() bool {
	return false
}

// Length gives the length of the line.
func (n Line) Length() float64 {
	delta := n.a.XY.Sub(n.b.XY)
	return math.Sqrt(delta.Dot(delta))
}
