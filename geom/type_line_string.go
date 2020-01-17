package geom

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"io"
	"sort"
	"unsafe"
)

// LineString is a curve defined by linear interpolation between a finite set
// of points. Each consecutive pair of points defines a line segment.
//
// Its assertions are:
//
// 1. It must contain at least 2 distinct points.
type LineString struct {
	lines []Line
}

// NewLineStringC creates a line string from the coordinates defining its
// points.
func NewLineStringC(pts []Coordinates, opts ...ConstructorOption) (LineString, error) {
	// Fewer lines than len(pts)-1 _may_ be used, however the normal case is
	// for there to be no coincident control points, so the full capacity would
	// be utilised.
	lines := make([]Line, 0, len(pts)-1)

	for i := 0; i < len(pts)-1; i++ {
		if pts[i].XY.Equals(pts[i+1].XY) {
			continue
		}
		ln, err := NewLineC(pts[i], pts[i+1], opts...)
		if err != nil {
			// Already checked to ensure pts[i] and pts[i+1] are not equal.
			panic(err)
		}
		lines = append(lines, ln)
	}
	if doCheapValidations(opts) && len(lines) == 0 {
		return LineString{}, errors.New("LineString must contain at least two distinct points")
	}
	return LineString{lines}, nil
}

// NewLineStringXY creates a line string from the XYs defining its points.
func NewLineStringXY(pts []XY, opts ...ConstructorOption) (LineString, error) {
	return NewLineStringC(oneDimXYToCoords(pts), opts...)
}

// AsGeometry converts this LineString into a Geometry.
func (s LineString) AsGeometry() Geometry {
	return Geometry{lineStringTag, unsafe.Pointer(&s)}
}

// StartPoint gives the first point of the line string.
func (s LineString) StartPoint() Point {
	return s.lines[0].StartPoint()
}

// EndPoint gives the last point of the line string.
func (s LineString) EndPoint() Point {
	return s.lines[len(s.lines)-1].EndPoint()
}

// NumPoints gives the number of control points in the line string.
func (s LineString) NumPoints() int {
	return len(s.lines) + 1
}

// PointN gives the nth (zero indexed) point in the line string. Panics if n is
// out of range with respect to the number of points.
func (s LineString) PointN(n int) Point {
	if n == s.NumPoints()-1 {
		return s.EndPoint()
	}
	return s.lines[n].StartPoint()
}

func (s LineString) AsText() string {
	return string(s.AppendWKT(nil))
}

func (s LineString) AppendWKT(dst []byte) []byte {
	dst = append(dst, []byte("LINESTRING")...)
	return s.appendWKTBody(dst)
}

func (s LineString) appendWKTBody(dst []byte) []byte {
	dst = append(dst, '(')
	for _, ln := range s.lines {
		dst = appendFloat(dst, ln.a.X)
		dst = append(dst, ' ')
		dst = appendFloat(dst, ln.a.Y)
		dst = append(dst, ',')
	}
	last := s.lines[len(s.lines)-1].b
	dst = appendFloat(dst, last.X)
	dst = append(dst, ' ')
	dst = appendFloat(dst, last.Y)
	return append(dst, ')')
}

type lineWithIndex struct {
	ln  Line
	idx int
}

