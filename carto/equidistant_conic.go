package carto

import (
	"github.com/peterstace/simplefeatures/geom"
)

type EquidistantConic struct {
	earthRadius  float64
	stdParallels [2]float64
	origin       geom.XY
}

func NewEquidistantConic(earthRadius float64) *EquidistantConic {
	return &EquidistantConic{
		earthRadius:  earthRadius,
		stdParallels: [2]float64{0, 45},
	}
}

func (c *EquidistantConic) SetStandardParallels(lat1, lat2 float64) *EquidistantConic {
	// TODO: check that lat1 and lat2 are not equal.
	// TODO: check that lat1 and lat2 are in a reasonable range.
	// TODO: swap lat1 and lat2 if lat1 > lat2 (?).
	c.stdParallels[0] = lat1
	c.stdParallels[1] = lat2
	return c
}

func (c *EquidistantConic) SetOrigin(lonLat geom.XY) *EquidistantConic {
	c.origin = lonLat
	return c
}

func (c *EquidistantConic) To(lonlat geom.XY) geom.XY {
	var (
		R = c.earthRadius

		φd  = lonlat.Y
		φ0d = c.origin.Y
		φ1d = c.stdParallels[0]
		φ2d = c.stdParallels[1]

		φr  = dtor(φd)
		φ0r = dtor(φ0d)
		φ1r = dtor(φ1d)
		φ2r = dtor(φ2d)

		λd  = lonlat.X
		λ0d = c.origin.X

		λr  = dtor(λd)
		λ0r = dtor(λ0d)
	)
	var (
		n  = (cos(φ1r) - cos(φ2r)) / (φ2r - φ1r)
		G  = cos(φ1r)/n + φ1r
		ρ0 = G - φ0r

		ρ = G - φr
		x = ρ * sin(n*(λr-λ0r))
		y = ρ0 - ρ*cos(n*(λr-λ0r))
	)
	return geom.XY{X: R * x, Y: R * y}
}

func (c *EquidistantConic) From(xy geom.XY) geom.XY {
	var (
		R = c.earthRadius

		x = xy.X / R
		y = xy.Y / R

		φ0d = c.origin.Y
		φ1d = c.stdParallels[0]
		φ2d = c.stdParallels[1]

		λ0d = c.origin.X
		λ0r = dtor(λ0d)

		φ0r = dtor(φ0d)
		φ1r = dtor(φ1d)
		φ2r = dtor(φ2d)
	)
	var (
		n  = (cos(φ1r) - cos(φ2r)) / (φ2r - φ1r)
		G  = cos(φ1r)/n + φ1r
		ρ0 = G - φ0r

		ρ = sign(n) * sqrt(x*x+(ρ0-y)*(ρ0-y))

		θ = atan(x / (ρ0 - y))

		φr = G - ρ
		λr = λ0r + θ/n
	)
	return geom.XY{X: rtod(λr), Y: rtod(φr)}
}
