package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestPolygonSummary(t *testing.T) {
	for _, tc := range []struct {
		name          string
		pointSequence [][]float64
		coordsType    geom.CoordinatesType
		wantSummary   string
	}{
		// Empty.
		{
			name:          "Empty polygon",
			pointSequence: [][]float64{},
			coordsType:    geom.DimXY,
			wantSummary:   "Polygon[XY] with 0 rings consisting of 0 total points",
		},

		// Basic single polygon without inner rings.
		{
			name: "XY square polygon",
			pointSequence: [][]float64{
				{-1, 1, 1, 1, 1, -1, -1, -1, -1, 1},
			},
			coordsType:  geom.DimXY,
			wantSummary: "Polygon[XY] with 1 ring consisting of 5 total points",
		},
		{
			name: "XYZ square polygon",
			pointSequence: [][]float64{
				{-1, 1, 0.5, 1, 1, 0.5, 1, -1, 0.5, -1, -1, 0.5, -1, 1, 0.5},
			},
			coordsType:  geom.DimXYZ,
			wantSummary: "Polygon[XYZ] with 1 ring consisting of 5 total points",
		},
		{
			name: "XYM square polygon",
			pointSequence: [][]float64{
				{-1, 1, 0.8, 1, 1, 0.8, 1, -1, 0.8, -1, -1, 0.8, -1, 1, 0.8},
			},
			coordsType:  geom.DimXYM,
			wantSummary: "Polygon[XYM] with 1 ring consisting of 5 total points",
		},
		{
			name: "XYMZ square polygon",
			pointSequence: [][]float64{
				{-1, 1, 0.5, 0.8, 1, 1, 0.5, 0.8, 1, -1, 0.5, 0.8, -1, -1, 0.5, 0.8, -1, 1, 0.5, 0.8},
			},
			coordsType:  geom.DimXYZM,
			wantSummary: "Polygon[XYZM] with 1 ring consisting of 5 total points",
		},

		// Polygon with single inner ring.
		{
			name: "XY 1 square within a square polygon",
			pointSequence: [][]float64{
				{-100, 100, 100, 100, 100, -100, -100, -100, -100, 100},
				{-1, 1, 1, 1, 1, -1, -1, -1, -1, 1},
			},
			coordsType:  geom.DimXY,
			wantSummary: "Polygon[XY] with 2 rings consisting of 10 total points",
		},
		{
			name: "XYZ 1 square within a square polygon",
			pointSequence: [][]float64{
				{-100, 100, 0.5, 100, 100, 0.5, 100, -100, 0.5, -100, -100, 0.5, -100, 100, 0.5},
				{-1, 1, 0.5, 1, 1, 0.5, 1, -1, 0.5, -1, -1, 0.5, -1, 1, 0.5},
			},
			coordsType:  geom.DimXYZ,
			wantSummary: "Polygon[XYZ] with 2 rings consisting of 10 total points",
		},
		{
			name: "XYM 1 square within a square polygon",
			pointSequence: [][]float64{
				{-100, 100, 0.8, 100, 100, 0.8, 100, -100, 0.8, -100, -100, 0.8, -100, 100, 0.8},
				{-1, 1, 0.8, 1, 1, 0.8, 1, -1, 0.8, -1, -1, 0.8, -1, 1, 0.8},
			},
			coordsType:  geom.DimXYM,
			wantSummary: "Polygon[XYM] with 2 rings consisting of 10 total points",
		},
		{
			name: "XYMZ 1 square within a square polygon",
			pointSequence: [][]float64{
				{-100, 100, 0.5, 0.8, 100, 100, 0.5, 0.8, 100, -100, 0.5, 0.8, -100, -100, 0.5, 0.8, -100, 100, 0.5, 0.8},
				{-1, 1, 0.5, 0.8, 1, 1, 0.5, 0.8, 1, -1, 0.5, 0.8, -1, -1, 0.5, 0.8, -1, 1, 0.5, 0.8},
			},
			coordsType:  geom.DimXYZM,
			wantSummary: "Polygon[XYZM] with 2 rings consisting of 10 total points",
		},

		// Polygon with multiple inner rings.
		{
			name: "XY 2 squares within a square polygon",
			pointSequence: [][]float64{
				{-100, 100, 100, 100, 100, -100, -100, -100, -100, 100},
				{-1, 1, 1, 1, 1, -1, -1, -1, -1, 1},
				{10, 10, 11, 10, 11, 9, 10, 9, 10, 10},
			},
			coordsType:  geom.DimXY,
			wantSummary: "Polygon[XY] with 3 rings consisting of 15 total points",
		},
		{
			name: "XYZ 2 squares within a square polygon",
			pointSequence: [][]float64{
				{-100, 100, 0.5, 100, 100, 0.5, 100, -100, 0.5, -100, -100, 0.5, -100, 100, 0.5},
				{-1, 1, 0.5, 1, 1, 0.5, 1, -1, 0.5, -1, -1, 0.5, -1, 1, 0.5},
				{10, 10, 0.5, 11, 10, 0.5, 11, 9, 0.5, 10, 9, 0.5, 10, 10, 0.5},
			},
			coordsType:  geom.DimXYZ,
			wantSummary: "Polygon[XYZ] with 3 rings consisting of 15 total points",
		},
		{
			name: "XYM 2 squares within a square polygon",
			pointSequence: [][]float64{
				{-100, 100, 0.8, 100, 100, 0.8, 100, -100, 0.8, -100, -100, 0.8, -100, 100, 0.8},
				{-1, 1, 0.8, 1, 1, 0.8, 1, -1, 0.8, -1, -1, 0.8, -1, 1, 0.8},
				{10, 10, 0.8, 11, 10, 0.8, 11, 9, 0.8, 10, 9, 0.8, 10, 10, 0.8},
			},
			coordsType:  geom.DimXYM,
			wantSummary: "Polygon[XYM] with 3 rings consisting of 15 total points",
		},
		{
			name: "XYMZ 2 squares within a square polygon",
			pointSequence: [][]float64{
				{-100, 100, 0.5, 0.8, 100, 100, 0.5, 0.8, 100, -100, 0.5, 0.8, -100, -100, 0.5, 0.8, -100, 100, 0.5, 0.8},
				{-1, 1, 0.5, 0.8, 1, 1, 0.5, 0.8, 1, -1, 0.5, 0.8, -1, -1, 0.5, 0.8, -1, 1, 0.5, 0.8},
				{10, 10, 0.5, 0.8, 11, 10, 0.5, 0.8, 11, 9, 0.5, 0.8, 10, 9, 0.5, 0.8, 10, 10, 0.5, 0.8},
			},
			coordsType:  geom.DimXYZM,
			wantSummary: "Polygon[XYZM] with 3 rings consisting of 15 total points",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var lineStrings []geom.LineString
			for _, lineStringCoord := range tc.pointSequence {
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
