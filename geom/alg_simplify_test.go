package geom_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestSimplify(t *testing.T) {
	for i, tc := range []struct {
		input     string
		threshold float64
		output    string
	}{
		// Points and MultiPoints pass through unchanged.
		{"POINT(0 1)", 0, "POINT(0 1)"},
		{"POINT(0 1)", 0.5, "POINT(0 1)"},
		{"POINT(0 1)", 1, "POINT(0 1)"},
		{"POINT(0 1)", 2, "POINT(0 1)"},
		{"POINT EMPTY", 0, "POINT EMPTY"},
		{"POINT EMPTY", 0.5, "POINT EMPTY"},
		{"POINT EMPTY", 1, "POINT EMPTY"},
		{"POINT EMPTY", 2, "POINT EMPTY"},
		{"MULTIPOINT(0 0,1 1)", 0.5, "MULTIPOINT(0 0,1 1)"},
		{"MULTIPOINT(0 0,1 1)", 2, "MULTIPOINT(0 0,1 1)"},

		// LineStrings
		{"LINESTRING(0 0,1 1)", 0.0, "LINESTRING(0 0,1 1)"},
		{"LINESTRING(0 0,1 1)", 1.0, "LINESTRING(0 0,1 1)"},
		{"LINESTRING(0 0,1 1)", 2.0, "LINESTRING(0 0,1 1)"},
		{"LINESTRING(0 0,1 1,2 0)", 0.5, "LINESTRING(0 0,1 1,2 0)"},
		{"LINESTRING(0 0,1 1,2 0)", 1.0, "LINESTRING(0 0,2 0)"},
		{"LINESTRING(0 0,1 1,2 0)", 1.5, "LINESTRING(0 0,2 0)"},

		{"LINESTRING(0 0,0 1,1 1,1 0)", 0.5, "LINESTRING(0 0,0 1,1 1,1 0)"},
		{"LINESTRING(0 0,0 1,1 1,1 0)", 0.9, "LINESTRING(0 0,0 1,1 0)"},
		{"LINESTRING(0 0,0 1,1 1,1 0)", 1.0, "LINESTRING(0 0,1 0)"},
		{"LINESTRING(0 0,0 1,1 1,1 0)", 1.5, "LINESTRING(0 0,1 0)"},

		{"LINESTRING(0 0,0 1,1 0,0 0)", 0.5, "LINESTRING(0 0,0 1,1 0,0 0)"},
		{"LINESTRING(0 0,0 1,1 0,0 0)", 1.0, "LINESTRING EMPTY"},
		{"LINESTRING(0 0,0 1,1 0,0 0)", 1.5, "LINESTRING EMPTY"},

		// Polygons
		{"POLYGON EMPTY", 1.0, "POLYGON EMPTY"},
		{"POLYGON((0 0,0 1,1 0,0 0))", 1.0, "POLYGON EMPTY"},
		{"POLYGON((0 0,0 2,2 0,0 0))", 1.0, "POLYGON((0 0,0 2,2 0,0 0))"},
		{"POLYGON((2 2,2 3,3 3,3 2,2 2))", 1.0, "POLYGON EMPTY"},
		{"POLYGON((0 0,0 5,5 5,5 0,0 0),(2 2,2 3,3 3,3 2,2 2)) ", 1.0, "POLYGON((0 0,0 5,5 5,5 0,0 0))"},

		// For MultiLineStrings, each child is treated separately.
		{"MULTILINESTRING((0 0,1 1),(0 0,0 1,1 1,1 0))", 1.5, "MULTILINESTRING((0 0,1 1),(0 0,1 0))"},
		{"MULTILINESTRING((0 0,0 1,1 0,0 0))", 1.5, "MULTILINESTRING EMPTY"},
		{"MULTILINESTRING((0 0,0 1,1 0,0 0),(0 0,0 1,1 1,1 0))", 1.5, "MULTILINESTRING((0 0,1 0))"},
		{"MULTILINESTRING((0 0,0 1,1 1,1 0),(0 0,0 1,1 0,0 0))", 1.5, "MULTILINESTRING((0 0,1 0))"},

		// MultiPolygons
		{"MULTIPOLYGON EMPTY", 1.0, "MULTIPOLYGON EMPTY"},
		{"MULTIPOLYGON(EMPTY)", 1.0, "MULTIPOLYGON EMPTY"},
		{"MULTIPOLYGON(EMPTY,EMPTY)", 1.0, "MULTIPOLYGON EMPTY"},
		{"MULTIPOLYGON(EMPTY,((0 0,0 2,2 2,2 0,0 0)))", 1.0, "MULTIPOLYGON(((0 0,0 2,2 2,2 0,0 0)))"},
		{"MULTIPOLYGON(((0 0,0 2,2 2,2 0,0 0)),EMPTY)", 1.0, "MULTIPOLYGON(((0 0,0 2,2 2,2 0,0 0)))"},
		{"MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)))", 1.0, "MULTIPOLYGON EMPTY"},
		{"MULTIPOLYGON(((0 0,0 2,2 2,2 0,0 0)))", 1.0, "MULTIPOLYGON(((0 0,0 2,2 2,2 0,0 0)))"},

		// GeometryCollections
		{"GEOMETRYCOLLECTION EMPTY", 1.0, "GEOMETRYCOLLECTION EMPTY"},
		{"GEOMETRYCOLLECTION(POLYGON EMPTY)", 1.0, "GEOMETRYCOLLECTION(POLYGON EMPTY)"},
		{"GEOMETRYCOLLECTION(POLYGON((0 0,0 1,1 1,1 0,0 0)))", 1.0, "GEOMETRYCOLLECTION(POLYGON EMPTY)"},
		{
			"GEOMETRYCOLLECTION(POINT(1 2),POLYGON((0 0,0 1,1 1,1 0,0 0)))",
			1.0,
			"GEOMETRYCOLLECTION(POINT(1 2),POLYGON EMPTY)",
		},

		// Z, M, and ZM
		{"POINT Z(0 1 10)", 0, "POINT Z(0 1 10)"},
		{"POINT M(0 1 10)", 0, "POINT M(0 1 10)"},
		{"POINT ZM(0 1 10 11)", 0, "POINT ZM(0 1 10 11)"},
		{"LINESTRING Z(0 0 10,1 1 20)", 0.0, "LINESTRING Z(0 0 10,1 1 20)"},
		{"LINESTRING M(0 0 10,1 1 20)", 0.0, "LINESTRING M(0 0 10,1 1 20)"},
		{"LINESTRING ZM(0 0 10 11,1 1 20 21)", 0.0, "LINESTRING ZM(0 0 10 11,1 1 20 21)"},
		{"LINESTRING Z(0 0 10,1 1 20,2 0 30)", 1.0, "LINESTRING Z(0 0 10,2 0 30)"},
		{"LINESTRING M(0 0 10,1 1 20,2 0 30)", 1.0, "LINESTRING M(0 0 10,2 0 30)"},
		{"LINESTRING ZM(0 0 10 11,1 1 20 21,2 0 30 31)", 1.0, "LINESTRING ZM(0 0 10 11,2 0 30 31)"},

		{"POLYGON Z((2 2 10,2 3 20,3 3 30,3 2 40,2 2 10))", 0.0, "POLYGON Z((2 2 10,2 3 20,3 3 30,3 2 40,2 2 10))"},
		{"POLYGON M((2 2 10,2 3 20,3 3 30,3 2 40,2 2 10))", 0.0, "POLYGON M((2 2 10,2 3 20,3 3 30,3 2 40,2 2 10))"},
		{"POLYGON ZM((2 2 10 11,2 3 20 21,3 3 30 31,3 2 40 41,2 2 10 11))", 0.0, "POLYGON ZM((2 2 10 11,2 3 20 21,3 3 30 31,3 2 40 41,2 2 10 11))"},

		// Empty Z, M, ZM
		{"POINT Z EMPTY", 0.1, "POINT Z EMPTY"},
		{"POINT M EMPTY", 0.1, "POINT M EMPTY"},
		{"POINT ZM EMPTY", 0.1, "POINT ZM EMPTY"},
		{"LINESTRING Z EMPTY", 0.1, "LINESTRING Z EMPTY"},
		{"LINESTRING M EMPTY", 0.1, "LINESTRING M EMPTY"},
		{"LINESTRING ZM EMPTY", 0.1, "LINESTRING ZM EMPTY"},
		{"POLYGON Z EMPTY", 0.1, "POLYGON Z EMPTY"},
		{"POLYGON M EMPTY", 0.1, "POLYGON M EMPTY"},
		{"POLYGON ZM EMPTY", 0.1, "POLYGON ZM EMPTY"},
		{"MULTIPOINT Z EMPTY", 0.1, "MULTIPOINT Z EMPTY"},
		{"MULTIPOINT M EMPTY", 0.1, "MULTIPOINT M EMPTY"},
		{"MULTIPOINT ZM EMPTY", 0.1, "MULTIPOINT ZM EMPTY"},
		{"MULTILINESTRING Z EMPTY", 0.1, "MULTILINESTRING Z EMPTY"},
		{"MULTILINESTRING M EMPTY", 0.1, "MULTILINESTRING M EMPTY"},
		{"MULTILINESTRING ZM EMPTY", 0.1, "MULTILINESTRING ZM EMPTY"},
		{"MULTIPOLYGON Z EMPTY", 0.1, "MULTIPOLYGON Z EMPTY"},
		{"MULTIPOLYGON M EMPTY", 0.1, "MULTIPOLYGON M EMPTY"},
		{"MULTIPOLYGON ZM EMPTY", 0.1, "MULTIPOLYGON ZM EMPTY"},
		{"GEOMETRYCOLLECTION Z EMPTY", 0.1, "GEOMETRYCOLLECTION Z EMPTY"},
		{"GEOMETRYCOLLECTION M EMPTY", 0.1, "GEOMETRYCOLLECTION M EMPTY"},
		{"GEOMETRYCOLLECTION ZM EMPTY", 0.1, "GEOMETRYCOLLECTION ZM EMPTY"},
		{"GEOMETRYCOLLECTION EMPTY", 0.1, "GEOMETRYCOLLECTION EMPTY"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			in := geomFromWKT(t, tc.input)
			want := geomFromWKT(t, tc.output)
			t.Logf("input:     %v", in.AsText())
			t.Logf("threshold: %v", tc.threshold)
			t.Logf("want:      %v", want.AsText())
			got, err := in.Simplify(tc.threshold)
			expectNoErr(t, err)
			t.Logf("got:       %v", got.AsText())
			expectGeomEq(t, got, want, geom.IgnoreOrder)
		})
	}
}

