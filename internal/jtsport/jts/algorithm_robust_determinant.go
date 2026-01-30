package jts

import "math"

// Implements an algorithm to compute the sign of a 2x2 determinant for double
// precision values robustly. It is a direct translation of code developed by
// Olivier Devillers.

// Algorithm_RobustDeterminant_SignOfDet2x2 computes the sign of the determinant
// of the 2x2 matrix with the given entries, in a robust way.
//
// Returns:
//
//	-1 if the determinant is negative
//	 1 if the determinant is positive
//	 0 if the determinant is 0
func Algorithm_RobustDeterminant_SignOfDet2x2(x1, y1, x2, y2 float64) int {
	sign := 1
	var swap float64
	var k float64

	// Testing null entries.
	if x1 == 0.0 || y2 == 0.0 {
		if y1 == 0.0 || x2 == 0.0 {
			return 0
		} else if y1 > 0 {
			if x2 > 0 {
				return -sign
			}
			return sign
		} else {
			if x2 > 0 {
				return sign
			}
			return -sign
		}
	}
	if y1 == 0.0 || x2 == 0.0 {
		if y2 > 0 {
			if x1 > 0 {
				return sign
			}
			return -sign
		} else {
			if x1 > 0 {
				return -sign
			}
			return sign
		}
	}

	// Making y coordinates positive and permuting the entries so that y2 is
	// the biggest one.
	if 0.0 < y1 {
		if 0.0 < y2 {
			if y1 > y2 {
				sign = -sign
				swap = x1
				x1 = x2
				x2 = swap
				swap = y1
				y1 = y2
				y2 = swap
			}
		} else {
			if y1 <= -y2 {
				sign = -sign
				x2 = -x2
				y2 = -y2
			} else {
				swap = x1
				x1 = -x2
				x2 = swap
				swap = y1
				y1 = -y2
				y2 = swap
			}
		}
	} else {
		if 0.0 < y2 {
			if -y1 <= y2 {
				sign = -sign
				x1 = -x1
				y1 = -y1
			} else {
				swap = -x1
				x1 = x2
				x2 = swap
				swap = -y1
				y1 = y2
				y2 = swap
			}
		} else {
			if y1 >= y2 {
				x1 = -x1
				y1 = -y1
				x2 = -x2
				y2 = -y2
			} else {
				sign = -sign
				swap = -x1
				x1 = -x2
				x2 = swap
				swap = -y1
				y1 = -y2
				y2 = swap
			}
		}
	}

	// Making x coordinates positive. If |x2| < |x1| one can conclude.
	if 0.0 < x1 {
		if 0.0 < x2 {
			if x1 > x2 {
				return sign
			}
		} else {
			return sign
		}
	} else {
		if 0.0 < x2 {
			return -sign
		} else {
			if x1 >= x2 {
				sign = -sign
				x1 = -x1
				x2 = -x2
			} else {
				return -sign
			}
		}
	}

	// All entries strictly positive x1 <= x2 and y1 <= y2.
	for {
		k = math.Floor(x2 / x1)
		x2 = x2 - k*x1
		y2 = y2 - k*y1

		// Testing if R (new U2) is in U1 rectangle.
		if y2 < 0.0 {
			return -sign
		}
		if y2 > y1 {
			return sign
		}

		// Finding R'.
		if x1 > x2+x2 {
			if y1 < y2+y2 {
				return sign
			}
		} else {
			if y1 > y2+y2 {
				return -sign
			} else {
				x2 = x1 - x2
				y2 = y1 - y2
				sign = -sign
			}
		}
		if y2 == 0.0 {
			if x2 == 0.0 {
				return 0
			}
			return -sign
		}
		if x2 == 0.0 {
			return sign
		}

		// Exchange 1 and 2 role.
		k = math.Floor(x1 / x2)
		x1 = x1 - k*x2
		y1 = y1 - k*y2

		// Testing if R (new U1) is in U2 rectangle.
		if y1 < 0.0 {
			return sign
		}
		if y1 > y2 {
			return -sign
		}

		// Finding R'.
		if x2 > x1+x1 {
			if y2 < y1+y1 {
				return -sign
			}
		} else {
			if y2 > y1+y1 {
				return sign
			} else {
				x1 = x2 - x1
				y1 = y2 - y1
				sign = -sign
			}
		}
		if y1 == 0.0 {
			if x1 == 0.0 {
				return 0
			}
			return sign
		}
		if x1 == 0.0 {
			return -sign
		}
	}
}

// Algorithm_RobustDeterminant_OrientationIndex returns the index of the
// direction of the point q relative to a vector specified by p1-p2.
//
// Returns:
//
//	 1 if q is counter-clockwise (left) from p1-p2
//	-1 if q is clockwise (right) from p1-p2
//	 0 if q is collinear with p1-p2
func Algorithm_RobustDeterminant_OrientationIndex(p1, p2, q *Geom_Coordinate) int {
	dx1 := p2.GetX() - p1.GetX()
	dy1 := p2.GetY() - p1.GetY()
	dx2 := q.GetX() - p2.GetX()
	dy2 := q.GetY() - p2.GetY()
	return Algorithm_RobustDeterminant_SignOfDet2x2(dx1, dy1, dx2, dy2)
}
