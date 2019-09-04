package geom_test

import (
	"reflect"
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
		"GEOMETRYCOLLECTION(LINESTRING(0 0,1 1))",
		"MULTIPOINT ((10 40),(40 30),(20 20),(30 10))",
		"MULTIPOINT (10 40,40 30,20 20,30 10)",
		"MULTIPOINT (10 40,(40 30), EMPTY)",
		"POINT (30 10)",
		"POINT (-30 -10)",
		"POINT EMPTY",
		"LINESTRING(30 10,10 30,40 40)",
		"POLYGON((30 10,40 40,20 40,10 20,30 10))",
		"POLYGON((35 10,45 45,15 40,10 20,35 10),(20 30,35 35,30 20,20 30))",
		"MULTILINESTRING((10 10,20 20,10 40),(40 40,30 30,40 20,30 10))",
		"MULTIPOLYGON(((30 20,45 40,10 40,30 20)),((15 5,40 10,10 20,5 10,15 5)))",
		"MULTIPOLYGON(((40 40,20 45,45 30,40 40)),((20 35,10 30,10 10,30 5,45 20,20 35),(30 20,20 15,20 25,30 20)))",
		"GEOMETRYCOLLECTION(POINT(4 6),LINESTRING(4 6,7 10))",
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g1 := geomFromWKT(t, wkt)
			g2 := geomFromWKT(t, string(g1.AsText()))
			if !reflect.DeepEqual(g1, g2) {
				t.Log("wkt", wkt)
				t.Logf("g1 %#v", g1)
				t.Logf("g2 %#v", g2)
				t.Errorf("not equal")
			}
		})
	}
}
