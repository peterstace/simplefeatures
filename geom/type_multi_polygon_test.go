package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestMultiPolygonSummary(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		multiPolyPointSequences [][][]float64
		coordsType              geom.CoordinatesType
		wantSummary             string
	}{
		// Empty.
		{
			name:                    "Empty polygon",
			multiPolyPointSequences: [][][]float64{},
			coordsType:              geom.DimXY,
			wantSummary:             "MultiPolygon[XY] with 0 polygons consisting of 0 total rings and 0 total points",
		},

		// Basic single polygon without inner rings.
		{
			name: "XY 1 square polygon",
			multiPolyPointSequences: [][][]float64{
				{
					{-1, 1, 1, 1, 1, -1, -1, -1, -1, 1},
				},
			},
			coordsType:  geom.DimXY,
			wantSummary: "MultiPolygon[XY] with 1 polygon consisting of 1 total ring and 5 total points",
		},
		{
			name: "XYZ 1 square polygon",
			multiPolyPointSequences: [][][]float64{
				{
					{-1, 1, 0.5, 1, 1, 0.5, 1, -1, 0.5, -1, -1, 0.5, -1, 1, 0.5},
				},
			},
			coordsType:  geom.DimXYZ,
			wantSummary: "MultiPolygon[XYZ] with 1 polygon consisting of 1 total ring and 5 total points",
		},
		{
			name: "XYM 1 square polygon",
			multiPolyPointSequences: [][][]float64{
				{
					{-1, 1, 0.8, 1, 1, 0.8, 1, -1, 0.8, -1, -1, 0.8, -1, 1, 0.8},
				},
			},
			coordsType:  geom.DimXYM,
			wantSummary: "MultiPolygon[XYM] with 1 polygon consisting of 1 total ring and 5 total points",
		},
		{
			name: "XYMZ 1 square polygon",
			multiPolyPointSequences: [][][]float64{
				{
					{-1, 1, 0.5, 0.8, 1, 1, 0.5, 0.8, 1, -1, 0.5, 0.8, -1, -1, 0.5, 0.8, -1, 1, 0.5, 0.8},
				},
			},
			coordsType:  geom.DimXYZM,
			wantSummary: "MultiPolygon[XYZM] with 1 polygon consisting of 1 total ring and 5 total points",
		},

		// Multiple basic polygon without inner rings.
		{
			name: "XY 2 square polygons",
			multiPolyPointSequences: [][][]float64{
				{
					{-1, 1, 1, 1, 1, -1, -1, -1, -1, 1},
				},
				{
					{9, 11, 11, 11, 11, 9, 9, 9, 9, 11},
				},
			},
			coordsType:  geom.DimXY,
			wantSummary: "MultiPolygon[XY] with 2 polygons consisting of 2 total rings and 10 total points",
		},
		{
			name: "XYZ 2 square polygons",
			multiPolyPointSequences: [][][]float64{
				{
					{-1, 1, 0.5, 1, 1, 0.5, 1, -1, 0.5, -1, -1, 0.5, -1, 1, 0.5},
				},
				{
					{9, 11, 0.5, 11, 11, 0.5, 11, 9, 0.5, 9, 9, 0.5, 9, 11, 0.5},
				},
			},
			coordsType:  geom.DimXYZ,
			wantSummary: "MultiPolygon[XYZ] with 2 polygons consisting of 2 total rings and 10 total points",
		},
		{
			name: "XYM 2 square polygons",
			multiPolyPointSequences: [][][]float64{
				{
					{-1, 1, 0.8, 1, 1, 0.8, 1, -1, 0.8, -1, -1, 0.8, -1, 1, 0.8},
				},
				{
					{9, 11, 0.8, 11, 11, 0.8, 11, 9, 0.8, 9, 9, 0.8, 9, 11, 0.8},
				},
			},
			coordsType:  geom.DimXYM,
			wantSummary: "MultiPolygon[XYM] with 2 polygons consisting of 2 total rings and 10 total points",
		},
		{
			name: "XYMZ 2 square polygons",
			multiPolyPointSequences: [][][]float64{
				{
					{-1, 1, 0.5, 0.8, 1, 1, 0.5, 0.8, 1, -1, 0.5, 0.8, -1, -1, 0.5, 0.8, -1, 1, 0.5, 0.8},
				},
				{
					{9, 11, 0.5, 0.8, 11, 11, 0.5, 0.8, 11, 9, 0.5, 0.8, 9, 9, 0.5, 0.8, 9, 11, 0.5, 0.8},
				},
			},
			coordsType:  geom.DimXYZM,
			wantSummary: "MultiPolygon[XYZM] with 2 polygons consisting of 2 total rings and 10 total points",
		},

		// Single polygons with multiple inner rings.
		{
			name: "XY 2 squares within a square polygon",
			multiPolyPointSequences: [][][]float64{
				{
					{-100, 100, 100, 100, 100, -100, -100, -100, -100, 100},
					{-1, 1, 1, 1, 1, -1, -1, -1, -1, 1},
					{10, 10, 11, 10, 11, 9, 10, 9, 10, 10},
				},
			},
			coordsType:  geom.DimXY,
			wantSummary: "MultiPolygon[XY] with 1 polygon consisting of 3 total rings and 15 total points",
		},
		{
			name: "XYZ 2 squares within a square polygon",
			multiPolyPointSequences: [][][]float64{
				{
					{-100, 100, 0.5, 100, 100, 0.5, 100, -100, 0.5, -100, -100, 0.5, -100, 100, 0.5},
					{-1, 1, 0.5, 1, 1, 0.5, 1, -1, 0.5, -1, -1, 0.5, -1, 1, 0.5},
					{10, 10, 0.5, 11, 10, 0.5, 11, 9, 0.5, 10, 9, 0.5, 10, 10, 0.5},
				},
			},
			coordsType:  geom.DimXYZ,
			wantSummary: "MultiPolygon[XYZ] with 1 polygon consisting of 3 total rings and 15 total points",
		},
		{
			name: "XYM 2 squares within a square polygon",
			multiPolyPointSequences: [][][]float64{
				{
					{-100, 100, 0.8, 100, 100, 0.8, 100, -100, 0.8, -100, -100, 0.8, -100, 100, 0.8},
					{-1, 1, 0.8, 1, 1, 0.8, 1, -1, 0.8, -1, -1, 0.8, -1, 1, 0.8},
					{10, 10, 0.8, 11, 10, 0.8, 11, 9, 0.8, 10, 9, 0.8, 10, 10, 0.8},
				},
			},
			coordsType:  geom.DimXYM,
			wantSummary: "MultiPolygon[XYM] with 1 polygon consisting of 3 total rings and 15 total points",
		},
		{
			name: "XYMZ 2 squares within a square polygon",
			multiPolyPointSequences: [][][]float64{
				{
					{-100, 100, 0.5, 0.8, 100, 100, 0.5, 0.8, 100, -100, 0.5, 0.8, -100, -100, 0.5, 0.8, -100, 100, 0.5, 0.8},
					{-1, 1, 0.5, 0.8, 1, 1, 0.5, 0.8, 1, -1, 0.5, 0.8, -1, -1, 0.5, 0.8, -1, 1, 0.5, 0.8},
					{10, 10, 0.5, 0.8, 11, 10, 0.5, 0.8, 11, 9, 0.5, 0.8, 10, 9, 0.5, 0.8, 10, 10, 0.5, 0.8},
				},
			},
			coordsType:  geom.DimXYZM,
			wantSummary: "MultiPolygon[XYZM] with 1 polygon consisting of 3 total rings and 15 total points",
		},

		// Multiple polygons with multiple inner rings.
		{
			name: "XY 2 squares within each of 2 square polygons",
			multiPolyPointSequences: [][][]float64{
				{
					{-100, 100, 100, 100, 100, -100, -100, -100, -100, 100},
					{-1, 1, 1, 1, 1, -1, -1, -1, -1, 1},
					{10, 10, 11, 10, 11, 9, 10, 9, 10, 10},
				},
				{
					{100, -100, 200, -100, 200, -200, 100, -200, 100, -100},
					{101, -101, 102, -101, 102, -102, 101, -102, 101, -101},
					{110, -110, 111, -110, 111, -111, 110, -111, 110, -110},
				},
			},
			coordsType:  geom.DimXY,
			wantSummary: "MultiPolygon[XY] with 2 polygons consisting of 6 total rings and 30 total points",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			polygons := make([]geom.Polygon, len(tc.multiPolyPointSequences))
			for i, polyPointsSequences := range tc.multiPolyPointSequences {
				lineStrings := make([]geom.LineString, len(polyPointsSequences))
				for j, coords := range polyPointsSequences {
					ls, err := geom.NewLineString(geom.NewSequence(coords, tc.coordsType))
					expectNoErr(t, err)
					lineStrings[j] = ls
				}
				p, err := geom.NewPolygonFromRings(lineStrings)
				expectNoErr(t, err)
				polygons[i] = p
			}

			mp, err := geom.NewMultiPolygonFromPolygons(polygons)
			expectNoErr(t, err)
			expectStringEq(t, mp.Summary(), tc.wantSummary)
			expectStringEq(t, mp.String(), tc.wantSummary)
		})
	}
}
