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
	coords OptionalCoordinates
}

// NewEmptyPoint creates a Point that is empty.
func NewEmptyPoint() Point {
	return Point{OptionalCoordinates{Empty: true}}
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
	return Point{coords: OptionalCoordinates{Value: c}}
}

// NewPointOC creates a new point given its OptionalCoordinates.
func NewPointOC(oc OptionalCoordinates, _ ...ConstructorOption) Point {
	return Point{coords: oc}
}

// Type return type string for Point
func (p Point) Type() string {
	return pointType
}

// AsGeometry converts this Point into a Geometry.
func (p Point) AsGeometry() Geometry {
	return Geometry{pointTag, unsafe.Pointer(&p)}
}

// XY gives the XY location of the point. The returned flag is set to true if
// and only if the point is non-empty.
func (p Point) XY() (XY, bool) {
	return p.coords.Value.XY, !p.coords.Empty
}

// Coordinates returns the coordinates of the point.
func (p Point) Coordinates() OptionalCoordinates {
	return p.coords
}

func (p Point) AsText() string {
	return string(p.AppendWKT(nil))
}

func (p Point) AppendWKT(dst []byte) []byte {
	dst = append(dst, "POINT"...)
	xy, ok := p.XY()
	if !ok {
		return append(dst, " EMPTY"...)
	}
	dst = append(dst, '(')
	dst = appendFloat(dst, xy.X)
	dst = append(dst, ' ')
	dst = appendFloat(dst, xy.Y)
	return append(dst, ')')
}

func (p Point) IsEmpty() bool {
	return p.coords.Empty
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

func (p Point) Envelope() (Envelope, bool) {
	xy, ok := p.XY()
	if !ok {
		return Envelope{}, false
	}
	return NewEnvelope(xy), true
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
	xy, ok := p.XY()
	if !ok {
		marsh.writeFloat64(math.NaN())
		marsh.writeFloat64(math.NaN())
	} else {
		marsh.writeFloat64(xy.X)
		marsh.writeFloat64(xy.Y)
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
	if !other.IsPoint() {
		return false
	}
	if p.IsEmpty() != other.IsEmpty() {
		return false
	}
	if p.IsEmpty() {
		return true
	}
	// No need to check returned flag, since we know that both Points are
	// non-empty.
	xyA, _ := p.XY()
	xyB, _ := other.AsPoint().XY()
	return newEqualsExactOptionSet(opts).eq(xyA, xyB)
}

// IsValid checks if this Point is valid, but there is not way to indicate if
// Point is valid, so this function always returns true
func (p Point) IsValid() bool {
	return true
}

// Centroid of a point it that point.
func (p Point) Centroid() Point {
	return p
}

// Reverse in the case of Point outputs the same point.
func (p Point) Reverse() Point {
	return p
}
