package geom

import (
	"database/sql/driver"
	"errors"
	"math"
	"unsafe"

	"github.com/peterstace/simplefeatures/rtree"
)

// LineString is a linear geometry defined by linear interpolation between a
// finite set of points. Its zero value is the empty line string. It is
// immutable after creation.
//
// A LineString must consist of either zero points (i.e. it is the empty line
// string), or it must have at least 2 points with distinct XY values.
type LineString struct {
	seq Sequence
}

// NewLineString creates a new LineString from a Sequence of points. The
// sequence must contain exactly 0 points, or at least 2 points with distinct
// XY values (otherwise an error is returned).
func NewLineString(seq Sequence, opts ...ConstructorOption) (LineString, error) {
	n := seq.Length()
	if skipValidations(opts) || n == 0 {
		return LineString{seq}, nil
	}

	// Valid non-empty LineStrings must have at least 2 *distinct* points.
	first := seq.GetXY(0)
	for i := 1; i < n; i++ {
		if seq.GetXY(i) != first {
			return LineString{seq}, nil
		}
	}
	return LineString{}, errors.New("non-empty LineStrings " +
		"must contain at least 2 points with distinct XY values")
}

// Type returns the GeometryType for a LineString
func (s LineString) Type() GeometryType {
	return TypeLineString
}

// AsGeometry converts this LineString into a Geometry.
func (s LineString) AsGeometry() Geometry {
	return Geometry{TypeLineString, unsafe.Pointer(&s)}
}

// StartPoint gives the first point of the LineString. If the LineString is
// empty then it returns the empty Point.
func (s LineString) StartPoint() Point {
	if s.IsEmpty() {
		return NewEmptyPoint(s.CoordinatesType())
	}
	return NewPoint(s.seq.Get(0))
}

// EndPoint gives the last point of the LineString. If the LineString is empty
// then it returns the empty Point.
func (s LineString) EndPoint() Point {
	if s.IsEmpty() {
		return NewEmptyPoint(s.CoordinatesType())
	}
	return NewPoint(s.seq.Get(s.seq.Length() - 1))
}

// AsText returns the WKT (Well Known Text) representation of this geometry.
func (s LineString) AsText() string {
	return string(s.AppendWKT(nil))
}

// AppendWKT appends the WKT (Well Known Text) representation of this geometry
// to the input byte slice.
func (s LineString) AppendWKT(dst []byte) []byte {
	dst = appendWKTHeader(dst, "LINESTRING", s.CoordinatesType())
	return s.appendWKTBody(dst)
}

func (s LineString) appendWKTBody(dst []byte) []byte {
	if s.IsEmpty() {
		return appendWKTEmpty(dst)
	}
	return appendWKTSequence(dst, s.seq, false, BitSet{})
}

// IsSimple returns true if this geometry contains no anomalous geometry
// points, such as self intersection or self tangency. LineStrings are simple
// if and only if the curve defined by the LineString doesn't pass through the
// same point twice (with the except of the two endpoints being coincident).
func (s LineString) IsSimple() bool {
	first, last, ok := firstAndLastLines(s.seq)
	if !ok {
		return true
	}

	// We need to track the index of the previous line segments, so that we can
	// ignore the case where adjacent line segments intersect at a point (due
	// to their construction). We can just subtract 1 from the current index,
	// since there could be duplicated vertices in the middle of the sequence.
	prev := -1

	var tree rtree.RTree
	defer tree.Recycle()
	n := s.seq.Length()
	for i := 0; i < n; i++ {
		ln, ok := getLine(s.seq, i)
		if !ok {
			continue
		}
		simple := true // assume simple until proven otherwise
		box := ln.envelope().box()
		tree.RangeSearch(box, func(j int) error {
			other, ok := getLine(s.seq, j)
			if !ok {
				// We previously were able to access line j (otherwise we
				// wouldn't have been able to put it into the tree). So if we
				// can't access it now, something has gone horribly wrong.
				panic("couldn't get line")
			}

			inter := ln.intersectLine(other)
			if inter.empty {
				return nil
			}
			if inter.ptA != inter.ptB {
				// Two overlapping line segments.
				simple = false
				return rtree.Stop
			}

			// The dimension must be 1. Since the intersection is between two
			// lines, the intersection must be a *single* point.

			if j == prev {
				// Adjacent lines will intersect at a point due to
				// construction, so this case is okay.
				return nil
			}

			// The first and last segment are allowed to intersect at a point,
			// so long as that point is the start of the first segment and the
			// end of the last segment (i.e. the line string is closed).
			if i == last && j == first && s.IsClosed() {
				return nil
			}

			// Any other point intersection (e.g. looping back on
			// itself at a point) is disallowed for simple linestrings.
			simple = false
			return rtree.Stop
		})

		if !simple {
			return false
		}
		tree.Insert(box, i)
		prev = i
	}
	return true
}

