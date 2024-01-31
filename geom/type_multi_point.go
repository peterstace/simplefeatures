package geom

import (
	"database/sql/driver"
	"fmt"
	"unsafe"
)

// MultiPoint is a 0-dimensional geometry that is a collection of points. Its
// zero value is the empty MultiPoint (i.e. a collection of zero points) with
// 2D coordinates type. It is immutable after creation.
type MultiPoint struct {
	// Invariant: ctype matches the coordinates type of each point.
	points []Point
	ctype  CoordinatesType
}

// NewMultiPoint creates a MultiPoint from a list of Points. The coordinate
// type of the MultiPoint is the lowest common coordinates type of its Points.
//
// It doesn't perform any validation on the result. The Validate method can be
// used to check the validity of the result if needed.
func NewMultiPoint(pts []Point) MultiPoint {
	if len(pts) == 0 {
		return MultiPoint{}
	}

	ctype := DimXYZM
	for _, p := range pts {
		ctype &= p.CoordinatesType()
	}

	forced := forceCoordinatesTypeOfPointSlice(pts, ctype)
	return MultiPoint{forced, ctype}
}

// Validate checks if the MultiPoint is valid. The only validation rule is that
// each point in the collection must be valid.
func (m MultiPoint) Validate() error {
	for i, pt := range m.points {
		if err := pt.Validate(); err != nil {
			return wrap(err, "validating point at index %d", i)
		}
	}
	return nil
}

// Type returns the GeometryType for a MultiPoint.
func (m MultiPoint) Type() GeometryType {
	return TypeMultiPoint
}

// AsGeometry converts this MultiPoint into a Geometry.
func (m MultiPoint) AsGeometry() Geometry {
	return Geometry{TypeMultiPoint, unsafe.Pointer(&m)}
}

// NumPoints gives the number of element points making up the MultiPoint.
func (m MultiPoint) NumPoints() int {
	return len(m.points)
}

// PointN gives the nth (zero indexed) Point.
func (m MultiPoint) PointN(n int) Point {
	return m.points[n]
}

// AsText returns the WKT (Well Known Text) representation of this geometry.
func (m MultiPoint) AsText() string {
	return string(m.AppendWKT(nil))
}

// AppendWKT appends the WKT (Well Known Text) representation of this geometry
// to the input byte slice.
func (m MultiPoint) AppendWKT(dst []byte) []byte {
	dst = appendWKTHeader(dst, "MULTIPOINT", m.CoordinatesType())
	if len(m.points) == 0 {
		return appendWKTEmpty(dst)
	}
	dst = append(dst, '(')
	for i, pt := range m.points {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = pt.appendWKTBody(dst)
	}
	return append(dst, ')')
}

// IsSimple returns true if this geometry contains no anomalous geometry
// points, such as self intersection or self tangency.  MultiPoints are simple
// if and only if no two of its points have equal XY coordinates.
func (m MultiPoint) IsSimple() bool {
	seen := make(map[XY]bool)
	for i := 0; i < m.NumPoints(); i++ {
		xy, ok := m.PointN(i).XY()
		if !ok {
			continue
		}
		if seen[xy] {
			return false
		}
		seen[xy] = true
	}
	return true
}

// IsEmpty return true if and only if this MultiPoint doesn't contain any
// Points, or only contains empty Points.
func (m MultiPoint) IsEmpty() bool {
	for _, pt := range m.points {
		if !pt.IsEmpty() {
			return false
		}
	}
	return true
}

// Envelope returns the Envelope that most tightly surrounds the geometry.
func (m MultiPoint) Envelope() Envelope {
	var env Envelope
	for _, pt := range m.points {
		env = env.ExpandToIncludeEnvelope(pt.Envelope())
	}
	return env
}

// Boundary returns the spatial boundary for this MultiPoint, which is always
// the empty set. This is represented by the empty GeometryCollection.
func (m MultiPoint) Boundary() GeometryCollection {
	return GeometryCollection{}
}

