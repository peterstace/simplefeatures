package jts

import "math"

// Algorithm_HCoordinate represents a homogeneous coordinate in a 2-D coordinate
// space. In JTS HCoordinates are used as a clean way of computing intersections
// between line segments.
type Algorithm_HCoordinate struct {
	X float64
	Y float64
	W float64
}

// Algorithm_HCoordinate_Intersection computes the (approximate) intersection
// point between two line segments using homogeneous coordinates.
//
// Note that this algorithm is not numerically stable; i.e. it can produce
// intersection points which lie outside the envelope of the line segments
// themselves. In order to increase the precision of the calculation input
// points should be normalized before passing them to this routine.
//
// Deprecated: use Algorithm_Intersection_Intersection instead.
func Algorithm_HCoordinate_Intersection(p1, p2, q1, q2 *Geom_Coordinate) (*Geom_Coordinate, error) {
	// Unrolled computation.
	px := p1.Y - p2.Y
	py := p2.X - p1.X
	pw := p1.X*p2.Y - p2.X*p1.Y

	qx := q1.Y - q2.Y
	qy := q2.X - q1.X
	qw := q1.X*q2.Y - q2.X*q1.Y

	x := py*qw - qy*pw
	y := qx*pw - px*qw
	w := px*qy - qx*py

	xInt := x / w
	yInt := y / w

	if math.IsNaN(xInt) || math.IsInf(xInt, 0) || math.IsNaN(yInt) || math.IsInf(yInt, 0) {
		return nil, Algorithm_NewNotRepresentableException()
	}

	return Geom_NewCoordinateWithXY(xInt, yInt), nil
}

// Algorithm_NewHCoordinate creates a new HCoordinate with default values.
func Algorithm_NewHCoordinate() *Algorithm_HCoordinate {
	return &Algorithm_HCoordinate{
		X: 0.0,
		Y: 0.0,
		W: 1.0,
	}
}

// Algorithm_NewHCoordinateWithXYW creates a new HCoordinate with the given
// x, y, and w values.
func Algorithm_NewHCoordinateWithXYW(x, y, w float64) *Algorithm_HCoordinate {
	return &Algorithm_HCoordinate{
		X: x,
		Y: y,
		W: w,
	}
}

// Algorithm_NewHCoordinateWithXY creates a new HCoordinate with the given
// x and y values (w defaults to 1.0).
func Algorithm_NewHCoordinateWithXY(x, y float64) *Algorithm_HCoordinate {
	return &Algorithm_HCoordinate{
		X: x,
		Y: y,
		W: 1.0,
	}
}

// Algorithm_NewHCoordinateFromCoordinate creates a new HCoordinate from a
// Coordinate.
func Algorithm_NewHCoordinateFromCoordinate(p *Geom_Coordinate) *Algorithm_HCoordinate {
	return &Algorithm_HCoordinate{
		X: p.X,
		Y: p.Y,
		W: 1.0,
	}
}

// Algorithm_NewHCoordinateFromHCoordinates creates a new HCoordinate which is
// the intersection of the lines defined by two HCoordinates.
func Algorithm_NewHCoordinateFromHCoordinates(p1, p2 *Algorithm_HCoordinate) *Algorithm_HCoordinate {
	return &Algorithm_HCoordinate{
		X: p1.Y*p2.W - p2.Y*p1.W,
		Y: p2.X*p1.W - p1.X*p2.W,
		W: p1.X*p2.Y - p2.X*p1.Y,
	}
}

// Algorithm_NewHCoordinateFromCoordinates constructs a homogeneous coordinate
// which is the intersection of the lines defined by the homogeneous coordinates
// represented by two Coordinates.
func Algorithm_NewHCoordinateFromCoordinates(p1, p2 *Geom_Coordinate) *Algorithm_HCoordinate {
	// Optimization when it is known that w = 1.
	return &Algorithm_HCoordinate{
		X: p1.Y - p2.Y,
		Y: p2.X - p1.X,
		W: p1.X*p2.Y - p2.X*p1.Y,
	}
}

// Algorithm_NewHCoordinateFrom4Coordinates creates a new HCoordinate which is
// the intersection point of two line segments defined by four Coordinates.
func Algorithm_NewHCoordinateFrom4Coordinates(p1, p2, q1, q2 *Geom_Coordinate) *Algorithm_HCoordinate {
	// Unrolled computation.
	px := p1.Y - p2.Y
	py := p2.X - p1.X
	pw := p1.X*p2.Y - p2.X*p1.Y

	qx := q1.Y - q2.Y
	qy := q2.X - q1.X
	qw := q1.X*q2.Y - q2.X*q1.Y

	return &Algorithm_HCoordinate{
		X: py*qw - qy*pw,
		Y: qx*pw - px*qw,
		W: px*qy - qx*py,
	}
}

// GetX returns the X ordinate of this HCoordinate.
func (h *Algorithm_HCoordinate) GetX() (float64, error) {
	a := h.X / h.W
	if math.IsNaN(a) || math.IsInf(a, 0) {
		return 0, Algorithm_NewNotRepresentableException()
	}
	return a, nil
}

// GetY returns the Y ordinate of this HCoordinate.
func (h *Algorithm_HCoordinate) GetY() (float64, error) {
	a := h.Y / h.W
	if math.IsNaN(a) || math.IsInf(a, 0) {
		return 0, Algorithm_NewNotRepresentableException()
	}
	return a, nil
}

// GetCoordinate returns a Coordinate for this HCoordinate.
func (h *Algorithm_HCoordinate) GetCoordinate() (*Geom_Coordinate, error) {
	x, err := h.GetX()
	if err != nil {
		return nil, err
	}
	y, err := h.GetY()
	if err != nil {
		return nil, err
	}
	return Geom_NewCoordinateWithXY(x, y), nil
}
