package geom

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
	case cp > 0:
		return leftTurn
	case cp < 0:
		return rightTurn
	default:
		return collinear
	}
}