// Value implements the database/sql/driver.Valuer interface by returning the
// WKB (Well Known Binary) representation of this Geometry.
func (m MultiPoint) Value() (driver.Value, error) {
	return m.AsBinary(), nil
}

// Scan implements the database/sql.Scanner interface by parsing the src value
// as WKB (Well Known Binary).
//
// If the WKB doesn't represent a MultiPoint geometry, then an error is returned.
//
// Geometry constraint validation is performed on the resultant geometry (an
// error will be returned if the geometry is invalid). If this validation isn't
// needed or is undesirable, then the WKB should be scanned into a byte slice
// and then UnmarshalWKB called manually (passing in NoValidate{}).
func (m *MultiPoint) Scan(src interface{}) error {
	return scanAsType(src, m)
}

// AsBinary returns the WKB (Well Known Text) representation of the geometry.
func (m MultiPoint) AsBinary() []byte {
	return m.AppendWKB(nil)
}

// AppendWKB appends the WKB (Well Known Text) representation of the geometry
// to the input slice.
func (m MultiPoint) AppendWKB(dst []byte) []byte {
	marsh := newWKBMarshaler(dst)
	marsh.writeByteOrder()
	marsh.writeGeomType(TypeMultiPoint, m.CoordinatesType())
	n := m.NumPoints()
	marsh.writeCount(n)
	for i := 0; i < n; i++ {
		pt := m.PointN(i)
		marsh.buf = pt.AppendWKB(marsh.buf)
	}
	return marsh.buf
}

// ConvexHull returns the geometry representing the smallest convex geometry
// that contains this geometry.
func (m MultiPoint) ConvexHull() Geometry {
	return convexHull(m.AsGeometry())
}

// MarshalJSON implements the encoding/json.Marshaler interface by encoding
// this geometry as a GeoJSON geometry object.
func (m MultiPoint) MarshalJSON() ([]byte, error) {
	var dst []byte
	dst = append(dst, `{"type":"MultiPoint","coordinates":[`...)
	first := true
	for _, pt := range m.points {
		c, ok := pt.Coordinates()
		if ok {
			if !first {
				dst = append(dst, ',')
			}
			first = false
			dst = appendGeoJSONCoordinate(dst, c)
		}
	}
	dst = append(dst, "]}"...)
	return dst, nil
}

// UnmarshalJSON implements the encoding/json.Unmarshaler interface by decoding
// the GeoJSON representation of a MultiPoint.
func (m *MultiPoint) UnmarshalJSON(buf []byte) error {
	return unmarshalGeoJSONAsType(buf, m)
}

// Coordinates returns the coordinates of the non-empty points represented by
// the MultiPoint.
func (m MultiPoint) Coordinates() Sequence {
	ctype := m.CoordinatesType()
	coords := make([]float64, 0, len(m.points)*ctype.Dimension())
	for _, pt := range m.points {
		if c, ok := pt.Coordinates(); ok {
			coords = c.appendFloat64s(coords)
		}
	}
	return NewSequence(coords, ctype)
}

// TransformXY transforms this MultiPoint into another MultiPoint according to fn.
func (m MultiPoint) TransformXY(fn func(XY) XY) MultiPoint {
	if len(m.points) == 0 {
		return MultiPoint{}.ForceCoordinatesType(m.CoordinatesType())
	}
	txPoints := make([]Point, len(m.points))
	for i, pt := range m.points {
		if c, ok := pt.Coordinates(); ok {
			c.XY = fn(c.XY)
			txPoints[i] = NewPoint(c)
		} else {
			txPoints[i] = pt
		}
	}
	return NewMultiPoint(txPoints)
}

