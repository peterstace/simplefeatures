package geom_test

import (
	"strconv"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestUnmarshalWKTValidGrammar(t *testing.T) {
	for _, tt := range []struct {
		name, wkt string
	}{
		{"empty point", "POINT EMPTY"},
		{"mixed case", "PoInT (1 1)"},
		{"upper case", "POINT (1 1)"},
		{"lower case", "point (1 1)"},
		{"no space between tag and coord", "point(1 1)"},
		{"exponent", "point (1e3 1.5e2)"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := UnmarshalWKT(strings.NewReader(tt.wkt))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestUnmarshalWKTInvalidGrammar(t *testing.T) {
	for _, tt := range []struct {
		name, wkt string
	}{
		{"NaN coord", "point(NaN NaN)"},
		{"+Inf coord", "point(+Inf +Inf)"},
		{"-Inf coord", "point(-Inf -Inf)"},
		{"left unbalanced point", "point ( 1 2"},
		{"right unbalanced point", "point 1 2 )"},
		{"point no parens", "point 1 1"},
		{"point no coords", "point ( )"},

		{"mixed empty", "LINESTRING(0 0, EMPTY, 2 2)"},
		{"foo internal point", "LINESTRING(0 0, foo, 2 2)"},
		{"line string no coords", "LINESTRING()"},

		{"polygon no coords", "POLYGON()"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := UnmarshalWKT(strings.NewReader(tt.wkt))
			if err == nil {
				t.Fatalf("expected error but got nil")
			} else {
				t.Logf("got error: %v", err)
			}
		})
	}
}

func TestMarshalUnmarshalWKT(t *testing.T) {
	for i, wkt := range []string{
		"POINT (30 10)",
		"POINT (-30 -10)",
		"POINT EMPTY",

		"LINESTRING(30 10,10 30,40 40)",
		"LINESTRING EMPTY",

		"POLYGON((30 10,40 40,20 40,10 20,30 10))",
		"POLYGON((35 10,45 45,15 40,10 20,35 10),(20 30,35 35,30 20,20 30))",
		"POLYGON EMPTY",

		"MULTIPOINT ((10 40),(40 30),(20 20),(30 10))",
		"MULTIPOINT (10 40,40 30,20 20,30 10)",
		"MULTIPOINT (10 40,(40 30), EMPTY)",
		"MULTIPOINT EMPTY",
		"MULTIPOINT(EMPTY)",

		"MULTILINESTRING((10 10,20 20,10 40),(40 40,30 30,40 20,30 10))",
		"MULTILINESTRING((1 2,3 4,5 6),EMPTY)",
		"MULTILINESTRING EMPTY",
		"MULTILINESTRING(EMPTY)",

		"MULTIPOLYGON EMPTY",
		"MULTIPOLYGON(((30 20,45 40,10 40,30 20)),((15 5,40 10,10 20,5 10,15 5)))",
		"MULTIPOLYGON(((40 40,20 45,45 30,40 40)),((20 35,10 30,10 10,30 5,45 20,20 35),(30 20,20 15,20 25,30 20)))",
		"MULTIPOLYGON(EMPTY,((20 35,10 30,10 10,30 5,45 20,20 35),(30 20,20 15,20 25,30 20)))",

		"GEOMETRYCOLLECTION EMPTY",
		"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION EMPTY)",
		"GEOMETRYCOLLECTION(POINT EMPTY)",
		"GEOMETRYCOLLECTION(LINESTRING EMPTY)",
		"GEOMETRYCOLLECTION(POLYGON EMPTY)",
		"GEOMETRYCOLLECTION(MULTIPOINT EMPTY)",
		"GEOMETRYCOLLECTION(MULTILINESTRING EMPTY)",
		"GEOMETRYCOLLECTION(MULTIPOLYGON EMPTY)",
		"GEOMETRYCOLLECTION(LINESTRING(0 0,1 1))",
		"GEOMETRYCOLLECTION(POINT(4 6),LINESTRING(4 6,7 10))",
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g1 := geomFromWKT(t, wkt)
			g2 := geomFromWKT(t, string(g1.AsText()))
			expectGeomEq(t, g1, g2)
		})
	}
}

func TestUnmarshalWKT(t *testing.T) {
	t.Run("multi line string containing an empty line string", func(t *testing.T) {
		g := geomFromWKT(t, "MULTILINESTRING((1 2,3 4),EMPTY,(5 6,7 8))")
		mls := g.AsMultiLineString()
		expectIntEq(t, mls.NumLineStrings(), 3)
		expectGeomEq(t,
			mls.LineStringN(0).AsGeometry(),
			geomFromWKT(t, "LINESTRING(1 2,3 4)"),
		)
		expectGeomEq(t,
			mls.LineStringN(1).AsGeometry(),
			geomFromWKT(t, "LINESTRING EMPTY"),
		)
		expectGeomEq(t,
			mls.LineStringN(2).AsGeometry(),
			geomFromWKT(t, "LINESTRING(5 6,7 8)"),
		)
	})
}

func TestAsTextEmpty(t *testing.T) {
	for i, tt := range []struct {
		g    Geometry
		want string
	}{
		{
			NewEmptyPoint().AsGeometry(),
			"POINT EMPTY",
		},
		{
			NewEmptyLineString().AsGeometry(),
			"LINESTRING EMPTY",
		},
		{
			NewEmptyPolygon().AsGeometry(),
			"POLYGON EMPTY",
		},
		{
			NewMultiPoint(nil).AsGeometry(),
			"MULTIPOINT EMPTY",
		},
		{
			NewMultiLineString(nil).AsGeometry(),
			"MULTILINESTRING EMPTY",
		},
		{
			NewEmptyMultiPolygon().AsGeometry(),
			"MULTIPOLYGON EMPTY",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := tt.g.AsText()
			expectStringEq(t, got, tt.want)
		})
	}
}
