package carto

import (
	. "github.com/peterstace/simplefeatures/geom"
)

// Orthographic is is a projection where the sphere is projected onto a tangent
// plane, with a point of perspective that is infinitely far away. It gives a
// view of the sphere as seen from outer space.
type Orthographic struct {
	radius float64
	λ0     float64
	cosφ0  float64
	sinφ0  float64
}

// NewOrthographic returns a new Orthographic projection with the given earth
// radius and projection origin. The projection has the least distortion near
// the origin.
func NewOrthographic(radius float64) *Orthographic {
	return &Orthographic{
		radius: radius,
		λ0:     0,
		cosφ0:  1, // φ0 = 0
		sinφ0:  0, // φ0 = 0
	}
}

func (m *Orthographic) SetOrigin(originLonLat XY) {
	m.λ0 = dtor(originLonLat.X)
	φ0 := dtor(originLonLat.Y)
	m.sinφ0 = sin(φ0)
	m.cosφ0 = cos(φ0)
}

// To converts a (longitude, latitude) pair to an orthographically project (x,
// y) pair. The units of the longitude and latitude are in degrees. The units
// of the x and y coordinates are the same as that used to specify the radius.
func (m *Orthographic) To(lonLat XY) XY {
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

// From converts an orthographically projected (x, y) pair to a (longitude,
// latitude) pair. The units of the longitude and latitude are in degrees.  The
// units of the x and y coordinates are the same as that used to specify the
// radius.
func (m *Orthographic) From(xy XY) XY {
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
