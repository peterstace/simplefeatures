package simplefeatures

import (
	"fmt"
	"sort"
)

func convexHullG(g Geometry) Geometry {
	if g.IsEmpty() {
		// special case to mirror postgis behaviour
		return g
	}
	pts := g.convexHullPointSet()
	hull := grahamScan(pts)
	switch len(hull) {
	case 0:
		return NewGeometryCollection(nil)
	case 1:
		return NewPoint(hull[0])
	case 2:
		ln, err := NewLine(
			Coordinates{hull[0]},
			Coordinates{hull[1]},
		)
		if err != nil {
			panic("bug in grahamScan routine - output 2 coincident points")
		}
		return ln
	default:
		coords := [][]Coordinates{make([]Coordinates, len(hull))}
		for i := range hull {
			coords[0][i] = Coordinates{XY: hull[i]}
		}
		poly, err := NewPolygonFromCoords(coords)
		if err != nil {
			panic(fmt.Errorf("bug in grahamScan routine - didn't produce a valid polygon: %v", err))
		}
		return poly
	}
}

type pointStack []XY

func (s *pointStack) push(p XY) {
	(*s) = append(*s, p)
}

func (s *pointStack) pop() XY {
	p := s.top()
	(*s) = (*s)[:len(*s)-1]
	return p
}

func (s *pointStack) top() XY {
	return (*s)[len(*s)-1]
}

func (s *pointStack) underTop() XY {
	return (*s)[len(*s)-2]
}

// grahamScan returns the convex hull of the input points. It will either
// represent the empty set (zero points), a point (one point), a line (2
// points), or a closed polygon (>= 3 points).
func grahamScan(pts []XY) []XY {
	if len(pts) <= 1 {
		return pts
	}

	sortByPolarAngle(pts)

	// Append the lowest-then-leftmost point so that the polygon will be closed.
	pts = append(pts, pts[0])

	var stack pointStack
	stack.push(pts[0])
	pts = pts[1:]
	for len(pts) > 0 && len(stack) < 2 {
		if !stack.top().Equals(pts[0]) {
			stack.push(pts[0])
		}
		pts = pts[1:]
	}

	for len(pts) > 0 {
		ori := orientation(stack.underTop(), stack.top(), pts[0])
		switch ori {
		case leftTurn:
			stack.push(pts[0])
		case collinear:
			if distanceSq(stack.underTop(), pts[0]).GT(distanceSq(stack.underTop(), stack.top())) {
				stack.pop()
				stack.push(pts[0])
			}
		default:
			stack.pop()
			if orientation(stack.underTop(), stack.top(), pts[0]) == collinear {
				stack.pop()
			}
			stack.push(pts[0])
		}
		pts = pts[1:]
	}
	return stack
}

// soryByPolarAngle sorts the points by their polar angle relative to the
// lowest-then-leftmost point.
func sortByPolarAngle(pts []XY) {
	// the lowest-then-leftmost (anchor) point comes first
	ltlp := ltl(pts)
	pts[ltlp], pts[0] = pts[0], pts[ltlp]
	anchor := pts[0]

	pts = pts[1:] // only sort the remaining points
	sort.Slice(pts, func(i, j int) bool {
		if anchor.Equals(pts[i]) {
			return true
		}
		if anchor.Equals(pts[j]) {
			return false
		}
		return orientation(anchor, pts[i], pts[j]) == leftTurn
	})
}

// ltl finds the index of the lowest-then-leftmost point.
func ltl(pts []XY) int {
	rpi := 0
	for i := 1; i < len(pts); i++ {
		if pts[i].Y.LT(pts[rpi].Y) || (pts[i].Y.Equals(pts[rpi].Y) && pts[i].X.LT(pts[rpi].X)) {
			rpi = i
		}
	}
	return rpi
}

type threePointOrientation int

const (
	// rightTurn indicates the orientation is right turn which is anticlockwise
	rightTurn threePointOrientation = iota + 1
	// collinear indicates three points are on the same line
	collinear
	// leftTurn indicates the orientation is left turn which is clockwise
	leftTurn
)

func (o threePointOrientation) String() string {
	switch o {
	case rightTurn:
		return "right turn"
	case collinear:
		return "collinear"
	case leftTurn:
		return "left turn"
	default:
		return "invalid orientation"
	}
}

// orientation checks if s is on the right hand side or left hand side of the line formed by p and q.
func orientation(p, q, s XY) threePointOrientation {
	cp := q.Sub(p).Cross(s.Sub(q))
	switch {
	case cp.GT(zero):
		return leftTurn
	case cp.LT(zero):
		return rightTurn
	default:
		return collinear
	}
}

// distanceSq gives the square of the distance between p and q.
func distanceSq(p, q XY) Scalar {
	pSubQ := p.Sub(q)
	return pSubQ.Dot(pSubQ)
}
