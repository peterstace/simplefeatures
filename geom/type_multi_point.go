package geom

import (
	"bytes"
	"database/sql/driver"
	"io"
	"unsafe"
)

// MultiPoint is a 0-dimensional geometry that is a collection of points. Its
// zero value is the empty MultiPoint (i.e. a collection of zero points) with
// 2D coordinates type. It is immutable after creation.
type MultiPoint struct {
	seq   Sequence
	empty BitSet
}

// NewMultiPointFromPoints creates a MultiPoint from a list of Points. The
// coordinate type of the MultiPoint is the lowest common coordinates type of
// its Points.
func NewMultiPointFromPoints(pts []Point, opts ...ConstructorOption) MultiPoint {
	if len(pts) == 0 {
		return MultiPoint{}
	}

	ctype := DimXYZM
	for _, p := range pts {
		ctype &= p.CoordinatesType()
	}

	var empty BitSet
	floats := make([]float64, 0, len(pts)*ctype.Dimension())
	for i, pt := range pts {
		c, ok := pt.Coordinates()
		if !ok {
			empty.Set(i, true)
		}
		floats = append(floats, c.X, c.Y)
		if ctype.Is3D() {
			floats = append(floats, c.Z)
		}
		if ctype.IsMeasured() {
			floats = append(floats, c.M)
		}
	}
	seq := NewSequence(floats, ctype)
	return NewMultiPointWithEmptyMask(seq, empty, opts...)
}

// NewMultiPoint creates a new MultiPoint from a sequence of Coordinates.
func NewMultiPoint(seq Sequence, opts ...ConstructorOption) MultiPoint {
	return MultiPoint{seq, BitSet{}}
}

// NewMultiPointWithEmptyMask creates a new MultiPoint from a sequence of
// coordinates. If there are any positions set in the BitSet, then these are
// used to indicate that the corresponding point in the sequence is an empty
// point.
func NewMultiPointWithEmptyMask(seq Sequence, empty BitSet, opts ...ConstructorOption) MultiPoint {
	return MultiPoint{
		seq,
		empty.Clone(), // clone so that the caller doesn't have access to the interal empty set
	}
}

// Type return type string for MultiPoint
func (m MultiPoint) Type() string {
	return multiPointType
}

// AsGeometry converts this MultiPoint into a Geometry.
func (m MultiPoint) AsGeometry() Geometry {
	return Geometry{multiPointTag, unsafe.Pointer(&m)}
}

// NumPoints gives the number of element points making up the MultiPoint.
func (m MultiPoint) NumPoints() int {
	return m.seq.Length()
}

// PointN gives the nth (zero indexed) Point.
func (m MultiPoint) PointN(n int) Point {
	if m.empty.Get(n) {
		return NewEmptyPoint(m.CoordinatesType())
	} else {
		c := m.seq.Get(n)
		return NewPoint(c)
	}
}

func (m MultiPoint) AsText() string {
	return string(m.AppendWKT(nil))
}

func (m MultiPoint) AppendWKT(dst []byte) []byte {
	dst = appendWKTHeader(dst, "MULTIPOINT", m.CoordinatesType())
	if m.NumPoints() == 0 {
		return appendWKTEmpty(dst)
	}
	return appendWKTSequence(dst, m.seq, true, m.empty)
}

// IsSimple returns true iff no two of its points are equal.
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

func (m MultiPoint) Intersection(g Geometry) (Geometry, error) {
	return intersection(m.AsGeometry(), g)
}

func (m MultiPoint) Intersects(g Geometry) bool {
	return hasIntersection(m.AsGeometry(), g)
}

func (m MultiPoint) IsEmpty() bool {
	for i := 0; i < m.NumPoints(); i++ {
		if !m.empty.Get(i) {
			return false
		}
	}
	return true
}

func (m MultiPoint) Envelope() (Envelope, bool) {
	var has bool
	var env Envelope
	for i := 0; i < m.NumPoints(); i++ {
		xy, ok := m.PointN(i).XY()
		if !ok {
			continue
		}
		if has {
			env = env.ExtendToIncludePoint(xy)
		} else {
			env = NewEnvelope(xy)
			has = true
		}
	}
	return env, has
}

func (m MultiPoint) Boundary() GeometryCollection {
	return GeometryCollection{}
}

func (m MultiPoint) Value() (driver.Value, error) {
	var buf bytes.Buffer
	err := m.AsBinary(&buf)
	return buf.Bytes(), err
}

func (m MultiPoint) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeMultiPoint, m.CoordinatesType())
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
	var dst []byte
	dst = append(dst, `{"type":"MultiPoint","coordinates":`...)
	dst = appendGeoJSONSequence(dst, m.seq, m.empty)
	dst = append(dst, '}')
	return dst, nil
}

// Coordinates returns the coordinates of the points represented by the
// MultiPoint. If a point has its corresponding bit set to true in the BitSet,
// then that point is empty.
func (m MultiPoint) Coordinates() (seq Sequence, empty BitSet) {
	// TODO: If we had a read-only BitSet, then we could avoid the clone here.
	return m.seq, m.empty.Clone()
}

// TransformXY transforms this MultiPoint into another MultiPoint according to fn.
func (m MultiPoint) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (MultiPoint, error) {
	transformed := transformSequence(m.seq, fn)
	return NewMultiPointWithEmptyMask(transformed, m.empty, opts...), nil
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
	return NewPointFromXY(sum.Scale(1 / float64(n)))
}

// Reverse in the case of MultiPoint outputs each component point in their
// original order.
func (m MultiPoint) Reverse() MultiPoint {
	return m
}

// CoordinatesType returns the CoordinatesType used to represent points making
// up the geometry.
func (m MultiPoint) CoordinatesType() CoordinatesType {
	return m.seq.CoordinatesType()
}

// ForceCoordinatesType returns a new MultiPoint with a different CoordinatesType. If a
// dimension is added, then new values are populated with 0.
func (m MultiPoint) ForceCoordinatesType(newCType CoordinatesType) MultiPoint {
	return MultiPoint{m.seq.ForceCoordinatesType(newCType), m.empty}
}

// Force2D returns a copy of the MultiPoint with Z and M values removed.
func (m MultiPoint) Force2D() MultiPoint {
	return m.ForceCoordinatesType(DimXY)
}
