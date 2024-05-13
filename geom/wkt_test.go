package geom_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
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
		{
			"multipoint with single empty point",
			"MULTIPOINT(EMPTY)",
		},
		{
			"multipoint with empty point and non-empty point, empty first",
			"MULTIPOINT(EMPTY,1 2)",
		},
		{
			"multipoint with empty point and non-empty point, empty second",
			"MULTIPOINT(1 2,EMPTY)",
		},
		{
			"multipoint with empty point and non-empty point with parens, empty first",
			"MULTIPOINT(EMPTY,(1 2))",
		},
		{
			"multipoint with empty point and non-empty point with parens, empty second",
			"MULTIPOINT((1 2),EMPTY)",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("WKT: %v", tt.wkt)
			_, err := geom.UnmarshalWKT(tt.wkt)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestUnmarshalWKTSyntaxErrors(t *testing.T) {
	for _, tt := range []struct {
		description, wkt string
		errorText        string
	}{
		{
			"missing EOF after WKT",
			"POINT(0 0) MORE",
			"unexpected token: 'MORE' (expected EOF)",
		},
		{
			"unknown geometry type",
			"LINESTRANG(0 0,1 1)",
			"unexpected token: 'LINESTRANG' (expected geometry tag)",
		},
		{
			"square opening bracket instead of paren",
			"POINT[0 0)",
			"unexpected token: '[' (expected 'EMPTY' or '(')",
		},
		{
			"square closing bracket instead of paren",
			"POINT(0 0]",
			"unexpected token: ']' (expected ')')",
		},
		{
			"- instead of ,",
			"LINESTRING(0 0-1 1)",
			"unexpected token: '-' (expected ')' or ',')",
		},
		{
			"invalid float64",
			"POINT(X Y)",
			`strconv.ParseFloat: parsing "X": invalid syntax`,
		},
		{
			"Inf not allowed",
			"POINT(Inf 0)",
			`invalid numeric literal: Inf`,
		},
		{
			"NaN not allowed",
			"POINT(NaN 0)",
			`invalid numeric literal: NaN`,
		},
		{
			"unexpected EOF",
			"POINT(0 0",
			"unexpected EOF",
		},
		{
			"token contains invalid octal digit",
			"POINT(08, 0)",
			"invalid token '08' (invalid digit '8' in octal literal)",
		},

		{
			"mixed empty",
			"LINESTRING(0 0, EMPTY, 2 2)",
			"strconv.ParseFloat: parsing \"EMPTY\": invalid syntax",
		},
		{
			"point no coords",
			"POINT()",
			"strconv.ParseFloat: parsing \")\": invalid syntax",
		},
		{
			"line string no coords",
			"LINESTRING()",
			"strconv.ParseFloat: parsing \")\": invalid syntax",
		},
		{
			"polygon no coords",
			"POLYGON()",
			"unexpected token: ')' (expected 'EMPTY' or '(')",
		},
		{
			"multi point no coords",
			"MULTIPOINT()",
			"strconv.ParseFloat: parsing \")\": invalid syntax",
		},
		{
			"multi linestring no coords",
			"MULTILINESTRING()",
			"unexpected token: ')' (expected 'EMPTY' or '(')",
		},
		{
			"multi polygon no coords",
			"MULTIPOLYGON()",
			"unexpected token: ')' (expected 'EMPTY' or '(')",
		},
		{
			"geometry collection no coords",
			"GEOMETRYCOLLECTION()",
			// This is a slightly unexpected error here. It occurs because we
			// take the ')' token as the geometry type and then immediately
			// peek to see if we got a Z, M, or ZM. But this triggers an
			// unexpected EOF error because there is no next token to peek. We
			// _could_ alter the code to give a more accurate error message,
			// but I think it's ok because this is an extreme edge case.
			"unexpected EOF",
		},
		{
			"multipoint with parenthesis around empty point",
			"MULTIPOINT((1 2),(EMPTY))",
			"strconv.ParseFloat: parsing \"EMPTY\": invalid syntax",
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			_, err := geom.UnmarshalWKT(tt.wkt)
			if err == nil {
				t.Fatalf("expected error but got nil")
			}
			want := "invalid WKT syntax: " + tt.errorText
			if err.Error() != want {
				t.Logf("got:  %q", err.Error())
				t.Logf("want: %q", want)
				t.Errorf("mismatch")
			}
		})
	}
}

