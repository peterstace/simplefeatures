package geom_test

import (
	"strconv"
	"testing"
)

func TestNormalize(t *testing.T) {
	for i, tc := range []struct {
		inputWKT string
		wantWKT  string
	}{
		{"POINT EMPTY", "POINT EMPTY"},
		{"POINT(1 2)", "POINT(1 2)"},

		{"MULTIPOINT EMPTY", "MULTIPOINT EMPTY"},
		{"MULTIPOINT(EMPTY)", "MULTIPOINT(EMPTY)"},
		{"MULTIPOINT(1 2)", "MULTIPOINT(1 2)"},
		{"MULTIPOINT(1 2,3 4)", "MULTIPOINT(1 2,3 4)"},
		{"MULTIPOINT(3 4,1 2)", "MULTIPOINT(1 2,3 4)"},
		{"MULTIPOINT(3 4,EMPTY)", "MULTIPOINT(EMPTY,3 4)"},
		{"MULTIPOINT(1 1,1 -1)", "MULTIPOINT(1 -1,1 1)"},
		{"MULTIPOINT(2 1,3 1,1 1)", "MULTIPOINT(1 1,2 1,3 1)"},

		{"LINESTRING EMPTY", "LINESTRING EMPTY"},
		{"LINESTRING(1 2,3 4)", "LINESTRING(1 2,3 4)"},
		{"LINESTRING(3 4,1 2)", "LINESTRING(1 2,3 4)"},
		{"LINESTRING(1 2,5 6,0 5,3 4)", "LINESTRING(1 2,5 6,0 5,3 4)"},
		{"LINESTRING(3 4,5 6,0 5,1 2)", "LINESTRING(1 2,0 5,5 6,3 4)"},
		{"LINESTRING(0 0,0 1,1 0,0 0)", "LINESTRING(0 0,0 1,1 0,0 0)"},
		{"LINESTRING(0 0,1 0,0 1,0 0)", "LINESTRING(0 0,0 1,1 0,0 0)"},

		{"MULTILINESTRING EMPTY", "MULTILINESTRING EMPTY"},
		{"MULTILINESTRING((3 4,1 2))", "MULTILINESTRING((1 2,3 4))"},
		{"MULTILINESTRING((5 6,7 8),(1 2,3 4))", "MULTILINESTRING((1 2,3 4),(5 6,7 8))"},
		{"MULTILINESTRING((5 6,7 8),EMPTY,(1 2,3 4))", "MULTILINESTRING(EMPTY,(1 2,3 4),(5 6,7 8))"},
		{"MULTILINESTRING((5 6,7 8),(1 2,5 6),(1 2,3 4))", "MULTILINESTRING((1 2,3 4),(1 2,5 6),(5 6,7 8))"},
		{"MULTILINESTRING((5 6,7 8),(1 2,3 4),(1 2,5 6))", "MULTILINESTRING((1 2,3 4),(1 2,5 6),(5 6,7 8))"},

		{"POLYGON EMPTY", "POLYGON EMPTY"},

		// Normalises outer ring orientation:
		{"POLYGON((0 0,0 1,1 0,0 0))", "POLYGON((0 0,1 0,0 1,0 0))"},
		{"POLYGON((0 0,1 0,0 1,0 0))", "POLYGON((0 0,1 0,0 1,0 0))"},

		// Normalises inner ring orientations:
		{
			"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1))",
			"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1))",
		},
		{
			"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1))",
			"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1))",
		},

		// Normalises ring starting points:
		{"POLYGON((0 0,1 0,0 1,0 0))", "POLYGON((0 0,1 0,0 1,0 0))"},
		{"POLYGON((1 0,0 1,0 0,1 0))", "POLYGON((0 0,1 0,0 1,0 0))"},
		{"POLYGON((0 1,0 0,1 0,0 1))", "POLYGON((0 0,1 0,0 1,0 0))"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := geomFromWKT(t, tc.inputWKT).Normalize()
			expectGeomEqWKT(t, got, tc.wantWKT)
		})
	}
}
