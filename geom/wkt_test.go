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

		{"mixed empty", "LINESTRING(0 0, EMPTY, 2 2)"},
		{"foo internal point", "LINESTRING(0 0, foo, 2 2)"},

		{"point no coords", "POINT()"},
		{"line string no coords", "LINESTRING()"},
		{"polygon no coords", "POLYGON()"},
		{"multi point no coords", "MULTIPOINT()"},
		{"multi linestring no coords", "MULTILINESTRING()"},
		{"multi polygon no coords", "MULTIPOLYGON()"},
		{"geometry collection no coords", "GEOMETRYCOLLECTION()"},
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
