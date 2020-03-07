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
	// Uses 2 Coordinates variables rather than a Sequence to avoid the
	// indirection involved with a Sequence.
	a, b  Coordinates
	ctype CoordinatesType
}

// NewLineC creates a line segment given the Coordinates of its two endpoints.
func NewLineC(a, b Coordinates, ctype CoordinatesType, opts ...ConstructorOption) (Line, error) {
	if !skipValidations(opts) && a.XY == b.XY {
		return Line{}, ValidationError{"Line endpoints must be distinct"}
	}
	return Line{a, b, ctype}, nil
}

// NewLineXY creates a line segment given the XYs of its two endpoints.
func NewLineXY(a, b XY, opts ...ConstructorOption) (Line, error) {
	return NewLineC(Coordinates{XY: a}, Coordinates{XY: b}, DimXY, opts...)
}

// Type return type string for Line
func (n Line) Type() string {
	// Line is not a standard type, use LineString as its type
	return lineStringType
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
func (n Line) NumPoints() int {
	return 2
}

// PointN returns the coordinates of the first point when i is 0, and the
// second point when i is 1. It panics if i is any other value.
func (n Line) PointN(i int) Coordinates {
	switch i {
	case 0:
		return n.StartPoint()
	case 1:
		return n.EndPoint()
	default:
		panic("i must be 0 or 1")
	}
}

func (n Line) AsText() string {
	return string(n.AppendWKT(nil))
}

func (n Line) AppendWKT(dst []byte) []byte {
	dst = appendWKTHeader(dst, "LINESTRING", n.ctype)
	dst = append(dst, '(')
	dst = appendWKTCoords(dst, n.a, n.ctype, false)
	dst = append(dst, ',')
	dst = appendWKTCoords(dst, n.b, n.ctype, false)
	dst = append(dst, ')')
	return dst
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

func (n Line) Envelope() Envelope {
	return NewEnvelope(n.a.XY, n.b.XY)
}

func (n Line) Boundary() MultiPoint {
	return NewMultiPointXY([]XY{n.a.XY, n.b.XY})
}

func (n Line) Value() (driver.Value, error) {
	var buf bytes.Buffer
	err := n.AsBinary(&buf)
	return buf.Bytes(), err
}

func (n Line) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeLineString, n.ctype)
	marsh.writeCount(2)
	marsh.writeCoordinates(n.a, n.ctype)
	marsh.writeCoordinates(n.b, n.ctype)
	return marsh.err
}

func (n Line) ConvexHull() Geometry {
	return convexHull(n.AsGeometry())
}

func (n Line) MarshalJSON() ([]byte, error) {
	var dst []byte
	dst = append(dst, `{"type":"LineString","coordinates":[`...)
	dst = appendGeoJSONCoordinate(dst, n.ctype, n.a)
	dst = append(dst, ',')
	dst = appendGeoJSONCoordinate(dst, n.ctype, n.b)
	dst = append(dst, "]}"...)
	return dst, nil
}

// Coordinates returns the coordinates of the start and end point of the Line.
func (n Line) Coordinates() Sequence {
	floats := make([]float64, 0, 2*n.ctype.Dimension())
	for _, c := range [2]Coordinates{n.a, n.b} {
		floats = append(floats, c.X, c.Y)
		if n.ctype.Is3D() {
			floats = append(floats, c.Z)
		}
		if n.ctype.IsMeasured() {
			floats = append(floats, c.M)
		}
	}
	return NewSequence(floats, n.ctype)
}

// TransformXY transforms this Line into another Line according to fn.
func (n Line) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (Line, error) {
	n.a.XY = fn(n.a.XY)
	n.b.XY = fn(n.b.XY)
	ln, err := NewLineC(n.a, n.b, n.ctype, opts...)
	return ln, err
}

// EqualsExact checks if this Line is exactly equal to another curve.
func (n Line) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	var otherSeq Sequence
	switch {
	case other.IsLine():
		otherSeq = other.AsLine().Coordinates()
	case other.IsLineString():
		otherSeq = other.AsLineString().Coordinates()
	default:
		return false
	}
	return curvesExactEqual(n.Coordinates(), otherSeq, opts)
}

// IsValid checks if this Line is valid
func (n Line) IsValid() bool {
	_, err := NewLineC(n.a, n.b, n.ctype)
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
	ls, err := NewLineStringFromSequence(n.Coordinates(), DisableAllValidations)
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
	return Line{n.b, n.a, n.ctype}
}

func (n Line) CoordinatesType() CoordinatesType {
	return n.ctype
}

func (n Line) Force(newCType CoordinatesType) Line {
	if n.ctype.Is3D() != newCType.Is3D() {
		n.a.Z = 0
		n.b.Z = 0
	}
	if n.ctype.IsMeasured() != newCType.IsMeasured() {
		n.a.M = 0
		n.b.M = 0
	}
	n.ctype = newCType
	return n
}
