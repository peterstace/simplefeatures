package geom_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
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
			_, err := UnmarshalWKT(tt.wkt)
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
			"invalid token: '08' (invalid digit '8' in octal literal)",
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
	} {
		t.Run(tt.description, func(t *testing.T) {
			_, err := UnmarshalWKT(tt.wkt)
			if err == nil {
				t.Fatalf("expected error but got nil")
			}
			if _, isSynErr := err.(geom.SyntaxError); !isSynErr {
				t.Fatalf("expected a SyntaxError but instead got %v", err)
			}
			if err.Error() != tt.errorText {
				t.Logf("got:  %q", err.Error())
				t.Logf("want: %q", tt.errorText)
				t.Errorf("mismatch")
			}
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
	t.Run("multipoints with and without parenthesised points", func(t *testing.T) {
		g1 := geomFromWKT(t, "MULTIPOINT((10 40),(40 30),(20 20),(30 10))")
		g2 := geomFromWKT(t, "MULTIPOINT(10 40,40 30,20 20,30 10)")
		expectGeomEq(t, g1, g2)
	})
}

func TestAsTextEmpty(t *testing.T) {
	for i, tt := range []struct {
		want string
		g    Geometry
	}{
		{"POINT EMPTY", Point{}.AsGeometry()},
		{"LINESTRING EMPTY", LineString{}.AsGeometry()},
		{"POLYGON EMPTY", Polygon{}.AsGeometry()},
		{"MULTIPOINT EMPTY", MultiPoint{}.AsGeometry()},
		{"MULTILINESTRING EMPTY", MultiLineString{}.AsGeometry()},
		{"MULTIPOLYGON EMPTY", MultiPolygon{}.AsGeometry()},
		{"GEOMETRYCOLLECTION EMPTY", GeometryCollection{}.AsGeometry()},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := tt.g.AsText()
			expectStringEq(t, got, tt.want)
		})
	}
}
