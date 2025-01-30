package carto

import "github.com/peterstace/simplefeatures/geom"

// AlbersEqualAreaConic allows projecting (longitude, latitude) coordinates to
// (x, y) pairs via the Albers equal area conic projection.
//
// The Albers equal area conic projection is a conic projection that is:
//   - Configured by setting two standard parallels.
//   - Equal area.
//   - Not conformal, but preserves shape locally at the standard parallels.
type AlbersEqualAreaConic struct {
	radius       float64
	origin       geom.XY
	stdParallels [2]float64
}

// NewAlbersEqualAreaConic returns a new AlbersEqualAreaConic projection with
// the given earth radius. The standard parallels are set to 30 and 60 degrees
// north.
func NewAlbersEqualAreaConic(earthRadius float64) *AlbersEqualAreaConic {
	return &AlbersEqualAreaConic{
		radius:       earthRadius,
		stdParallels: [2]float64{30, 60},
	}
}

// SetStandardParallels sets the standard parallels of the projection to the
// given latitudes expressed in degrees.
func (c *AlbersEqualAreaConic) SetStandardParallels(lat1, lat2 float64) {
	c.stdParallels[0] = lat1
	c.stdParallels[1] = lat2
}

// SetOrigin sets the origin of the projection to the given (longitude,
// latitude) pair. The origin has projected coordinates (0, 0).
func (c *AlbersEqualAreaConic) SetOrigin(origin geom.XY) {
	c.origin = origin
}

// Forward converts a (longitude, latitude) pair expressed in degrees to a
// projected (x, y) pair.
func (c *AlbersEqualAreaConic) Forward(lonlat geom.XY) geom.XY {
	var (
		R  = c.radius
		φ  = dtor(lonlat.Y)
		φ0 = dtor(c.origin.Y)
		φ1 = dtor(c.stdParallels[0])
		φ2 = dtor(c.stdParallels[1])
		λ  = dtor(lonlat.X)
		λ0 = dtor(c.origin.X)
	)
	var (
		n  = (sin(φ1) + sin(φ2)) / 2
		θ  = n * (λ - λ0)
		C  = sq(cos(φ1)) + 2*n*sin(φ1)
		ρ  = R * sqrt(C-2*n*sin(φ)) / n
		ρ0 = R * sqrt(C-2*n*sin(φ0)) / n
	)
	var (
		x = ρ * sin(θ)
		y = ρ0 - ρ*cos(θ)
	)
	return geom.XY{X: x, Y: y}
}

// Reverse converts a projected (x, y) pair to a (longitude, latitude) pair
// expressed in degrees.
func (c *AlbersEqualAreaConic) Reverse(xy geom.XY) geom.XY {
	var (
		R  = c.radius
		x  = xy.X
		y  = xy.Y
		φ0 = dtor(c.origin.Y)
		φ1 = dtor(c.stdParallels[0])
		φ2 = dtor(c.stdParallels[1])
		λ0 = dtor(c.origin.X)
	)
	var (
		n  = (sin(φ1) + sin(φ2)) / 2
		C  = sq(cos(φ1)) + 2*n*sin(φ1)
		ρ0 = R * sqrt(C-2*n*sin(φ0)) / n
		ρ  = R * sqrt(sq(x)+sq(ρ0-y))
		θ  = atan(x / (ρ0 - y))
	)
	var (
		φ = asin((C - ρ*ρ*n*n) / (2 * n))
		λ = λ0 + θ/n
	)
	return geom.XY{X: rtod(λ), Y: rtod(φ)}
}
