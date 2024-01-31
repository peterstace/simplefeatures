package geom_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestDensifyEmpty(t *testing.T) {
	for _, empty := range []geom.Geometry{
		geom.Point{}.AsGeometry(),
		geom.LineString{}.AsGeometry(),
		geom.Polygon{}.AsGeometry(),
		geom.MultiPoint{}.AsGeometry(),
		geom.MultiLineString{}.AsGeometry(),
		geom.MultiPolygon{}.AsGeometry(),
		geom.GeometryCollection{}.AsGeometry(),
	} {
		t.Run(empty.String(), func(t *testing.T) {
			for _, ct := range []geom.CoordinatesType{
				geom.DimXY,
				geom.DimXYZ,
				geom.DimXYM,
				geom.DimXYZM,
			} {
				t.Run(ct.String(), func(t *testing.T) {
					input := empty.ForceCoordinatesType(ct)
					got := input.Densify(1.0)
					expectGeomEq(t, got, input)
				})
			}
		})
	}
}

func TestDensify(t *testing.T) {
	for i, tc := range []struct {
		input   string
		maxDist float64
		want    string
	}{
		// LineString with a single segment (tests threshold logic):
		{"LINESTRING(0 0,1 0)", 2.0, "LINESTRING(0 0,1 0)"},
		{"LINESTRING(0 0,1 0)", 1.0, "LINESTRING(0 0,1 0)"},
		{"LINESTRING(0 0,1 0)", 0.9, "LINESTRING(0 0,0.5 0,1 0)"},
		{"LINESTRING(0 0,1 0)", 0.5, "LINESTRING(0 0,0.5 0,1 0)"},
		{"LINESTRING(0 0,1 0)", 0.4, "LINESTRING(0 0,0.3333333333333333 0,0.6666666666666666 0,1 0)"},
		{"LINESTRING(0 0,1 0)", 0.3, "LINESTRING(0 0,0.25 0,0.5 0,0.75 0,1 0)"},
		{"LINESTRING(0 0,1 0)", 0.25, "LINESTRING(0 0,0.25 0,0.5 0,0.75 0,1 0)"},

		// LineString with Z/M/ZM:
		{"LINESTRING(1 2,3 4,5 6)", 1.5, "LINESTRING(1 2,2 3,3 4,4 5,5 6)"},
		{"LINESTRING M(1 2 10,3 4 11,5 6 12)", 1.5, "LINESTRING M(1 2 10,2 3 10.5,3 4 11,4 5 11.5,5 6 12)"},
		{"LINESTRING Z(1 2 10,3 4 11,5 6 12)", 1.5, "LINESTRING Z(1 2 10,2 3 10.5,3 4 11,4 5 11.5,5 6 12)"},
		{"LINESTRING ZM(1 2 10 20,3 4 11 21,5 6 12 22)", 1.5, "LINESTRING ZM(1 2 10 20,2 3 10.5 20.5,3 4 11 21,4 5 11.5 21.5,5 6 12 22)"},

		// LineString where each segment is broken into a different number of parts:
		{"LINESTRING(0 0,2 0,2 1)", 0.5, "LINESTRING(0 0,0.5 0,1 0,1.5 0,2 0,2 0.5,2 1)"},

		// Other geometry types:
		{"POINT(0 0)", 1.0, "POINT(0 0)"},
		{"MULTIPOINT((0 0),(1 1))", 1.0, "MULTIPOINT((0 0),(1 1))"},
		{"MULTILINESTRING((0 0,1 1),(2 2,3 3))", 1.0, "MULTILINESTRING((0 0,0.5 0.5,1 1),(2 2,2.5 2.5,3 3))"},
		{"POLYGON((0 0,0 1,1 0,0 0))", 1.0, "POLYGON((0 0,0 1,0.5 0.5,1 0,0 0))"},
		{"MULTIPOLYGON(((0 0,0 1,1 0,0 0)))", 1.0, "MULTIPOLYGON(((0 0,0 1,0.5 0.5,1 0,0 0)))"},
		{"GEOMETRYCOLLECTION(POINT(0 0),LINESTRING(0 0,1 1))", 1.0, "GEOMETRYCOLLECTION(POINT(0 0),LINESTRING(0 0,0.5 0.5,1 1))"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			input := geomFromWKT(t, tc.input)
			got := input.Densify(tc.maxDist)
			expectGeomEqWKT(t, got, tc.want)
		})
	}
}

func TestDensifyInvalidMaxDist(t *testing.T) {
	for i, tc := range []struct {
		input   string
		maxDist float64
	}{
		{"LINESTRING(0 0,1 0)", -1},
		{"LINESTRING(0 0,1 0)", 0},
		{"POINT(0 0)", -1},
		{"POINT(0 0)", 0},
		{"MULTIPOINT((0 0))", -1},
		{"MULTIPOINT((0 0))", 0},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			input := geomFromWKT(t, tc.input)
			expectPanics(t, func() { input.Densify(tc.maxDist) })
		})
	}
}
