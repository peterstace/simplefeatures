package carto

import "github.com/peterstace/simplefeatures/geom"

// Sinusoidal allows projecting (longitude, latitude) coordinates to (x, y)
// pairs via the sinusoidal projection.
//
// The sinusoidal projection is a pseudocylindrical projection that is:
//   - Configured by setting the central meridian.
//   - Equal area.
//   - Not conformal, but preserves shape locally along the central meridian
//     and equator.
//   - Not equidistant, but preserves distance locally along all parallels and
//     the central meridian.
type Sinusoidal struct {
	radius float64
	λ0     float64
}

// NewSinusoidal returns a new [Sinusoidal] projection with the given earth
// radius.
func NewSinusoidal(earthRadius float64) *Sinusoidal {
	return &Sinusoidal{
		radius: earthRadius,
		λ0:     0,
	}
}

// SetCentralMeridian sets the central meridian of the projection to the given
// longitude expressed in degrees.
func (c *Sinusoidal) SetCentralMeridian(lon float64) {
	c.λ0 = dtor(lon)
}

// Forward converts a (longitude, latitude) pair expressed in degrees to a
// projected (x, y) pair.
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

// Reverse converts a projected (x, y) pair to a (longitude, latitude) pair
// expressed in degrees.
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
