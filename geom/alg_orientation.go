package geom

// threePointOrientation describes the relationship between 3 distinct points
// in the plane. The relationships as described in words (i.e. 'left', 'right',
// 'collinear') assume that the X axis increases from left to right, and the Y
// axis increases from bottom to top.
type threePointOrientation int

const (
	// rightTurn indicates that the last point is to the right of the line
	// formed by traversing from the first point to the second point.
	//
	//   A---B
	//        \
	//         C
	rightTurn threePointOrientation = iota + 1

	// collinear indicates that the three points are on the same line.
	//
	//   A---B---C
	collinear

	// rightTurn indicates that the last point is to the left of the line
	// formed by traversing from the first point to the second point.
	//
	//         C
	//        /
	//   A---B
	leftTurn
)

// orientation calculates the 3 point orientation between p, q, and s.
func orientation(p, q, s XY) threePointOrientation {
	cp := q.Sub(p).Cross(s.Sub(q))
	switch {
	case cp > 0:
		return leftTurn
	case cp < 0:
		return rightTurn
	default:
		return collinear
	}
}
