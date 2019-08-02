package geom

import (
	"fmt"
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

func (e Envelope) Min() XY {
	return e.min
}

func (e Envelope) Max() XY {
	return e.max
}

func (e Envelope) Extend(point XY) Envelope {
	return Envelope{
		min: XY{e.min.X.Min(point.X), e.min.Y.Min(point.Y)},
		max: XY{e.max.X.Max(point.X), e.max.Y.Max(point.Y)},
	}
}

func (e Envelope) Union(other Envelope) Envelope {
	return Envelope{
		min: XY{e.min.X.Min(other.min.X), e.min.Y.Min(other.min.Y)},
		max: XY{e.max.X.Max(other.max.X), e.max.Y.Max(other.max.Y)},
	}
}

func (e Envelope) IntersectsPoint(p XY) bool {
	return p.X.GTE(e.min.X) && p.X.LTE(e.max.X) && p.Y.GTE(e.min.Y) && p.Y.LTE(e.max.Y)
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
