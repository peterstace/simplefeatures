package simplefeatures

import (
	"errors"
	"strconv"
)

// LineString is a curve defined by linear interpolation between a finite set
// of points. Each consecutive pair of points defines a line segment. It must
// contain at least 2 distinct points.
type LineString struct {
	pts []Coordinates
}

// NewLineString creates a line string from the coordinates defining its
// points.
func NewLineString(pts []Coordinates) (LineString, error) {
	// Empty line string.
	if len(pts) == 0 {
		return LineString{}, nil
	}

	for _, pt := range pts {
		if err := pt.Validate(); err != nil {
			return LineString{}, err
		}
	}

	// Must have at least 2 distinct points.
	var twoDistinct bool
	for _, pt := range pts[1:] {
		if pt.XY != pts[0].XY {
			twoDistinct = true
			break
		}
	}
	if !twoDistinct {
		return LineString{}, errors.New("LineString must contain either zero or at least two distinct points")
	}

	return LineString{pts}, nil
}

func (s LineString) AsText() []byte {
	return s.AppendWKT(nil)
}

func (s LineString) AppendWKT(dst []byte) []byte {
	dst = append(dst, []byte("LINESTRING")...)
	if len(s.pts) == 0 {
		dst = append(dst, ' ')
	}
	return s.appendWKTBody(dst)
}

func (s LineString) appendWKTBody(dst []byte) []byte {
	if len(s.pts) == 0 {
		return append(dst, []byte("EMPTY")...)
	}
	dst = append(dst, '(')
	for i, pt := range s.pts {
		dst = strconv.AppendFloat(dst, pt.X, 'f', -1, 64)
		dst = append(dst, ' ')
		dst = strconv.AppendFloat(dst, pt.Y, 'f', -1, 64)
		if i != len(s.pts)-1 {
			dst = append(dst, ',')
		}
	}
	return append(dst, ')')
}

// IsSimple returns true iff the curve defined by the LineString doesn't pass
// through the same point twice (with the possible exception of the two
// endpoints).
func (s LineString) IsSimple() bool {
	// 1. Build Lines
	// 2. Check for pairwise intersection.
	//  a. Point is allowed if lines adjacent.
	//  b. Start to end is allowed if first and last line.
	var lines []Line
	n := len(s.pts)
	for i := 0; i < n-1; i++ {
		if s.pts[i].XY == s.pts[i+1].XY {
			continue
		}
		ln, err := NewLine(s.pts[i], s.pts[i+1])
		if err != nil {
			panic(err)
		}
		lines = append(lines, ln)
	}
	n = len(lines)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			// TODO: it would be better to use Dimension() and Equals() here
			// once those are implemented.
			intersection := lines[i].Intersection(lines[j])
			switch intersection := intersection.(type) {
			case Point:
				if i+1 == j {
					// Adjacent lines will intersect at a point due to
					// construction, so this case is okay.
				} else if i == 0 && j == n-1 {
					// The first and last segment are allowed to intersect at a
					// point, so long as that point is the start of the first
					// segment and the end of the last segment (i.e. a linear
					// ring).
					if intersection.coords.XY != lines[i].a.XY || intersection.coords.XY != lines[j].b.XY {
						return false
					}
				} else {
					// Any other point intersection (e.g. looping back on
					// itself at a point) is disallowed for simple linestrings.
					return false
				}
			case Line:
				return false
			case GeometryCollection:
				if len(intersection.geoms) != 0 {
					return false
				}
			default:
				panic("unexpected intersection type")
			}
		}
	}
	return true
}

// parrallel tests if two line segments are parrallel with each other. The
// first line segment is [a,b] and the second is [c,d].
func parallel(a, b, c, d XY) bool {
	return (b.Y-a.Y)*(d.X-c.X) == (d.Y-c.Y)*(b.X-a.X)
}

// intersect checks if two line segments intersect. The first line segment is
// [a,b] and the second is [c,d].
func intersect(a, b, c, d XY) bool {
	p := ((c.Y-d.Y)*(a.X-c.X) + (d.X-c.X)*(a.Y-c.Y)) / ((d.X-c.X)*(a.Y-b.Y) - (a.X-b.X)*(d.Y-c.Y))
	q := ((a.Y-b.Y)*(a.X-c.X) + (b.X-a.X)*(a.Y-c.Y)) / ((d.X-c.X)*(a.Y-b.Y) - (a.X-b.X)*(d.Y-c.Y))
	return p >= 0 && p <= 1 && q >= 0 && q <= 1
}

func (s LineString) IsClosed() bool {
	if len(s.pts) == 0 {
		return false
	}
	return s.pts[0] == s.pts[len(s.pts)-1]
}

func (s LineString) Intersection(Geometry) Geometry {
	panic("not implemented")
}

func (s LineString) IsEmpty() bool {
	return len(s.pts) == 0
}
