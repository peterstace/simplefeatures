package geom

import (
	"fmt"
	"math"

	"github.com/peterstace/simplefeatures/rtree"
)

// Envelope is an axis-aligned rectangle (also known as an Axis Aligned
// Bounding Box or Minimum Bounding Rectangle). It usually represents a 2D area
// with non-zero width and height, but can also represent degenerate cases
// where the width or height (or both) are zero. Its bounds are validated so as
// to not be NaN or +/- Infinity.
type Envelope struct {
	min XY
	max XY
}

// NewEnvelope returns the smallest envelope that contains all provided points.
// It returns an error if any of the XYs contains NaN or +/- Infinity
// coordinates.
func NewEnvelope(first XY, others ...XY) (Envelope, error) {
	if err := first.validate(); err != nil {
		return Envelope{}, err
	}
	env := Envelope{
		min: first,
		max: first,
	}
	for _, xy := range others {
		var err error
		env, err = env.ExtendToIncludeXY(xy)
		if err != nil {
			return Envelope{}, err
		}
	}
	return env, nil
}

// EnvelopeFromGeoms returns the smallest envelope that contains all points
// contained by the provided geometries, provided that at least one non-empty
// geometry is given. If no non-empty geometries are given, then the returned
// flag is set to false.
func EnvelopeFromGeoms(geoms ...Geometry) (Envelope, bool) {
	envs := make([]Envelope, 0, len(geoms))
	for _, g := range geoms {
		env, ok := g.Envelope()
		if ok {
			envs = append(envs, env)
		}
	}
	if len(envs) == 0 {
		return Envelope{}, false
	}
	env := envs[0]
	for _, e := range envs[1:] {
		env = env.ExpandToIncludeEnvelope(e)
	}
	return env, true
}

// AsGeometry returns the envelope as a Geometry. In the regular case where the
// envelope covers some area, then a Polygon geometry is returned. In
// degenerate cases where the envelope only covers a line or a point, a
// LineString or Point geometry is returned.
func (e Envelope) AsGeometry() Geometry {
	if e.min == e.max {
		return e.min.asUncheckedPoint().AsGeometry()
	}

	if e.min.X == e.max.X || e.min.Y == e.max.Y {
		ln := line{e.min, e.max}
		return ln.asLineString().AsGeometry()
	}

	floats := [...]float64{
		e.min.X, e.min.Y,
		e.min.X, e.max.Y,
		e.max.X, e.max.Y,
		e.max.X, e.min.Y,
		e.min.X, e.min.Y,
	}
	seq := NewSequence(floats[:], DimXY)
	ls, err := NewLineString(seq)
	if err != nil {
		panic(fmt.Sprintf("constructing geometry from envelope: %v", err))
	}
	poly, err := NewPolygon([]LineString{ls})
	if err != nil {
		panic(fmt.Sprintf("constructing geometry from envelope: %v", err))
	}
	return poly.AsGeometry()
}

// Min returns the point in the envelope with the minimum X and Y values.
func (e Envelope) Min() XY {
	return e.min
}

// Max returns the point in the envelope with the maximum X and Y values.
func (e Envelope) Max() XY {
	return e.max
}

// ExtendToIncludeXY returns the smallest envelope that contains all of the
// points in this envelope along with the provided point. It gives an error if
// the XY contains NaN or +/- Infinite coordinates.
func (e Envelope) ExtendToIncludeXY(xy XY) (Envelope, error) {
	if err := xy.validate(); err != nil {
		return Envelope{}, err
	}
	return e.uncheckedExtend(xy), nil
}

// uncheckedExtend extends the envelope in the same manner as
// ExtendToIncludeXY but doesn't validate the XY. It should only be used
// when the XY doesn't come directly from user input.
func (e Envelope) uncheckedExtend(xy XY) Envelope {
	return Envelope{
		min: XY{fastMin(e.min.X, xy.X), fastMin(e.min.Y, xy.Y)},
		max: XY{fastMax(e.max.X, xy.X), fastMax(e.max.Y, xy.Y)},
	}
}

// ExpandToIncludeEnvelope returns the smallest envelope that contains all of
// the points in this envelope and another envelope.
func (e Envelope) ExpandToIncludeEnvelope(other Envelope) Envelope {
	return Envelope{
		min: XY{fastMin(e.min.X, other.min.X), fastMin(e.min.Y, other.min.Y)},
		max: XY{fastMax(e.max.X, other.max.X), fastMax(e.max.Y, other.max.Y)},
	}
}

// Contains returns true iff this envelope contains the given point.
func (e Envelope) Contains(p XY) bool {
	return p.validate() == nil &&
		p.X >= e.min.X && p.X <= e.max.X &&
		p.Y >= e.min.Y && p.Y <= e.max.Y
}

// Intersects returns true iff this envelope has any points in common with
// another envelope.
func (e Envelope) Intersects(o Envelope) bool {
	return true &&
		(e.min.X <= o.max.X) && (e.max.X >= o.min.X) &&
		(e.min.Y <= o.max.Y) && (e.max.Y >= o.min.Y)
}

// Center returns the center point of the envelope.
func (e Envelope) Center() XY {
	return e.min.Add(e.max).Scale(0.5)
}

// Covers returns true iff and only if this envelope entirely covers another
// envelope (i.e. every point in the other envelope is contained within this
// envelope).
func (e Envelope) Covers(o Envelope) bool {
	return true &&
		e.min.X <= o.min.X && e.min.Y <= o.min.Y &&
		e.max.X >= o.max.X && e.max.Y >= o.max.Y
}

// Width returns the difference between the maximum and minimum X coordinates
// of the envelope.
func (e Envelope) Width() float64 {
	return e.max.X - e.min.X
}

// Height returns the difference between the maximum and minimum X coordinates
// of the envelope.
func (e Envelope) Height() float64 {
	return e.max.Y - e.min.Y
}

// Area returns the area covered by the envelope.
func (e Envelope) Area() float64 {
	return (e.max.X - e.min.X) * (e.max.Y - e.min.Y)
}

// Distance calculates the shortest distance between this envelope and another
// envelope. If the envelopes intersect with each other, then the returned
// distance is 0.
func (e Envelope) Distance(o Envelope) float64 {
	dx := fastMax(0, fastMax(o.min.X-e.max.X, e.min.X-o.max.X))
	dy := fastMax(0, fastMax(o.min.Y-e.max.Y, e.min.Y-o.max.Y))
	return math.Sqrt(dx*dx + dy*dy)
}

func (e Envelope) box() rtree.Box {
	return rtree.Box{
		MinX: e.min.X,
		MinY: e.min.Y,
		MaxX: e.max.X,
		MaxY: e.max.Y,
	}
}
