package simplefeatures

import (
	"errors"
)

// LinearRing is a LineString that is constainted to be closed (has the same
// start and end point) and simple (doesn't self intersect).
type LinearRing struct {
	ls LineString
}

var _ Geometry = LinearRing{}

func NewLinearRing(pts []Coordinates) (LinearRing, error) {
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

func (r LinearRing) AsText() []byte {
	return r.AppendWKT(nil)
}

func (r LinearRing) AppendWKT(dst []byte) []byte {
	return r.ls.AppendWKT(dst)
}

// IsSimple always returns true. Simplicity is one of the LinearRing constraints.
func (r LinearRing) IsSimple() bool {
	panic("not implemented")
}

func (r LinearRing) Intersection(Geometry) Geometry {
	panic("not implemented")
}
