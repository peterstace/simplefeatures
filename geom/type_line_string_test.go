package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestLineStringSummary(t *testing.T) {
	for _, tc := range []struct {
		name string
		coords []float64
		coordsType geom.CoordinatesType
		wantSummary string
	}{
		{name: "XY 2-point line", coords: []float64{0, 0, 1, 1}, coordsType: geom.DimXY, wantSummary: "LineString[XY] with 2 points"},
		{name: "XYZ 2-point line", coords: []float64{0, 0, 0.5, 1, 1, 0.5}, coordsType: geom.DimXYZ, wantSummary: "LineString[XYZ] with 2 points"},
		{name: "XYM 2-point line", coords: []float64{0, 0, 0.8, 1, 1, 0.8}, coordsType: geom.DimXYM, wantSummary: "LineString[XYM] with 2 points"},
		{name: "XYZM 2-point line", coords: []float64{0, 0, 0.5, 0.8, 1, 1, 0.5, 0.8}, coordsType: geom.DimXYZM, wantSummary: "LineString[XYZM] with 2 points"},

		{name: "XY 0-point line", coords: nil, coordsType: geom.DimXY, wantSummary: "LineString[XY] with 0 points"},
		{name: "XYZ 0-point line", coords: nil, coordsType: geom.DimXYZ, wantSummary: "LineString[XYZ] with 0 points"},
		{name: "XYM 0-point line", coords: nil, coordsType: geom.DimXYM, wantSummary: "LineString[XYM] with 0 points"},
		{name: "XYZM 0-point line", coords: nil, coordsType: geom.DimXYZM, wantSummary: "LineString[XYZM] with 0 points"},
	}{
		t.Run(tc.name, func(t *testing.T) {
			g, err := geom.NewLineString(geom.NewSequence(tc.coords, tc.coordsType))
			expectNoErr(t, err)
			expectStringEq(t, g.Summary(), tc.wantSummary)
			expectStringEq(t, g.String(), tc.wantSummary)
		})
	}
}
