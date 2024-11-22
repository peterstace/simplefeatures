package carto

import (
	"github.com/peterstace/simplefeatures/geom"
)

type AzimuthalEquidistant struct {
	radius       float64
	originLonLat geom.XY
}

func NewAzimuthalEquidistant(earthRadius float64) *AzimuthalEquidistant {
	return &AzimuthalEquidistant{
		radius:       earthRadius,
		originLonLat: geom.XY{},
	}
}

func (a *AzimuthalEquidistant) SetOrigin(origin geom.XY) {
	a.originLonLat = origin
}

func (a *AzimuthalEquidistant) To(lonLat geom.XY) geom.XY {
	R := a.radius
	λd := lonLat.X
	φd := lonLat.Y
	λr := dtor(λd)
	φr := dtor(φd)
	λ0r := dtor(a.originLonLat.X)
	φ0r := dtor(a.originLonLat.Y)

	ρ := R * acos(sin(φ0r)*sin(φr)+cos(φ0r)*cos(φr)*cos(λr-λ0r))
	θ := atan2(
		cos(φr)*sin(λr-λ0r),
		cos(φ0r)*sin(φr)-sin(φ0r)*cos(φr)*cos(λr-λ0r),
	)
	return geom.XY{
		X: ρ * sin(θ),
		Y: ρ * cos(θ),
	}
}

func (a *AzimuthalEquidistant) From(xy geom.XY) geom.XY {
	R := a.radius
	x := xy.X
	y := xy.Y
	λ0r := dtor(a.originLonLat.X)
	φ0r := dtor(a.originLonLat.Y)

	ρ := sqrt(x*x + y*y)
	φr := asin(cos(ρ/R)*sin(φ0r) + (y*sin(ρ/R)*cos(φ0r))/ρ)
	λr := λ0r + atan2(
		x*sin(ρ/R),
		ρ*cos(φ0r)*cos(ρ/R)-y*sin(φ0r)*sin(ρ/R),
	)
	λd := rtod(λr)
	φd := rtod(φr)
	return geom.XY{X: λd, Y: φd}
}
