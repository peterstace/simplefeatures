package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestPointSummary(t *testing.T) {
	for _, tc := range []struct {
		p geom.Point
		wantSummary string
	}{
		{p: geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 0, Y: 0}, Type: geom.DimXY}), wantSummary: "Point[XY] with 1 point"},
		{p: geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 0, Y: 0}, Z: 0.5, Type: geom.DimXYZ}), wantSummary: "Point[XYZ] with 1 point"},
		{p: geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 0, Y: 0}, M: 0.8, Type: geom.DimXYM}), wantSummary: "Point[XYM] with 1 point"},
		{p: geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 0, Y: 0}, Z: 0.5, M: 0.8, Type: geom.DimXYZM}), wantSummary: "Point[XYZM] with 1 point"},
		{p: geom.NewEmptyPoint(geom.DimXY), wantSummary: "Point[XY] with 0 points"},
		{p: geom.NewEmptyPoint(geom.DimXYZ), wantSummary: "Point[XYZ] with 0 points"},
		{p: geom.NewEmptyPoint(geom.DimXYM), wantSummary: "Point[XYM] with 0 points"},
		{p: geom.NewEmptyPoint(geom.DimXYZM), wantSummary: "Point[XYZM] with 0 points"},
	}{
		t.Run(tc.wantSummary, func(t *testing.T) {
			expectStringEq(t, tc.p.Summary(), tc.wantSummary)
			expectStringEq(t, tc.p.String(), tc.wantSummary)
		})
	}
}