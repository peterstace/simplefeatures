package geom_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestGeometryCollectionSummary(t *testing.T) {
	ls, err := geom.NewLineString(geom.NewSequence([]float64{1, 2, 3, 4}, geom.DimXY))
	expectNoErr(t, err)

	for _, tc := range []struct {
		name        string
		geoms       []geom.Geometry
		coordsType  geom.CoordinatesType
		wantSummary string
	}{
		{
			name:        "Empty",
			wantSummary: "GeometryCollection[XY] with 0 child geometries consisting of 0 total points",
		},

		// Single point.
		{
			name: "XY single point",
			geoms: []geom.Geometry{
				geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 0, Y: 0}, Type: geom.DimXY}).AsGeometry(),
			},
			coordsType:  geom.DimXY,
			wantSummary: "GeometryCollection[XY] with 1 child geometry consisting of 1 total point",
		},
		{
			name: "XYZ single point",
			geoms: []geom.Geometry{
				geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 0, Y: 0}, Z: 0.5, Type: geom.DimXYZ}).AsGeometry(),
			},
			coordsType:  geom.DimXYZ,
			wantSummary: "GeometryCollection[XYZ] with 1 child geometry consisting of 1 total point",
		},
		{
			name: "XYM single point",
			geoms: []geom.Geometry{
				geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 0, Y: 0}, M: 0.8, Type: geom.DimXYM}).AsGeometry(),
			},
			coordsType:  geom.DimXYZ,
			wantSummary: "GeometryCollection[XYM] with 1 child geometry consisting of 1 total point",
		},
		{
			name: "XYZM single point",
			geoms: []geom.Geometry{
				geom.NewPoint(geom.Coordinates{XY: geom.XY{X: 0, Y: 0}, Z: 0.5, M: 0.8, Type: geom.DimXYZM}).AsGeometry(),
			},
			coordsType:  geom.DimXYZM,
			wantSummary: "GeometryCollection[XYZM] with 1 child geometry consisting of 1 total point",
		},

		// Multiple geometries and points.
		{
			name:        "XY single line string",
			geoms:       []geom.Geometry{ls.AsGeometry()},
			coordsType:  geom.DimXY,
			wantSummary: "GeometryCollection[XY] with 1 child geometry consisting of 2 total points",
		},
		{
			name:        "XY 2 line strings",
			geoms:       []geom.Geometry{ls.AsGeometry(), ls.AsGeometry()},
			coordsType:  geom.DimXY,
			wantSummary: "GeometryCollection[XY] with 2 child geometries consisting of 4 total points",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := geom.NewGeometryCollection(tc.geoms)
			expectStringEq(t, g.Summary(), tc.wantSummary)
			expectStringEq(t, g.String(), tc.wantSummary)
		})
	}
}
