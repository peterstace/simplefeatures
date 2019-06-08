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
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := UnmarshalWKT(strings.NewReader(tt.wkt))
			if err == nil {
				t.Fatalf("expected error but got nil")
			}
		})
	}
}

func TestUnmarshalWKTValid(t *testing.T) {
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
			want: NewPoint(30, 10),
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
				NewPoint(30, 10),
				NewPoint(10, 30),
				NewPoint(40, 40),
			})),
		},
		/*
			{
				name: "basic polygon (wikipedia)",
				wkt:  "POLYGON ((30 10, 40 40, 20 40, 10 20, 30 10))",
				want: must(NewPolygon(must(NewLinearRing([]Point{
					NewPoint(30, 10),
					NewPoint(40, 40),
					NewPoint(20, 40),
					NewPoint(10, 20),
					NewPoint(30, 10),
				})).(LinearRing))),
			},
		*/
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalWKT(strings.NewReader(tt.wkt))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("want=%#v got=%#v", got, tt.want)
			}
		})
	}
}
