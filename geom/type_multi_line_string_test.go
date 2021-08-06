package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestMultiLineStringSummary(t *testing.T) {
	for _, tc := range []struct {
		name             string
		lineStringCoords [][]float64
		coordsType       geom.CoordinatesType
		wantSummary      string
	}{
		{
			name:             "Empty",
			lineStringCoords: [][]float64{},
			coordsType:       geom.DimXY,
			wantSummary:      "MultiLineString[XY] with 0 linestrings consisting of 0 total points",
		},
		{
			name: "XY single 2-point lines",
			lineStringCoords: [][]float64{
				{0, 0, 1, 1},
				{0, 0, -1, -1},
			},
			coordsType:  geom.DimXY,
			wantSummary: "MultiLineString[XY] with 2 linestrings consisting of 4 total points",
		},
		{
			name: "XY multiple 2-point lines",
			lineStringCoords: [][]float64{
				{0, 0, 1, 1},
				{0, 0, -1, -1},
			},
			coordsType:  geom.DimXY,
			wantSummary: "MultiLineString[XY] with 2 linestrings consisting of 4 total points",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lineStrings := make([]geom.LineString, len(tc.lineStringCoords))
			for i, coords := range tc.lineStringCoords {
				ls, err := geom.NewLineString(geom.NewSequence(coords, tc.coordsType))
				expectNoErr(t, err)
				lineStrings[i] = ls
			}
			g := geom.NewMultiLineStringFromLineStrings(lineStrings)
			expectStringEq(t, g.Summary(), tc.wantSummary)
			expectStringEq(t, g.String(), tc.wantSummary)
		})
	}
}
