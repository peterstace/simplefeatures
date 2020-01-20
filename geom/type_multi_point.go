package geom

import (
	"bytes"
	"database/sql/driver"
	"io"
	"unsafe"
)

// MultiPoint is a 0-dimensional geometric collection of points. The points are
// not connected or ordered.
//
// Its assertions are:
//
// 1. It must be made up of 0 or more valid Points.
type MultiPoint struct {
	pts []Point
}

func NewMultiPoint(pts []Point, opts ...ConstructorOption) MultiPoint {
	return MultiPoint{pts}
}

// NewMultiPointOC creates a new MultiPoint consisting of a Point for each
// non-empty OptionalCoordinate.
func NewMultiPointOC(coords []OptionalCoordinates, opts ...ConstructorOption) MultiPoint {
	var pts []Point
	for _, c := range coords {
		if c.Empty {
			continue
		}
		pt := NewPointC(c.Value, opts...)
		pts = append(pts, pt)
	}
	return NewMultiPoint(pts, opts...)
}

// NewMultiPointC creates a new MultiPoint consisting of a point for each coordinate.
func NewMultiPointC(coords []Coordinates, opts ...ConstructorOption) MultiPoint {
	var pts []Point
	for _, c := range coords {
		pt := NewPointC(c, opts...)
		pts = append(pts, pt)
	}
	return NewMultiPoint(pts, opts...)
}

// NewMultiPointXY creates a new MultiPoint consisting of a point for each XY.
func NewMultiPointXY(pts []XY, opts ...ConstructorOption) MultiPoint {
	return NewMultiPointC(oneDimXYToCoords(pts))
}

// AsGeometry converts this MultiPoint into a Geometry.
func (m MultiPoint) AsGeometry() Geometry {
	return Geometry{multiPointTag, unsafe.Pointer(&m)}
}

// NumPoints gives the number of element points making up the MultiPoint.
func (m MultiPoint) NumPoints() int {
	return len(m.pts)
}

// PointN gives the nth (zero indexed) Point.
func (m MultiPoint) PointN(n int) Point {
	return m.pts[n]
}

func (m MultiPoint) AsText() string {
	return string(m.AppendWKT(nil))
}

func (m MultiPoint) AppendWKT(dst []byte) []byte {
	dst = append(dst, []byte("MULTIPOINT")...)
	if len(m.pts) == 0 {
		return append(dst, []byte(" EMPTY")...)
	}
	dst = append(dst, '(')
	for i, pt := range m.pts {
		dst = pt.appendWKTBody(dst)
		if i != len(m.pts)-1 {
			dst = append(dst, ',')
		}
	}
	return append(dst, ')')
}

// IsSimple returns true iff no two of its points are equal.
func (m MultiPoint) IsSimple() bool {
	seen := make(map[XY]bool)
	for _, p := range m.pts {
		if seen[p.coords.XY] {
			return false
		}
		seen[p.coords.XY] = true
	}
	return true
}

func (m MultiPoint) Intersection(g Geometry) (Geometry, error) {
	return intersection(m.AsGeometry(), g)
}

func (m MultiPoint) Intersects(g Geometry) bool {
	return hasIntersection(m.AsGeometry(), g)
}

func (m MultiPoint) IsEmpty() bool {
	return len(m.pts) == 0
}

func (m MultiPoint) Equals(other Geometry) (bool, error) {
	return equals(m.AsGeometry(), other)
}

func (m MultiPoint) Envelope() (Envelope, bool) {
	if len(m.pts) == 0 {
		return Envelope{}, false
	}
	env := NewEnvelope(m.pts[0].coords.XY)
	for _, pt := range m.pts[1:] {
		env = env.ExtendToIncludePoint(pt.coords.XY)
	}
	return env, true
}

func (m MultiPoint) Boundary() Geometry {
	// This is a little bit more complicated than it really has to be (it just
	// has to always return an empty set). However, this is the behavour of
	// Postgis.
	if m.IsEmpty() {
		return m.AsGeometry()
	}
	return NewGeometryCollection(nil).AsGeometry()
}

func (m MultiPoint) Value() (driver.Value, error) {
	var buf bytes.Buffer
	err := m.AsBinary(&buf)
	return buf.Bytes(), err
}

func (m MultiPoint) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeMultiPoint)
	n := m.NumPoints()
	marsh.writeCount(n)
	for i := 0; i < n; i++ {
		pt := m.PointN(i)
		marsh.setErr(pt.AsBinary(w))
	}
	return marsh.err
}

// ConvexHull finds the convex hull of the set of points. This may either be
// the empty set, a single point, a line, or a polygon.
func (m MultiPoint) ConvexHull() Geometry {
	return convexHull(m.AsGeometry())
}

func (m MultiPoint) MarshalJSON() ([]byte, error) {
	return marshalGeoJSON("MultiPoint", m.Coordinates())
}

// Coordinates returns the coordinates of the points represented by the
// MultiPoint.
func (m MultiPoint) Coordinates() []Coordinates {
	coords := make([]Coordinates, len(m.pts))
	for i := range coords {
		coords[i] = m.pts[i].Coordinates()
	}
	return coords
}

// TransformXY transforms this MultiPoint into another MultiPoint according to fn.
func (m MultiPoint) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (Geometry, error) {
	coords := m.Coordinates()
	transform1dCoords(coords, fn)
	return NewMultiPointC(coords, opts...).AsGeometry(), nil
}

// EqualsExact checks if this MultiPoint is exactly equal to another MultiPoint.
func (m MultiPoint) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	return other.IsMultiPoint() &&
		multiPointExactEqual(m, other.AsMultiPoint(), opts)
}

// IsValid checks if this MultiPoint is valid. However, there is no way to indicate
// whether or not MultiPoint is valid, so this function always returns true
func (m MultiPoint) IsValid() bool {
	return true
}

// Reverse in the case of MultiPoint outputs each component point in their original order.
func (m MultiPoint) Reverse() MultiPoint {
	coords := make([]Coordinates, len(m.pts))
	// Form the reversed slice.
	for i := 0; i < len(m.pts); i++ {
		coords[i] = m.pts[i].Coordinates()
	}
	return NewMultiPointC(coords)
}
