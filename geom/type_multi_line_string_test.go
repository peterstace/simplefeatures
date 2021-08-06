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
		// Empty.
		{
			name:             "Empty",
			lineStringCoords: [][]float64{},
			coordsType:       geom.DimXY,
			wantSummary:      "MultiLineString[XY] with 0 linestrings consisting of 0 total points",
		},

		// Single line string.
		{
			name: "XY single 2-point lines",
			lineStringCoords: [][]float64{
				{0, 0, 1, 1},
			},
			coordsType:  geom.DimXY,
			wantSummary: "MultiLineString[XY] with 1 linestring consisting of 2 total points",
		},
		{
			name: "XYZ single 2-point lines",
			lineStringCoords: [][]float64{
				{0, 0, 0.5, 1, 1, 0.5},
			},
			coordsType:  geom.DimXYZ,
			wantSummary: "MultiLineString[XYZ] with 1 linestring consisting of 2 total points",
		},
		{
			name: "XYM single 2-point lines",
			lineStringCoords: [][]float64{
				{0, 0, 0.8, 1, 1, 0.8},
			},
			coordsType:  geom.DimXYM,
			wantSummary: "MultiLineString[XYM] with 1 linestring consisting of 2 total points",
		},
		{
			name: "XYZM single 2-point lines",
			lineStringCoords: [][]float64{
				{0, 0, 0.5, 0.8, 1, 1, 0.5, 0.8},
			},
			coordsType:  geom.DimXYZM,
			wantSummary: "MultiLineString[XYZM] with 1 linestring consisting of 2 total points",
		},

		// Multiple line strings.
		{
			name: "XY multiple 2-point lines",
			lineStringCoords: [][]float64{
				{0, 0, 1, 1},
				{0, 0, -1, -1},
			},
			coordsType:  geom.DimXY,
			wantSummary: "MultiLineString[XY] with 2 linestrings consisting of 4 total points",
		},
		{
			name: "XYZ single 2-point lines",
			lineStringCoords: [][]float64{
				{0, 0, 0.5, 1, 1, 0.5},
				{0, 0, 0.5, -1, -1, 0.5},
			},
			coordsType:  geom.DimXYZ,
			wantSummary: "MultiLineString[XYZ] with 2 linestrings consisting of 4 total points",
		},
		{
			name: "XYM single 2-point lines",
			lineStringCoords: [][]float64{
				{0, 0, 0.8, 1, 1, 0.8},
				{0, 0, 0.8, -1, -1, 0.8},
			},
			coordsType:  geom.DimXYM,
			wantSummary: "MultiLineString[XYM] with 2 linestrings consisting of 4 total points",
		},
		{
			name: "XYZM single 2-point lines",
			lineStringCoords: [][]float64{
				{0, 0, 0.5, 0.8, 1, 1, 0.5, 0.8},
				{0, 0, 0.5, 0.8, -1, -1, 0.5, 0.8},
			},
			coordsType:  geom.DimXYZM,
			wantSummary: "MultiLineString[XYZM] with 2 linestrings consisting of 4 total points",
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
