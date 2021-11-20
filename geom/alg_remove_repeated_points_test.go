package geom_test

import (
	"strconv"
	"testing"
)

func TestRemoveRepeatedPoints(t *testing.T) {
	for i, tt := range []struct {
		input      string
		wantOutput string
	}{
		{"POINT EMPTY", "POINT EMPTY"},
		{"POINT(1 2)", "POINT(1 2)"},

		{"MULTIPOINT EMPTY", "MULTIPOINT EMPTY"},
		{"MULTIPOINT(EMPTY)", "MULTIPOINT(EMPTY)"},
		{"MULTIPOINT(EMPTY,EMPTY)", "MULTIPOINT(EMPTY)"},
		{"MULTIPOINT(EMPTY,1 2)", "MULTIPOINT(EMPTY,1 2)"},
		{"MULTIPOINT(1 2,1 2)", "MULTIPOINT(1 2)"},
		{"MULTIPOINT(1 2,3 3,1 2)", "MULTIPOINT(1 2,3 3)"},
		{"MULTIPOINT(3 3,1 2,3 3,1 2)", "MULTIPOINT(3 3,1 2)"},

		{"LINESTRING(0 0,1 1)", "LINESTRING(0 0,1 1)"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			inputG := geomFromWKT(t, tt.input)
			gotG := inputG.RemoveRepeatedPoints()
			expectGeomEqWKT(t, gotG, tt.wantOutput)
		})
	}
}
