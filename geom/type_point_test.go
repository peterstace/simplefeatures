package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestPointSummary(t *testing.T) {
	for _, tc := range []struct {
		name string
		p geom.Point
		wantSummary string
	}{
		{p: geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 135, Y: -35}, Type: geom.DimXY}), wantSummary: "Point[XY] with 1 point"},
		{p: geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 135, Y: -35}, Type: geom.DimXYZ}), wantSummary: "Point[XYZ] with 1 point"},
		{p: geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 135, Y: -35}, Type: geom.DimXYM}), wantSummary: "Point[XYM] with 1 point"},
		{p: geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 135, Y: -35}, Type: geom.DimXYZM}), wantSummary: "Point[XYZM] with 1 point"},
		{p: geom.NewEmptyPoint(geom.DimXY), wantSummary: "Point[XY] with 0 points"},
		{p: geom.NewEmptyPoint(geom.DimXYZ), wantSummary: "Point[XYZ] with 0 points"},
		{p: geom.NewEmptyPoint(geom.DimXYM), wantSummary: "Point[XYM] with 0 points"},
		{p: geom.NewEmptyPoint(geom.DimXYZM), wantSummary: "Point[XYZM] with 0 points"},
	}{
		t.Run(tc.name, func(t *testing.T) {
			expectStringEq(t, tc.p.Summary(), tc.wantSummary)
			expectStringEq(t, tc.p.String(), tc.wantSummary)
		})
	}
}