// Centroid gives the centroid of the coordinates of the MultiPoint.
func (m MultiPoint) Centroid() Point {
	var sum XY
	var n int
	for i := 0; i < m.NumPoints(); i++ {
		xy, ok := m.PointN(i).XY()
		if ok {
			sum = sum.Add(xy)
			n++
		}
	}
	if n == 0 {
		return NewEmptyPoint(DimXY)
	}
	return sum.Scale(1 / float64(n)).AsPoint()
}

// Reverse in the case of MultiPoint outputs each component point in their
// original order.
func (m MultiPoint) Reverse() MultiPoint {
	return m
}

// CoordinatesType returns the CoordinatesType used to represent points making
// up the geometry.
func (m MultiPoint) CoordinatesType() CoordinatesType {
	return m.ctype
}

// ForceCoordinatesType returns a new MultiPoint with a different CoordinatesType. If a
// dimension is added, then new values are populated with 0.
func (m MultiPoint) ForceCoordinatesType(newCType CoordinatesType) MultiPoint {
	newPoints := forceCoordinatesTypeOfPointSlice(m.points, newCType)
	return MultiPoint{newPoints, newCType}
}

// forceCoordinatesTypeOfPointSlice creates a new slice of Points, each forced
// to a new coordinates type.
func forceCoordinatesTypeOfPointSlice(pts []Point, ctype CoordinatesType) []Point {
	cp := make([]Point, len(pts))
	for i, pt := range pts {
		cp[i] = pt.ForceCoordinatesType(ctype)
	}
	return cp
}

// Force2D returns a copy of the MultiPoint with Z and M values removed.
func (m MultiPoint) Force2D() MultiPoint {
	return m.ForceCoordinatesType(DimXY)
}

// PointOnSurface returns one of the Points in the Collection.
func (m MultiPoint) PointOnSurface() Point {
	nearest := newNearestPointAccumulator(m.Centroid())
	n := m.NumPoints()
	for i := 0; i < n; i++ {
		nearest.consider(m.PointN(i).Force2D())
	}
	return nearest.point
}

func (m MultiPoint) asXYs() []XY {
	xys := make([]XY, 0, len(m.points))
	for _, pt := range m.points {
		if xy, ok := pt.XY(); ok {
			xys = append(xys, xy)
		}
	}
	return xys
}

// Dump returns the MultiPoint represented as a Point slice.
func (m MultiPoint) Dump() []Point {
	pts := make([]Point, len(m.points))
	copy(pts, m.points)
	return pts
}

// DumpCoordinates returns the non-empty points in a MultiPoint represented as
// a Sequence.
func (m MultiPoint) DumpCoordinates() Sequence {
	ctype := m.CoordinatesType()
	nonEmpty := make([]float64, 0, len(m.points)*ctype.Dimension())
	for _, pt := range m.points {
		if c, ok := pt.Coordinates(); ok {
			nonEmpty = c.appendFloat64s(nonEmpty)
		}
	}
	seq := NewSequence(nonEmpty, ctype)
	return seq
}

// Summary returns a text summary of the MultiPoint following a similar format to https://postgis.net/docs/ST_Summary.html.
func (m MultiPoint) Summary() string {
	var pointSuffix string
	numPoints := m.NumPoints()
	if numPoints != 1 {
		pointSuffix = "s"
	}
	return fmt.Sprintf("%s[%s] with %d point%s", m.Type(), m.CoordinatesType(), numPoints, pointSuffix)
}

// String returns the string representation of the MultiPoint.
func (m MultiPoint) String() string {
	return m.Summary()
}

// SnapToGrid returns a copy of the MultiPoint with all coordinates snapped to
// a base 10 grid.
//
// The grid spacing is specified by the number of decimal places to round to
// (with negative decimal places being allowed). E.g., a decimalPlaces value of
// 2 would cause all coordinates to be rounded to the nearest 0.01, and a
// decimalPlaces of -1 would cause all coordinates to be rounded to the nearest
// 10.
func (m MultiPoint) SnapToGrid(decimalPlaces int) MultiPoint {
	return m.TransformXY(snapToGridXY(decimalPlaces))
}
