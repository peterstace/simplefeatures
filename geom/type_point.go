package geom

import (
	"database/sql/driver"
	"fmt"
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

// NewPoint creates a new point given its Coordinates.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewPoint(c Coordinates) Point {
	return Point{c, true}
}

// Validate checks if the Point is valid. For it to be valid, it must be empty
// or not have NaN or Inf XY values.
func (p Point) Validate() error {
	if !p.full {
		return nil
	}
	return p.coords.XY.validate()
}

// NewEmptyPoint creates a Point that is empty.
func NewEmptyPoint(ctype CoordinatesType) Point {
	return Point{Coordinates{Type: ctype}, false}
}

// Type returns the GeometryType for a Point.
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
	return p.appendWKTBody(dst)
}

func (p Point) appendWKTBody(dst []byte) []byte {
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

// Envelope returns the envelope best fitting the Point (either an empty
// envelope, or an envelope covering a single point).
func (p Point) Envelope() Envelope {
	if xy, ok := p.XY(); ok {
		return Envelope{}.ExpandToIncludeXY(xy)
	}
	return Envelope{}
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

// Scan implements the database/sql.Scanner interface by parsing the src value
// as WKB (Well Known Binary).
//
// If the WKB doesn't represent a Point geometry, then an error is returned.
//
// Geometry constraint validation is performed on the resultant geometry (an
// error will be returned if the geometry is invalid). If this validation isn't
// needed or is undesirable, then the WKB should be scanned into a byte slice
// and then UnmarshalWKB called manually (passing in NoValidate{}).
func (p *Point) Scan(src interface{}) error {
	return scanAsType(src, p)
}

// AsBinary returns the WKB (Well Known Text) representation of the geometry.
func (p Point) AsBinary() []byte {
	return p.AppendWKB(nil)
}

// AppendWKB appends the WKB (Well Known Text) representation of the geometry
// to the input slice.
func (p Point) AppendWKB(dst []byte) []byte {
	marsh := newWKBMarshaler(dst)
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

// MarshalJSON implements the encoding/json.Marshaler interface by encoding
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

// UnmarshalJSON implements the encoding/json.Unmarshaler interface by decoding
// the GeoJSON representation of a Point.
func (p *Point) UnmarshalJSON(buf []byte) error {
	return unmarshalGeoJSONAsType(buf, p)
}

// TransformXY transforms this Point into another Point according to fn.
func (p Point) TransformXY(fn func(XY) XY) Point {
	if !p.full {
		return p
	}
	newC := p.coords
	newC.XY = fn(newC.XY)
	return NewPoint(newC)
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
	return NewMultiPoint([]Point{p})
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

// DumpCoordinates returns a Sequence representing the point. For an empty
// Point, the Sequence will be empty. For a non-empty Point, the Sequence will
// contain the single set of coordinates representing the point.
func (p Point) DumpCoordinates() Sequence {
	ctype := p.CoordinatesType()
	var floats []float64
	coords, ok := p.Coordinates()
	if ok {
		n := ctype.Dimension()
		floats = coords.appendFloat64s(make([]float64, 0, n))
	}
	seq := NewSequence(floats, ctype)
	seq.assertNoUnusedCapacity()
	return seq
}

// Summary returns a text summary of the Point following a similar format to https://postgis.net/docs/ST_Summary.html.
func (p Point) Summary() string {
	var pointSuffix string
	numPoints := 1
	if p.IsEmpty() {
		numPoints = 0
		pointSuffix = "s"
	}
	return fmt.Sprintf("%s[%s] with %d point%s", p.Type(), p.CoordinatesType(), numPoints, pointSuffix)
}

// String returns the string representation of the Point.
func (p Point) String() string {
	return p.Summary()
}

// SnapToGrid returns a copy of the Point with all coordinates snapped to a
// base 10 grid.
//
// The grid spacing is specified by the number of decimal places to round to
// (with negative decimal places being allowed). E.g., a decimalPlaces value of
// 2 would cause all coordinates to be rounded to the nearest 0.01, and a
// decimalPlaces of -1 would cause all coordinates to be rounded to the nearest
// 10.
func (p Point) SnapToGrid(decimalPlaces int) Point {
	return p.TransformXY(snapToGridXY(decimalPlaces))
}
