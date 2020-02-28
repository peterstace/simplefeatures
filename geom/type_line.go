package geom

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"io"
	"math"
	"unsafe"
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
	if !skipValidations(opts) && a.XY == b.XY {
		return Line{}, fmt.Errorf("line endpoints must be distinct: %v", a.XY)
	}
	return Line{a, b}, nil
}

// NewLineXY creates a line segment given the XYs of its two endpoints.
func NewLineXY(a, b XY, opts ...ConstructorOption) (Line, error) {
	return NewLineC(Coordinates{a}, Coordinates{b}, opts...)
}

// AsGeometry converts this Line into a Geometry.
func (n Line) AsGeometry() Geometry {
	return Geometry{lineTag, unsafe.Pointer(&n)}
}

// StartPoint gives the coordinates of the first control point of the line.
func (n Line) StartPoint() Coordinates {
	return n.a
}

// EndPoint gives the coordinates of the second (last) control point of the line.
func (n Line) EndPoint() Coordinates {
	return n.b
}

// NumPoints always returns 2.
func (Line) NumPoints() int {
	return 2
}

// PointN returns the coordinates of the first point when n is 0, and the
// second point when n is 1. It panics if n is any other value.
func (ln Line) PointN(n int) Coordinates {
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
	dst = append(dst, "LINESTRING("...)
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
	return intersection(n.AsGeometry(), g)
}

func (n Line) Intersects(g Geometry) bool {
	return hasIntersection(n.AsGeometry(), g)
}

func (n Line) Equals(other Geometry) (bool, error) {
	return equals(n.AsGeometry(), other)
}

func (n Line) Envelope() Envelope {
	return NewEnvelope(n.a.XY, n.b.XY)
}

func (n Line) Boundary() MultiPoint {
	return NewMultiPoint([]Point{
		NewPointXY(n.a.XY),
		NewPointXY(n.b.XY),
	})
}

func (n Line) Value() (driver.Value, error) {
	var buf bytes.Buffer
	err := n.AsBinary(&buf)
	return buf.Bytes(), err
}

func (n Line) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeLineString)
	marsh.writeCount(2)
	marsh.writeFloat64(n.StartPoint().X)
	marsh.writeFloat64(n.StartPoint().Y)
	marsh.writeFloat64(n.EndPoint().X)
	marsh.writeFloat64(n.EndPoint().Y)
	return marsh.err
}

func (n Line) ConvexHull() Geometry {
	return convexHull(n.AsGeometry())
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
	ln, err := NewLineC(coords[0], coords[1], opts...)
	return ln.AsGeometry(), err
}

// EqualsExact checks if this Line is exactly equal to another curve.
func (n Line) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	var c curve
	switch {
	case other.IsLine():
		c = other.AsLine()
	case other.IsLineString():
		c = other.AsLineString()
	default:
		return false
	}
	return curvesExactEqual(n, c, opts)
}

// IsValid checks if this Line is valid
func (n Line) IsValid() bool {
	_, err := NewLineC(n.a, n.b)
	return err == nil
}

// Length gives the length of the line.
func (n Line) Length() float64 {
	delta := n.a.XY.Sub(n.b.XY)
	return math.Sqrt(delta.Dot(delta))
}

func (n Line) Centroid() Point {
	return NewPointF((n.a.XY.X+n.b.XY.X)/2, (n.a.XY.Y+n.b.XY.Y)/2)
}

// AsLineString is a helper function that converts this Line into a LineString.
func (n Line) AsLineString() LineString {
	ls, err := NewLineStringC(n.Coordinates(), DisableAllValidations)
	if err != nil {
		// Cannot occur due to construction. A valid Line will always be a
		// valid LineString.
		msg := fmt.Sprintf("implementation error: Could not convert "+
			"Line to LineString. Line=%v, Err=%v.", n, err)
		panic(msg)
	}
	return ls
}

// Reverse in the case of Line outputs the coordinates in reverse order.
func (n Line) Reverse() Line {
	return Line{n.b, n.a}
}
