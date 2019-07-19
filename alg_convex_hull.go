package simplefeatures

import (
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
		coords := make([][]Coordinates, 1)
		coords[0] = make([]Coordinates, len(hull))
		for i := range hull {
			coords[0][i] = Coordinates{XY: hull[i]}
		}
		poly, err := NewPolygonFromCoords(coords)
		if err != nil {
			panic("bug in grahamScan routine - didn't produce a valid polygon")
		}
		return poly
	}
}

// grahamScan returns the convex hull of the input points.
func grahamScan(ps []XY) []XY {
	if len(ps) < 3 {
		return nil
	}

	sortByPolarAngle(ps)

	resultStack := make([]XY, 2, len(ps))
	copy(resultStack, ps[:2])
	toDoStack := make([]XY, len(ps)-2)
	copy(toDoStack, ps[2:])

	for i := 0; i < len(toDoStack); i++ {
		ori := orientation(resultStack[len(resultStack)-2], resultStack[len(resultStack)-1], toDoStack[i])
		switch {
		case ori == leftTurn:
			resultStack = append(resultStack, toDoStack[i])
		default:
			resultStack = resultStack[:len(resultStack)-1]
			resultStack = append(resultStack, toDoStack[i])
		}
	}

	return append(resultStack, resultStack[0])
}

// soryByPolarAngle sorts the points by their polar angle
func sortByPolarAngle(ps []XY) {
	ltlp := ltl(ps)

	// swap the ltl point with first point
	ps[ltlp], ps[0] = ps[0], ps[ltlp]

	virtualPoint := ps[0]

	sort.Slice(ps, func(i, j int) bool {
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

// ltl stands for lowest-then-leftmost points. It returns the index of lowest-then-leftmost point
func ltl(ps []XY) int {
	rpi := 0

	for i := 1; i < len(ps); i++ {
		if ps[i].Y.AsFloat() < ps[rpi].Y.AsFloat() ||
			(ps[i].Y.AsFloat() == ps[rpi].Y.AsFloat() &&
				ps[i].X.AsFloat() < ps[rpi].X.AsFloat()) {
			rpi = i
		}
	}

	return rpi
}

const (
	// rightTurn indicates the orientation is right turn which is anticlockwise
	rightTurn = iota
	// collinear indicates three points are on the same line
	collinear
	// leftTurn indicates the orientation is left turn which is clockwise
	leftTurn
)

// orientation checks if s is on the right hand side or left hand side of the line formed by p and q
// if it returns -1 which means there is an unexpected result.
func orientation(p, q, s XY) int {
	cp := crossProduct(p, q, s)
	switch {
	case cp.GT(zero):
		return leftTurn
	case cp.Equals(zero):
		return collinear
	case cp.LT(zero):
		return rightTurn
	default:
		return -1
	}
}

// crossProduct implements Heron's formula which returns the 2 times of area formed by p, q and s
//         | p.X p.Y 1 |
// 2 * S = | q.X q.Y 1 |
//         | s.X s.Y 1 |
// when p, q and s are clockwise, the return value is negative
// when p, q and s are anticlockwise, the return value is positive
func crossProduct(p, q, s XY) Scalar {
	return q.Sub(p).Cross(s.Sub(q))
}

// distance give the length of p an q
func distanceSq(p, q XY) Scalar {
	return p.Sub(q).Dot(p.Sub(q))
}
