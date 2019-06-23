package simplefeatures

import (
	"errors"
	"strconv"
)

// LineString is a curve defined by linear interpolation between a finite set
// of points. Each consecutive pair of points defines a line segment. It must
// contain at least 2 distinct points.
type LineString struct {
	lines []Line
}

// NewLineString creates a line string from the coordinates defining its
// points.
func NewLineString(pts []Coordinates) (LineString, error) {
	// Must have at least 2 distinct points.
	err := errors.New("LineString must contain at least two distinct points")
	if len(pts) == 0 {
		return LineString{}, err
	}
	var twoDistinct bool
	for _, pt := range pts[1:] {
		if !xyeq(pt.XY, pts[0].XY) {
			twoDistinct = true
			break
		}
	}
	if !twoDistinct {
		return LineString{}, err
	}

	var lines []Line
	for i := 0; i < len(pts)-1; i++ {
		if xyeq(pts[i].XY, pts[i+1].XY) {
			continue
		}
		ln, err := NewLine(pts[i], pts[i+1])
		if err != nil {
			panic(err)
		}
		lines = append(lines, ln)
	}

	return LineString{lines}, nil
}

func (s LineString) AsText() []byte {
	return s.AppendWKT(nil)
}

func (s LineString) AppendWKT(dst []byte) []byte {
	dst = append(dst, []byte("LINESTRING")...)
	return s.appendWKTBody(dst)
}

func (s LineString) appendWKTBody(dst []byte) []byte {
	dst = append(dst, '(')
	for _, ln := range s.lines {
		dst = strconv.AppendFloat(dst, ln.a.X.AsFloat(), 'f', -1, 64)
		dst = append(dst, ' ')
		dst = strconv.AppendFloat(dst, ln.a.Y.AsFloat(), 'f', -1, 64)
		dst = append(dst, ',')
	}
	last := s.lines[len(s.lines)-1].b
	dst = strconv.AppendFloat(dst, last.X.AsFloat(), 'f', -1, 64)
	dst = append(dst, ' ')
	dst = strconv.AppendFloat(dst, last.Y.AsFloat(), 'f', -1, 64)
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
				aPt, err := NewPointFromCoords(s.lines[i].a)
				if err != nil {
					panic(err)
				}
				bPt, err := NewPointFromCoords(s.lines[j].b)
				if err != nil {
					panic(err)
				}
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
	return xyeq(s.lines[0].a.XY, s.lines[len(s.lines)-1].b.XY)
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
