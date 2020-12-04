package geom

import (
	"database/sql/driver"
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

// Type returns the GeometryType for a MultiPoint
func (m MultiPoint) Type() GeometryType {
	return TypeMultiPoint
}

// AsGeometry converts this MultiPoint into a Geometry.
func (m MultiPoint) AsGeometry() Geometry {
	return Geometry{TypeMultiPoint, unsafe.Pointer(&m)}
}

// NumPoints gives the number of element points making up the MultiPoint.
func (m MultiPoint) NumPoints() int {
	return m.seq.Length()
}

// PointN gives the nth (zero indexed) Point.
func (m MultiPoint) PointN(n int) Point {
	if m.empty.Get(n) {
		return NewEmptyPoint(m.CoordinatesType())
	}
	c := m.seq.Get(n)
	return NewPoint(c)
}

// AsText returns the WKT (Well Known Text) representation of this geometry.
func (m MultiPoint) AsText() string {
	return string(m.AppendWKT(nil))
}

// AppendWKT appends the WKT (Well Known Text) representation of this geometry
// to the input byte slice.
func (m MultiPoint) AppendWKT(dst []byte) []byte {
	dst = appendWKTHeader(dst, "MULTIPOINT", m.CoordinatesType())
	if m.NumPoints() == 0 {
		return appendWKTEmpty(dst)
	}
	return appendWKTSequence(dst, m.seq, true, m.empty)
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

// Intersects return true if and only if this geometry intersects with the
// other, i.e. they shared at least one common point.
func (m MultiPoint) Intersects(g Geometry) bool {
	return hasIntersection(m.AsGeometry(), g)
}

// IsEmpty return true if and only if this MultiPoint doesn't contain any
// Points, or only contains empty Points.
func (m MultiPoint) IsEmpty() bool {
	for i := 0; i < m.NumPoints(); i++ {
		if !m.empty.Get(i) {
			return false
		}
	}
	return true
}

// Envelope returns the Envelope that most tightly surrounds the geometry. If
// the geometry is empty, then false is returned.
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

// AsBinary returns the WKB (Well Known Text) representation of the geometry.
func (m MultiPoint) AsBinary() []byte {
	return m.AppendWKB(nil)
}

// AppendWKB appends the WKB (Well Known Text) representation of the geometry
// to the input slice.
func (m MultiPoint) AppendWKB(dst []byte) []byte {
	marsh := newWKBMarshaller(dst)
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

// MarshalJSON implements the encoding/json.Marshaller interface by encoding
// this geometry as a GeoJSON geometry object.
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
	n := m.seq.Length()
	xys := make([]XY, 0, n)
	for i := 0; i < n; i++ {
		if !m.empty.Get(i) {
			xys = append(xys, m.seq.GetXY(i))
		}
	}
	return xys
}
