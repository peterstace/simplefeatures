package geom

import (
	"database/sql/driver"
	"errors"
	"io"
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

// NewLinearRingC builds a LinearRing from a sequence of coordinates.
func NewLinearRingC(pts []Coordinates, opts ...ConstructorOption) (LinearRing, error) {
	ls, err := NewLineStringC(pts, opts...)
	if err != nil {
		return LinearRing{}, err
	}
	if doCheapValidations(opts) && !ls.IsClosed() {
		return LinearRing{}, errors.New("linear rings must be closed")
	}
	if doExpensiveValidations(opts) && !ls.IsSimple() {
		return LinearRing{}, errors.New("linear rings must be simple")
	}
	return LinearRing{ls}, nil
}

// StartPoint gives the first point of the linear ring.
func (r LinearRing) StartPoint() Point {
	return r.ls.StartPoint()
}

// EndPoint gives the last point of the linear ring. Because linear rings are
// closed, this is by definition the same as the start point.
func (r LinearRing) EndPoint() Point {
	return r.ls.EndPoint()
}

// NumPoints gives the number of control points in the linear ring.
func (r LinearRing) NumPoints() int {
	return r.ls.NumPoints()
}

// PointN gives the nth (zero indexed) point in the line string. Panics if n is
// out of range with respect to the number of points.
func (r LinearRing) PointN(n int) Point {
	return r.ls.PointN(n)
}

func (r LinearRing) AsText() string {
	return string(r.AppendWKT(nil))
}

func (r LinearRing) AppendWKT(dst []byte) []byte {
	return r.ls.AppendWKT(dst)
}

// IsSimple always returns true. Simplicity is one of the LinearRing constraints.
func (r LinearRing) IsSimple() bool {
	return true
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
	// Same behaviour as Postgis, but could be any empty set.
	return NewMultiPoint(nil)
}

func (r LinearRing) Value() (driver.Value, error) {
	return wkbAsBytes(r)
}

func (r LinearRing) AsBinary(w io.Writer) error {
	return r.ls.AsBinary(w)
}

// ConvexHull returns the convex hull of the LinearRing, which is always a
// Polygon.
func (r LinearRing) ConvexHull() Geometry {
	return convexHull(r)
}

func (r LinearRing) convexHullPointSet() []XY {
	return r.ls.convexHullPointSet()
}

func (r LinearRing) MarshalJSON() ([]byte, error) {
	return r.ls.MarshalJSON()
}

// LinearRing returns the coordinates of the points making up the LinearRings.
func (r LinearRing) Coordinates() []Coordinates {
	return r.ls.Coordinates()
}

// TransformXY transforms this LinearRing into another LinearRing according to fn.
func (r LinearRing) TransformXY(fn func(XY) XY, opts ...ConstructorOption) (Geometry, error) {
	coords := r.Coordinates()
	transform1dCoords(coords, fn)
	return NewLinearRingC(coords, opts...)
}

// EqualsExact checks if this LinearRing is exactly equal to another curve.
func (r LinearRing) EqualsExact(other Geometry, opts ...EqualsExactOption) bool {
	c, ok := other.(curve)
	return ok && curvesExactEqual(r, c, opts)
}

// Valid checks if this LinearRing is valid
func (r LinearRing) Valid() bool {
	_, err := NewLinearRingC(r.Coordinates())
	return err == nil
}
