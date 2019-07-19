package simplefeatures

import (
	"sort"
)

const (
	// Right indicates the orientation is right turn which is anticlockwise
	Right = iota
	// Collinear indicates three points are on the same line
	Collinear
	// Left indicates the orientation is left turn which is clockwise
	Left
)

// graphamScan returns a convex hull.
func grahamScan(ps []XY) []XY {
	if len(ps) < 3 {
		return nil
	}
	sortedPoints := sortByPolarAngle(ps)

	s := make([]XY, 2, len(sortedPoints))
	copy(s, sortedPoints[:2])
	t := make([]XY, len(sortedPoints)-2)
	copy(t, sortedPoints[2:])

	for i := 0; i < len(t); i++ {
		ori := orientation(s[len(s)-2], s[len(s)-1], t[i])
		switch {
		case ori == Left:
			s = append(s, t[i])
		default:
			s = s[:len(s)-1]
			s = append(s, t[i])
		}
	}

	return append(s, s[0])
}

// soryByPolarAngle sorts the points by their polar angle
func sortByPolarAngle(ps []XY) []XY {
	ltlp := ltl(ps)

	// swap the ltl point with first point
	ps[ltlp], ps[0] = ps[0], ps[ltlp]

	virtualPoint := ps[0]

	sort.Slice(ps, func(i, j int) bool {
		if i == 0 {
			return false
		}
		ori := orientation(virtualPoint, ps[i], ps[j])

		if ori == Collinear {
			return distanceSq(virtualPoint, ps[i]).LT(distanceSq(virtualPoint, ps[j]))
		}

		return ori == Left
	})

	return ps
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

// orientation checks if s is on the right hand side or left hand side of the line formed by p and q
// if it returns -1 which means there is an unexpected result.
func orientation(p, q, s XY) int {
	cp := crossProduct(p, q, s)
	switch {
	case cp.GT(NewScalarFromFloat64(0)):
		return Left
	case cp.Equals(NewScalarFromFloat64(0)):
		return Collinear
	case cp.LT(NewScalarFromFloat64(0)):
		return Right
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
	// return p.X.Mul(q.Y).Sub(p.Y.Mul(q.X)).Add(q.X.Mul(s.Y)).Sub(q.Y.Mul(s.X)).Mul(s.X.Mul(p.Y)).Sub(s.Y.Mul(p.X))
	// return p.X.AsFloat()*q.Y.AsFloat() - p.Y.AsFloat()*q.X.AsFloat() +
	// 	q.X.AsFloat()*s.Y.AsFloat() - q.Y.AsFloat()*s.X.AsFloat() +
	// 	s.X.AsFloat()*p.Y.AsFloat() - s.Y.AsFloat()*p.X.AsFloat()
}

// distance give the length of p an q
func distanceSq(p, q XY) Scalar {
	return p.Sub(q).Dot(p.Sub(q))
	// return math.Sqrt(
	// 	(p.X.AsFloat()-q.X.AsFloat())*(p.X.AsFloat()-q.X.AsFloat()) +
	// 		(p.Y.AsFloat()-q.Y.AsFloat())*(p.Y.AsFloat()-q.Y.AsFloat()),
	// )
}
