package carto

import "github.com/peterstace/simplefeatures/geom"

// LambertConformalConic allows projecting (longitude, latitude) coordinates to
// (x, y) pairs via the Lambert conformal conic projection.
//
// The Lambert conformal conic projection is a conic projection that is:
//   - Configured by setting two standard parallels.
//   - Conformal. Shape is preserved locally at all points.
//   - Not equal area, but preserves area locally at the standard parallels.
//   - Not equidistant, but preserves distance locally along the standard
//     parallels.
type LambertConformalConic struct {
	radius       float64
	origin       geom.XY
	stdParallels [2]float64
}

// NewLambertConformalConic returns a new [LambertConformalConic] projection with
// the given earth radius.
func NewLambertConformalConic(earthRadius float64) *LambertConformalConic {
	return &LambertConformalConic{
		radius:       earthRadius,
		origin:       geom.XY{X: 0, Y: 0},
		stdParallels: [2]float64{0, 0},
	}
}

// SetOrigin sets the origin of the projection to the given (longitude,
// latitude) pair. The origin have projected coordinates (0, 0).
func (c *LambertConformalConic) SetOrigin(origin geom.XY) {
	c.origin = origin
}

// SetStandardParallels sets the standard parallels of the projection to the
// given latitudes expressed in degrees.
func (c *LambertConformalConic) SetStandardParallels(lat1, lat2 float64) {
	c.stdParallels[0] = lat1
	c.stdParallels[1] = lat2
}

// Forward converts a (longitude, latitude) pair expressed in degrees to a
// projected (x, y) pair.
func (c *LambertConformalConic) Forward(lonlat geom.XY) geom.XY {
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
		ρ  = R * F * pow(cot(π/4+φ/2), n)
		ρ0 = R * F * pow(cot(π/4+φ0/2), n)
	)
	return geom.XY{
		X: ρ * sin(n*(λ-λ0)),
		Y: ρ0 - ρ*cos(n*(λ-λ0)),
	}
}

// Reverse converts a projected (x, y) pair to a (longitude, latitude) pair
// expressed in degrees.
func (c *LambertConformalConic) Reverse(xy geom.XY) geom.XY {
	var (
		R  = c.radius
		x  = xy.X
		y  = xy.Y
		φ0 = dtor(c.origin.Y)
		λ0 = dtor(c.origin.X)
		φ1 = dtor(c.stdParallels[0])
		φ2 = dtor(c.stdParallels[1])
	)
	var (
		n  = ln(cos(φ1)*sec(φ2)) / ln(tan(π/4+φ2/2)*cot(π/4+φ1/2))
		F  = cos(φ1) * pow(tan(π/4+φ1/2), n) / n
		ρ0 = R * F * pow(cot(π/4+φ0/2), n)
	)
	var (
		ρ = sign(n) * sqrt(sq(x)+sq(ρ0-y))
		θ = atan(x / (ρ0 - y))
	)
	var (
		φ = 2*atan(pow(R*F/ρ, 1/n)) - π/2
		λ = λ0 + θ/n
	)
	return geom.XY{X: rtod(λ), Y: rtod(φ)}
}
