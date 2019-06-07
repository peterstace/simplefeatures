package simplefeatures

import "errors"

// LinearRing is a LineString that is both closed (has the same start and end
// point) and simple (doesn't self intersect).
type LinearRing struct {
	ls LineString
}

func NewLinearRing(pts []Point) (LinearRing, error) {
	ls, err := NewLineString(pts)
	if err != nil {
		return LinearRing{}, err
	}
	if !ls.IsClosed() {
		return LinearRing{}, errors.New("linear rings must be closed")
	}
	if !ls.IsSimple() {
		return LinearRing{}, errors.New("linear rings must be simple")
	}
	return LinearRing{ls}, nil
}