// IsSimple returns true iff the curve defined by the LineString doesn't pass
// through the same point twice (with the possible exception of the two
// endpoints being coincident).
func (s LineString) IsSimple() bool {
	// A line sweep algorithm is used, where a vertical line is swept over X
	// values (from lowest to highest). We only have consider line segments
	// that have overlapping X values when performing pairwise intersection
	// tests.
	//
	// 1. Create slice of segments, sorted by their min X coordinate.
	// 2. Loop over each segment.
	//    a. Remove any elements from the heap that have their max X less than the minX of the current segment.
	//    b. Check to see if the new element intersects with any elements in the heap.
	//    c. Insert the current element into the heap.

	n := len(s.lines)
	unprocessed := seq(n)
	sort.Slice(unprocessed, func(i, j int) bool {
		return minX(s.lines[unprocessed[i]]) < minX(s.lines[unprocessed[j]])
	})

	active := intHeap{less: func(i, j int) bool {
		return maxX(s.lines[i]) < maxX(s.lines[j])
	}}

	for _, current := range unprocessed {
		currentX := minX(s.lines[current])
		for len(active.data) != 0 && maxX(s.lines[active.data[0]]) < currentX {
			active.pop()
		}
		for _, other := range active.data {
			intersects, dim := hasIntersectionLineWithLine(s.lines[current], s.lines[other])
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
	return s.lines[0].a.XY.Equals(s.lines[len(s.lines)-1].b.XY)
}

func (s LineString) Intersection(g Geometry) (Geometry, error) {
	return intersection(s.AsGeometry(), g)
}

func (s LineString) Intersects(g Geometry) bool {
	return hasIntersection(s.AsGeometry(), g)
}

func (s LineString) IsEmpty() bool {
	return false
}

func (s LineString) Equals(other Geometry) (bool, error) {
	return equals(s.AsGeometry(), other)
}

func (s LineString) Envelope() (Envelope, bool) {
	env := NewEnvelope(s.lines[0].a.XY)
	for _, line := range s.lines {
		env = env.ExtendToIncludePoint(line.b.XY)
	}
	return env, true
}

func (s LineString) Boundary() Geometry {
	if s.IsClosed() {
		// Same behaviour as Postgis, but could instead be any other empty set.
		return NewMultiPoint(nil).AsGeometry()
	}
	return NewMultiPoint([]Point{
		NewPointXY(s.lines[0].a.XY),
		NewPointXY(s.lines[len(s.lines)-1].b.XY),
	}).AsGeometry()
}

func (s LineString) Value() (driver.Value, error) {
	var buf bytes.Buffer
	err := s.AsBinary(&buf)
	return buf.Bytes(), err
}

func (s LineString) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeLineString)
	n := s.NumPoints()
	marsh.writeCount(n)
	for i := 0; i < n; i++ {
		marsh.writeFloat64(s.PointN(i).XY().X)
		marsh.writeFloat64(s.PointN(i).XY().Y)
	}
	return marsh.err
}

func (s LineString) ConvexHull() Geometry {
	return convexHull(s.AsGeometry())
}

func (s LineString) MarshalJSON() ([]byte, error) {
	return marshalGeoJSON("LineString", s.Coordinates())
}

// Coordinates returns the coordinates of each point along the LineString.
func (s LineString) Coordinates() []Coordinates {
	n := s.NumPoints()
	coords := make([]Coordinates, n)
	for i := range coords {
		coords[i] = s.PointN(i).Coordinates()
	}
	return coords
}

// TransformXY transforms this LineString into another LineString according to fn.
func (s LineString) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (Geometry, error) {
	coords := s.Coordinates()
	transform1dCoords(coords, fn)
	ls, err := NewLineStringC(coords, opts...)
	return ls.AsGeometry(), err
}

// EqualsExact checks if this LineString is exactly equal to another curve.
func (s LineString) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	var c curve
	switch {
	case other.IsLine():
		c = other.AsLine()
	case other.IsLineString():
		c = other.AsLineString()
	default:
		return false
	}
	return curvesExactEqual(s, c, opts)
}

// IsValid checks if this LineString is valid
func (s LineString) IsValid() bool {
	_, err := NewLineStringC(s.Coordinates())
	return err == nil
}

// IsRing returns true iff this LineString is both simple and closed (i.e. is a
// linear ring).
func (s LineString) IsRing() bool {
	return s.IsSimple() && s.IsClosed()
}

// Length gives the length of the line string.
func (s LineString) Length() float64 {
	var sum float64
	for _, ln := range s.lines {
		sum += ln.Length()
	}
	return sum
}

// AsMultiLineString is a convinience function that converts this LineString
// into a MultiLineString.
func (s LineString) AsMultiLineString() MultiLineString {
	return NewMultiLineString([]LineString{s})
}
