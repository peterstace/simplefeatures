package carto

import (
	"github.com/peterstace/simplefeatures/geom"
)

// WebMercator is a variant of the Web Mercator projection that is used for web
// maps. The projection maps between (latitude, longitude) pairs expressed in
// degrees, and (x, y) pairs. The x and y coordinates are in the range 0 to
// 2^zoom, where zoom is the zoom level of the map.
//
// The x coordinate increases from left to right, and the y coordinate
// increases from top to bottom.
//
// It is:
//   - Conformal (shape is preserved locally at all points).
//   - Not equal area.
//   - Not equidistant.
type WebMercator struct {
	zoom int
}

// NewWebMercator returns a new [WebMercator] projection with the given zoom.
func NewWebMercator(zoom int) *WebMercator {
	return &WebMercator{zoom}
}

// Forward converts a (longitude, latitude) pair expressed in degrees to a
// projected (x, y) pair.
func (m *WebMercator) Forward(lonlat geom.XY) geom.XY {
	var (
		λd = lonlat.X
		φ  = dtor(lonlat.Y)
		P  = float64(int(1) << m.zoom)
	)
	return geom.XY{
		X: (λd + 180) / 360 * P,
		Y: (π - ln(tan(π/4+φ/2))) * P / (2 * π),
	}
}

// Reverse converts a projected (x, y) pair to a (longitude, latitude) pair
// expressed in degrees.
func (m *WebMercator) Reverse(xy geom.XY) geom.XY {
	var (
		x = xy.X
		y = xy.Y
		P = float64(int(1) << m.zoom)
	)
	var (
		λd = x/P*360 - 180
		φr = 2 * (atan(exp(π-2*π*y/P)) - π/4)
	)
	return geom.XY{
		X: λd,
		Y: rtod(φr),
	}
}
