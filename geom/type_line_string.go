package geom

import (
	"bytes"
	"database/sql/driver"
	"io"
	"math"
	"sort"
	"unsafe"
)

// LineString is a curve defined by linear interpolation between a finite set
// of points. Its zero value is the empty line string.
//
// Each consecutive pair of points defines a line segment. It must contain
// either zero points (i.e. is the empty LineString) or it must contain at
// least 2 distinct points.
type LineString struct {
	seq Sequence
}

// NewEmptyLineString gives the empty LineString. It is equivalent to calling
// NewLineStringC with a zero length coordinates argument.
func NewEmptyLineString(ctype CoordinatesType) LineString {
	return LineString{NewSequence(nil, ctype)}
}

func NewLineStringFromSequence(seq Sequence, opts ...ConstructorOption) (LineString, error) {
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
	return LineString{}, ValidationError{
		"non-empty LineStrings must contain at least 2 distinct points"}
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
	ctype := s.CoordinatesType()
	if s.IsEmpty() {
		return NewEmptyPoint(ctype)
	}
	return NewPointC(s.seq.Get(0), ctype)
}

// EndPoint gives the last point of the LineString. If the LineString is empty
// then it returns the empty Point.
func (s LineString) EndPoint() Point {
	ctype := s.CoordinatesType()
	if s.IsEmpty() {
		return NewEmptyPoint(ctype)
	}
	return NewPointC(s.seq.Get(s.seq.Length()-1), ctype)
}

func (s LineString) AsText() string {
	return string(s.AppendWKT(nil))
}

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

// IsSimple returns true iff the curve defined by the LineString doesn't pass
// through the same point twice (with the possible exception of the two
// endpoints being coincident).
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
	iter := newLineStringIterator(s)
	for iter.next() {
		lines = append(lines, iter.line())
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

func (s LineString) IsClosed() bool {
	return !s.IsEmpty() && s.seq.GetXY(0) == s.seq.GetXY(s.seq.Length()-1)
}

func (s LineString) Intersection(g Geometry) (Geometry, error) {
	return intersection(s.AsGeometry(), g)
}

func (s LineString) Intersects(g Geometry) bool {
	return hasIntersection(s.AsGeometry(), g)
}

func (s LineString) IsEmpty() bool {
	return s.seq.Length() == 0
}

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
	return NewMultiPointFromSequence(NewSequenceNoCopy(fs, XYOnly), BitSet{})
}

func (s LineString) Value() (driver.Value, error) {
	var buf bytes.Buffer
	err := s.AsBinary(&buf)
	return buf.Bytes(), err
}

func (s LineString) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeLineString, s.CoordinatesType())
	marsh.writeSequence(s.seq)
	return marsh.err
}

func (s LineString) ConvexHull() Geometry {
	return convexHull(s.AsGeometry())
}

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
	ls, err := NewLineStringFromSequence(transformed, opts...)
	return ls, err
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
	_, err := NewLineStringFromSequence(s.Coordinates())
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
		return NewEmptyPoint(XYOnly)
	}
	return NewPointXY(sumXY.Scale(1.0 / sumLength))
}

func sumCentroidAndLengthOfLineString(s LineString) (sumXY XY, sumLength float64) {
	iter := newLineStringIterator(s)
	for iter.next() {
		line := iter.line()
		length := line.Length()
		cent, ok := line.Centroid().XY()
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
	mls, err := NewMultiLineString([]LineString{s})
	if err != nil {
		// Because there is only a single line string, this can't panic due to
		// mixed coordinate type.
		panic(err)
	}
	return mls
}

// Reverse in the case of LineString outputs the coordinates in reverse order.
func (s LineString) Reverse() LineString {
	rev, err := NewLineStringFromSequence(s.seq.Reverse())
	if err != nil {
		panic("Reverse of an existing LineString should not fail")
	}
	return rev
}

func (s LineString) CoordinatesType() CoordinatesType {
	return s.seq.CoordinatesType()
}
