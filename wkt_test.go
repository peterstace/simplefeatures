package simplefeatures_test

import (
	"reflect"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures"
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

func TestUnmarshalWKTPopulate(t *testing.T) {
	must := func(g Geometry, err error) Geometry {
		if err != nil {
			t.Fatalf("could not create geometry: %v", err)
		}
		return g
	}
	for _, tt := range []struct {
		name string
		wkt  string
		want Geometry
	}{
		{
			name: "basic point (wikipedia)",
			wkt:  "POINT (30 10)",
			want: must(NewPoint(30, 10)),
		},
		{
			name: "empty point",
			wkt:  "POINT EMPTY",
			want: NewEmptyPoint(),
		},
		{
			name: "basic line string (wikipedia)",
			wkt:  "LINESTRING (30 10, 10 30, 40 40)",
			want: must(NewLineString([]Point{
				must(NewPoint(30, 10)).(Point),
				must(NewPoint(10, 30)).(Point),
				must(NewPoint(40, 40)).(Point),
			})),
		},
		{
			name: "basic polygon (wikipedia)",
			wkt:  "POLYGON ((30 10, 40 40, 20 40, 10 20, 30 10))",
			want: must(NewPolygon(must(NewLinearRing([]Point{
				must(NewPoint(30, 10)).(Point),
				must(NewPoint(40, 40)).(Point),
				must(NewPoint(20, 40)).(Point),
				must(NewPoint(10, 20)).(Point),
				must(NewPoint(30, 10)).(Point),
			})).(LinearRing))),
		},
		{
			name: "polygon with hole (wikipedia)",
			wkt:  "POLYGON ((35 10, 45 45, 15 40, 10 20, 35 10), (20 30, 35 35, 30 20, 20 30))",
			want: must(NewPolygon(
				must(NewLinearRing([]Point{
					must(NewPoint(35, 10)).(Point),
					must(NewPoint(45, 45)).(Point),
					must(NewPoint(15, 40)).(Point),
					must(NewPoint(10, 20)).(Point),
					must(NewPoint(35, 10)).(Point),
				})).(LinearRing),
				must(NewLinearRing([]Point{
					must(NewPoint(20, 30)).(Point),
					must(NewPoint(35, 35)).(Point),
					must(NewPoint(30, 20)).(Point),
					must(NewPoint(20, 30)).(Point),
				})).(LinearRing),
			)),
		},
		{
			name: "basic multipoint (wikipedia)",
			wkt:  "MULTIPOINT ((10 40), (40 30), (20 20), (30 10))",
			want: must(NewMultiPoint([]Point{
				must(NewPoint(10, 40)).(Point),
				must(NewPoint(40, 30)).(Point),
				must(NewPoint(20, 20)).(Point),
				must(NewPoint(30, 10)).(Point),
			})),
		},
		{
			name: "basic multipoint without parens (wikipedia)",
			wkt:  "MULTIPOINT (10 40, 40 30, 20 20, 30 10)",
			want: must(NewMultiPoint([]Point{
				must(NewPoint(10, 40)).(Point),
				must(NewPoint(40, 30)).(Point),
				must(NewPoint(20, 20)).(Point),
				must(NewPoint(30, 10)).(Point),
			})),
		},
		{
			name: "mixed style multipoint",
			wkt:  "MULTIPOINT (10 40, (40 30), EMPTY)",
			want: must(NewMultiPoint([]Point{
				must(NewPoint(10, 40)).(Point),
				must(NewPoint(40, 30)).(Point),
				NewEmptyPoint(),
			})),
		},
		{
			name: "multi line string (wikipedia)",
			wkt:  "MULTILINESTRING ((10 10, 20 20, 10 40), (40 40, 30 30, 40 20, 30 10))",
			want: must(NewMultiLineString([]LineString{
				must(NewLineString([]Point{
					must(NewPoint(10, 10)).(Point),
					must(NewPoint(20, 20)).(Point),
					must(NewPoint(10, 40)).(Point),
				})).(LineString),
				must(NewLineString([]Point{
					must(NewPoint(40, 40)).(Point),
					must(NewPoint(30, 30)).(Point),
					must(NewPoint(40, 20)).(Point),
					must(NewPoint(30, 10)).(Point),
				})).(LineString),
			})),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalWKT(strings.NewReader(tt.wkt))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("want=%#v got=%#v", tt.want, got)
			}
		})
	}
}
