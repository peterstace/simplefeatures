package carto

import "github.com/peterstace/simplefeatures/geom"

type LambertConformalConic struct {
	radius       float64
	origin       geom.XY
	stdParallels [2]float64
}

func NewLambertConformalConic(earthRadius float64) *LambertConformalConic {
	return &LambertConformalConic{
		radius:       earthRadius,
		origin:       geom.XY{0, 0},
		stdParallels: [2]float64{0, 0},
	}
}

func (c *LambertConformalConic) SetOrigin(origin geom.XY) {
	c.origin = origin
}

func (c *LambertConformalConic) SetStandardParallels(lat1, lat2 float64) {
	c.stdParallels[0] = lat1
	c.stdParallels[1] = lat2
}

func (c *LambertConformalConic) To(lonlat geom.XY) geom.XY {
	var (
		R  = c.radius
		φ  = dtor(lonlat.Y)
		λ  = dtor(lonlat.X)
		φ0 = dtor(c.origin.Y)
		λ0 = dtor(c.origin.X)
		φ1 = dtor(c.stdParallels[0])
		φ2 = dtor(c.stdParallels[1])
	)
	var (
		n  = ln(cos(φ1)*sec(φ2)) / ln(tan(π/4+φ2/2)*cot(π/4+φ1/2))
		F  = cos(φ1) * pow(tan(π/4+φ1/2), n) / n
		ρ  = R * F / pow(cot(π/4+φ/2), n)
		ρ0 = R * F / pow(cot(π/4+φ0/2), n)
	)
	var (
		x = ρ * sin(n*(λ-λ0))
		y = ρ0 - ρ*cos(n*(λ-λ0))
	)
	return xy(x, y)
}
