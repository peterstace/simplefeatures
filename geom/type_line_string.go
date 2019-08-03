package geom

import (
	"database/sql/driver"
	"errors"
	"io"
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
func NewLineStringC(pts []Coordinates) (LineString, error) {
	var lines []Line
	for i := 0; i < len(pts)-1; i++ {
		if pts[i].XY.Equals(pts[i+1].XY) {
			continue
		}
		ln := must(NewLineC(pts[i], pts[i+1])).(Line)
		lines = append(lines, ln)
	}
	if len(lines) == 0 {
		return LineString{}, errors.New("LineString must contain at least two distinct points")
	}
	return LineString{lines}, nil
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
		dst = ln.a.X.appendAsFloat(dst)
		dst = append(dst, ' ')
		dst = ln.a.Y.appendAsFloat(dst)
		dst = append(dst, ',')
	}
	last := s.lines[len(s.lines)-1].b
	dst = last.X.appendAsFloat(dst)
	dst = append(dst, ' ')
	dst = last.Y.appendAsFloat(dst)
	return append(dst, ')')
}

// IsSimple returns true iff the curve defined by the LineString doesn't pass
// through the same point twice (with the possible exception of the two
// endpoints).
func (s LineString) IsSimple() bool {
	// 1. Check for pairwise intersection.
	//  a. Point is allowed if lines adjacent.
	//  b. Start to end is allowed if first and last line.
	n := len(s.lines)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			intersection := s.lines[i].Intersection(s.lines[j])
			if intersection.IsEmpty() {
				continue
			}
			if intersection.Dimension() >= 1 {
				// two overlapping line segments
				return false
			}
			// The intersection must be a single point.
			if i+1 == j {
				// Adjacent lines will intersect at a point due to
				// construction, so this case is okay.
				continue
			}
			if i == 0 && j == n-1 {
				// The first and last segment are allowed to intersect at a
				// point, so long as that point is the start of the first
				// segment and the end of the last segment (i.e. a linear
				// ring).
				aPt := NewPointC(s.lines[i].a)
				bPt := NewPointC(s.lines[j].b)
				if !intersection.Equals(aPt) || !intersection.Equals(bPt) {
					return false
				}
			} else {
				// Any other point intersection (e.g. looping back on
				// itself at a point) is disallowed for simple linestrings.
				return false
			}
		}
	}
	return true
}

func (s LineString) IsClosed() bool {
	return s.lines[0].a.XY.Equals(s.lines[len(s.lines)-1].b.XY)
}

func (s LineString) Intersection(g Geometry) Geometry {
	return intersection(s, g)
}

func (s LineString) IsEmpty() bool {
	return false
}

func (s LineString) Dimension() int {
	return 1
}

func (s LineString) Equals(other Geometry) bool {
	return equals(s, other)
}

func (s LineString) Envelope() (Envelope, bool) {
	env := NewEnvelope(s.lines[0].a.XY)
	for _, line := range s.lines {
		env = env.Extend(line.b.XY)
	}
	return env, true
}

func (s LineString) Boundary() Geometry {
	if s.IsClosed() {
		// Same behaviour as Postgis, but could instead be any other empty set.
		return NewMultiPoint(nil)
	}
	return NewMultiPoint([]Point{
		NewPointXY(s.lines[0].a.XY),
		NewPointXY(s.lines[len(s.lines)-1].b.XY),
	})
}

func (s LineString) Value() (driver.Value, error) {
	return s.AsText(), nil
}

func (s LineString) AsBinary(w io.Writer) error {
	marsh := newWKBMarshaller(w)
	marsh.writeByteOrder()
	marsh.writeGeomType(wkbGeomTypeLineString)
	n := s.NumPoints()
	marsh.writeCount(n)
	for i := 0; i < n; i++ {
		marsh.writeFloat64(s.PointN(i).XY().X.AsFloat())
		marsh.writeFloat64(s.PointN(i).XY().Y.AsFloat())
	}
	return marsh.err
}

func (s LineString) ConvexHull() Geometry {
	return convexHull(s)
}

func (s LineString) convexHullPointSet() []XY {
	n := s.NumPoints()
	points := make([]XY, n)
	for i := 0; i < n; i++ {
		points[i] = s.PointN(i).XY()
	}
	return points
}

func (s LineString) MarshalJSON() ([]byte, error) {
	n := s.NumPoints()
	coords := make([]Coordinates, n)
	for i := 0; i < n; i++ {
		coords[i] = s.PointN(i).Coordinates()
	}
	return marshalGeoJSON("LineString", coords)
}
