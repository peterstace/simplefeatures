package geom

import (
	"fmt"
	"sort"
)

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

func convexHull(g Geometry) Geometry {
	if g.IsEmpty() {
		// special case to mirror postgis behaviour
		return g
	}
	pts := g.convexHullPointSet()
	// TODO: length may not be a good signal to indicate
	// how to convert XY to Geometry. Need to revisit
	switch len(pts) {
	case 0:
		return NewGeometryCollection(nil)
	case 1:
		return NewPointXY(pts[0])
	case 2:
		if pts[0].Equals(pts[1]) {
			return NewPointXY(pts[0])
		}
		ln, err := NewLineC(
			Coordinates{pts[0]},
			Coordinates{pts[1]},
		)
		if err != nil {
			panic(err)
		}
		return ln
	}

	// TODO: check if the points are all the same
	isSame := true
	for i := 1; i < len(pts); i++ {
		if !pts[0].Equals(pts[i]) {
			isSame = false
			break
		}
	}
	if isSame {
		return NewPointXY(pts[0])
	}

	cl, ok := collinearLine(pts)
	if ok {
		return cl
	}

	hull := grahamScan(pts)
	coords := [][]Coordinates{make([]Coordinates, len(hull))}
	for i := range hull {
		coords[0][i] = Coordinates{XY: hull[i]}
	}
	poly, err := NewPolygonC(coords)
	if err != nil {
		panic(fmt.Errorf("bug in grahamScan routine - didn't produce a valid polygon: %v", err))
	}
	return poly
}

// grahamScan returns the convex hull of the input points. It will either
// represent the empty set (zero points), a point (one point), a line (2
// points), or a closed polygon (>= 3 points).
func grahamScan(ps []XY) []XY {

	sortByPolarAngle(ps)

	// Append the lowest-then-leftmost point so that the polygon will be closed.
	resultStack := make([]XY, 0, len(ps))
	resultStack = append(resultStack, ps[0], ps[0])
	toDoStack := make([]XY, len(ps)-1)
	copy(toDoStack, ps[1:])

	for len(toDoStack) > 0 {
		toBeCompared := toDoStack[0]
		ori := orientation(resultStack[len(resultStack)-2], resultStack[len(resultStack)-1], toBeCompared)
		switch ori {
		case leftTurn:
			resultStack = append(resultStack, toBeCompared)
			toDoStack = toDoStack[1:]
		case collinear:
			if distanceSq(resultStack[len(resultStack)-2], resultStack[len(resultStack)-1]).LT(distanceSq(resultStack[len(resultStack)-2], toBeCompared)) {
				resultStack = resultStack[:len(resultStack)-1]
				resultStack = append(resultStack, toBeCompared)
			}
			toDoStack = toDoStack[1:]
		default:
			resultStack = resultStack[:len(resultStack)-1]
		}
	}

	resultStack = append(resultStack, resultStack[0])

	if resultStack[0] == resultStack[1] {
		resultStack = resultStack[1:]
	}

	return resultStack
}

// sortByPolarAngle sorts the points by their polar angle relative to the
// lowest-then-leftmost anchor point.
func sortByPolarAngle(ps []XY) {
	// the lowest-then-leftmost (anchor) point comes first
	ltlp := ltl(ps)
	ps[ltlp], ps[0] = ps[0], ps[ltlp]
	virtualPoint := ps[0]

	ps = ps[1:] // only sort the remaining points
	sort.SliceStable(ps, func(i, j int) bool {
		if i == 0 {
			return false
		}

		ori := orientation(virtualPoint, ps[i], ps[j])

		if ori == collinear {
			return distanceSq(virtualPoint, ps[i]).LT(distanceSq(virtualPoint, ps[j]))
		}

		return ori == leftTurn
	})
}

// ltl finds the index of the lowest-then-leftmost point.
func ltl(ps []XY) int {
	rpi := 0
	for i := 1; i < len(ps); i++ {
		if ps[i].Y.LT(ps[rpi].Y) ||
			(ps[i].Y.Equals(ps[rpi].Y) &&
				ps[i].X.LT(ps[rpi].X)) {
			rpi = i
		}
	}
	return rpi
}

// distanceSq gives the square of the distance between p and q.
func distanceSq(p, q XY) Scalar {
	pSubQ := p.Sub(q)
	return pSubQ.Dot(pSubQ)
}

func collinearLine(pts []XY) (Geometry, bool) {
	if len(pts) < 2 {
		return nil, false
	}
	ps := make([]XY, len(pts))
	copy(ps, pts)

	startPoint := ps[ltl(ps)]
	// check collinear
	sort.Slice(ps, func(i, j int) bool {
		return distanceSq(startPoint, ps[i]).LT(distanceSq(startPoint, ps[j]))
	})

	if ps[0].Equals(ps[len(ps)-1]) {
		return nil, false
	}

	for i := 1; i < len(ps); i++ {
		if orientation(ps[0], ps[len(ps)-1], ps[i]) != collinear {
			return nil, false
		}
	}

	// already check that if the initial point is as same as the end point.
	// Ignore error here.
	cl, _ := NewLineC(
		Coordinates{ps[0]},
		Coordinates{ps[len(ps)-1]},
	)

	return cl, true
}
