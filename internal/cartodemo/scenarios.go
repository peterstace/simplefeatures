package main

import (
	"math"

	"github.com/peterstace/simplefeatures/carto"
	"github.com/peterstace/simplefeatures/geom"
)

const (
	earthRadius = 6371000
	earthCircum = 2 * pi * earthRadius
	pi          = math.Pi
)

var (
	fullWorldMask = rectangle(
		xy(-180, +90),
		xy(+180, -90),
	)
	sqrt2 = math.Sqrt(2)
)

func xy(x, y float64) geom.XY {
	return geom.XY{X: x, Y: y}
}

var scenarios = map[string]worldProjectionFixture{
	"web_mercator": {
		proj:      carto.NewWebMercator(0).Forward,
		worldMask: fullWorldMask,
		mapMask:   rectangle(xy(0, 0), xy(1, 1)),
		mapFlipY:  true,
	},
	"equirectangular_marinus": func() worldProjectionFixture {
		p := carto.NewEquirectangular(earthRadius)
		p.SetStandardParallels(36)
		cos36 := math.Cos(36 * pi / 180)
		return worldProjectionFixture{
			proj:      p.Forward,
			worldMask: fullWorldMask,
			mapMask: rectangle(
				xy(-0.5*earthCircum*cos36, +0.25*earthCircum),
				xy(+0.5*earthCircum*cos36, -0.25*earthCircum),
			),
		}
	}(),
	"lambert_cylindrical_equal_area": {
		proj:      carto.NewLambertCylindricalEqualArea(earthRadius).Forward,
		worldMask: fullWorldMask,
		mapMask: rectangle(
			xy(-0.5*earthCircum, +0.25*earthCircum*2/pi),
			xy(+0.5*earthCircum, -0.25*earthCircum*2/pi),
		),
	},
	"sinusoidal": func() worldProjectionFixture {
		var lhs, rhs []float64
		for lat := -90.0; lat <= 90; lat++ {
			lhs = append(lhs,
				-0.5*earthCircum*math.Cos(lat*pi/180),
				lat/90*0.25*earthCircum,
			)
			rhs = append(rhs,
				+0.5*earthCircum*math.Cos(lat*pi/180),
				-lat/90*0.25*earthCircum,
			)
		}
		all := append(lhs, rhs...)
		mapMask := geom.NewPolygonXY(all)
		return worldProjectionFixture{
			proj:      carto.NewSinusoidal(earthRadius).Forward,
			worldMask: fullWorldMask,
			mapMask:   mapMask,
		}
	}(),
	"orthographic_north_pole": func() worldProjectionFixture {
		proj := carto.NewOrthographic(earthRadius)
		proj.SetCenter(geom.XY{X: 15, Y: 90})
		return worldProjectionFixture{
			proj:      proj.Forward,
			worldMask: geom.NewSingleRingPolygonXY(-180, 0, 180, 0, 180, 90, -180, 90, -180, 0),
			mapMask:   circle(xy(0, 0), earthRadius),
			paddingPx: 2,
		}
	}(),
	"azimuthal_equidistant": func() worldProjectionFixture {
		proj := carto.NewAzimuthalEquidistant(earthRadius)
		proj.SetCenter(geom.XY{X: 0, Y: 90})
		return worldProjectionFixture{
			proj:      proj.Forward,
			worldMask: geom.NewSingleRingPolygonXY(-180, 90, +180, 90, +180, -89.99, -180, -89.99, -180, 90),
			mapMask:   circle(xy(0, 0), earthCircum/2),
			paddingPx: 2,
		}
	}(),
	"equidistant_conic_europe":       equidistantConic(40, 55, xy(15, 47)),
	"lambert_conformal_conic_europe": lambertConformalConic(40, 55, xy(15, 47), -30, 90),
	"albers_equal_area_conic_europe": albersEqualAreaConic(xy(15, 47), 40, 55),
}

func equidistantConic(stdParallel1, stdParallel2 float64, origin geom.XY) worldProjectionFixture {
	p := carto.NewEquidistantConic(earthRadius)
	p.SetStandardParallels(stdParallel1, stdParallel2)
	p.SetOrigin(origin)

	const eps = 0.1
	mapMask := geom.NewSingleRingPolygonXY(
		-180+eps, -90+eps,
		-180+eps, +90-eps,
		+180-eps, +90-eps,
		+180-eps, -90+eps,
		-180+eps, -90+eps,
	)
	mapMask = mapMask.Densify(0.1)
	mapMask = mapMask.TransformXY(p.Forward)

	return worldProjectionFixture{
		proj:         p.Forward,
		worldMask:    fullWorldMask,
		mapMask:      mapMask,
		stdParallels: [2]float64{stdParallel1, stdParallel2},
		paddingPx:    2,
	}
}

func lambertConformalConic(
	stdParallel1, stdParallel2 float64,
	origin geom.XY,
	minLat, maxLat float64,
) worldProjectionFixture {
	p := carto.NewLambertConformalConic(earthRadius)
	p.SetOrigin(origin)
	p.SetStandardParallels(stdParallel1, stdParallel2)

	worldMask := geom.NewSingleRingPolygonXY(
		-180, minLat,
		-180, maxLat,
		+180, maxLat,
		+180, minLat,
		-180, minLat,
	)

	mapMask := worldMask.Densify(0.1).TransformXY(p.Forward)
	return worldProjectionFixture{
		proj:         p.Forward,
		worldMask:    worldMask,
		mapMask:      mapMask,
		stdParallels: [2]float64{stdParallel1, stdParallel2},
		paddingPx:    2,
	}
}

func albersEqualAreaConic(origin geom.XY, stdParallel1, stdParallel2 float64) worldProjectionFixture {
	p := carto.NewAlbersEqualAreaConic(earthRadius)
	p.SetOrigin(origin)
	p.SetStandardParallels(stdParallel1, stdParallel2)

	worldMask := geom.NewSingleRingPolygonXY(
		-180, -90,
		-180, +90,
		+180, +90,
		+180, -90,
		-180, -90,
	)
	mapMask := worldMask.Densify(0.1).TransformXY(p.Forward)

	return worldProjectionFixture{
		proj:         p.Forward,
		worldMask:    worldMask,
		mapMask:      mapMask,
		stdParallels: [2]float64{stdParallel1, stdParallel2},
		paddingPx:    2,
	}
}
