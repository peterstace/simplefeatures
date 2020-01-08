package geom

import (
	"bytes"
	"container/heap"
	"database/sql/driver"
	"errors"
	"io"
	"math"
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
	var lines []Line
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

type lineWithIndexHeap []lineWithIndex

func (h *lineWithIndexHeap) Len() int {
	return len(*h)
}
func (h *lineWithIndexHeap) Less(i, j int) bool {
	return maxX((*h)[i].ln) < maxX((*h)[j].ln)
}
func (h *lineWithIndexHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}
func (h *lineWithIndexHeap) Push(x interface{}) {
	*h = append(*h, x.(lineWithIndex))
}
func (h *lineWithIndexHeap) Pop() interface{} {
	e := (*h)[len(*h)-1]
	*h = (*h)[:len(*h)-1]
	return e
}

func minX(ln Line) float64 {
	return math.Min(ln.StartPoint().XY().X, ln.EndPoint().XY().X)
}

func maxX(ln Line) float64 {
	return math.Max(ln.StartPoint().XY().X, ln.EndPoint().XY().X)
}

// IsSimple returns true iff the curve defined by the LineString doesn't pass
// through the same point twice (with the possible exception of the two
// endpoints being coincident).
func (s LineString) IsSimple() bool {
	// 1. Create slice of segments along with their index.
	// 2. Loop over each segment.
	//    a. Remove any elements from the heap that have their max X less than the minX of the current segment.
	//    b. Check to see if the new element intersects with any elements in the heap.
	//    c. Insert the element into the heap.

	n := len(s.lines)
	unprocessed := make([]lineWithIndex, n)
	for i, ln := range s.lines {
		unprocessed[i] = lineWithIndex{ln, i}
	}
	sort.Slice(unprocessed, func(i, j int) bool {
		return maxX(unprocessed[i].ln) < maxX(unprocessed[j].ln)
	})

	var active lineWithIndexHeap
	for _, current := range unprocessed {
		currentX := minX(current.ln)
		for len(active) != 0 && maxX(active[0].ln) < currentX {
			heap.Pop(&active)
		}
		for _, other := range active {
			intersection := mustIntersection(current.ln.AsGeometry(), other.ln.AsGeometry())
			if intersection.IsEmpty() {
				continue
			}
			if intersection.Dimension() >= 1 {
				// two overlapping line segments
				return false
			}
			// The intersection must be a single point.
			if abs(current.idx-other.idx) == 1 {
				// Adjacent lines will intersect at a point due to
				// construction, so this case is okay.
				continue
			}

			// The first and last segment are allowed to intersect at a
			// point, so long as that point is the start of the first
			// segment and the end of the last segment (i.e. a linear
			// ring).
			if (current.idx == 0 && other.idx == n-1) || (current.idx == n-1 && other.idx == 0) {
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
		heap.Push(&active, current)
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
