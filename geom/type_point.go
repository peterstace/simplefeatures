package geom

import (
	"database/sql/driver"
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

// Type returns the GeometryType for a Point
func (p Point) Type() GeometryType {
	return TypePoint
}

// AsGeometry converts this Point into a Geometry.
func (p Point) AsGeometry() Geometry {
	return Geometry{TypePoint, unsafe.Pointer(&p)}
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

// AsText returns the WKT (Well Known Text) representation of this geometry.
func (p Point) AsText() string {
	return string(p.AppendWKT(nil))
}

// AppendWKT appends the WKT (Well Known Text) representation of this geometry
// to the input byte slice.
func (p Point) AppendWKT(dst []byte) []byte {
	dst = appendWKTHeader(dst, "POINT", p.coords.Type)
	if !p.full {
		return appendWKTEmpty(dst)
	}
	return appendWKTCoords(dst, p.coords, true)
}

// IsEmpty returns true if and only if this Point is the empty Point.
func (p Point) IsEmpty() bool {
	return !p.full
}

// IsSimple returns true if this geometry contains no anomalous geometry
// points, such as self intersection or self tangency. Points are always
// simple, so this method always return true.
func (p Point) IsSimple() bool {
	return true
}

// Envelope returns a zero area Envelope covering the Point. If the Point is
// empty, then false is returned.
func (p Point) Envelope() (Envelope, bool) {
	xy, ok := p.XY()
	if !ok {
		return Envelope{}, false
	}
	return NewEnvelope(xy), true
}

// Boundary returns the spatial boundary for this Point, which is always the
// empty set. This is represented by the empty GeometryCollection.
func (p Point) Boundary() GeometryCollection {
	return GeometryCollection{}
}

// Value implements the database/sql/driver.Valuer interface by returning the
// WKB (Well Known Binary) representation of this Geometry.
func (p Point) Value() (driver.Value, error) {
	return p.AsBinary(), nil
}

// AsBinary returns the WKB (Well Known Text) representation of the geometry.
func (p Point) AsBinary() []byte {
	return p.AppendWKB(nil)
}

// AppendWKB appends the WKB (Well Known Text) representation of the geometry
// to the input slice.
func (p Point) AppendWKB(dst []byte) []byte {
	marsh := newWKBMarshaller(dst)
	marsh.writeByteOrder()
	marsh.writeGeomType(TypePoint, p.CoordinatesType())
	if !p.full {
		p.coords.X = math.NaN()
		p.coords.Y = math.NaN()
		p.coords.Z = math.NaN()
		p.coords.M = math.NaN()
	}
	marsh.writeCoordinates(p.coords)
	return marsh.buf
}

// ConvexHull returns the geometry representing the smallest convex geometry
// that contains this geometry.
func (p Point) ConvexHull() Geometry {
	return convexHull(p.AsGeometry())
}

// MarshalJSON implements the encoding/json.Marshaller interface by encoding
// this geometry as a GeoJSON geometry object.
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
	return NewMultiPointFromPoints([]Point{p})
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

// PointOnSurface returns the original Point.
func (p Point) PointOnSurface() Point {
	return p.Force2D()
}

func (p Point) asXYs() []XY {
	if xy, ok := p.XY(); ok {
		return []XY{xy}
	}
	return nil
}
