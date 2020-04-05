package geom

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"io"
	"math"
	"unsafe"
)

// Line is a linear geometry that represents a single line segment between two
// points that have distinct XY values. It is immutable after creation.
type Line struct {
	// Uses 2 Coordinates variables rather than a Sequence to avoid the
	// indirection involved with a Sequence.
	a, b Coordinates
}

// NewLine creates a line segment given the Coordinates of its two endpoints.
// An error is returned if the XY values of the coordinates are not distinct.
func NewLine(a, b Coordinates, opts ...ConstructorOption) (Line, error) {
	ctype := a.Type & b.Type
	// TODO: Would be better to have a ForceCoordinateType function on Coordinates.
	a.Type = ctype
	b.Type = ctype

	if !skipValidations(opts) && a.XY == b.XY {
		return Line{}, fmt.Errorf("line endpoints must have distinct XY values: %v", a.XY)
	}
	return Line{a, b}, nil
}

// NewLineFromXY creates a line segment given the XYs of its two endpoints. An
// error is returned if the XY values are not distinct.
func NewLineFromXY(a, b XY, opts ...ConstructorOption) (Line, error) {
	return NewLine(
		Coordinates{XY: a, Type: DimXY},
		Coordinates{XY: b, Type: DimXY},
		opts...,
	)
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

// AsText returns the WKT (Well Known Text) representation of this geometry.
func (n Line) AsText() string {
	return string(n.AppendWKT(nil))
}

// AppendWKT appends the WKT (Well Known Text) representation of this geometry
// to the input byte slice.
func (n Line) AppendWKT(dst []byte) []byte {
	dst = appendWKTHeader(dst, "LINESTRING", n.a.Type)
	dst = append(dst, '(')
	dst = appendWKTCoords(dst, n.a, false)
	dst = append(dst, ',')
	dst = appendWKTCoords(dst, n.b, false)
	dst = append(dst, ')')
	return dst
}

// IsSimple returns true if this geometry contains no anomalous geometry
// points, such as self intersection or self tangency.  Lines are always
// simple, so this method always returns true.
func (n Line) IsSimple() bool {
	return true
}

// Intersection calculates the of this geometry and another, i.e. the portion
// of the two geometries that are shared. It is not implemented for all
// geometry pairs, and returns an error for those cases.
func (n Line) Intersection(g Geometry) (Geometry, error) {
	return intersection(n.AsGeometry(), g)
}

// Intersects return true if and only if this geometry intersects with the
// other, i.e. they shared at least one common point.
func (n Line) Intersects(g Geometry) bool {
	return hasIntersection(n.AsGeometry(), g)
}

// Envelope returns the Envelope that most tightly surrounds the Line.
func (n Line) Envelope() Envelope {
	return NewEnvelope(n.a.XY, n.b.XY)
}

// Boundary returns the spatial boundary of this Line. This is the MultiPoint
// collection containing the two endpoints of the Line.
func (n Line) Boundary() MultiPoint {
	return NewMultiPoint(
		NewSequence([]float64{
			n.a.XY.X, n.a.XY.Y,
			n.b.XY.X, n.b.XY.Y,
		}, DimXY),
	)
}

// Value implements the database/sql/driver.Valuer interface by returning the
// WKB (Well Known Binary) representation of this Geometry.
func (n Line) Value() (driver.Value, error) {
	var buf bytes.Buffer
	err := n.AsBinary(&buf)
	return buf.Bytes(), err
}

// AsBinary writes the WKB (Well Known Binary) representation of the geometry
// to the writer.
func (n Line) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeLineString, n.a.Type)
	marsh.writeCount(2)
	marsh.writeCoordinates(n.a)
	marsh.writeCoordinates(n.b)
	return marsh.err
}

// ConvexHull returns the geometry representing the smallest convex geometry
// that contains this geometry.
func (n Line) ConvexHull() Geometry {
	return convexHull(n.AsGeometry())
}

// MarshalJSON implements the encoding/json.Marshaller interface by encoding
// this geometry as a GeoJSON geometry object.
func (n Line) MarshalJSON() ([]byte, error) {
	var dst []byte
	dst = append(dst, `{"type":"LineString","coordinates":[`...)
	dst = appendGeoJSONCoordinate(dst, n.a)
	dst = append(dst, ',')
	dst = appendGeoJSONCoordinate(dst, n.b)
	dst = append(dst, "]}"...)
	return dst, nil
}

// Coordinates returns the coordinates of the start and end point of the Line.
func (n Line) Coordinates() Sequence {
	ctype := n.a.Type
	floats := make([]float64, 0, 2*ctype.Dimension())
	for _, c := range [2]Coordinates{n.a, n.b} {
		floats = append(floats, c.X, c.Y)
		if ctype.Is3D() {
			floats = append(floats, c.Z)
		}
		if ctype.IsMeasured() {
			floats = append(floats, c.M)
		}
	}
	return NewSequence(floats, ctype)
}

// TransformXY transforms this Line into another Line according to fn.
func (n Line) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (Line, error) {
	return NewLineFromXY(
		fn(n.a.XY),
		fn(n.b.XY),
		opts...,
	)
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

// Length gives the length of the line.
func (n Line) Length() float64 {
	delta := n.a.XY.Sub(n.b.XY)
	return math.Sqrt(delta.Dot(delta))
}

// Centroid retruns the centroid of this Line, which is always its midpoint.
func (n Line) Centroid() Point {
	return NewPointFromXY(n.a.XY.Add(n.b.XY).Scale(0.5))
}

// AsLineString is a helper function that converts this Line into a LineString.
func (n Line) AsLineString() LineString {
	ls, err := NewLineString(n.Coordinates(), DisableAllValidations)
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

// CoordinatesType returns the CoordinatesType used to represent points making
// up the geometry.
func (n Line) CoordinatesType() CoordinatesType {
	return n.a.Type
}

// ForceCoordinatesType returns a new Line with a different CoordinatesType. If a dimension is
// added, then new values are populated with 0.
func (n Line) ForceCoordinatesType(newCType CoordinatesType) Line {
	if n.a.Type.Is3D() != newCType.Is3D() {
		n.a.Z = 0
		n.b.Z = 0
	}
	if n.a.Type.IsMeasured() != newCType.IsMeasured() {
		n.a.M = 0
		n.b.M = 0
	}
	n.a.Type = newCType
	n.b.Type = newCType
	return n
}
