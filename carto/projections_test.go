package carto_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/carto"
	"github.com/peterstace/simplefeatures/geom"
)

type projection interface {
	To(lonlat geom.XY) geom.XY
	From(xy geom.XY) geom.XY
}

type projectionSubTest struct {
	lotLat    geom.XY
	projected geom.XY
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
					geom.XY{0, 0},
					geom.XY{0.5, 0.5},
				},
				{
					geom.XY{151.20756306500027, -33.86648215268569},
					geom.XY{0.9200210085138897, 0.6000844973286593},
				},
			},
		},
		{
			name:      "WebMercator21",
			proj:      carto.NewWebMercator(21),
			threshold: 1.0 / (1 << 16), // 1/65536th of a tile.
			subtests: []projectionSubTest{
				{
					geom.XY{151.20756306500027, -33.86648215268569},
					geom.XY{1929423.8980469208, 1258468.4037417925},
				},
			},
		},
		{
			name: "OrthographicAtSydney",
			proj: carto.NewOrthographic(
				carto.WGS84EllipsoidMeanRadiusM,
				geom.XY{151, -34},
			),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{
				{
					geom.XY{151, -34},
					geom.XY{0, 0},
				},
				{
					// 1km north of origin.
					geom.XY{151, -33.99100679628548},
					geom.XY{0, 1000},
				},
				{
					// ~100km south west of origin.
					geom.XY{150.29102511044510493, -34.68753125394282932},
					geom.XY{-64821.441153708925, -76672.52425247061},
				},
			},
		},
		{
			name: "LambertCylindricalEqual",
			proj: carto.NewLambertCylindricalEqualArea(
				carto.WGS84EllipsoidMeanRadiusM, 0,
			),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{{
				geom.XY{151, -34},
				geom.XY{1.679045703992921e7, -3.5626228929251498e6},
			}},
		},
		{
			name:      "LambertCylindricalEqualAtSydney",
			proj:      carto.NewLambertCylindricalEqualArea(carto.WGS84EllipsoidMeanRadiusM, 151),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{{
				geom.XY{151, -34},
				geom.XY{0, -3.5626228929251498e6},
			}},
		},
		{
			name:      "Sinusoidal",
			proj:      carto.NewSinusoidal(carto.WGS84EllipsoidMeanRadiusM, 0),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{{
				geom.XY{151, -34},
				geom.XY{1.3919919746472625e+07, -3.780632710977439e+06},
			}},
		},
		{
			name:      "SinusoidalAtSydney",
			proj:      carto.NewSinusoidal(carto.WGS84EllipsoidMeanRadiusM, 151),
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{{
				geom.XY{151, -34},
				geom.XY{0, -3.780632710977439e+06},
			}},
		},
		{
			name:      "Equirectangular - Plate Carree",
			proj:      &carto.Equirectangular{Radius: carto.WGS84EllipsoidMeanRadiusM},
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{{
				geom.XY{151, -34},
				geom.XY{1.679045703992921e+07, -3.780632710977439e+06},
			}},
		},
		{
			name:      "Equirectangular - Marinus of Tyre",
			proj:      &carto.Equirectangular{Radius: carto.WGS84EllipsoidMeanRadiusM, StandardParallels: 36},
			threshold: 1e-3, // 1mm
			subtests: []projectionSubTest{{
				// Gibraltar, ~480km west of 0 degrees and at ~36 degrees latitude.
				geom.XY{-5.34660683624621225, 36.1335656729737309},
				geom.XY{-480973.8495682527, 4.0178747161028227e+06},
			}},
		},
	} {
		t.Run(pc.name, func(t *testing.T) {
			for i, st := range pc.subtests {
				t.Run(strconv.Itoa(i), func(t *testing.T) {
					t.Run("To", func(t *testing.T) {
						got := pc.proj.To(st.lotLat)
						expectXYWithinTolerance(t, got, st.projected, pc.threshold)
					})
					t.Run("From", func(t *testing.T) {
						got := pc.proj.From(st.projected)
						const threshold = 1e-8 // 1e-8 degrees is about 1mm.
						expectXYWithinTolerance(t, got, st.lotLat, threshold)
					})
				})
			}
		})
	}
}
