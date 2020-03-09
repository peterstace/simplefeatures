package geom

import (
	"bytes"
	"database/sql/driver"
	"io"
	"math"
	"unsafe"
)

// Point is a zero dimensional geometry that represents a single location in a
// coordinate space. It is immutable after creation.
//
// The Point may be empty.
//
// The zero value of Point is a 2D empty Point.
type Point struct {
	coords Coordinates
	full   bool
}

// NewPoint creates a new point gives its Coordinates.
func NewPoint(c Coordinates, _ ...ConstructorOption) Point {
	return Point{c, true}
}

// NewEmptyPoint creates a Point that is empty.
func NewEmptyPoint(ctype CoordinatesType) Point {
	return Point{Coordinates{Type: ctype}, false}
}

// NewPointFromXY creates a new point from an XY.
func NewPointFromXY(xy XY, _ ...ConstructorOption) Point {
	return Point{Coordinates{XY: xy, Type: DimXY}, true}
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
	return p.coords.XY, p.full
}

// Coordinates returns the coordinates of the point. The returned flag is set
// to true if and only if the point is non-empty.
func (p Point) Coordinates() (Coordinates, bool) {
	return p.coords, p.full
}

func (p Point) AsText() string {
	return string(p.AppendWKT(nil))
}

func (p Point) AppendWKT(dst []byte) []byte {
	dst = appendWKTHeader(dst, "POINT", p.coords.Type)
	if !p.full {
		return appendWKTEmpty(dst)
	}
	return appendWKTCoords(dst, p.coords, true)
}

func (p Point) IsEmpty() bool {
	return !p.full
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
	return GeometryCollection{}
}

func (p Point) Value() (driver.Value, error) {
	var buf bytes.Buffer
	err := p.AsBinary(&buf)
	return buf.Bytes(), err
}

func (p Point) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypePoint, p.CoordinatesType())
	if !p.full {
		p.coords.X = math.NaN()
		p.coords.Y = math.NaN()
		p.coords.Z = math.NaN()
		p.coords.M = math.NaN()
	}
	marsh.writeCoordinates(p.coords)
	return marsh.err
}

// ConvexHull returns the convex hull of this Point, which is always the same
// point.
func (p Point) ConvexHull() Geometry {
	return convexHull(p.AsGeometry())
}

func (p Point) MarshalJSON() ([]byte, error) {
	var dst []byte
	dst = append(dst, `{"type":"Point","coordinates":`...)
	if p.full {
		dst = appendGeoJSONCoordinate(dst, p.coords)
	} else {
		dst = append(dst, '[', ']')
	}
	return append(dst, '}'), nil
}

// TransformXY transforms this Point into another Point according to fn.
func (p Point) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (Point, error) {
	if !p.full {
		return p, nil
	}
	newC := p.coords
	newC.XY = fn(newC.XY)
	return NewPoint(newC, opts...), nil
}

// EqualsExact checks if this Point is exactly equal to another Point.
func (p Point) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	if !other.IsPoint() {
		return false
	}
	if p.CoordinatesType() != other.CoordinatesType() {
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
	return newEqualsExactOptionSet(opts).eq(p.coords, other.AsPoint().coords)
}

// IsValid checks if this Point is valid, but there is not way to indicate if
// Point is valid, so this function always returns true
func (p Point) IsValid() bool {
	return true
}

// Centroid of a point is that point.
func (p Point) Centroid() Point {
	return p.Force2D()
}

// Reverse in the case of Point outputs the same point.
func (p Point) Reverse() Point {
	return p
}

// AsMultiPoint is a convenience function that converts this Point into a
// MultiPoint.
func (p Point) AsMultiPoint() MultiPoint {
	mp, err := NewMultiPointFromPoints([]Point{p}, p.CoordinatesType())
	if err != nil {
		// Cannot occur due to construction.
		panic(err)
	}
	return mp
}

// CoordinatesType returns the CoordinatesType used to represent the Point.
func (p Point) CoordinatesType() CoordinatesType {
	return p.coords.Type
}

// ForceCoordinatesType returns a new Point with a different CoordinatesType. If a dimension
// is added, then new values are populated with 0.
func (p Point) ForceCoordinatesType(newCType CoordinatesType) Point {
	if !p.full {
		return NewEmptyPoint(newCType)
	}
	if newCType.Is3D() != p.coords.Type.Is3D() {
		p.coords.Z = 0
	}
	if newCType.IsMeasured() != p.coords.Type.IsMeasured() {
		p.coords.M = 0
	}
	p.coords.Type = newCType
	return p
}

// Force2D returns a copy of the Point with Z and M values removed.
func (p Point) Force2D() Point {
	return p.ForceCoordinatesType(DimXY)
}
