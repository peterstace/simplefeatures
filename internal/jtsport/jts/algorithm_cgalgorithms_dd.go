package jts

import "math"

// Implements basic computational geometry algorithms using DD arithmetic.
type Algorithm_CGAlgorithmsDD struct{}

// A value which is safely greater than the relative round-off error in
// double-precision numbers.
const algorithm_CGAlgorithmsDD_dpSafeEpsilon = 1e-15

// Algorithm_CGAlgorithmsDD_OrientationIndex returns the index of the direction of the
// point q relative to a vector specified by p1-p2.
//
// Returns:
//
//	 1 if q is counter-clockwise (left) from p1-p2
//	-1 if q is clockwise (right) from p1-p2
//	 0 if q is collinear with p1-p2
func Algorithm_CGAlgorithmsDD_OrientationIndex(p1, p2, q *Geom_Coordinate) int {
	return Algorithm_CGAlgorithmsDD_OrientationIndexFloat64(p1.GetX(), p1.GetY(), p2.GetX(), p2.GetY(), q.GetX(), q.GetY())
}

// Algorithm_CGAlgorithmsDD_OrientationIndexFloat64 returns the index of the direction of
// the point q relative to a vector specified by p1-p2.
//
// Returns:
//
//	 1 if q is counter-clockwise (left) from p1-p2
//	-1 if q is clockwise (right) from p1-p2
//	 0 if q is collinear with p1-p2
func Algorithm_CGAlgorithmsDD_OrientationIndexFloat64(p1x, p1y, p2x, p2y, qx, qy float64) int {
	index := algorithm_CGAlgorithmsDD_orientationIndexFilter(p1x, p1y, p2x, p2y, qx, qy)
	if index <= 1 {
		return index
	}

	dx1 := Math_DD_ValueOfFloat64(p2x).SelfAddFloat64(-p1x)
	dy1 := Math_DD_ValueOfFloat64(p2y).SelfAddFloat64(-p1y)
	dx2 := Math_DD_ValueOfFloat64(qx).SelfAddFloat64(-p2x)
	dy2 := Math_DD_ValueOfFloat64(qy).SelfAddFloat64(-p2y)

	return dx1.SelfMultiply(dy2).SelfSubtract(dy1.SelfMultiply(dx2)).Signum()
}

// Algorithm_CGAlgorithmsDD_SignOfDet2x2 computes the sign of the determinant of the 2x2
// matrix with the given DD entries.
//
// Returns:
//
//	-1 if the determinant is negative,
//	 1 if the determinant is positive,
//	 0 if the determinant is 0.
func Algorithm_CGAlgorithmsDD_SignOfDet2x2(x1, y1, x2, y2 *Math_DD) int {
	det := x1.Multiply(y2).SelfSubtract(y1.Multiply(x2))
	return det.Signum()
}

// Algorithm_CGAlgorithmsDD_SignOfDet2x2Float64 computes the sign of the determinant of
// the 2x2 matrix with the given float64 entries.
//
// Returns:
//
//	-1 if the determinant is negative,
//	 1 if the determinant is positive,
//	 0 if the determinant is 0.
func Algorithm_CGAlgorithmsDD_SignOfDet2x2Float64(dx1, dy1, dx2, dy2 float64) int {
	x1 := Math_DD_ValueOfFloat64(dx1)
	y1 := Math_DD_ValueOfFloat64(dy1)
	x2 := Math_DD_ValueOfFloat64(dx2)
	y2 := Math_DD_ValueOfFloat64(dy2)

	det := x1.Multiply(y2).SelfSubtract(y1.Multiply(x2))
	return det.Signum()
}

// algorithm_CGAlgorithmsDD_orientationIndexFilter is a filter for computing the
// orientation index of three coordinates.
//
// If the orientation can be computed safely using standard DP arithmetic, this
// routine returns the orientation index. Otherwise, a value i > 1 is returned.
// In this case the orientation index must be computed using some other more
// robust method. The filter is fast to compute, so can be used to avoid the use
// of slower robust methods except when they are really needed, thus providing
// better average performance.
//
// Uses an approach due to Jonathan Shewchuk, which is in the public domain.
//
// Returns:
//
//	the orientation index if it can be computed safely
//	i > 1 if the orientation index cannot be computed safely
func algorithm_CGAlgorithmsDD_orientationIndexFilter(pax, pay, pbx, pby, pcx, pcy float64) int {
	var detsum float64

	detleft := (pax - pcx) * (pby - pcy)
	detright := (pay - pcy) * (pbx - pcx)
	det := detleft - detright

	if detleft > 0.0 {
		if detright <= 0.0 {
			return algorithm_CGAlgorithmsDD_signum(det)
		}
		detsum = detleft + detright
	} else if detleft < 0.0 {
		if detright >= 0.0 {
			return algorithm_CGAlgorithmsDD_signum(det)
		}
		detsum = -detleft - detright
	} else {
		return algorithm_CGAlgorithmsDD_signum(det)
	}

	errbound := algorithm_CGAlgorithmsDD_dpSafeEpsilon * detsum
	if (det >= errbound) || (-det >= errbound) {
		return algorithm_CGAlgorithmsDD_signum(det)
	}

	return 2
}

func algorithm_CGAlgorithmsDD_signum(x float64) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

// Algorithm_CGAlgorithmsDD_Intersection computes an intersection point between two lines
// using DD arithmetic. If the lines are parallel (either identical or separate)
// a nil value is returned.
func Algorithm_CGAlgorithmsDD_Intersection(p1, p2, q1, q2 *Geom_Coordinate) *Geom_Coordinate {
	px := Math_NewDDFromFloat64(p1.GetY()).SelfSubtract(Math_NewDDFromFloat64(p2.GetY()))
	py := Math_NewDDFromFloat64(p2.GetX()).SelfSubtract(Math_NewDDFromFloat64(p1.GetX()))
	pw := Math_NewDDFromFloat64(p1.GetX()).SelfMultiply(Math_NewDDFromFloat64(p2.GetY())).SelfSubtract(Math_NewDDFromFloat64(p2.GetX()).SelfMultiply(Math_NewDDFromFloat64(p1.GetY())))

	qx := Math_NewDDFromFloat64(q1.GetY()).SelfSubtract(Math_NewDDFromFloat64(q2.GetY()))
	qy := Math_NewDDFromFloat64(q2.GetX()).SelfSubtract(Math_NewDDFromFloat64(q1.GetX()))
	qw := Math_NewDDFromFloat64(q1.GetX()).SelfMultiply(Math_NewDDFromFloat64(q2.GetY())).SelfSubtract(Math_NewDDFromFloat64(q2.GetX()).SelfMultiply(Math_NewDDFromFloat64(q1.GetY())))

	x := py.Multiply(qw).SelfSubtract(qy.Multiply(pw))
	y := qx.Multiply(pw).SelfSubtract(px.Multiply(qw))
	w := px.Multiply(qy).SelfSubtract(qx.Multiply(py))

	xInt := x.SelfDivide(w).DoubleValue()
	yInt := y.SelfDivide(w).DoubleValue()

	if math.IsNaN(xInt) || math.IsInf(xInt, 0) || math.IsNaN(yInt) || math.IsInf(yInt, 0) {
		return nil
	}

	return Geom_NewCoordinateWithXY(xInt, yInt)
}
