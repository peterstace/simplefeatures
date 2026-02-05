package geom_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/test"
)

func TestPreparedGeometry(t *testing.T) {
	for i, tt := range []struct {
		wkt1, wkt2 string
	}{
		// Point vs point
		{"POINT(1 2)", "POINT(1 2)"},
		{"POINT(1 2)", "POINT(3 4)"},

		// Point vs linestring
		{"POINT(0.5 0)", "LINESTRING(0 0,1 0)"},
		{"POINT(0 0)", "LINESTRING(0 0,1 0)"},
		{"POINT(0 1)", "LINESTRING(0 0,1 0)"},

		// Point vs polygon
		{"POINT(0.5 0.5)", "POLYGON((0 0,1 0,1 1,0 1,0 0))"},
		{"POINT(0 0)", "POLYGON((0 0,1 0,1 1,0 1,0 0))"},
		{"POINT(5 5)", "POLYGON((0 0,1 0,1 1,0 1,0 0))"},

		// Linestring vs linestring
		{"LINESTRING(0 0,1 1)", "LINESTRING(0 1,1 0)"},
		{"LINESTRING(0 0,1 0)", "LINESTRING(1 0,2 0)"},
		{"LINESTRING(0 0,2 0)", "LINESTRING(1 0,3 0)"},
		{"LINESTRING(0 0,1 0)", "LINESTRING(2 0,3 0)"},

		// Linestring vs polygon
		{"LINESTRING(0 0,2 2)", "POLYGON((0 0,1 0,1 1,0 1,0 0))"},
		{"LINESTRING(0.25 0.25,0.75 0.75)", "POLYGON((0 0,1 0,1 1,0 1,0 0))"},
		{"LINESTRING(5 5,6 6)", "POLYGON((0 0,1 0,1 1,0 1,0 0))"},
		{"LINESTRING(0 0,1 0)", "POLYGON((0 0,1 0,1 1,0 1,0 0))"},

		// Polygon vs polygon
		{"POLYGON((0 0,3 0,3 3,0 3,0 0))", "POLYGON((1 1,2 1,2 2,1 2,1 1))"},
		{"POLYGON((1 1,2 1,2 2,1 2,1 1))", "POLYGON((0 0,3 0,3 3,0 3,0 0))"},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "POLYGON((1 0,3 0,3 2,1 2,1 0))"},
		{"POLYGON((0 0,1 0,1 1,0 1,0 0))", "POLYGON((1 0,2 0,2 1,1 1,1 0))"},
		{"POLYGON((0 0,1 0,1 1,0 1,0 0))", "POLYGON((2 2,3 2,3 3,2 3,2 2))"},
		{"POLYGON((0 0,1 0,1 1,0 1,0 0))", "POLYGON((0 0,1 0,1 1,0 1,0 0))"},

		// Empty geometries
		{"POINT EMPTY", "POINT EMPTY"},
		{"POINT EMPTY", "POINT(1 2)"},
		{"POLYGON((0 0,1 0,1 1,0 1,0 0))", "POINT EMPTY"},

		// Polygon with hole
		{"POLYGON((0 0,10 0,10 10,0 10,0 0),(3 3,7 3,7 7,3 7,3 3))", "POINT(5 5)"},
		{"POLYGON((0 0,10 0,10 10,0 10,0 0),(3 3,7 3,7 7,3 7,3 3))", "POINT(1 1)"},

		// Multi-geometries
		{"MULTIPOINT((0 0),(1 1))", "POINT(0 0)"},
		{"MULTILINESTRING((0 0,1 0),(0 1,1 1))", "POINT(0.5 0)"},
		{"MULTIPOLYGON(((0 0,1 0,1 1,0 1,0 0)),((2 2,3 2,3 3,2 3,2 2)))", "POINT(0.5 0.5)"},

		// Geometry collection as prepared geometry
		{"GEOMETRYCOLLECTION(POINT(0 0),LINESTRING(1 1,2 2),POLYGON((3 3,6 3,6 6,3 6,3 3)))", "POINT(0 0)"},
		{"GEOMETRYCOLLECTION(POINT(0 0),LINESTRING(1 1,2 2),POLYGON((3 3,6 3,6 6,3 6,3 3)))", "POINT(4 4)"},
		{"GEOMETRYCOLLECTION(POINT(0 0),LINESTRING(1 1,2 2),POLYGON((3 3,6 3,6 6,3 6,3 3)))", "POINT(9 9)"},
		{"GEOMETRYCOLLECTION(POINT(0 0),LINESTRING(1 1,2 2),POLYGON((3 3,6 3,6 6,3 6,3 3)))", "LINESTRING(4 4,5 5)"},

		// Geometry collection as test geometry
		{"POLYGON((0 0,10 0,10 10,0 10,0 0))", "GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(2 2,3 3))"},
		{"POLYGON((0 0,10 0,10 10,0 10,0 0))", "GEOMETRYCOLLECTION(POINT(1 1),POINT(20 20))"},

		// Both geometry collections
		{"GEOMETRYCOLLECTION(POINT(0 0),POLYGON((1 1,4 1,4 4,1 4,1 1)))", "GEOMETRYCOLLECTION(POINT(2 2),LINESTRING(2 2,3 3))"},
		{"GEOMETRYCOLLECTION(POINT(0 0),POLYGON((1 1,4 1,4 4,1 4,1 1)))", "GEOMETRYCOLLECTION(POINT(20 20),LINESTRING(20 20,30 30))"},

		// Empty geometry collection
		{"GEOMETRYCOLLECTION EMPTY", "POINT(1 1)"},
		{"POINT(1 1)", "GEOMETRYCOLLECTION EMPTY"},
		{"GEOMETRYCOLLECTION EMPTY", "GEOMETRYCOLLECTION EMPTY"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g1 := test.FromWKT(t, tt.wkt1)
			g2 := test.FromWKT(t, tt.wkt2)

			pg, err := geom.Prepare(g1)
			test.NoErr(t, err)

			type predicate struct {
				name string
				got  func() (bool, error)
				want func() (bool, error)
			}

			predicates := []predicate{
				{
					name: "Intersects",
					got:  func() (bool, error) { return pg.Intersects(g2) },
					want: func() (bool, error) { return geom.Intersects(g1, g2), nil },
				},
				{
					name: "Contains",
					got:  func() (bool, error) { return pg.Contains(g2) },
					want: func() (bool, error) { return geom.Contains(g1, g2) },
				},
				{
					name: "CoveredBy",
					got:  func() (bool, error) { return pg.CoveredBy(g2) },
					want: func() (bool, error) { return geom.CoveredBy(g1, g2) },
				},
				{
					name: "Covers",
					got:  func() (bool, error) { return pg.Covers(g2) },
					want: func() (bool, error) { return geom.Covers(g1, g2) },
				},
				{
					name: "Disjoint",
					got:  func() (bool, error) { return pg.Disjoint(g2) },
					want: func() (bool, error) { return geom.Disjoint(g1, g2) },
				},
				{
					name: "Overlaps",
					got:  func() (bool, error) { return pg.Overlaps(g2) },
					want: func() (bool, error) { return geom.Overlaps(g1, g2) },
				},
				{
					name: "Touches",
					got:  func() (bool, error) { return pg.Touches(g2) },
					want: func() (bool, error) { return geom.Touches(g1, g2) },
				},
				{
					name: "Within",
					got:  func() (bool, error) { return pg.Within(g2) },
					want: func() (bool, error) { return geom.Within(g1, g2) },
				},
			}

			for _, pred := range predicates {
				t.Run(pred.name, func(t *testing.T) {
					got, gotErr := pred.got()
					test.NoErr(t, gotErr)
					want, wantErr := pred.want()
					test.NoErr(t, wantErr)
					test.Eq(t, got, want)
				})
			}
		})
	}
}

