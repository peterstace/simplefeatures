package carto

import (
	. "github.com/peterstace/simplefeatures/geom"
)

// Orthographic allows projecting (longitude, latitude) coordinates to (x, y)
// pairs via the orthographic projection.
//
// The orthographic is a projection where the sphere is projected onto a
// tangent plane, with a point of perspective that is infinitely far away. It
// gives a view of the sphere as seen from outer space.
//
// It is:
//   - Configured by setting the center of the projection.
//   - Not conformal, equal area or equidistant, but preserves shape, area, and
//     distance locally at the center of the projection.
type Orthographic struct {
	radius float64
	λ0     float64
	cosφ0  float64
	sinφ0  float64
}

// NewOrthographic returns a new Orthographic projection with the given earth
// radius.
func NewOrthographic(radius float64) *Orthographic {
	return &Orthographic{
		radius: radius,
		λ0:     0,
		cosφ0:  1, // φ0 = 0
		sinφ0:  0, // φ0 = 0
	}
}

// SetCenterLonLat sets the center of the projection to the given (longitude,
// latitude) pair. The center have projected coordinates (0, 0) and be the
// center of the circular map.
func (m *Orthographic) SetCenter(centerLonLat XY) {
	m.λ0 = dtor(centerLonLat.X)
	φ0 := dtor(centerLonLat.Y)
	m.sinφ0 = sin(φ0)
	m.cosφ0 = cos(φ0)
}

// Forward converts a (longitude, latitude) pair expressed in degrees to a
// projected (x, y) pair.
func (m *Orthographic) Forward(lonLat XY) XY {
	var (
		R     = m.radius
		λ     = dtor(lonLat.X)
		φ     = dtor(lonLat.Y)
		λ0    = m.λ0
		cosφ0 = m.cosφ0
		sinφ0 = m.sinφ0
	)
	return XY{
		X: R * cos(φ) * sin(λ-λ0),
		Y: R * (cosφ0*sin(φ) - sinφ0*cos(φ)*cos(λ-λ0)),
	}
}

// Reverse converts a projected (x, y) pair to a (longitude, latitude) pair
// expressed in degrees.
func (m *Orthographic) Reverse(xy XY) XY {
	var (
		R     = m.radius
		x     = xy.X
		y     = xy.Y
		λ0    = m.λ0
		cosφ0 = m.cosφ0
		sinφ0 = m.sinφ0
	)
	var (
		ρ = xy.Length()
		c = asin(ρ / R)
		φ = asin(cos(c)*sinφ0 + y*sin(c)*cosφ0/ρ)
		λ = λ0 + atan(x*sin(c)/(ρ*cos(c)*cosφ0-y*sin(c)*sinφ0))
	)
	return rtodxy(λ, φ)
}
