package simplefeatures

import (
	"database/sql/driver"
	"fmt"
	"io"
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

// StartPoint gives the first point of the line.
func (n Line) StartPoint() Point {
	return NewPointFromCoords(n.a)
}

// EndPoint gives the second (last) point of the line.
func (n Line) EndPoint() Point {
	return NewPointFromCoords(n.b)
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
	dst = n.a.X.appendAsFloat(dst)
	dst = append(dst, ' ')
	dst = n.a.Y.appendAsFloat(dst)
	dst = append(dst, ',')
	dst = n.b.X.appendAsFloat(dst)
	dst = append(dst, ' ')
	dst = n.b.Y.appendAsFloat(dst)
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

func (n Line) Boundary() Geometry {
	return NewMultiPoint([]Point{
		NewPoint(n.a.XY),
		NewPoint(n.b.XY),
	})
}

func (n Line) Value() (driver.Value, error) {
	return n.AsText(), nil
}

func (n Line) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeLineString)
	marsh.writeCount(2)
	marsh.writeFloat64(n.StartPoint().XY().X.AsFloat())
	marsh.writeFloat64(n.StartPoint().XY().Y.AsFloat())
	marsh.writeFloat64(n.EndPoint().XY().X.AsFloat())
	marsh.writeFloat64(n.EndPoint().XY().Y.AsFloat())
	return marsh.err
}

func (n Line) MarshalJSON() ([]byte, error) {
	return marshalGeoJSON("LineString", []Coordinates{
		n.StartPoint().Coordinates(),
		n.EndPoint().Coordinates(),
	})
}
