package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestPolygonSummary(t *testing.T) {
	for _, tc := range []struct {
		name             string
		lineStringCoords [][]float64
		coordsType       geom.CoordinatesType
		wantSummary      string
	}{
		// Basic single polygon without inner rings.
		{
			name: "XY square polygon",
			lineStringCoords: [][]float64{
				{0, 0, 1, 0, 1, -1, 0, -1, 0, 0},
			},
			coordsType: geom.DimXY,
			wantSummary: "Polygon[XY] with 0 rings consisting of 5 total points",
		},
		{
			name: "XYZ square polygon",
			lineStringCoords: [][]float64{
				{0, 0, 0.5, 1, 0, 0.5, 1, -1, 0.5, 0, -1, 0.5, 0, 0, 0.5},
			},
			coordsType: geom.DimXYZ,
			wantSummary: "Polygon[XYZ] with 0 rings consisting of 5 total points",
		},
		{
			name: "XYM square polygon",
			lineStringCoords: [][]float64{
				{0, 0, 0.8, 1, 0, 0.8, 1, -1, 0.8, 0, -1, 0.8, 0, 0, 0.8},
			},
			coordsType: geom.DimXYM,
			wantSummary: "Polygon[XYM] with 0 rings consisting of 5 total points",
		},
		{
			name: "XYMZ square polygon",
			lineStringCoords: [][]float64{
				{0, 0, 0.5, 0.8, 1, 0, 0.5, 0.8, 1, -1, 0.5, 0.8, 0, -1, 0.5, 0.8, 0, 0, 0.5, 0.8},
			},
			coordsType: geom.DimXYZM,
			wantSummary: "Polygon[XYZM] with 0 rings consisting of 5 total points",
		},

		// Polygons with inner ring.
		{
			name: "XY 1 square within a square polygon",
			lineStringCoords: [][]float64{
				{-1, 1, 2, 1, 2, -2, -1, -2, -1, 1},
				{0, 0, 1, 0, 1, -1, 0, -1, 0, 0},
			},
			coordsType: geom.DimXY,
			wantSummary: "Polygon[XY] with 1 ring consisting of 10 total points",
		},
		{
			name: "XYZ 1 square within a square polygon",
			lineStringCoords: [][]float64{
				{-1, 1, 0.5, 2, 1, 0.5, 2, -2, 0.5, -1, -2, 0.5, -1, 1, 0.5},
				{0, 0, 0.5, 1, 0, 0.5, 1, -1, 0.5, 0, -1, 0.5, 0, 0, 0.5},
			},
			coordsType: geom.DimXYZ,
			wantSummary: "Polygon[XYZ] with 1 ring consisting of 10 total points",
		},
		{
			name: "XYM 1 square within a square polygon",
			lineStringCoords: [][]float64{
				{-1, 1, 0.8, 2, 1, 0.8, 2, -2, 0.8, -1, -2, 0.8, -1, 1, 0.8},
				{0, 0, 0.8, 1, 0, 0.8, 1, -1, 0.8, 0, -1, 0.8, 0, 0, 0.8},
			},
			coordsType: geom.DimXYM,
			wantSummary: "Polygon[XYM] with 1 ring consisting of 10 total points",
		},
		{
			name: "XYMZ 1 square within a square polygon",
			lineStringCoords: [][]float64{
				{-1, 1, 0.5, 0.8, 2, 1, 0.5, 0.8, 2, -2, 0.5, 0.8, -1, -2, 0.5, 0.8, -1, 1, 0.5, 0.8},
				{0, 0, 0.5, 0.8, 1, 0, 0.5, 0.8, 1, -1, 0.5, 0.8, 0, -1, 0.5, 0.8, 0, 0, 0.5, 0.8},
			},
			coordsType: geom.DimXYZM,
			wantSummary: "Polygon[XYZM] with 1 ring consisting of 10 total points",
		},

		// Polygons with inner rings.
		{
			name: "XY 2 squares within a square polygon",
			lineStringCoords: [][]float64{
				{-10, 10, 20, 10, 20, -20, -10, -20, -10, 10},
				{0, 0, 1, 0, 1, -1, 0, -1, 0, 0},
				{1, -1, 2, -1, 2, -2, 1, -2, 1, -1},
			},
			coordsType: geom.DimXY,
			wantSummary: "Polygon[XY] with 2 rings consisting of 15 total points",
		},
		{
			name: "XYZ 2 squares within a square polygon",
			lineStringCoords: [][]float64{
				{-10, 10, 0.5, 20, 10, 0.5, 20, -20, 0.5, -10, -20, 0.5, -10, 10, 0.5},
				{0, 0, 0.5, 1, 0, 0.5, 1, -1, 0.5, 0, -1, 0.5, 0, 0, 0.5},
				{1, -1, 0.5, 2, -1, 0.5, 2, -2, 0.5, 1, -2, 0.5, 1, -1, 0.5},
			},
			coordsType: geom.DimXYZ,
			wantSummary: "Polygon[XYZ] with 2 rings consisting of 15 total points",
		},
		{
			name: "XYM 2 squares within a square polygon",
			lineStringCoords: [][]float64{
				{-10, 10, 0.8, 20, 10, 0.8, 20, -20, 0.8, -10, -20, 0.8, -10, 10, 0.8},
				{0, 0, 0.8, 1, 0, 0.8, 1, -1, 0.8, 0, -1, 0.8, 0, 0, 0.8},
				{1, -1, 0.8, 2, -1, 0.8, 2, -2, 0.8, 1, -2, 0.8, 1, -1, 0.8},
			},
			coordsType: geom.DimXYM,
			wantSummary: "Polygon[XYM] with 2 rings consisting of 15 total points",
		},
		{
			name: "XYMZ 2 squares within a square polygon",
			lineStringCoords: [][]float64{
				{-10, 10, 0.5, 0.8, 20, 10, 0.5, 0.8, 20, -20, 0.5, 0.8, -10, -20, 0.5, 0.8, -10, 10, 0.5, 0.8},
				{0, 0, 0.5, 0.8, 1, 0, 0.5, 0.8, 1, -1, 0.5, 0.8, 0, -1, 0.5, 0.8, 0, 0, 0.5, 0.8},
				{1, -1, 0.5, 0.8, 2, -1, 0.5, 0.8, 2, -2, 0.5, 0.8, 1, -2, 0.5, 0.8, 1, -1, 0.5, 0.8},
			},
			coordsType: geom.DimXYZM,
			wantSummary: "Polygon[XYZM] with 2 rings consisting of 15 total points",
		},
	}{
		t.Run(tc.name, func(t *testing.T) {
			var lineStrings []geom.LineString
			for _, lineStringCoord := range tc.lineStringCoords {
				ls, err := geom.NewLineString(geom.NewSequence(lineStringCoord, tc.coordsType))
				expectNoErr(t, err)
				lineStrings = append(lineStrings, ls)
			}
			p, err := geom.NewPolygonFromRings(lineStrings)
			expectNoErr(t, err)
			expectStringEq(t, p.Summary(), tc.wantSummary)
			expectStringEq(t, p.String(), tc.wantSummary)
		})
	}
}

