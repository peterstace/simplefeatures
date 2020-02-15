package geom

import (
	"bytes"
	"database/sql/driver"
	"io"
	"math"
	"unsafe"
)

// Point is a 0-dimensional geometry, and represents a single location in a
// coordinate space.
//
// The Point may be empty.
//
// There aren't any assertions about what constitutes a valid point.
type Point struct {
	coords Coordinates
	empty  bool
}

// NewEmptyPoint creates a Point that is empty.
func NewEmptyPoint() Point {
	return Point{empty: true}
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

// NewPointOC creates a new point given its OptionalCoordinates.
func NewPointOC(oc OptionalCoordinates, _ ...ConstructorOption) Point {
	return Point{coords: oc.Value, empty: oc.Empty}
}

// AsGeometry converts this Point into a Geometry.
func (p Point) AsGeometry() Geometry {
	return Geometry{pointTag, unsafe.Pointer(&p)}
}

// XY gives the XY location of the point.
func (p Point) XY() XY {
	return p.coords.XY
}

// Coordinates returns the coordinates of the point.
func (p Point) Coordinates() OptionalCoordinates {
	return OptionalCoordinates{Empty: p.empty, Value: p.coords}
}

func (p Point) AsText() string {
	return string(p.AppendWKT(nil))
}

func (p Point) AppendWKT(dst []byte) []byte {
	dst = append(dst, "POINT"...)
	if p.IsEmpty() {
		return append(dst, " EMPTY"...)
	}
	dst = append(dst, '(')
	dst = appendFloat(dst, p.coords.X)
	dst = append(dst, ' ')
	dst = appendFloat(dst, p.coords.Y)
	return append(dst, ')')
}

func (p Point) IsEmpty() bool {
	return p.empty
}

func (p Point) IsSimple() bool {
	return true
}

func (p Point) Intersection(g Geometry) (Geometry, error) {
	return intersection(p.AsGeometry(), g)
}

func (p Point) Intersects(g Geometry) bool {
	return hasIntersection(p.AsGeometry(), g)
}

func (p Point) Equals(other Geometry) (bool, error) {
	return equals(p.AsGeometry(), other)
}

func (p Point) Envelope() (Envelope, bool) {
	if p.IsEmpty() {
		return Envelope{}, false
	}
	return NewEnvelope(p.coords.XY), true
}

func (p Point) Boundary() GeometryCollection {
	return NewGeometryCollection(nil)
}

func (p Point) Value() (driver.Value, error) {
	var buf bytes.Buffer
	err := p.AsBinary(&buf)
	return buf.Bytes(), err
}

func (p Point) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypePoint)
	if p.IsEmpty() {
		marsh.writeFloat64(math.NaN())
		marsh.writeFloat64(math.NaN())
	} else {
		marsh.writeFloat64(p.coords.X)
		marsh.writeFloat64(p.coords.Y)
	}
	return marsh.err
}

// ConvexHull returns the convex hull of this Point, which is always the same
// point.
func (p Point) ConvexHull() Geometry {
	return convexHull(p.AsGeometry())
}

func (p Point) MarshalJSON() ([]byte, error) {
	return marshalGeoJSON("Point", p.Coordinates())
}

// TransformXY transforms this Point into another Point according to fn.
func (p Point) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (Geometry, error) {
	coords := p.Coordinates()
	if !coords.Empty {
		coords.Value.XY = fn(coords.Value.XY)
	}
	return NewPointOC(coords, opts...).AsGeometry(), nil
}

// EqualsExact checks if this Point is exactly equal to another Point.
func (p Point) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	return other.IsPoint() &&
		newEqualsExactOptionSet(opts).eq(p.XY(), other.AsPoint().XY())
}

// IsValid checks if this Point is valid, but there is not way to indicate if
// Point is valid, so this function always returns true
func (p Point) IsValid() bool {
	return true
}

// Reverse in the case of Point outputs the same point.
func (p Point) Reverse() Point {
	return p
}
