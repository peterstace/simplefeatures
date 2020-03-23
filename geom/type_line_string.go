package geom

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"io"
	"math"
	"sort"
	"unsafe"
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

// Type return type string for LineString
func (s LineString) Type() string {
	return lineStringType
}

// AsGeometry converts this LineString into a Geometry.
func (s LineString) AsGeometry() Geometry {
	return Geometry{lineStringTag, unsafe.Pointer(&s)}
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
	// A line sweep algorithm is used, where a vertical line is swept over X
	// values (from lowest to highest). We only have to consider line segments
	// that have overlapping X values when performing pairwise intersection
	// tests.
	//
	// 1. Create slice of segments, sorted by their min X coordinate.
	// 2. Loop over each segment.
	//    a. Remove any elements from the heap that have their max X less than the minX of the current segment.
	//    b. Check to see if the new element intersects with any elements in the heap.
	//    c. Insert the current element into the heap.

	if s.IsEmpty() {
		return true
	}

	lines := make([]Line, 0, s.seq.Length()-1)
	for i := 0; i < s.seq.Length(); i++ {
		ln, ok := getLine(s.seq, i)
		if ok {
			lines = append(lines, ln)
		}
	}

	n := len(lines)
	unprocessed := intSequence(n)
	sort.Slice(unprocessed, func(i, j int) bool {
		return minX(lines[unprocessed[i]]) < minX(lines[unprocessed[j]])
	})

	active := intHeap{less: func(i, j int) bool {
		return maxX(lines[i]) < maxX(lines[j])
	}}

	for _, current := range unprocessed {
		currentX := minX(lines[current])
		for len(active.data) != 0 && maxX(lines[active.data[0]]) < currentX {
			active.pop()
		}
		for _, other := range active.data {
			intersects, dim := hasIntersectionLineWithLine(lines[current], lines[other])
			if !intersects {
				continue
			}
			if dim >= 1 {
				// Two overlapping line segments.
				return false
			}

			// The dimension must be 1. Since the intersection is between two
			// Lines, the intersection must be a *single* point.

			if abs(current-other) == 1 {
				// Adjacent lines will intersect at a point due to
				// construction, so this case is okay.
				continue
			}

			// The first and last segment are allowed to intersect at a point,
			// so long as that point is the start of the first segment and the
			// end of the last segment (i.e. the line string is closed).
			if (current == 0 && other == n-1) || (current == n-1 && other == 0) {
				if s.IsClosed() {
					continue
				} else {
					return false
				}
			}

			// Any other point intersection (e.g. looping back on
			// itself at a point) is disallowed for simple linestrings.
			return false
		}
		active.push(current)
	}
	return true
}

// IsClosed returns true if and only if this LineString is not empty and its
// start and end points are coincident.
func (s LineString) IsClosed() bool {
	return !s.IsEmpty() && s.seq.GetXY(0) == s.seq.GetXY(s.seq.Length()-1)
}

// Intersection calculates the of this geometry and another, i.e. the portion
// of the two geometries that are shared. It is not implemented for all
// geometry pairs, and returns an error for those cases.
func (s LineString) Intersection(g Geometry) (Geometry, error) {
	return intersection(s.AsGeometry(), g)
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
	var buf bytes.Buffer
	err := s.AsBinary(&buf)
	return buf.Bytes(), err
}

// AsBinary writes the WKB (Well Known Binary) representation of the geometry
// to the writer.
func (s LineString) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeLineString, s.CoordinatesType())
	marsh.writeSequence(s.seq)
	return marsh.err
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
	var otherSeq Sequence
	switch {
	case other.IsLine():
		otherSeq = other.AsLine().Coordinates()
	case other.IsLineString():
		otherSeq = other.AsLineString().Coordinates()
	default:
		return false
	}
	return curvesExactEqual(s.Coordinates(), otherSeq, opts)
}

// IsValid checks if this LineString is valid
func (s LineString) IsValid() bool {
	_, err := NewLineString(s.Coordinates())
	return err == nil
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
		length := ln.Length()
		cent, ok := ln.Centroid().XY()
		if ok {
			sumXY = sumXY.Add(cent.Scale(length))
			sumLength += length
		}
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
