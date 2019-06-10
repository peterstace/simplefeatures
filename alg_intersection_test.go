package simplefeatures_test

import (
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures"
)

func TestIntersectionLineWithLine(t *testing.T) {
	for _, tt := range []struct {
		in1, in2, out string
	}{
		{"LINESTRING(0 0,0 1)", "LINESTRING(0 0,1 0)", "POINT(0 0)"},
		{"LINESTRING(0 1,1 1)", "LINESTRING(1 0,1 1)", "POINT(1 1)"},
		{"LINESTRING(0 1,0 0)", "LINESTRING(0 0,1 0)", "POINT(0 0)"},
		{"LINESTRING(0 0,0 1)", "LINESTRING(1 0,0 0)", "POINT(0 0)"},
		{"LINESTRING(0 0,1 0)", "LINESTRING(1 0,2 0)", "POINT(1 0)"},
		{"LINESTRING(0 0,1 0)", "LINESTRING(2 0,3 0)", "GEOMETRYCOLLECTION EMPTY"},
		{"LINESTRING(1 0,2 0)", "LINESTRING(0 0,3 0)", "LINESTRING(1 0,2 0)"},
		{"LINESTRING(0 0,0 1)", "LINESTRING(1 0,1 1)", "GEOMETRYCOLLECTION EMPTY"},
		{"LINESTRING(0 0,1 1)", "LINESTRING(1 0,0 1)", "POINT(0.5 0.5)"},
		{"LINESTRING(1 1,2 2)", "LINESTRING(0 0,3 3)", "LINESTRING(1 1,2 2)"},
	} {
		in1g, err := UnmarshalWKT(strings.NewReader(tt.in1))
		if err != nil {
			t.Fatalf("could not unmarshal wkt: %v", err)
		}
		in2g, err := UnmarshalWKT(strings.NewReader(tt.in2))
		if err != nil {
			t.Fatalf("could not unmarshal wkt: %v", err)
		}
		result := in1g.Intersection(in2g)
		got := string(result.AsText())
		if got != tt.out {
			t.Errorf("got=%v want=%v", got, tt.out)
		}
	}
}
