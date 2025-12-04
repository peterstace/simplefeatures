package carto

import "github.com/peterstace/simplefeatures/geom"

// LambertCylindricalEqualArea allows projecting (longitude, latitude)
// coordinates to (x, y) pairs via the Lambert cylindrical equal area
// projection.
//
// The Lambert cylindrical equal area projection is a cylindrical projection
// that is:
//   - Configured by setting the central meridian.
//   - Equal area.
//   - Not conformal, but preserves shape locally along the equator.
//   - Not equidistant, but preserves distance along the equator.
type LambertCylindricalEqualArea struct {
	radius float64
	λ0     float64
}

// NewLambertCylindricalEqualArea returns a new [LambertCylindricalEqualArea]
// projection with the given earth radius.
func NewLambertCylindricalEqualArea(radius float64) *LambertCylindricalEqualArea {
	return &LambertCylindricalEqualArea{
		radius: radius,
		λ0:     0,
	}
}

// SetCentralMeridian sets the central meridian of the projection to the given
// longitude expressed in degrees.
func (c *LambertCylindricalEqualArea) SetCentralMeridian(lon float64) {
	c.λ0 = dtor(lon)
}

// Forward converts a (longitude, latitude) pair expressed in degrees to a
// projected (x, y) pair.
func (c *LambertCylindricalEqualArea) Forward(lonLat geom.XY) geom.XY {
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

// Reverse converts a projected (x, y) pair to a (longitude, latitude) pair
// expressed in degrees.
func (c *LambertCylindricalEqualArea) Reverse(xy geom.XY) geom.XY {
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
