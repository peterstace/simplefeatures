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
			intersection := lines[i].Intersection(lines[j])
			// TODO: need equals
		}
	}

	/*
		// TODO: I'm not happy with this algorithm. It doesn't feel very elegant. A
		// better approach may be to implement it in terms of the Line type.
		n := len(s.pts)
		for i := 0; i < n-2; i++ {
			if s.pts[i].XY == s.pts[i+1].XY || s.pts[i+1].XY == s.pts[i+2].XY {
				continue
			}
			if parallel(
				s.pts[i].XY,
				s.pts[i+1].XY,
				s.pts[i+1].XY,
				s.pts[i+2].XY,
			) {
				return false
			}
		}
		for i := 0; i < n-3; i++ {
			for j := i + 2; j < n-1; j++ {
				if i == 0 && j == n-2 && s.pts[0].XY == s.pts[n-1].XY {
					continue
				}
				if intersect(
					s.pts[i].XY,
					s.pts[i+1].XY,
					s.pts[j].XY,
					s.pts[j+1].XY,
				) {
					return false
				}
			}
		}
		return true
	*/
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
