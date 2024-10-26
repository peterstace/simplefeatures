package carto

import "github.com/peterstace/simplefeatures/geom"

type LambertCylindricalEqualArea struct {
	radius float64
	λ0     float64
}

func NewLambertCylindricalEqualArea(radius float64) *LambertCylindricalEqualArea {
	return &LambertCylindricalEqualArea{
		radius: radius,
		λ0:     0,
	}
}

func (c *LambertCylindricalEqualArea) SetCentralMeridian(lon float64) {
	c.λ0 = dtor(lon)
}

func (c *LambertCylindricalEqualArea) To(lonLat geom.XY) geom.XY {
	var (
		R  = c.radius
		λ  = dtor(lonLat.X)
		λ0 = c.λ0
		φ  = dtor(lonLat.Y)
	)
	return geom.XY{
		X: R * (λ - λ0),
		Y: R * sin(φ),
	}
}

func (c *LambertCylindricalEqualArea) From(xy geom.XY) geom.XY {
	var (
		R  = c.radius
		x  = xy.X
		y  = xy.Y
		λ0 = c.λ0
	)
	var (
		λ = x/R + λ0
		φ = asin(y / R)
	)
	return rtodxy(λ, φ)
}
