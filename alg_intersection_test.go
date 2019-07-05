package simplefeatures_test

import (
	"strconv"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures"
)

func TestIntersection(t *testing.T) {
	for i, tt := range []struct {
		in1, in2, out string
	}{
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
		{"LINESTRING(1 0,0 1)", "LINESTRING(0 1,1 0)", "LINESTRING(1 0,0 1)"},
		{"LINESTRING(1 0,0 1)", "LINESTRING(1 0,0 1)", "LINESTRING(1 0,0 1)"},
		{"LINESTRING(0 0,1 1)", "LINESTRING(1 1,0 0)", "LINESTRING(0 0,1 1)"},
		{"LINESTRING(0 0,1 1)", "LINESTRING(0 0,1 1)", "LINESTRING(0 0,1 1)"},
		{"LINESTRING(0 0,0 1)", "LINESTRING(0 1,0 0)", "LINESTRING(0 0,0 1)"},
		{"LINESTRING(0 0,0 1)", "LINESTRING(0 0,0 1)", "LINESTRING(0 0,0 1)"},
		{"LINESTRING(0 0,1 0)", "LINESTRING(1 0,0 0)", "LINESTRING(0 0,1 0)"},
		{"LINESTRING(0 0,1 0)", "LINESTRING(0 0,1 0)", "LINESTRING(0 0,1 0)"},
		{"LINESTRING(1 1,2 2)", "LINESTRING(0 0,3 3)", "LINESTRING(1 1,2 2)"},
		{"LINESTRING(3 1,2 2)", "LINESTRING(1 3,2 2)", "POINT(2 2)"},

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
				result := in1g.Intersection(in2g)
				got := string(result.AsText())
				if got != tt.out {
					t.Errorf("\ninput1: %s\ninput2: %s\nwant:   %v\ngot:    %v", tt.in1, tt.in2, tt.out, got)
				}
			})

			t.Run("reversed", func(t *testing.T) {
				result := in2g.Intersection(in1g)
				got := string(result.AsText())
				if got != tt.out {
					t.Errorf("\ninput1: %s\ninput2: %s\nwant:   %v\ngot:    %v", tt.in2, tt.in1, tt.out, got)
				}
			})
		})
	}
}
