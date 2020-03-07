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
	full   bool
	ctype  CoordinatesType
}

// NewEmptyPoint creates a Point that is empty.
func NewEmptyPoint(ctype CoordinatesType) Point {
	return Point{Coordinates{}, false, ctype}
}

// NewPointXY creates a new point from an XY.
func NewPointXY(xy XY, _ ...ConstructorOption) Point {
	return Point{Coordinates{XY: xy}, true, XYOnly}
}

// NewPointF creates a new point from float64 x and y values.
func NewPointF(x, y float64, _ ...ConstructorOption) Point {
	return NewPointXY(XY{x, y})
}

// NewPointC creates a new point gives its Coordinates.
func NewPointC(c Coordinates, ctype CoordinatesType, _ ...ConstructorOption) Point {
	return Point{c, true, ctype}
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
	dst = appendWKTHeader(dst, "POINT", p.ctype)
	if !p.full {
		return appendWKTEmpty(dst)
	}
	return appendWKTCoords(dst, p.coords, p.ctype, true)
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
	marsh.writeGeomType(wkbGeomTypePoint, p.ctype)
	if !p.full {
		p.coords.X = math.NaN()
		p.coords.Y = math.NaN()
		p.coords.Z = math.NaN()
		p.coords.M = math.NaN()
	}
	marsh.writeCoordinates(p.coords, p.ctype)
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
		dst = appendGeoJSONCoordinate(dst, p.ctype, p.coords)
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
	return NewPointC(newC, p.ctype, opts...), nil
}

// EqualsExact checks if this Point is exactly equal to another Point.
func (p Point) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	if !other.IsPoint() {
		return false
	}
	if p.ctype != other.CoordinatesType() {
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
	return newEqualsExactOptionSet(opts).eq(p.coords, other.AsPoint().coords, p.ctype)
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
	var empty BitSet
	floats := make([]float64, 2, 4)
	if p.full {
		floats[0] = p.coords.X
		floats[1] = p.coords.Y
	}
	if p.full && p.ctype.Is3D() {
		floats = append(floats, p.coords.Z)
	}
	if p.full && p.ctype.IsMeasured() {
		floats = append(floats, p.coords.M)
	}
	seq := NewSequenceNoCopy(floats, p.CoordinatesType())
	return NewMultiPointFromSequence(seq, empty)
}

func (p Point) CoordinatesType() CoordinatesType {
	return p.ctype
}

func (p Point) Force2D() Point {
	xy, ok := p.XY()
	if ok {
		return NewPointXY(xy)
	} else {
		return NewEmptyPoint(XYOnly)
	}
}
