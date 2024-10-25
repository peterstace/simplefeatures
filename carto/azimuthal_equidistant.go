package carto

import (
	"github.com/peterstace/simplefeatures/geom"
)

type AzimuthalEquidistant struct {
	Radius       float64
	OriginLonLat geom.XY
}

func (a *AzimuthalEquidistant) To(lonLat geom.XY) geom.XY {
	R := a.Radius
	λd := lonLat.X
	φd := lonLat.Y
	λr := dtor(λd)
	φr := dtor(φd)
	λ0r := dtor(a.OriginLonLat.X)
	φ0r := dtor(a.OriginLonLat.Y)

	// From https://en.wikipedia.org/wiki/Azimuthal_equidistant_projection,
	// with some slight algebraic rearrangement.
	ρ := R * acos(sin(φ0r)*sin(φr)+cos(φ0r)*cos(φr)*cos(λr-λ0r))
	θ := atan2(
		cos(φr)*sin(λr-λ0r),
		cos(φ0r)*sin(φr)-sin(φ0r)*cos(φr)*cos(λr-λ0r),
	)
	x := ρ * sin(θ)
	y := ρ * cos(θ)
	return xy(x, y)
}

func (a *AzimuthalEquidistant) From(xy geom.XY) geom.XY {
	R := a.Radius
	x := xy.X
	y := xy.Y
	λ0r := dtor(a.OriginLonLat.X)
	φ0r := dtor(a.OriginLonLat.Y)

	// From https://en.wikipedia.org/wiki/Azimuthal_equidistant_projection,
	// with some slight algebraic rearrangement.
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
