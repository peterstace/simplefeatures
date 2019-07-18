package simplefeatures

import "sort"

const (
	// Right indicates the orientation is right turn which is anticlockwise
	Right = iota
	// Collinear indicates three points are on the same line
	Collinear
	// Left indicates the orientation is left turn which is clockwise
	Left
)

// graphamScan returns a convex hull.
func grahamScan(ps []Point) []Point {
	if len(ps) < 3 {
		return nil
	}
	// TODO: handle the case that three points are collinear
	sortedPoints := sortByPolarAngle(ps)

	s := make([]Point, 2, len(sortedPoints))
	copy(s, sortedPoints[:2])
	t := make([]Point, len(sortedPoints)-2)
	copy(t, sortedPoints[2:])

	for i := 0; i < len(t); i++ {
		if orientation(s[len(s)-2], s[len(s)-1], t[i]) == Left {
			s = append(s, t[i])
		} else {
			s = s[:len(s)-1]
			s = append(s, t[i])
		}
	}

	return append(s, s[0])
}

func sortByPolarAngle(ps []Point) []Point {
	ltlp := ltl(ps)

	// swap the ltl point with first point
	ps[ltlp], ps[0] = ps[0], ps[ltlp]

	virtualPoint := ps[0]

	sort.Slice(ps, func(i, j int) bool {
		if i == 0 {
			return false
		}
		return orientation(virtualPoint, ps[i], ps[j]) == Left
	})

	return ps
}

// ltl stands for lowest-then-leftmost points. It returns the index of lowest-then-leftmost point
func ltl(ps []Point) int {
	rpi := 0

	for i := 1; i < len(ps); i++ {
		if ps[i].XY().Y.AsFloat() < ps[rpi].XY().Y.AsFloat() ||
			(ps[i].XY().Y.AsFloat() == ps[rpi].XY().Y.AsFloat() &&
				ps[i].XY().X.AsFloat() < ps[rpi].XY().X.AsFloat()) {
			rpi = i
		}
	}

	return rpi
}

// orientation checks if s is on the right hand side or left hand side of the line formed by p and q
// if it returns -1 which means there is an unexpected result.
func orientation(p, q, s Point) int {
	cp := crossProduct(p, q, s)
	switch {
	case cp > 0:
		return Left
	case cp == 0:
		return Collinear
	case cp < 0:
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
func crossProduct(p, q, s Point) float64 {
	return p.XY().X.AsFloat()*q.XY().Y.AsFloat() - p.XY().Y.AsFloat()*q.XY().X.AsFloat() +
		q.XY().X.AsFloat()*s.XY().Y.AsFloat() - q.XY().Y.AsFloat()*s.XY().X.AsFloat() +
		s.XY().X.AsFloat()*p.XY().Y.AsFloat() - s.XY().Y.AsFloat()*p.XY().X.AsFloat()
}
