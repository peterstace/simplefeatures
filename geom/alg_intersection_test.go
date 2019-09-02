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
		{"POINT(3 3)", "LINESTRING(0 0,2 2)", "POINT EMPTY"},
		{"POINT(-1 -1)", "LINESTRING(0 0,2 2)", "POINT EMPTY"},
		{"POINT(0 2)", "LINESTRING(0 0,2 2)", "POINT EMPTY"},
		{"POINT(2 0)", "LINESTRING(0 0,2 2)", "POINT EMPTY"},
		{"POINT(0 3.14)", "LINESTRING(0 0,0 4)", "POINT(0 3.14)"},
		{"POINT(1 0.25)", "LINESTRING(0 0,4 1)", "POINT(1 0.25)"},
		{"POINT(2 0.5)", "LINESTRING(0 0,4 1)", "POINT(2 0.5)"},

		// Point/LineString
		{"POINT(0 0)", "LINESTRING(1 0,2 1,3 0)", "POINT EMPTY"},
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
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(1 0)", "MULTIPOINT EMPTY"},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(1 0,0 1)", "MULTIPOINT EMPTY"},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(0.5 0.5)", "POINT(0.5 0.5)"},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(0 0)", "POINT(0 0)"},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(0.5 0.5,1 0)", "POINT(0.5 0.5)"},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(1 1,0 1)", "POINT(1 1)"},

		// LineString/LineString -- most test cases covered by LR/LR
		{"LINESTRING(0 0,1 0,1 1,0 1)", "LINESTRING(1 1,2 1,2 2,1 2)", "POINT(1 1)"},

		// LineString/LinearRing -- most test cases covered by LR/LR
		{"LINESTRING(0 0,1 0,1 1,0 1)", "LINEARRING(1 1,2 1,2 2,1 2,1 1)", "POINT(1 1)"},

		// LinearRing/LinearRing
		{"LINEARRING(0 0,1 0,1 1,0 1,0 0)", "LINEARRING(2 2,3 2,3 3,2 3,2 2)", "GEOMETRYCOLLECTION EMPTY"},
		{"LINEARRING(0 0,1 0,1 1,0 1,0 0)", "LINEARRING(1 1,2 1,2 2,1 2,1 1)", "POINT(1 1)"},
		{"LINEARRING(0 0,1 0,1 1,0 1,0 0)", "LINEARRING(1 0,2 0,2 1,1 1,1 0)", "LINESTRING(1 0,1 1)"},
		{"LINEARRING(0 0,1 0,0 1,0 0)", "LINEARRING(1 0,1 1,0 1,1 0)", "LINESTRING(0 1,1 0)"},
		{"LINEARRING(0 0,1 0,1 1,0 1,0 0)", "LINEARRING(0.5 0.5,1.5 0.5,1.5 1.5,0.5 1.5,0.5 0.5)", "MULTIPOINT((0.5 1),(1 0.5))"},
		{"LINEARRING(0 0,1 0,1 1,0 1,0 0)", "LINEARRING(1 0,2 0,2 1,1 1,1.5 0.5,1 0.5,1 0)", "GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(1 0,1 0.5))"},

		// MultiPoint/MultiPoint
		{"MULTIPOINT EMPTY", "MULTIPOINT EMPTY", "MULTIPOINT EMPTY"},
		{"MULTIPOINT EMPTY", "MULTIPOINT((1 2))", "MULTIPOINT EMPTY"},
		{"MULTIPOINT((1 2))", "MULTIPOINT((1 2))", "POINT(1 2)"},
		{"MULTIPOINT((1 2))", "MULTIPOINT((1 2),(1 2))", "POINT(1 2)"},
		{"MULTIPOINT((1 2))", "MULTIPOINT((1 2),(3 4))", "POINT(1 2)"},
		{"MULTIPOINT((3 4),(1 2))", "MULTIPOINT((1 2),(3 4))", "MULTIPOINT((1 2),(3 4))"},
		{"MULTIPOINT((3 4),(1 2))", "MULTIPOINT((1 4),(2 2))", "MULTIPOINT EMPTY"},

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
		{"MULTILINESTRING((0 0,1 0,1 1,0 1))", "LINEARRING(1 1,2 1,2 2,1 2,1 1)", "POINT(1 1)"},
		{"MULTILINESTRING((0 0,1 0,1 1,0 1))", "MULTILINESTRING((1 1,2 1,2 2,1 2,1 1))", "POINT(1 1)"},
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
				got := in1g.Intersection(in2g)
				if !got.EqualsExact(geomFromWKT(t, tt.out), IgnoreOrder) {
					t.Errorf("\ninput1: %s\ninput2: %s\nwant:   %v\ngot:    %v", tt.in1, tt.in2, tt.out, got.AsText())
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
				got := in2g.Intersection(in1g)
				if !got.EqualsExact(geomFromWKT(t, tt.out), IgnoreOrder) {
					t.Errorf("\ninput1: %s\ninput2: %s\nwant:   %v\ngot:    %v", tt.in2, tt.in1, tt.out, got.AsText())
				}
			})
		})
	}
}
