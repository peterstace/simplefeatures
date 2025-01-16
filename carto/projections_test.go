package carto_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/carto"
	"github.com/peterstace/simplefeatures/geom"
)

type projection interface {
	Forward(lonlat geom.XY) geom.XY
	Reverse(xy geom.XY) geom.XY
}

type projectionSubTest struct {
	lotLat    geom.XY
	projected geom.XY
}

func xy(x, y float64) geom.XY {
	return geom.XY{X: x, Y: y}
}

func TestProjections(t *testing.T) {
	for _, pc := range []struct {
		name      string
		proj      projection
		threshold float64
		subtests  []projectionSubTest
	}{
		{
			name:      "WebMercator0",
			proj:      carto.NewWebMercator(0),
			threshold: 1.0 / (1 << 16), // 1/65536th of a tile.
			subtests: []projectionSubTest{
				{
					xy(0, 0),
					xy(0.5, 0.5),
				},
				{
					xy(151.20756306500027, -33.86648215268569),
					xy(0.9200210085138897, 0.6000844973286593),
				},
			},
		},
		{
			name:      "WebMercator21",
			proj:      carto.NewWebMercator(21),
			threshold: 1.0 / (1 << 16), // 1/65536th of a tile.
			subtests: []projectionSubTest{
				{
					xy(151.20756306500027, -33.86648215268569),
					xy(1929423.8980469208, 1258468.4037417925),
				},
			},
		},
		{
			name: "OrthographicAtSydney",
			proj: func() projection {
				p := carto.NewOrthographic(
					carto.WGS84EllipsoidMeanRadiusM,
				)
				p.SetCenter(xy(151, -34))
				return p
			}(),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{
				{
					xy(151, -34),
					xy(0, 0),
				},
				{
					// 1km north of origin.
					xy(151, -33.99100679628548),
					xy(0, 1000),
				},
				{
					// ~100km south west of origin.
					xy(150.29102511044510493, -34.68753125394282932),
					xy(-64821.441153708925, -76672.52425247061),
				},
			},
		},
		{
			name: "LambertCylindricalEqual",
			proj: carto.NewLambertCylindricalEqualArea(
				carto.WGS84EllipsoidMeanRadiusM,
			),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{{
				xy(151, -34),
				xy(1.679045703992921e7, -3.5626228929251498e6),
			}},
		},
		{
			name: "LambertCylindricalEqualAtSydney",
			proj: func() projection {
				p := carto.NewLambertCylindricalEqualArea(carto.WGS84EllipsoidMeanRadiusM)
				p.SetCentralMeridian(151)
				return p
			}(),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{{
				xy(151, -34),
				xy(0, -3.5626228929251498e6),
			}},
		},
		{
			name:      "Sinusoidal",
			proj:      carto.NewSinusoidal(carto.WGS84EllipsoidMeanRadiusM),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{{
				xy(151, -34),
				xy(1.3919919746472625e+07, -3.780632710977439e+06),
			}},
		},
		{
			name: "SinusoidalAtSydney",
			proj: func() projection {
				p := carto.NewSinusoidal(carto.WGS84EllipsoidMeanRadiusM)
				p.SetCentralMeridian(151)
				return p
			}(),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{{
				xy(151, -34),
				xy(0, -3.780632710977439e+06),
			}},
		},
		{
			name: "Equirectangular - Plate Carree",
			proj: func() projection {
				return carto.NewEquirectangular(carto.WGS84EllipsoidMeanRadiusM)
			}(),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{{
				xy(151, -34),
				xy(1.679045703992921e+07, -3.780632710977439e+06),
			}},
		},
		{
			name: "Equirectangular - Marinus of Tyre",
			proj: func() projection {
				p := carto.NewEquirectangular(carto.WGS84EllipsoidMeanRadiusM)
				p.SetStandardParallels(36)
				return p
			}(),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{{
				// Gibraltar, ~480km west of 0 degrees and at ~36 degrees latitude.
				xy(-5.34660683624621225, 36.1335656729737309),
				xy(-480973.8495682527, 4.0178747161028227e+06),
			}},
		},
		{
			name: "Azimuthal Equidistant - North Pole",
			proj: func() projection {
				p := carto.NewAzimuthalEquidistant(
					carto.WGS84EllipsoidMeanRadiusM,
				)
				p.SetCenter(xy(0, 90))
				return p
			}(),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{
				{ // Hamburg:
					xy(9.988519873740467, 53.434757149649016),
					xy(705229.5, -4004246.7),
				},
				{ // Port Moresby:
					xy(147.1827863021232, -9.36844599194037),
					xy(5988277, 9285859),
				},
			},
		},
		{
			name: "Azimuthal Equidistant - Africa",
			proj: func() projection {
				p := carto.NewAzimuthalEquidistant(
					carto.WGS84EllipsoidMeanRadiusM,
				)
				p.SetCenter(xy(0, 0))
				return p
			}(),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{
				{ // Cape Town:
					xy(18.483735820900083, -33.95848592499432),
					xy(1805674, -3835659),
				},
			},
		},
		{
			name: "Equidistant Conic - South America",
			proj: func() projection {
				p := carto.NewEquidistantConic(
					carto.WGS84EllipsoidMeanRadiusM,
				)
				p.SetOrigin(xy(-60, -32))
				p.SetStandardParallels(-5, -42)
				return p
			}(),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{
				{ // Rio de Janeiro:
					xy(-43.2, -22.8),
					xy(1629961.7759447654, 929251.645477184),
				},
				{ // Baltimore:
					xy(-76.6, 39.3),
					xy(-2392910.752006106, 7792228.9404544085),
				},
			},
		},
		{
			name: "Equidistant Conic - North Asia",
			proj: func() projection {
				p := carto.NewEquidistantConic(
					carto.WGS84EllipsoidMeanRadiusM,
				)
				p.SetOrigin(xy(95, 30))
				p.SetStandardParallels(15, 65)
				return p
			}(),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{
				{ // Beijing:
					xy(116.44497408510593, 39.890737551498475),
					xy(1643407.6, 1292149.5),
				},
			},
		},
		{
			name: "Lambert Conformal Conic - Canada",
			proj: func() projection {
				p := carto.NewLambertConformalConic(
					carto.WGS84EllipsoidMeanRadiusM,
				)
				p.SetOrigin(xy(-96, 40))
				p.SetStandardParallels(50, 70)
				return p
			}(),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{
				{ // Toronto:
					xy(-79.3832, 43.6532),
					xy(1353292.7229285287, 590902.0666354574),
				},
				{ // Vancouver:
					xy(-123.1216, 49.2827),
					xy(-1916086.3118012992, 1453088.303860319),
				},
			},
		},
		{
			name: "Albers Equal Area Conic - Australia",
			proj: func() projection {
				p := carto.NewAlbersEqualAreaConic(
					carto.WGS84EllipsoidMeanRadiusM,
				)
				p.SetOrigin(xy(132, 0))
				p.SetStandardParallels(-18, -36)
				return p
			}(),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{
				{ // Sydney:
					xy(151.2146821, -33.8574973),
					xy(1757815.279206157, -3843578.921069043),
				},
				{ // Perth:
					xy(115.5397172, -31.9949202),
					xy(-1534150.6162269458, -3601473.816874394),
				},
			},
		},
	} {
		t.Run(pc.name, func(t *testing.T) {
			for i, st := range pc.subtests {
				t.Run(strconv.Itoa(i), func(t *testing.T) {
					t.Run("Forward", func(t *testing.T) {
						got := pc.proj.Forward(st.lotLat)
						expectXYWithinTolerance(t, got, st.projected, pc.threshold)
					})
					t.Run("Reverse", func(t *testing.T) {
						got := pc.proj.Reverse(st.projected)
						const threshold = 1e-8 // 1e-8 degrees is about 1mm.
						expectXYWithinTolerance(t, got, st.lotLat, threshold)
					})
				})
			}
		})
	}
}

func expectXYWithinTolerance(tb testing.TB, got, want geom.XY, tolerance float64) {
	tb.Helper()
	if delta := math.Abs(got.Sub(want).Length()); delta > tolerance {
		tb.Errorf("\ngot:  %v\nwant: %v\n", got, want)
	}
}
