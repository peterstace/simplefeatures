package geom

import "math"

// WebMercator is a variant of the Web Mercator projection that is used for web
// maps. The projection maps between (latitude, longitude) pairs expressed in
// degrees, and (x, y) pairs. The x and y coordinates are in the range [0,
// 2^zoom], where zoom is the zoom level of the map.
//
// The x coordinate ranges from left to right, and the y coordinate increases
// from top to bottom.
type WebMercator struct {
	zoom int
}

// NewWebMercator returns a new WebMercator projection with the given zoom.
func NewWebMercator(zoom int) *WebMercator {
	return &WebMercator{zoom}
}

// To converts a (longitude, latitude) pair to a Web Mercator (x, y) pair.
func (m *WebMercator) To(lonlat XY) XY {
	λd := lonlat.X
	φd := lonlat.Y
	φr := dtor(φd)
	P := float64(int(1) << m.zoom)

	// Directly from https://en.wikipedia.org/wiki/Web_Mercator_projection.
	x := (λd + 180) / 360 * P
	y := (π - ln(tan(π/4+φr/2))) * P / (2 * π)
	return XY{x, y}
}

// From converts a Web Mercator (x, y) pair to a (longitude, latitude) pair.
func (m *WebMercator) From(xy XY) XY {
	x := xy.X
	y := xy.Y
	P := float64(int(1) << m.zoom)

	// Deduced from https://en.wikipedia.org/wiki/Web_Mercator_projection via
	// inverting the equations.
	λd := x/P*360 - 180
	φr := 2 * (atan(exp(π-2*π*y/P)) - π/4)
	return XY{λd, rtod(φr)}
}

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
	return XY{x, y}
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
	return XY{rtod(λr), rtod(φr)}
}

func dtor(d float64) float64 {
	return d * π / 180
}

func rtod(r float64) float64 {
	return r * 180 / π
}

// These are redefined with a shorter name to make the formulas more readable.
const (
	π = math.Pi
)

// These are redefined with shorter names to make the formulas more readable.
var (
	tan  = math.Tan
	ln   = math.Log
	sin  = math.Sin
	cos  = math.Cos
	atan = math.Atan
	asin = math.Asin
	exp  = math.Exp
)

const (
	WGS84EllipsoidEquatorialRadiusM = 6378137.0
	WGS84EllipsoidPolarRadiusM      = 6356752.314245
	WGS84EllipsoidMeanRadiusM       = (2*WGS84EllipsoidEquatorialRadiusM + WGS84EllipsoidPolarRadiusM) / 3
)
