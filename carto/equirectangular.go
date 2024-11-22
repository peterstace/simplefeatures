package carto

import "github.com/peterstace/simplefeatures/geom"

// Equirectangular allows projecting (longitude, latitude) coordinates to (x, y)
// pairs via the equirectangular projection.
type Equirectangular struct {
	λ0     float64
	cosφ1  float64
	radius float64
}

// NewEquirectangular returns a new Equirectangular projection with the given
// earth radius.
func NewEquirectangular(earthRadius float64) *Equirectangular {
	return &Equirectangular{
		λ0:     0,
		cosφ1:  1, // φ1 = 0 (equator)
		radius: earthRadius,
	}
}

// SetCentralMeridian sets the central meridian of the projection to the given
// longitude expressed in degrees.
func (e *Equirectangular) SetCentralMeridian(lon float64) {
	e.λ0 = dtor(lon)
}

// SetStandardParallels sets the standard parallels of the projection to the
// given latitudes expressed in degrees.
func (e *Equirectangular) SetStandardParallels(lat float64) {
	φ1 := dtor(lat)
	e.cosφ1 = cos(φ1)
}

// Forward converts a (longitude, latitude) pair expressed in degrees to a
// projected (x, y) pair.
func (e *Equirectangular) Forward(lonLat geom.XY) geom.XY {
	var (
		R     = e.radius
		λ     = dtor(lonLat.X)
		φ     = dtor(lonLat.Y)
		λ0    = e.λ0
		cosφ1 = e.cosφ1
	)
	return geom.XY{
		X: R * (λ - λ0) * cosφ1,
		Y: R * φ,
	}
}

// Reverse converts a projected (x, y) pair to a (longitude, latitude) pair
// expressed in degrees.
func (e *Equirectangular) Reverse(xy geom.XY) geom.XY {
	var (
		R     = e.radius
		x     = xy.X
		y     = xy.Y
		λ0    = e.λ0
		cosφ1 = e.cosφ1
	)
	var (
		λ = x/(R*cosφ1) + λ0
		φ = y / R
	)
	return rtodxy(λ, φ)
}
