package carto

import "github.com/peterstace/simplefeatures/geom"

type Sinusoidal struct {
	radius          float64
	centralMeridian float64
}

func NewSinusoidal(radius, centralMeridian float64) *Sinusoidal {
	return &Sinusoidal{radius, centralMeridian}
}

func (c *Sinusoidal) To(lonLat geom.XY) geom.XY {
	var (
		R   = c.radius
		λd  = lonLat.X
		λ0d = c.centralMeridian
		φd  = lonLat.Y
		λr  = dtor(λd)
		λ0r = dtor(λ0d)
		φr  = dtor(φd)
	)
	var (
		x = R * cos(φr) * (λr - λ0r)
		y = R * φr
	)
	return geom.XY{X: x, Y: y}
}

func (c *Sinusoidal) From(xy geom.XY) geom.XY {
	var (
		R   = c.radius
		λ0d = c.centralMeridian
		λ0r = dtor(λ0d)
		x   = xy.X
		y   = xy.Y
	)
	var (
		λr = x/(R*cos(y/R)) + λ0r
		φr = y / R
		λd = rtod(λr)
		φd = rtod(φr)
	)
	return geom.XY{X: λd, Y: φd}
}
