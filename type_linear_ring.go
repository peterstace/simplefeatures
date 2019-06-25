package simplefeatures

import (
	"errors"
)

// LinearRing is a LineString that is constainted to be closed and simple.
//
// Its assertions are:
//
// 1. It must be a valid LineString.
//
// 2. It must be closed, i.e. the start and end points must be identical.
//
// 3. It must be simple, i.e. it must not self intersect (except for the start
// and end points, which must intersect).
//
type LinearRing struct {
	ls LineString
}

// NewLinearRing builds a LinearRing from a sequence of coordinates.
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

func (r LinearRing) Intersection(g Geometry) Geometry {
	return intersection(r, g)
}

// IsEmpty always returns false. LinearRings cannot be empty due to their
// assertions, in particular that LinearRings must be closed.
func (r LinearRing) IsEmpty() bool {
	return false
}

func (r LinearRing) Dimension() int {
	return 1
}

func (r LinearRing) Equals(other Geometry) bool {
	return equals(r, other)
}

func (r LinearRing) Envelope() (Envelope, bool) {
	env := NewEnvelope(r.ls.lines[0].a.XY)
	for _, line := range r.ls.lines[1:] {
		env = env.Extend(line.a.XY)
	}
	return env, true
}

func (r LinearRing) Boundary() Geometry {
	panic("not implemented")
}
