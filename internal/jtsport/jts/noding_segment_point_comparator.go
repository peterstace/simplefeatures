package jts

// Noding_SegmentPointComparator implements a robust method of comparing the
// relative position of two points along the same segment. The coordinates are
// assumed to lie "near" the segment. This means that this algorithm will only
// return correct results if the input coordinates have the same precision and
// correspond to rounded values of exact coordinates lying on the segment.

// Noding_SegmentPointComparator_Compare compares two Coordinates for their
// relative position along a segment lying in the specified Octant.
//
// Returns -1 if p0 occurs first, 0 if the two nodes are equal, 1 if p1 occurs
// first.
func Noding_SegmentPointComparator_Compare(octant int, p0, p1 *Geom_Coordinate) int {
	// Nodes can only be equal if their coordinates are equal.
	if p0.Equals2D(p1) {
		return 0
	}

	xSign := noding_SegmentPointComparator_relativeSign(p0.X, p1.X)
	ySign := noding_SegmentPointComparator_relativeSign(p0.Y, p1.Y)

	switch octant {
	case 0:
		return noding_SegmentPointComparator_compareValue(xSign, ySign)
	case 1:
		return noding_SegmentPointComparator_compareValue(ySign, xSign)
	case 2:
		return noding_SegmentPointComparator_compareValue(ySign, -xSign)
	case 3:
		return noding_SegmentPointComparator_compareValue(-xSign, ySign)
	case 4:
		return noding_SegmentPointComparator_compareValue(-xSign, -ySign)
	case 5:
		return noding_SegmentPointComparator_compareValue(-ySign, -xSign)
	case 6:
		return noding_SegmentPointComparator_compareValue(-ySign, xSign)
	case 7:
		return noding_SegmentPointComparator_compareValue(xSign, -ySign)
	default:
		panic("invalid octant value")
	}
}

func noding_SegmentPointComparator_relativeSign(x0, x1 float64) int {
	if x0 < x1 {
		return -1
	}
	if x0 > x1 {
		return 1
	}
	return 0
}

func noding_SegmentPointComparator_compareValue(compareSign0, compareSign1 int) int {
	if compareSign0 < 0 {
		return -1
	}
	if compareSign0 > 0 {
		return 1
	}
	if compareSign1 < 0 {
		return -1
	}
	if compareSign1 > 0 {
		return 1
	}
	return 0
}
