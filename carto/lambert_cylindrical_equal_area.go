package carto

import "github.com/peterstace/simplefeatures/geom"

type LambertCylindricalEqualArea struct {
	radius     float64
	centralLon float64
}

func NewLambertCylindricalEqualArea(radius, centralLon float64) *LambertCylindricalEqualArea {
	return &LambertCylindricalEqualArea{radius, centralLon}
}

func (c *LambertCylindricalEqualArea) To(lonLat geom.XY) geom.XY {
	var (
		R   = c.radius
		λd  = lonLat.X
		λ0d = c.centralLon
		φd  = lonLat.Y
		λr  = dtor(λd)
		λ0r = dtor(λ0d)
		φr  = dtor(φd)
	)
	var (
		x = R * (λr - λ0r)
		y = R * sin(φr)
	)
	return geom.XY{X: x, Y: y}
}

func (c *LambertCylindricalEqualArea) From(xy geom.XY) geom.XY {
	var (
		R   = c.radius
		λ0d = c.centralLon
		λ0r = dtor(λ0d)
		x   = xy.X
		y   = xy.Y
	)
	var (
		λr = x/R + λ0r
		φr = asin(y / R)
		λd = rtod(λr)
		φd = rtod(φr)
	)
	return geom.XY{X: λd, Y: φd}
}
