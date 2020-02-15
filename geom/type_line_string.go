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

// LineString is a curve defined by linear interpolation between a finite set
// of points. Its zero value is the empty line string.
//
// Each consecutive pair of points defines a line segment. It must contain
// either zero points (i.e. is the empty LineString) or it must contain at
// least 2 distinct points.
type LineString struct {
	// coords have been deduplicated such that no two consecutive coordinates
	// are coincident. This allows quick calculation of Line segments.
	coords []Coordinates

	// points are indexes into coords, and retain consecutive coincident
	// points. This is so that information about the original points making up
	// the LineString are retained.
	points []int
}

// NewEmptyLineString gives the empty LineString. It is equivalent to calling
// NewLineStringC with a zero length coordinates argument.
func NewEmptyLineString() LineString {
	return LineString{}
}

// NewLineStringC creates a line string from the coordinates defining its
// points.
func NewLineStringC(pts []Coordinates, opts ...ConstructorOption) (LineString, error) {
	coords := make([]Coordinates, 0, len(pts)) // may not use full capacity
	points := make([]int, len(pts))

	for i := range pts {
		if len(coords) == 0 || pts[i].XY != coords[len(coords)-1].XY {
			coords = append(coords, pts[i])
		}
		points[i] = len(coords) - 1
	}
	if doCheapValidations(opts) && len(coords) == 1 {
		return LineString{}, errors.New("LineString must either contain zero points or at least two distinct points")
	}
	return LineString{coords, points}, nil
}

// NewLineStringXY creates a line string from the XYs defining its points.
func NewLineStringXY(pts []XY, opts ...ConstructorOption) (LineString, error) {
	return NewLineStringC(oneDimXYToCoords(pts), opts...)
}

// AsGeometry converts this LineString into a Geometry.
func (s LineString) AsGeometry() Geometry {
	return Geometry{lineStringTag, unsafe.Pointer(&s)}
}

// StartPoint gives the first point of the LineString. If the LineString is
// empty then it returns the empty Point.
func (s LineString) StartPoint() Point {
	if s.IsEmpty() {
		return NewEmptyPoint()
	}
	return NewPointC(s.coords[s.points[0]])
}

// EndPoint gives the last point of the LineString. If the LineString is empty
// then it returns the empty Point.
func (s LineString) EndPoint() Point {
	if s.IsEmpty() {
		return NewEmptyPoint()
	}
	return NewPointC(s.coords[s.points[len(s.points)-1]])
}

// NumPoints gives the number of control points in the line string.
func (s LineString) NumPoints() int {
	return len(s.points)
}

// PointN gives the coordinates of the nth (zero indexed) point in the line
// string. Panics if n is out of range with respect to the number of points.
func (s LineString) PointN(n int) Coordinates {
	return s.coords[s.points[n]]
}

// NumLines gives the number of Line segments that make up the LineString.
func (s LineString) NumLines() int {
	return max(0, len(s.coords)-1)
}

// LineN gives the nth (zero indexed) Line in the LineString. Panics if n is
// out of range with respect to the number of lines.
func (s LineString) LineN(n int) Line {
	// Line is constructed directly here, rather than via NewLineC. This is
	// because LineN is called in a tight loop in many places, and skipping the
	// constructor significantly speeds up the benchmarks.
	//
	// The two coordinates are guaranteed to not be coincident due to the way
	// that the coords slice is constructed, so this is safe.
	return Line{s.coords[n], s.coords[n+1]}
}

func (s LineString) AsText() string {
	return string(s.AppendWKT(nil))
}

func (s LineString) AppendWKT(dst []byte) []byte {
	dst = append(dst, "LINESTRING"...)
	if s.IsEmpty() {
		dst = append(dst, ' ')
	}
	return s.appendWKTBody(dst)
}

