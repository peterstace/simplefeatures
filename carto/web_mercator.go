package carto

import (
	"github.com/peterstace/simplefeatures/geom"
)

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

// Forward converts a (longitude, latitude) pair to a Web Mercator (x, y) pair.
func (m *WebMercator) Forward(lonlat geom.XY) geom.XY {
	λd := lonlat.X
	φd := lonlat.Y
	φr := dtor(φd)
	P := float64(int(1) << m.zoom)

	x := (λd + 180) / 360 * P
	y := (π - ln(tan(π/4+φr/2))) * P / (2 * π)
	return geom.XY{X: x, Y: y}
}

// Reverse converts a Web Mercator (x, y) pair to a (longitude, latitude) pair.
func (m *WebMercator) Reverse(xy geom.XY) geom.XY {
	x := xy.X
	y := xy.Y
	P := float64(int(1) << m.zoom)

	λd := x/P*360 - 180
	φr := 2 * (atan(exp(π-2*π*y/P)) - π/4)
	return geom.XY{X: λd, Y: rtod(φr)}
}
