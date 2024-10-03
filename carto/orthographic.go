package carto

import (
	. "github.com/peterstace/simplefeatures/geom"
)

// Orthographic is is a projection where the sphere is projected onto a tangent
// plane, with a point of perspective that is infinitely far away. It gives a
// view of the sphere as seen from outer space.
type Orthographic struct {
	radius       float64
	originLonLat XY
}

// NewOrthographic returns a new Orthographic projection with the given earth
// radius and projection origin. The projection has the least distortion near
// the origin.
func NewOrthographic(radius float64, originLonLat XY) *Orthographic {
	return &Orthographic{radius, originLonLat}
}

// To converts a (longitude, latitude) pair to an orthographically project (x,
// y) pair. The units of the longitude and latitude are in degrees. The units
// of the x and y coordinates are the same as that used to specify the radius.
func (m *Orthographic) To(lonLat XY) XY {
	R := m.radius
	λd := lonLat.X
	φd := lonLat.Y
	λr := dtor(λd)
	φr := dtor(φd)
	λ0r := dtor(m.originLonLat.X)
	φ0r := dtor(m.originLonLat.Y)

	// Directly from https://en.wikipedia.org/wiki/Orthographic_map_projection.
	x := R * cos(φr) * sin(λr-λ0r)
	y := R * (cos(φ0r)*sin(φr) - sin(φ0r)*cos(φr)*cos(λr-λ0r))
	return XY{X: x, Y: y}
}

// From converts an orthographically projected (x, y) pair to a (longitude,
// latitude) pair. The units of the longitude and latitude are in degrees.  The
// units of the x and y coordinates are the same as that used to specify the
// radius.
func (m *Orthographic) From(xy XY) XY {
	R := m.radius
	x := xy.X
	y := xy.Y
	λ0r := dtor(m.originLonLat.X)
	φ0r := dtor(m.originLonLat.Y)

	// Directly from https://en.wikipedia.org/wiki/Orthographic_map_projection.
	ρ := xy.Length()
	c := asin(ρ / R)
	φr := asin(cos(c)*sin(φ0r) + y*sin(c)*cos(φ0r)/ρ)
	λr := λ0r + atan(x*sin(c)/(ρ*cos(c)*cos(φ0r)-y*sin(c)*sin(φ0r)))
	return XY{X: rtod(λr), Y: rtod(φr)}
}
