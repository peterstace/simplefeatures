package geom_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestSimplify(t *testing.T) {
	for i, tc := range []struct {
		input     string
		threshold float64
		output    string
	}{
		// Points and MultiPoints pass through unchanged.
		{"POINT(0 1)", 0, "POINT(0 1)"},
		{"POINT(0 1)", 0.5, "POINT(0 1)"},
		{"POINT(0 1)", 1, "POINT(0 1)"},
		{"POINT(0 1)", 2, "POINT(0 1)"},
		{"POINT EMPTY", 0, "POINT EMPTY"},
		{"POINT EMPTY", 0.5, "POINT EMPTY"},
		{"POINT EMPTY", 1, "POINT EMPTY"},
		{"POINT EMPTY", 2, "POINT EMPTY"},
		{"MULTIPOINT(0 0,1 1)", 0.5, "MULTIPOINT(0 0,1 1)"},
		{"MULTIPOINT(0 0,1 1)", 2, "MULTIPOINT(0 0,1 1)"},

		{"LINESTRING(0 0,1 1)", 0.0, "LINESTRING(0 0,1 1)"},
		{"LINESTRING(0 0,1 1)", 1.0, "LINESTRING(0 0,1 1)"},
		{"LINESTRING(0 0,1 1)", 2.0, "LINESTRING(0 0,1 1)"},
		{"LINESTRING(0 0,1 1,2 0)", 0.5, "LINESTRING(0 0,1 1,2 0)"},
		{"LINESTRING(0 0,1 1,2 0)", 1.0, "LINESTRING(0 0,2 0)"},
		{"LINESTRING(0 0,1 1,2 0)", 1.5, "LINESTRING(0 0,2 0)"},

		{"LINESTRING(0 0,0 1,1 1,1 0)", 0.5, "LINESTRING(0 0,0 1,1 1,1 0)"},
		{"LINESTRING(0 0,0 1,1 1,1 0)", 0.9, "LINESTRING(0 0,0 1,1 0)"},
		{"LINESTRING(0 0,0 1,1 1,1 0)", 1.0, "LINESTRING(0 0,1 0)"},
		{"LINESTRING(0 0,0 1,1 1,1 0)", 1.5, "LINESTRING(0 0,1 0)"},

		{"LINESTRING(0 0,0 1,1 0,0 0)", 0.5, "LINESTRING(0 0,0 1,1 0,0 0)"},
		{"LINESTRING(0 0,0 1,1 0,0 0)", 1.0, "LINESTRING EMPTY"},
		{"LINESTRING(0 0,0 1,1 0,0 0)", 1.5, "LINESTRING EMPTY"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			in := geomFromWKT(t, tc.input)
			want := geomFromWKT(t, tc.output)
			t.Logf("input:     %v", in.AsText())
			t.Logf("threshold: %v", tc.threshold)
			t.Logf("want:      %v", want.AsText())
			got, err := geom.Simplify(in, tc.threshold)
			expectNoErr(t, err)
			t.Logf("got:       %v", got.AsText())
			expectGeomEq(t, got, want, geom.IgnoreOrder)
		})
	}
}
