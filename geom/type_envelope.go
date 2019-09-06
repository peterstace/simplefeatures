package geom

import (
	"fmt"
	"math"
)

type Envelope struct {
	min XY
	max XY
}

func NewEnvelope(first XY, others ...XY) Envelope {
	env := Envelope{
		min: first,
		max: first,
	}
	for _, pt := range others {
		env = env.Extend(pt)
	}
	return env
}

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
		env = env.Union(e)
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

func (e Envelope) Min() XY {
	return e.min
}

func (e Envelope) Max() XY {
	return e.max
}

func (e Envelope) Extend(point XY) Envelope {
	return Envelope{
		min: XY{math.Min(e.min.X, point.X), math.Min(e.min.Y, point.Y)},
		max: XY{math.Max(e.max.X, point.X), math.Max(e.max.Y, point.Y)},
	}
}

func (e Envelope) Union(other Envelope) Envelope {
	return Envelope{
		min: XY{math.Min(e.min.X, other.min.X), math.Min(e.min.Y, other.min.Y)},
		max: XY{math.Max(e.max.X, other.max.X), math.Max(e.max.Y, other.max.Y)},
	}
}

func (e Envelope) IntersectsPoint(p XY) bool {
	return p.X >= e.min.X && p.X <= e.max.X && p.Y >= e.min.Y && p.Y <= e.max.Y
}

func (e Envelope) Intersects(o Envelope) bool {
	return (e.min.X <= o.max.X) && (e.max.X >= o.min.X) &&
		(e.min.Y <= o.max.Y) && (e.max.Y >= o.min.Y)
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
