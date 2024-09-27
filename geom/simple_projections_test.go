package geom_test

import (
	"strconv"
	"testing"

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
			proj:      geom.NewWebMercator(0),
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
			proj:      geom.NewWebMercator(21),
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
			proj: geom.NewOrthographic(
				geom.WGS84EllipsoidMeanRadiusM,
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
