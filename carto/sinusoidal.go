package carto

import "github.com/peterstace/simplefeatures/geom"

type Sinusoidal struct {
	radius float64
	λ0     float64
}

func NewSinusoidal(radius float64) *Sinusoidal {
	return &Sinusoidal{
		radius: radius,
		λ0:     0,
	}
}

func (c *Sinusoidal) SetCentralMeridian(lon float64) {
	c.λ0 = dtor(lon)
}

func (c *Sinusoidal) Forward(lonLat geom.XY) geom.XY {
	var (
		R  = c.radius
		λ0 = c.λ0
		λ  = dtor(lonLat.X)
		φ  = dtor(lonLat.Y)
	)
	return geom.XY{
		X: R * cos(φ) * (λ - λ0),
		Y: R * φ,
	}
}

func (c *Sinusoidal) Reverse(xy geom.XY) geom.XY {
	var (
		R  = c.radius
		λ0 = c.λ0
		x  = xy.X
		y  = xy.Y
	)
	var (
		λ = x/(R*cos(y/R)) + λ0
		φ = y / R
	)
	return rtodxy(λ, φ)
}