func TestSimplifyErrorCases(t *testing.T) {
	for i, tc := range []struct {
		wkt       string
		threshold float64
	}{
		// Simplification results in the inner and outer rings intersecting.
		{"POLYGON((0 0,0 1,-0.5 1.5,0 2,0 3,3 3,3 0,0 0),(-0.1 1.5,2 2,2 1,-0.1 1.5))", 0.5},

		// Reproduces a bug. The outer ring becomes invalid after simplification.
		{
			`POLYGON((
				151.1897065219023 -33.87468129434335,
				151.191808198953 -33.8734269493667,
				151.19232406823 -33.8738879421183,
				151.19237538770165 -33.873935599348954,
				151.192324067988 -33.8738879424094,
				151.1897065219023 -33.87468129434335
			))`,
			1e-5,
		},

		// Second case for "outer ring becomes invalid after simplification".
		// The outer ring becomes invalid because the Ramer-Douglas-Peucker
		// algorithm causes the ring (when considered as a LineString) to
		// become self-intersected.
		{`POLYGON((0 0,0 1,0.9 1,1 1.1,1.1 1,2 1,2 0,1 1.05,0 0))`, 0.2},

		// Inner ring becomes invalid after simplification.
		{
			`POLYGON(
				(-1 -1,-1 3,3 3,3 -1,-1 -1),
				(0 0,0 1,0.9 1,1 1.1,1.1 1,2 1,2 0,1 1.05,0 0)
			)`,
			0.2,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			in := geomFromWKT(t, tc.wkt)
			_, err := in.Simplify(tc.threshold)
			var want geom.ValidationError
			expectErrAs(t, err, &want)
		})
	}
}