func TestUnmarshalWKT(t *testing.T) {
	t.Run("multi line string containing an empty line string", func(t *testing.T) {
		g := geomFromWKT(t, "MULTILINESTRING((1 2,3 4),EMPTY,(5 6,7 8))")
		mls := g.MustAsMultiLineString()
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
	t.Run("multipoints with and without parenthesised points", func(t *testing.T) {
		g1 := geomFromWKT(t, "MULTIPOINT((10 40),(40 30),(20 20),(30 10))")
		g2 := geomFromWKT(t, "MULTIPOINT(10 40,40 30,20 20,30 10)")
		expectGeomEq(t, g1, g2)
	})
}

func TestAsTextEmpty(t *testing.T) {
	for i, tt := range []struct {
		want string
		g    geom.Geometry
	}{
		{"POINT EMPTY", geom.Point{}.AsGeometry()},
		{"LINESTRING EMPTY", geom.LineString{}.AsGeometry()},
		{"POLYGON EMPTY", geom.Polygon{}.AsGeometry()},
		{"MULTIPOINT EMPTY", geom.MultiPoint{}.AsGeometry()},
		{"MULTILINESTRING EMPTY", geom.MultiLineString{}.AsGeometry()},
		{"MULTIPOLYGON EMPTY", geom.MultiPolygon{}.AsGeometry()},
		{"GEOMETRYCOLLECTION EMPTY", geom.GeometryCollection{}.AsGeometry()},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := tt.g.AsText()
			expectStringEq(t, got, tt.want)
		})
	}
}

func TestInconsistentDimensionTypeInWKT(t *testing.T) {
	for _, tc := range []struct {
		allow bool
		wkt   string
	}{
		{true, "GEOMETRYCOLLECTION (POINT (1 2))"},
		{true, "GEOMETRYCOLLECTION (POINT (1 2), POINT(2 3))"},

		{true, "GEOMETRYCOLLECTION M(POINT M(1 2 0))"},
		{true, "GEOMETRYCOLLECTION Z(POINT Z(1 2 0))"},
		{true, "GEOMETRYCOLLECTION ZM(POINT ZM(1 2 0 0))"},

		{false, "GEOMETRYCOLLECTION M(POINT (1 2))"},
		{false, "GEOMETRYCOLLECTION M(POINT Z(1 2 0))"},
		{false, "GEOMETRYCOLLECTION M(POINT ZM(1 2 0 0))"},

		{false, "GEOMETRYCOLLECTION Z(POINT (1 2))"},
		{false, "GEOMETRYCOLLECTION Z(POINT M(1 2 0))"},
		{false, "GEOMETRYCOLLECTION Z(POINT ZM(1 2 0 0))"},

		{false, "GEOMETRYCOLLECTION ZM(POINT (1 2))"},
		{false, "GEOMETRYCOLLECTION ZM(POINT M(1 2 0))"},
		{false, "GEOMETRYCOLLECTION ZM(POINT Z(1 2 0))"},

		{false, "GEOMETRYCOLLECTION (POINT (1 2), POINT Z(2 3 0))"},
		{false, "GEOMETRYCOLLECTION (POINT (1 2), POINT M(2 3 0))"},
		{false, "GEOMETRYCOLLECTION (POINT (1 2), POINT ZM(2 3 0 0))"},

		// These forms are accepted by PostGIS, but banned by the OGC spec.
		// Simplefeatures follows the PostGIS behaviour (since it's a
		// reasonable extension).
		{true, "GEOMETRYCOLLECTION (POINT Z(1 2 0))"},
		{true, "GEOMETRYCOLLECTION (POINT M(1 2 0))"},
		{true, "GEOMETRYCOLLECTION (POINT ZM(1 2 0 0))"},
	} {
		t.Run(tc.wkt, func(t *testing.T) {
			_, err := geom.UnmarshalWKT(tc.wkt)
			if tc.allow {
				expectNoErr(t, err)
				return
			}
			expectErr(t, err)
			expectSubstring(t, err.Error(), "mixed dimensions in geometry collection")
		})
	}
}
