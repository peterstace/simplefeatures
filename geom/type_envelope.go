package geom

import (
	"fmt"
	"math"
)

// Envelope is an axis-aligned rectangle (also known as an Axis Aligned
// Bounding Box or Minimum Bounding Rectangle). It usually represents a 2D area
// with non-zero width and height, but can also represent degenerate cases
// where the width or height (or both) are zero.
type Envelope struct {
	min XY
	max XY
}

// NewEnvelope returns the smallest envelope that contains all provided points.
func NewEnvelope(first XY, others ...XY) Envelope {
	env := Envelope{
		min: first,
		max: first,
	}
	for _, pt := range others {
		env = env.ExtendToIncludePoint(pt)
	}
	return env
}

// mustEnvelope gets the envelope from a Geometry. If it's not defined (because
// the geometry is empty), then it panics.
func mustEnvelope(g Geometry) Envelope {
	env, ok := g.Envelope()
	if !ok {
		panic(fmt.Sprintf("mustEnvelope but envelope not defined: %s", string(g.AsText())))
	}
	return env
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
// Line or Point geometry is returned.
func (e Envelope) AsGeometry() Geometry {
	if e.min.Equals(e.max) {
		return NewPointXY(e.min)
	}
	var err error
	var g Geometry
	if e.min.X == e.max.X || e.min.Y == e.max.Y {
		g, err = NewLineC(Coordinates{XY: e.min}, Coordinates{XY: e.max})
	} else {
		g, err = NewPolygonC([][]Coordinates{{
			{XY: XY{e.min.X, e.min.Y}},
			{XY: XY{e.min.X, e.max.Y}},
			{XY: XY{e.max.X, e.max.Y}},
			{XY: XY{e.max.X, e.min.Y}},
			{XY: XY{e.min.X, e.min.Y}},
		}})
	}
	if err != nil {
		panic(fmt.Sprintf("constructing geometry from envelope: %v", err))
	}
	return g
}

// Min returns the point in the envelope with the minimum X and Y values.
func (e Envelope) Min() XY {
	return e.min
}

// Max returns the point in the envelope with the maximum X and Y values.
func (e Envelope) Max() XY {
	return e.max
}

// ExtendToIncludePoint returns the smallest envelope that contains all of the
// points in this envelope along with the provided point.
func (e Envelope) ExtendToIncludePoint(point XY) Envelope {
	return Envelope{
		min: XY{math.Min(e.min.X, point.X), math.Min(e.min.Y, point.Y)},
		max: XY{math.Max(e.max.X, point.X), math.Max(e.max.Y, point.Y)},
	}
}

// ExpandToIncludeEnvelope returns the smallest envelope that contains all of
// the points in this envelope and another envelope.
func (e Envelope) ExpandToIncludeEnvelope(other Envelope) Envelope {
	return Envelope{
		min: XY{math.Min(e.min.X, other.min.X), math.Min(e.min.Y, other.min.Y)},
		max: XY{math.Max(e.max.X, other.max.X), math.Max(e.max.Y, other.max.Y)},
	}
}

// Contains returns true iff this envelope contains the given point.
func (e Envelope) Contains(p XY) bool {
	return true &&
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

// Width returns the difference between the maximum and minimum Y coordinates
// of the envelope.
func (e Envelope) Height() float64 {
	return e.max.Y - e.min.Y
}

// Area returns the area covered by the envelope.
func (e Envelope) Area() float64 {
	return (e.max.X - e.min.X) * (e.max.Y - e.min.Y)
}

// ExpandBy calculates an expanded version of this envelope. The expansion
// amount can be controlled independently in the x and y dimensions (controlled
// by the x and y parameters). Positive values increase the size of the
// envelope and negative amounts decrease the size of the envelope.
func (e Envelope) ExpandBy(x, y float64) (Envelope, bool) {
	delta := XY{x, y}
	env := Envelope{
		min: e.min.Sub(delta),
		max: e.max.Add(delta),
	}
	if env.min.X > env.max.X || env.min.Y > env.max.Y {
		return Envelope{}, false
	}
	return env, true
}

// Distance calculates the stortest distance between this envelope and another
// envelope. If the envelopes intersect with each other, then the returned
// distance is 0.
func (e Envelope) Distance(o Envelope) float64 {
	dx := math.Max(0, math.Max(o.min.X-e.max.X, e.min.X-o.max.X))
	dy := math.Max(0, math.Max(o.min.Y-e.max.Y, e.min.Y-o.max.Y))
	return math.Sqrt(dx*dx + dy*dy)
}