func TestPreparedGeometryContainsProperly(t *testing.T) {
	for i, tt := range []struct {
		wkt1, wkt2 string
		want       bool
	}{
		// Point in polygon interior: true
		{
			wkt1: "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			wkt2: "POINT(0.5 0.5)",
			want: true,
		},
		// Point on polygon boundary: false (key distinction from Contains)
		{
			wkt1: "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			wkt2: "POINT(0 0)",
			want: false,
		},
		// Point on polygon edge: false
		{
			wkt1: "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			wkt2: "POINT(0.5 0)",
			want: false,
		},
		// Polygon properly containing another polygon
		{
			wkt1: "POLYGON((0 0,10 0,10 10,0 10,0 0))",
			wkt2: "POLYGON((1 1,2 1,2 2,1 2,1 1))",
			want: true,
		},
		// Equal polygons: false (boundary points shared)
		{
			wkt1: "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			wkt2: "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			want: false,
		},
		// Both empty
		{
			wkt1: "POINT EMPTY",
			wkt2: "POINT EMPTY",
			want: false,
		},
		// One empty
		{
			wkt1: "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			wkt2: "POINT EMPTY",
			want: false,
		},
		// Disjoint
		{
			wkt1: "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			wkt2: "POINT(5 5)",
			want: false,
		},
		// Polygon with shared boundary edge: false
		{
			wkt1: "POLYGON((0 0,2 0,2 2,0 2,0 0))",
			wkt2: "POLYGON((1 0,2 0,2 1,1 1,1 0))",
			want: false,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g1 := test.FromWKT(t, tt.wkt1)
			g2 := test.FromWKT(t, tt.wkt2)

			pg, err := geom.Prepare(g1)
			test.NoErr(t, err)

			got, err := pg.ContainsProperly(g2)
			test.NoErr(t, err)
			test.Eq(t, got, tt.want)
		})
	}
}

func TestPreparedGeometryMultipleEvaluations(t *testing.T) {
	pg, err := geom.Prepare(test.FromWKT(t, "POLYGON((0 0,10 0,10 10,0 10,0 0))"))
	test.NoErr(t, err)

	tests := []struct {
		wkt  string
		want bool
	}{
		{"POINT(5 5)", true},
		{"POINT(15 15)", false},
		{"LINESTRING(1 1,2 2)", true},
		{"LINESTRING(11 11,12 12)", false},
		{"POLYGON((1 1,2 1,2 2,1 2,1 1))", true},
		{"POLYGON((20 20,21 20,21 21,20 21,20 20))", false},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := test.FromWKT(t, tt.wkt)
			got, err := pg.Intersects(g)
			test.NoErr(t, err)
			test.Eq(t, got, tt.want)
		})
	}
}