// IsClosed returns true if and only if this LineString is not empty and its
// start and end points are coincident.
func (s LineString) IsClosed() bool {
	return !s.IsEmpty() && s.seq.GetXY(0) == s.seq.GetXY(s.seq.Length()-1)
}

// Intersects return true if and only if this geometry intersects with the
// other, i.e. they shared at least one common point.
func (s LineString) Intersects(g Geometry) bool {
	return hasIntersection(s.AsGeometry(), g)
}

// IsEmpty returns true if and only if this LineString is the empty LineString.
// The empty LineString is defined by a zero length coordinates sequence.
func (s LineString) IsEmpty() bool {
	return s.seq.Length() == 0
}

// Envelope returns the Envelope that most tightly surrounds the geometry. If
// the geometry is empty, then false is returned.
func (s LineString) Envelope() (Envelope, bool) {
	n := s.seq.Length()
	if n == 0 {
		return Envelope{}, false
	}
	env := NewEnvelope(s.seq.GetXY(0))
	for i := 1; i < n; i++ {
		env = env.ExtendToIncludePoint(s.seq.GetXY(i))
	}
	return env, true
}

// Boundary returns the spatial boundary of this LineString. For closed
// LineStrings (i.e. LineStrings where the start and end points have the same
// XY value), this is the empty MultiPoint. For non-closed LineStrings, this is
// the MultiPoint containing the two endpoints of the LineString.
func (s LineString) Boundary() MultiPoint {
	var fs []float64
	if !s.IsClosed() {
		xy1 := s.seq.GetXY(0)
		xy2 := s.seq.GetXY(s.seq.Length() - 1)
		fs = []float64{
			xy1.X, xy1.Y,
			xy2.X, xy2.Y,
		}
	}
	return NewMultiPoint(NewSequence(fs, DimXY))
}

// Value implements the database/sql/driver.Valuer interface by returning the
// WKB (Well Known Binary) representation of this Geometry.
func (s LineString) Value() (driver.Value, error) {
	return s.AsBinary(), nil
}

// AsBinary returns the WKB (Well Known Text) representation of the geometry.
func (s LineString) AsBinary() []byte {
	return s.AppendWKB(nil)
}

// AppendWKB appends the WKB (Well Known Text) representation of the geometry
// to the input slice.
func (s LineString) AppendWKB(dst []byte) []byte {
	marsh := newWKBMarshaller(dst)
	marsh.writeByteOrder()
	marsh.writeGeomType(TypeLineString, s.CoordinatesType())
	marsh.writeSequence(s.seq)
	return marsh.buf
}

// ConvexHull returns the geometry representing the smallest convex geometry
// that contains this geometry.
func (s LineString) ConvexHull() Geometry {
	return convexHull(s.AsGeometry())
}

// MarshalJSON implements the encoding/json.Marshaller interface by encoding
// this geometry as a GeoJSON geometry object.
func (s LineString) MarshalJSON() ([]byte, error) {
	var dst []byte
	dst = append(dst, `{"type":"LineString","coordinates":`...)
	dst = appendGeoJSONSequence(dst, s.seq, BitSet{})
	dst = append(dst, '}')
	return dst, nil
}

