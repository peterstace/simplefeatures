package carto

import (
	"github.com/peterstace/simplefeatures/geom"
)

// AzimuthalEquidistant allows projecting (longitude, latitude) coordinates to
// (x, y) pairs via the azimuthal equidistant projection.
//
// The azimuthal equidistant projection is a projection that is:
//   - Configured by setting a center point.
//   - Equidistant. Distances from the center point are correctly scaled.
//   - Azimuthal. Directions from the center point are correctly preserved.
//   - Not conformal, but preserves shape locally at the center point.
//   - Not equal area, but preserves area locally at the center point.
type AzimuthalEquidistant struct {
	radius       float64
	centerLonLat geom.XY
}

// NewAzimuthalEquidistant returns a new [AzimuthalEquidistant] projection with
// the given earth radius.
func NewAzimuthalEquidistant(earthRadius float64) *AzimuthalEquidistant {
	return &AzimuthalEquidistant{
		radius:       earthRadius,
		centerLonLat: geom.XY{},
	}
}

// SetCenterLonLat sets the center of the projection to the given (longitude,
// latitude) pair. The center have projected coordinates (0, 0) and be the
// center of the circular map.
func (a *AzimuthalEquidistant) SetCenter(centerLonLat geom.XY) {
	a.centerLonLat = centerLonLat
}

// Forward converts a (longitude, latitude) pair expressed in degrees to a
// projected (x, y) pair.
func (a *AzimuthalEquidistant) Forward(lonLat geom.XY) geom.XY {
	R := a.radius
	λd := lonLat.X
	φd := lonLat.Y
	λr := dtor(λd)
	φr := dtor(φd)
	λ0r := dtor(a.centerLonLat.X)
	φ0r := dtor(a.centerLonLat.Y)

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

// Reverse converts a projected (x, y) pair to a (longitude, latitude) pair
// expressed in degrees.
func (a *AzimuthalEquidistant) Reverse(xy geom.XY) geom.XY {
	R := a.radius
	x := xy.X
	y := xy.Y
	λ0r := dtor(a.centerLonLat.X)
	φ0r := dtor(a.centerLonLat.Y)

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