func (s LineString) appendWKTBody(dst []byte) []byte {
	if s.IsEmpty() {
		return append(dst, "EMPTY"...)
	}

	dst = append(dst, '(')
	for i, ptIdx := range s.points {
		if i > 0 {
			dst = append(dst, ',')
		}
		c := s.coords[ptIdx]
		dst = appendFloat(dst, c.X)
		dst = append(dst, ' ')
		dst = appendFloat(dst, c.Y)
	}
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
	// values (from lowest to highest). We only have to consider line segments
	// that have overlapping X values when performing pairwise intersection
	// tests.
	//
	// 1. Create slice of segments, sorted by their min X coordinate.
	// 2. Loop over each segment.
	//    a. Remove any elements from the heap that have their max X less than the minX of the current segment.
	//    b. Check to see if the new element intersects with any elements in the heap.
	//    c. Insert the current element into the heap.

	n := s.NumLines()
	unprocessed := seq(n)
	sort.Slice(unprocessed, func(i, j int) bool {
		return minX(s.LineN(unprocessed[i])) < minX(s.LineN(unprocessed[j]))
	})

	active := intHeap{less: func(i, j int) bool {
		return maxX(s.LineN(i)) < maxX(s.LineN(j))
	}}

	for _, current := range unprocessed {
		currentX := minX(s.LineN(current))
		for len(active.data) != 0 && maxX(s.LineN(active.data[0])) < currentX {
			active.pop()
		}
		for _, other := range active.data {
			intersects, dim := hasIntersectionLineWithLine(s.LineN(current), s.LineN(other))
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
	return !s.IsEmpty() && s.StartPoint().XY() == s.EndPoint().XY()
}

func (s LineString) Intersection(g Geometry) (Geometry, error) {
	return intersection(s.AsGeometry(), g)
}

func (s LineString) Intersects(g Geometry) bool {
	return hasIntersection(s.AsGeometry(), g)
}

func (s LineString) IsEmpty() bool {
	return len(s.coords) == 0
}

func (s LineString) Equals(other Geometry) (bool, error) {
	return equals(s.AsGeometry(), other)
}

func (s LineString) Envelope() (Envelope, bool) {
	if s.IsEmpty() {
		return Envelope{}, false
	}
	env := NewEnvelope(s.coords[0].XY)
	for _, c := range s.coords[1:] {
		env = env.ExtendToIncludePoint(c.XY)
	}
	return env, true
}

func (s LineString) Boundary() MultiPoint {
	var pts []Point
	if !s.IsClosed() {
		pts = append(pts, s.StartPoint(), s.EndPoint())
	}
	return NewMultiPoint(pts)
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
		marsh.writeFloat64(s.PointN(i).X)
		marsh.writeFloat64(s.PointN(i).Y)
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
	tmp := make([]Coordinates, len(s.coords))
	copy(tmp, s.coords)
	return tmp
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
	return s.IsClosed() && s.IsSimple()
}

// Length gives the length of the line string.
func (s LineString) Length() float64 {
	var sum float64
	for i := 0; i+1 < len(s.coords); i++ {
		dx := s.coords[i].X - s.coords[i+1].X
		dy := s.coords[i].Y - s.coords[i+1].Y
		sum += math.Sqrt(dx*dx + dy*dy)
	}
	return sum
}

// AsMultiLineString is a convinience function that converts this LineString
// into a MultiLineString.
func (s LineString) AsMultiLineString() MultiLineString {
	return NewMultiLineString([]LineString{s})
}

// Reverse in the case of LineString outputs the coordinates in reverse order.
func (s LineString) Reverse() LineString {
	coords := s.Coordinates()
	// Reverse the slice.
	for left, right := 0, len(coords)-1; left < right; left, right = left+1, right-1 {
		coords[left], coords[right] = coords[right], coords[left]
	}
	s2, err := NewLineStringC(coords)
	if err != nil {
		panic("Reverse of an existing LineString should not fail")
	}
	return s2
}