// Coordinates returns the coordinates of each point along the LineString.
func (s LineString) Coordinates() Sequence {
	return s.seq
}

// TransformXY transforms this LineString into another LineString according to fn.
func (s LineString) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (LineString, error) {
	transformed := transformSequence(s.seq, fn)
	return NewLineString(transformed, opts...)
}

// EqualsExact checks if this LineString is exactly equal to another curve.
func (s LineString) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	return other.IsLineString() && curvesExactEqual(
		s.Coordinates(),
		other.AsLineString().Coordinates(),
		opts,
	)
}

// IsRing returns true iff this LineString is both simple and closed (i.e. is a
// linear ring).
func (s LineString) IsRing() bool {
	return s.IsClosed() && s.IsSimple()
}

// Length gives the length of the line string.
func (s LineString) Length() float64 {
	var sum float64
	n := s.seq.Length()
	for i := 0; i+1 < n; i++ {
		xyA := s.seq.GetXY(i)
		xyB := s.seq.GetXY(i + 1)
		delta := xyA.Sub(xyB)
		sum += math.Sqrt(delta.Dot(delta))
	}
	return sum
}

// Centroid gives the centroid of the coordinates of the line string.
func (s LineString) Centroid() Point {
	sumXY, sumLength := sumCentroidAndLengthOfLineString(s)
	if sumLength == 0 {
		return NewEmptyPoint(DimXY)
	}
	return NewPointFromXY(sumXY.Scale(1.0 / sumLength))
}

func sumCentroidAndLengthOfLineString(s LineString) (sumXY XY, sumLength float64) {
	seq := s.Coordinates()
	for i := 0; i < seq.Length(); i++ {
		ln, ok := getLine(seq, i)
		if !ok {
			continue
		}
		length := ln.length()
		cent := ln.centroid()
		sumXY = sumXY.Add(cent.Scale(length))
		sumLength += length
	}
	return sumXY, sumLength
}

// AsMultiLineString is a convenience function that converts this LineString
// into a MultiLineString.
func (s LineString) AsMultiLineString() MultiLineString {
	return NewMultiLineStringFromLineStrings([]LineString{s})
}

// Reverse in the case of LineString outputs the coordinates in reverse order.
func (s LineString) Reverse() LineString {
	return LineString{s.seq.Reverse()}
}

// CoordinatesType returns the CoordinatesType used to represent points making
// up the geometry.
func (s LineString) CoordinatesType() CoordinatesType {
	return s.seq.CoordinatesType()
}

// ForceCoordinatesType returns a new LineString with a different CoordinatesType. If a
// dimension is added, then new values are populated with 0.
func (s LineString) ForceCoordinatesType(newCType CoordinatesType) LineString {
	return LineString{s.seq.ForceCoordinatesType(newCType)}
}

// Force2D returns a copy of the LineString with Z and M values removed.
func (s LineString) Force2D() LineString {
	return s.ForceCoordinatesType(DimXY)
}

func (s LineString) asLines() []line {
	n := s.seq.Length()
	lines := make([]line, 0, max(0, n-1))
	for i := 0; i < n; i++ {
		ln, ok := getLine(s.seq, i)
		if ok {
			lines = append(lines, ln)
		}
	}
	return lines
}

// PointOnSurface returns a Point on the LineString.
func (s LineString) PointOnSurface() Point {
	// Look for the control point on the LineString (other than the first and
	// last point) that is closest to the centroid.
	n := s.seq.Length()
	nearest := newNearestPointAccumulator(s.Centroid())
	for i := 1; i < n-1; i++ {
		candidate := NewPointFromXY(s.seq.GetXY(i))
		nearest.consider(candidate)
	}
	if !nearest.point.IsEmpty() {
		return nearest.point
	}

	// Consider the star/end points if we don't have anything yet.
	nearest.consider(s.StartPoint().Force2D())
	nearest.consider(s.EndPoint().Force2D())
	return nearest.point

}
