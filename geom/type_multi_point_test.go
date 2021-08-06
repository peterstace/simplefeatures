package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestMultiPointSummary(t *testing.T) {
	for _, tc := range []struct {
		name        string
		ps          []geom.Point
		wantSummary string
	}{
		{
			name:        "Empty",
			wantSummary: "MultiPoint[XY] with 0 points",
		},

		// Single point.
		{
			name: "XY single point",
			ps: []geom.Point{
				geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 0, Y: 0}, Type: geom.DimXY}),
			},
			wantSummary: "MultiPoint[XY] with 1 point",
		},
		{
			name: "XYZ single point",
			ps: []geom.Point{
				geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 0, Y: 0}, Z: 0.5, Type: geom.DimXYZ}),
			},
			wantSummary: "MultiPoint[XYZ] with 1 point",
		},
		{
			name: "XYM single point",
			ps: []geom.Point{
				geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 0, Y: 0}, M: 0.8, Type: geom.DimXYM}),
			},
			wantSummary: "MultiPoint[XYM] with 1 point",
		},

		// Multiple points.
		{
			name: "XY 2 points",
			ps: []geom.Point{
				geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 0, Y: 0}, Type: geom.DimXY}),
				geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 1, Y: 1}, Type: geom.DimXY}),
			},
			wantSummary: "MultiPoint[XY] with 2 points",
		},
		{
			name: "XYZ 2 points",
			ps: []geom.Point{
				geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 0, Y: 0}, Z: 0.5, Type: geom.DimXYZ}),
				geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 1, Y: 1}, Z: 0.5, Type: geom.DimXYZ}),
			},
			wantSummary: "MultiPoint[XYZ] with 2 points",
		},
		{
			name: "XYM 2 points",
			ps: []geom.Point{
				geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 0, Y: 0}, M: 0.8, Type: geom.DimXYM}),
				geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 1, Y: 1}, M: 0.8, Type: geom.DimXYM}),
			},
			wantSummary: "MultiPoint[XYM] with 2 points",
		},
		{
			name: "XY 2 points same coordinates",
			ps: []geom.Point{
				geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 0, Y: 0}, Type: geom.DimXY}),
				geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 0, Y: 0}, Type: geom.DimXY}),
			},
			wantSummary: "MultiPoint[XY] with 2 points",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := geom.NewMultiPointFromPoints(tc.ps)
			expectStringEq(t, g.Summary(), tc.wantSummary)
			expectStringEq(t, g.String(), tc.wantSummary)
		})
	}
}
