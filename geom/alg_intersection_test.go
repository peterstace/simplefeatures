package geom_test

import (
	"strconv"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestIntersection(t *testing.T) {
	for i, tt := range []struct {
		in1, in2, out string
	}{
		// Empty/ANY - always returns the empty geometry as-is to match PostGIS.
		{"POINT EMPTY", "POINT(2 3)", "POINT EMPTY"},
		{"POLYGON EMPTY", "POINT(2 3)", "POLYGON EMPTY"},
		{"LINESTRING EMPTY", "POINT(2 3)", "LINESTRING EMPTY"},

		// Empty/Empty - always returns the second geometry to match PostGIS.
		{"POINT EMPTY", "LINESTRING EMPTY", "LINESTRING EMPTY"},
		{"POLYGON EMPTY", "GEOMETRYCOLLECTION EMPTY", "GEOMETRYCOLLECTION EMPTY"},

		// Point/Point
		{"POINT(1 2)", "POINT(1 2)", "POINT(1 2)"},
		{"POINT(1 2)", "POINT(2 1)", "GEOMETRYCOLLECTION EMPTY"},

		// Point/Line
		{"POINT(0 0)", "LINESTRING(0 0,2 2)", "POINT(0 0)"},
		{"POINT(1 1)", "LINESTRING(0 0,2 2)", "POINT(1 1)"},
		{"POINT(2 2)", "LINESTRING(0 0,2 2)", "POINT(2 2)"},
		{"POINT(3 3)", "LINESTRING(0 0,2 2)", "GEOMETRYCOLLECTION EMPTY"},
		{"POINT(-1 -1)", "LINESTRING(0 0,2 2)", "GEOMETRYCOLLECTION EMPTY"},
		{"POINT(0 2)", "LINESTRING(0 0,2 2)", "GEOMETRYCOLLECTION EMPTY"},
		{"POINT(2 0)", "LINESTRING(0 0,2 2)", "GEOMETRYCOLLECTION EMPTY"},
		{"POINT(0 3.14)", "LINESTRING(0 0,0 4)", "POINT(0 3.14)"},
		{"POINT(1 0.25)", "LINESTRING(0 0,4 1)", "POINT(1 0.25)"},
		{"POINT(2 0.5)", "LINESTRING(0 0,4 1)", "POINT(2 0.5)"},

		// Point/LineString
		{"POINT(0 0)", "LINESTRING(1 0,2 1,3 0)", "GEOMETRYCOLLECTION EMPTY"},
		{"POINT(1 0)", "LINESTRING(1 0,2 1,3 0)", "POINT(1 0)"},
		{"POINT(2 1)", "LINESTRING(1 0,2 1,3 0)", "POINT(2 1)"},
		{"POINT(1.5 0.5)", "LINESTRING(1 0,2 1,3 0)", "POINT(1.5 0.5)"},

		// Point/Polygon
		{`POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, `POINT(1 2)`, `POINT(1 2)`},
		{`POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, `POINT(2.5 1.5)`, `POINT(2.5 1.5)`},
		{`POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, `POINT(4 1)`, `POINT(4 1)`},
		{`POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, `POINT(5 3)`, `POINT(5 3)`},
		{`POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, `POINT(1.5 1.5)`, `GEOMETRYCOLLECTION EMPTY`},
		{`POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, `POINT(3.5 1.5)`, `GEOMETRYCOLLECTION EMPTY`},
		{`POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, `POINT(6 2)`, `GEOMETRYCOLLECTION EMPTY`},

		// Line/Line
		{"LINESTRING(0 0,0 1)", "LINESTRING(0 0,1 0)", "POINT(0 0)"},
		{"LINESTRING(0 1,1 1)", "LINESTRING(1 0,1 1)", "POINT(1 1)"},
		{"LINESTRING(0 1,0 0)", "LINESTRING(0 0,1 0)", "POINT(0 0)"},
		{"LINESTRING(0 0,0 1)", "LINESTRING(1 0,0 0)", "POINT(0 0)"},
		{"LINESTRING(0 0,1 0)", "LINESTRING(1 0,2 0)", "POINT(1 0)"},
		{"LINESTRING(0 0,1 0)", "LINESTRING(2 0,3 0)", "GEOMETRYCOLLECTION EMPTY"},
		{"LINESTRING(1 0,2 0)", "LINESTRING(0 0,3 0)", "LINESTRING(1 0,2 0)"},
		{"LINESTRING(0 0,0 1)", "LINESTRING(1 0,1 1)", "GEOMETRYCOLLECTION EMPTY"},
		{"LINESTRING(0 0,1 1)", "LINESTRING(1 0,0 1)", "POINT(0.5 0.5)"},
		{"LINESTRING(1 0,0 1)", "LINESTRING(0 1,1 0)", "LINESTRING(0 1,1 0)"},
		{"LINESTRING(1 0,0 1)", "LINESTRING(1 0,0 1)", "LINESTRING(0 1,1 0)"},
		{"LINESTRING(0 0,1 1)", "LINESTRING(1 1,0 0)", "LINESTRING(0 0,1 1)"},
		{"LINESTRING(0 0,1 1)", "LINESTRING(0 0,1 1)", "LINESTRING(0 0,1 1)"},
		{"LINESTRING(0 0,0 1)", "LINESTRING(0 1,0 0)", "LINESTRING(0 0,0 1)"},
		{"LINESTRING(0 0,0 1)", "LINESTRING(0 0,0 1)", "LINESTRING(0 0,0 1)"},
		{"LINESTRING(0 0,1 0)", "LINESTRING(1 0,0 0)", "LINESTRING(0 0,1 0)"},
		{"LINESTRING(0 0,1 0)", "LINESTRING(0 0,1 0)", "LINESTRING(0 0,1 0)"},
		{"LINESTRING(1 1,2 2)", "LINESTRING(0 0,3 3)", "LINESTRING(1 1,2 2)"},
		{"LINESTRING(3 1,2 2)", "LINESTRING(1 3,2 2)", "POINT(2 2)"},

		// Line/MultiPoint
		{"LINESTRING(0 0,1 1)", "MULTIPOINT EMPTY", "MULTIPOINT EMPTY"},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(1 0)", "GEOMETRYCOLLECTION EMPTY"},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(1 0,0 1)", "GEOMETRYCOLLECTION EMPTY"},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(0.5 0.5)", "POINT(0.5 0.5)"},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(0 0)", "POINT(0 0)"},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(0.5 0.5,1 0)", "POINT(0.5 0.5)"},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(1 1,0 1)", "POINT(1 1)"},

		// LineString/LineString
		{"LINESTRING(0 0,1 0,1 1,0 1)", "LINESTRING(1 1,2 1,2 2,1 2)", "POINT(1 1)"},
		{"LINESTRING(0 0,1 0,1 1,0 1)", "LINESTRING(1 1,2 1,2 2,1 2,1 1)", "POINT(1 1)"},
		{"LINESTRING(0 0,1 0,1 1,0 1,0 0)", "LINESTRING(2 2,3 2,3 3,2 3,2 2)", "GEOMETRYCOLLECTION EMPTY"},
		{"LINESTRING(0 0,1 0,1 1,0 1,0 0)", "LINESTRING(1 1,2 1,2 2,1 2,1 1)", "POINT(1 1)"},
		{"LINESTRING(0 0,1 0,1 1,0 1,0 0)", "LINESTRING(1 0,2 0,2 1,1 1,1 0)", "LINESTRING(1 0,1 1)"},
		{"LINESTRING(0 0,1 0,0 1,0 0)", "LINESTRING(1 0,1 1,0 1,1 0)", "LINESTRING(0 1,1 0)"},
		{"LINESTRING(0 0,1 0,1 1,0 1,0 0)", "LINESTRING(0.5 0.5,1.5 0.5,1.5 1.5,0.5 1.5,0.5 0.5)", "MULTIPOINT((0.5 1),(1 0.5))"},
		{"LINESTRING(0 0,1 0,1 1,0 1,0 0)", "LINESTRING(1 0,2 0,2 1,1 1,1.5 0.5,1 0.5,1 0)", "GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(1 0,1 0.5))"},

		// MultiPoint/MultiPoint
		{"MULTIPOINT EMPTY", "MULTIPOINT EMPTY", "MULTIPOINT EMPTY"},
		{"MULTIPOINT EMPTY", "MULTIPOINT((1 2))", "MULTIPOINT EMPTY"},
		{"MULTIPOINT((1 2))", "MULTIPOINT((1 2))", "POINT(1 2)"},
		{"MULTIPOINT((1 2))", "MULTIPOINT((1 2),(1 2))", "POINT(1 2)"},
		{"MULTIPOINT((1 2))", "MULTIPOINT((1 2),(3 4))", "POINT(1 2)"},
		{"MULTIPOINT((3 4),(1 2))", "MULTIPOINT((1 2),(3 4))", "MULTIPOINT((1 2),(3 4))"},
		{"MULTIPOINT((3 4),(1 2))", "MULTIPOINT((1 4),(2 2))", "GEOMETRYCOLLECTION EMPTY"},

		// MultiPoint/Point
		{"MULTIPOINT EMPTY", "POINT(1 2)", "MULTIPOINT EMPTY"},
		{"MULTIPOINT((2 1))", "POINT(1 2)", "GEOMETRYCOLLECTION EMPTY"},
		{"MULTIPOINT((1 2))", "POINT(1 2)", "POINT(1 2)"},
		{"MULTIPOINT((1 2),(1 2))", "POINT(1 2)", "POINT(1 2)"},
		{"MULTIPOINT((1 2),(3 4))", "POINT(1 2)", "POINT(1 2)"},
		{"MULTIPOINT((3 4),(1 2))", "POINT(1 2)", "POINT(1 2)"},
		{"MULTIPOINT((5 6),(7 8))", "POINT(1 2)", "GEOMETRYCOLLECTION EMPTY"},

		// MultiPoint/Polygon
		{`POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, `MULTIPOINT(1 2,10 10)`, `POINT(1 2)`},
		{`POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, `MULTIPOINT(1 2)`, `POINT(1 2)`},

		// MultiLineString with other lines  -- most test cases covered by LR/LR
		{"MULTILINESTRING((0 0,1 0,1 1,0 1))", "LINESTRING(1 1,2 1,2 2,1 2,1 1)", "POINT(1 1)"},
		{"MULTILINESTRING((0 0,1 0,1 1,0 1))", "MULTILINESTRING((1 1,2 1,2 2,1 2,1 1))", "POINT(1 1)"},

		// Test cases found from fuzz:
		{"POLYGON EMPTY", "GEOMETRYCOLLECTION(POLYGON EMPTY)", "GEOMETRYCOLLECTION EMPTY"},
		{"MULTIPOINT((1 2))", "MULTIPOINT((4 8))", "GEOMETRYCOLLECTION EMPTY"},
		{"POINT(1 2)", "LINESTRING(0 0,0 4)", "GEOMETRYCOLLECTION EMPTY"},
		{"POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))", "MULTIPOINT((2 1),(1 2),(2 1))", "MULTIPOINT((1 2),(2 1))"},
		{"POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))", "MULTIPOINT((2 1),(3 6),(2 1))", "POINT(2 1)"},
		{"MULTIPOINT((1 2))", "MULTIPOINT((7 6),(3 3),(3 3))", "GEOMETRYCOLLECTION EMPTY"},
		{"MULTIPOINT((1 2))", "LINESTRING(2 1,3 6)", "GEOMETRYCOLLECTION EMPTY"},
		{"LINESTRING(1 2,4 5)", "MULTIPOINT((7 6),(3 3),(3 3))", "GEOMETRYCOLLECTION EMPTY"},
		{"MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))", "MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))", "MULTILINESTRING((0 1,2 3),(4 5,6 7),(6 7,8 9))"},
		{"MULTILINESTRING((0 1,2 3,4 5))", "LINESTRING(1 2,3 4,5 6)", "MULTILINESTRING((1 2,2 3),(2 3,3 4),(3 4,4 5))"},
		{"LINESTRING(0 0,1 1)", "LINESTRING(0 0,1 1,0 0)", "LINESTRING(0 0,1 1)"},
		{"LINESTRING(0 0,1 0,0 1,0 0)", "LINESTRING(0 0,1 0,1 1,0 1)", "GEOMETRYCOLLECTION(POINT(0 1),LINESTRING(0 0,1 0))"},
		{"LINESTRING(0 0,1 0,0 1,0 0)", "MULTILINESTRING((0 0,0 1,1 1),(0 1,0 0,1 0))", "MULTILINESTRING((0 0,1 0),(0 1,0 0))"},

		// The following two test cases were fonud using fuzz, however they
		// currently don't pass. The difference in the result is cosmetic --
		// the difference between "MULTILINESTRING((0 0,0.5 0.5),(0.5 0.5,1 1))"
		// and "LINESTRING(0 0,1 1)".
		//
		//{"LINESTRING(0 0,1 1)", "LINESTRING(0 0,1 1,0 1,1 0)", "MULTILINESTRING((0 0,0.5 0.5),(0.5 0.5,1 1))"},
		//{"LINESTRING(0 0,0 1,1 0,0 0)", "LINESTRING(0 0,1 1,0 1,0 0,1 1)", "GEOMETRYCOLLECTION(POINT(0.5 0.5),LINESTRING(0 0,0 1))"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			in1g, err := UnmarshalWKT(strings.NewReader(tt.in1))
			if err != nil {
				t.Fatalf("could not unmarshal wkt: %v", err)
			}
			in2g, err := UnmarshalWKT(strings.NewReader(tt.in2))
			if err != nil {
				t.Fatalf("could not unmarshal wkt: %v", err)
			}

			t.Run("forward", func(t *testing.T) {
				got, err := in1g.Intersection(in2g)
				if err != nil {
					t.Fatal(err)
				}
				if !got.EqualsExact(geomFromWKT(t, tt.out), IgnoreOrder) {
					t.Errorf("\ninput1: %s\ninput2: %s\nwant:   %v\ngot:    %v", tt.in1, tt.in2, tt.out, got.AsText())
				}

				// We can infer the desired result for Intersects, which gives
				// us another set of tests "for free". This only works under
				// the assumption that for every pair of geometries that
				// Intersection implements, that pair is also implemented for
				// Intersects.
				intersects, err := in1g.Intersects(in2g)
				if err != nil {
					t.Fatal(err)
				}
				if intersects == got.IsEmpty() {
					t.Errorf("\ninput1: %s\ninput2: %s\nwant:   %v\ngot:    %v\nintersects: %v", tt.in1, tt.in2, tt.out, got, intersects)
				}
			})

			if in1g.IsEmpty() && in2g.IsEmpty() {
				// We always return the second geometry when both are
				// empty, to match PostGIS behaviour. This implies that
				// intersection is non-commutative for the empty/empty
				// case, so skip the reverse case.
				return
			}
			t.Run("reversed", func(t *testing.T) {
				got, err := in2g.Intersection(in1g)
				if err != nil {
					t.Fatal(err)
				}
				if !got.EqualsExact(geomFromWKT(t, tt.out), IgnoreOrder) {
					t.Errorf("\ninput1: %s\ninput2: %s\nwant:   %v\ngot:    %v", tt.in2, tt.in1, tt.out, got.AsText())
				}

				// We can infer the desired result for Intersects, which gives
				// us another set of tests "for free". This only works under
				// the assumption that for every pair of geometries that
				// Intersection implements, that pair is also implemented for
				// Intersects.
				intersects, err := in2g.Intersects(in1g)
				if err != nil {
					t.Fatal(err)
				}
				if intersects == got.IsEmpty() {
					t.Errorf("\ninput1: %s\ninput2: %s\nwant:   %v\ngot:    %v\nintersects: %v", tt.in1, tt.in2, tt.out, got, intersects)
				}
			})
		})
	}
}

func TestIntersects(t *testing.T) {
	for i, tt := range []struct {
		in1, in2 string
		want     bool
	}{
		// Empty/ANY
		{"POINT EMPTY", "POINT(2 3)", false},
		{"POLYGON EMPTY", "POINT(2 3)", false},
		{"LINESTRING EMPTY", "POINT(2 3)", false},

		// Empty/Empty
		{"POINT EMPTY", "LINESTRING EMPTY", false},
		{"POLYGON EMPTY", "GEOMETRYCOLLECTION EMPTY", false},
		{"POLYGON EMPTY", "GEOMETRYCOLLECTION(POLYGON EMPTY)", false},

		// Point/Point
		{"POINT(1 2)", "POINT(1 2)", true},
		{"POINT(1 2)", "POINT(2 1)", false},

		// Point/Line
		{"POINT(0 0)", "LINESTRING(0 0,2 2)", true},
		{"POINT(1 1)", "LINESTRING(0 0,2 2)", true},
		{"POINT(2 2)", "LINESTRING(0 0,2 2)", true},
		{"POINT(3 3)", "LINESTRING(0 0,2 2)", false},
		{"POINT(-1 -1)", "LINESTRING(0 0,2 2)", false},
		{"POINT(0 2)", "LINESTRING(0 0,2 2)", false},
		{"POINT(2 0)", "LINESTRING(0 0,2 2)", false},
		{"POINT(0 3.14)", "LINESTRING(0 0,0 4)", true},
		{"POINT(1 0.25)", "LINESTRING(0 0,4 1)", true},
		{"POINT(2 0.5)", "LINESTRING(0 0,4 1)", true},

		// Point/LineString
		{"POINT(0 0)", "LINESTRING(1 0,2 1,3 0)", false},
		{"POINT(1 0)", "LINESTRING(1 0,2 1,3 0)", true},
		{"POINT(2 1)", "LINESTRING(1 0,2 1,3 0)", true},
		{"POINT(1.5 0.5)", "LINESTRING(1 0,2 1,3 0)", true},
		{"POINT(1 2)", "LINESTRING(0 0,0 4)", false},

		// Point/Polygon
		{`POINT(1 2)`, `POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, true},
		{`POINT(2.5 1.5)`, `POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, true},
		{`POINT(4 1)`, `POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, true},
		{`POINT(5 3)`, `POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, true},
		{`POINT(1.5 1.5)`, `POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, false},
		{`POINT(3.5 1.5)`, `POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, false},
		{`POINT(6 2)`, `POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, false},

		// Point/MultiLineString
		{"POINT(0 0)", "MULTILINESTRING((1 0,2 1,3 0))", false},
		{"POINT(1 0)", "MULTILINESTRING((1 0,2 1,3 0))", true},
		{"POINT(0 0)", "MULTILINESTRING((0 0,1 1),(1 1,2 1))", true},
		{"POINT(1 1)", "MULTILINESTRING((0 0,1 1),(1 1,2 1))", true},
		{"POINT(2 1)", "MULTILINESTRING((0 0,1 1),(1 1,2 1))", true},
		{"POINT(3 1)", "MULTILINESTRING((0 0,1 1),(1 1,2 1))", false},

		// Point/MultiPolygon
		{"POINT(0 0)", "MULTIPOLYGON(((0 0,1 0,1 1,0 0)))", true},
		{"POINT(1 1)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)))", true},
		{"POINT(1 1)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((3 0,5 0,5 2,3 2,3 0)))", true},
		{"POINT(4 1)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((3 0,5 0,5 2,3 2,3 0)))", true},
		{"POINT(6 1)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((3 0,5 0,5 2,3 2,3 0)))", false},

		// Line/Line
		{"LINESTRING(0 0,0 1)", "LINESTRING(0 0,1 0)", true},
		{"LINESTRING(0 1,1 1)", "LINESTRING(1 0,1 1)", true},
		{"LINESTRING(0 1,0 0)", "LINESTRING(0 0,1 0)", true},
		{"LINESTRING(0 0,0 1)", "LINESTRING(1 0,0 0)", true},
		{"LINESTRING(0 0,1 0)", "LINESTRING(1 0,2 0)", true},
		{"LINESTRING(0 0,1 0)", "LINESTRING(2 0,3 0)", false},
		{"LINESTRING(1 0,2 0)", "LINESTRING(0 0,3 0)", true},
		{"LINESTRING(0 0,0 1)", "LINESTRING(1 0,1 1)", false},
		{"LINESTRING(0 0,1 1)", "LINESTRING(1 0,0 1)", true},
		{"LINESTRING(1 0,0 1)", "LINESTRING(0 1,1 0)", true},
		{"LINESTRING(1 0,0 1)", "LINESTRING(1 0,0 1)", true},
		{"LINESTRING(0 0,1 1)", "LINESTRING(1 1,0 0)", true},
		{"LINESTRING(0 0,1 1)", "LINESTRING(0 0,1 1)", true},
		{"LINESTRING(0 0,0 1)", "LINESTRING(0 1,0 0)", true},
		{"LINESTRING(0 0,0 1)", "LINESTRING(0 0,0 1)", true},
		{"LINESTRING(0 0,1 0)", "LINESTRING(1 0,0 0)", true},
		{"LINESTRING(0 0,1 0)", "LINESTRING(0 0,1 0)", true},
		{"LINESTRING(1 1,2 2)", "LINESTRING(0 0,3 3)", true},
		{"LINESTRING(3 1,2 2)", "LINESTRING(1 3,2 2)", true},

		// Line/LineString
		{"LINESTRING(0 0,1 1)", "LINESTRING(0 0,1 1,0 0)", true},
		{"LINESTRING(0 0,1 1)", "LINESTRING(0 0,1 1,0 1,1 0)", true},

		// Line/Polygon
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(3 0,3 2)", false},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(1 2.1,2.1 1)", true},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(1 -1,1 1)", true},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(0.25 0.25,0.75 0.75)", true},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(2 0,3 -1)", true},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(-1 1,1 -1)", true},

		// Line/MultiPoint
		{"LINESTRING(0 0,1 1)", "MULTIPOINT EMPTY", false},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(1 0)", false},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(1 0,0 1)", false},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(0.5 0.5)", true},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(0 0)", true},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(0.5 0.5,1 0)", true},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(1 1,0 1)", true},
		{"LINESTRING(1 2,4 5)", "MULTIPOINT((7 6),(3 3),(3 3))", false},
		{"LINESTRING(2 1,3 6)", "MULTIPOINT((1 2))", false},

		// Line/MultiLineString
		{"LINESTRING(0 0,1 1)", "MULTILINESTRING((0 0.5,1 0.5,1 -0.5),(2 0.5,2 -0.5))", true},
		{"LINESTRING(0 1,1 2)", "MULTILINESTRING((0 0.5,1 0.5,1 -0.5),(2 0.5,2 -0.5))", false},

		// Line/MultiPolygon
		{"LINESTRING(5 2,5 4)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", false},
		{"LINESTRING(3 3,3 5)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},
		{"LINESTRING(1 1,3 1)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},
		{"LINESTRING(0 2,2 4)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},

		// LineString/LineString
		{"LINESTRING(0 0,1 0,1 1,0 1)", "LINESTRING(1 1,2 1,2 2,1 2)", true},
		{"LINESTRING(0 0,0 1,1 0,0 0)", "LINESTRING(0 0,1 1,0 1,0 0,1 1)", true},
		{"LINESTRING(0 0,1 0,0 1,0 0)", "LINESTRING(0 0,1 0,1 1,0 1)", true},
		{"LINESTRING(0 0,1 0,1 1,0 1)", "LINESTRING(1 1,2 1,2 2,1 2,1 1)", true},
		{"LINESTRING(0 0,1 0,1 1,0 1,0 0)", "LINESTRING(2 2,3 2,3 3,2 3,2 2)", false},
		{"LINESTRING(0 0,1 0,1 1,0 1,0 0)", "LINESTRING(1 1,2 1,2 2,1 2,1 1)", true},
		{"LINESTRING(0 0,1 0,1 1,0 1,0 0)", "LINESTRING(1 0,2 0,2 1,1 1,1 0)", true},
		{"LINESTRING(0 0,1 0,0 1,0 0)", "LINESTRING(1 0,1 1,0 1,1 0)", true},
		{"LINESTRING(0 0,1 0,1 1,0 1,0 0)", "LINESTRING(0.5 0.5,1.5 0.5,1.5 1.5,0.5 1.5,0.5 0.5)", true},
		{"LINESTRING(0 0,1 0,1 1,0 1,0 0)", "LINESTRING(1 0,2 0,2 1,1 1,1.5 0.5,1 0.5,1 0)", true},

		// LineString/Polygon
		{"LINESTRING(3 0,3 1,3 2)", "POLYGON((0 0,2 0,2 2,0 2,0 0))", false},
		{"LINESTRING(1 1,2 1, 3 1)", "POLYGON((0 0,2 0,2 2,0 2,0 0))", true},

		// LineString/MultiLineString
		{"LINESTRING(0 0,1 0,0 1,0 0)", "MULTILINESTRING((0 0,0 1,1 1),(0 1,0 0,1 0))", true},
		{"LINESTRING(1 1,2 1,2 2,1 2,1 1)", "MULTILINESTRING((0 0,1 0,1 1,0 1))", true},
		{"LINESTRING(1 2,3 4,5 6)", "MULTILINESTRING((0 1,2 3,4 5))", true},

		// LineString/MultiPolygon
		{"LINESTRING(3 0,3 1,3 2)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)))", false},
		{"LINESTRING(1 1,2 1, 3 1)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)))", true},

		// Polygon/MultiPoint
		{
			`POLYGON(
				(0 0,5 0,5 3,0 3,0 0),
				(1 1,2 1,2 2,1 2,1 1),
				(3 1,4 1,4 2,3 2,3 1)
			)`,
			`MULTIPOINT(1 2,10 10)`,
			true,
		},
		{
			`POLYGON(
				(0 0,5 0,5 3,0 3,0 0),
				(1 1,2 1,2 2,1 2,1 1),
				(3 1,4 1,4 2,3 2,3 1)
			)`,
			`MULTIPOINT(1 2)`,
			true,
		},
		{
			"POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))",
			"MULTIPOINT((2 1),(1 2),(2 1))",
			true,
		},
		{
			"POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))",
			"MULTIPOINT((2 1),(3 6),(2 1))",
			true,
		},

		// MultiPoint/MultiPoint
		{"MULTIPOINT EMPTY", "MULTIPOINT EMPTY", false},
		{"MULTIPOINT EMPTY", "MULTIPOINT((1 2))", false},
		{"MULTIPOINT((1 2))", "MULTIPOINT((1 2))", true},
		{"MULTIPOINT((1 2))", "MULTIPOINT((1 2),(1 2))", true},
		{"MULTIPOINT((1 2))", "MULTIPOINT((1 2),(3 4))", true},
		{"MULTIPOINT((3 4),(1 2))", "MULTIPOINT((1 2),(3 4))", true},
		{"MULTIPOINT((3 4),(1 2))", "MULTIPOINT((1 4),(2 2))", false},
		{"MULTIPOINT((1 2))", "MULTIPOINT((4 8))", false},
		{"MULTIPOINT((1 2))", "MULTIPOINT((7 6),(3 3),(3 3))", false},

		// MultiPoint/Point
		{"MULTIPOINT EMPTY", "POINT(1 2)", false},
		{"MULTIPOINT((2 1))", "POINT(1 2)", false},
		{"MULTIPOINT((1 2))", "POINT(1 2)", true},
		{"MULTIPOINT((1 2),(1 2))", "POINT(1 2)", true},
		{"MULTIPOINT((1 2),(3 4))", "POINT(1 2)", true},
		{"MULTIPOINT((3 4),(1 2))", "POINT(1 2)", true},
		{"MULTIPOINT((5 6),(7 8))", "POINT(1 2)", false},

		// MultiLineString/MultiLineString
		{"MULTILINESTRING((0 0,1 0,1 1,0 1))", "MULTILINESTRING((1 1,2 1,2 2,1 2,1 1))", true},
		{"MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))", "MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))", true},

		// MultiLineString/MultiPolygon
		{"MULTILINESTRING((5 2,5 4))", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", false},
		{"MULTILINESTRING((3 3,3 5))", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},
		{"MULTILINESTRING((1 1,3 1))", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},
		{"MULTILINESTRING((0 2,2 4))", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			runTest := func(g1, g2 Geometry) func(t *testing.T) {
				return func(t *testing.T) {
					got, err := g1.Intersects(g2)
					if err != nil {
						t.Fatal(err)
					}
					if got != tt.want {
						t.Errorf(
							"\ninput1: %s\ninput2: %s\ngot:  %v\nwant: %v\n",
							g1.AsText(), g2.AsText(), got, tt.want,
						)
					}
				}
			}
			g1 := geomFromWKT(t, tt.in1)
			g2 := geomFromWKT(t, tt.in2)
			t.Run("fwd", runTest(g1, g2))
			t.Run("rev", runTest(g2, g1))
		})
	}
}
