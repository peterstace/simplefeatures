package geom_test

import (
	"strconv"
	"testing"
)

func TestPointOnSurface(t *testing.T) {
	for i, tt := range []struct {
		inputWKT  string
		outputWKT string
	}{
		{"POINT EMPTY", "POINT EMPTY"},
		{"POINT(1 2)", "POINT(1 2)"},
		{"LINESTRING EMPTY", "POINT EMPTY"},
		{"LINESTRING(1 2,3 4)", "POINT(1 2)"},
		{"LINESTRING(3 4,1 2)", "POINT(3 4)"},
		{"LINESTRING(5 3,2 8,1 0)", "POINT(2 8)"},
		{"LINESTRING(1 1,2 2,100 100)", "POINT(2 2)"},
		{"LINESTRING(1 1,2 2,100 100,101 101)", "POINT(2 2)"},
		{"LINESTRING(1 1,2 2,100 100,101 101,102 102)", "POINT(100 100)"},
		{"LINESTRING(1 1,100 100,101 101,102 102,103 103)", "POINT(100 100)"},
		{"LINESTRING(1 1,1 1,2 2)", "POINT(1 1)"},
		{"LINESTRING(1 1,1 1,1 1,2 2)", "POINT(1 1)"},
		{"LINESTRING(1 1,1 1,1 1,1 1,2 2)", "POINT(1 1)"},
		{"LINESTRING(1 1,2 2,2 2)", "POINT(2 2)"},
		{"LINESTRING(1 1,2 2,2 2,2 2)", "POINT(2 2)"},
		{"LINESTRING(1 1,2 2,2 2,2 2,2 2)", "POINT(2 2)"},
		{"LINESTRING(6 6,7 7,8 8,9 9,10 10,1 1,2 2,3 3,4 4,11 11)", "POINT(7 7)"},
		{"LINESTRING(6 6,7 7,6 6,8 8,9 9,10 10,1 1,2 2,3 3,4 4,11 11)", "POINT(6 6)"},
		{"LINESTRING(1 1,0 0,1 1,2 2,1 1)", "POINT(1 1)"},
		{"POLYGON EMPTY", "POINT EMPTY"},
		{"POLYGON((0 0,0 1,1 1,1 0,0 0))", "POINT(0.5 0.5)"},
		{"POLYGON((0 0,0 1,1 0,0 0))", "POINT(0.25 0.5)"},
		{"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1))", "POINT(0.5 1.5)"},
		{"POLYGON((0 0,4 0,4 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1))", "POINT(3 1.5)"},
		{"POLYGON((0 0,0 1,0 2,2 2,2 0,0 0))", "POINT(1 1.5)"},
		{"POLYGON((0 0,0 1,0 1.5,0 2,2 2,2 0,0 0))", "POINT(1 1.25)"},
		{"POLYGON((0 0,0 0.5,0 1,0 2,2 2,2 0,0 0))", "POINT(1 1.5)"},
		{"POLYGON((0 0,1 0,1 2,2 1.5,3 2,3 3,0 3,0 0))", "POINT(0.5 1.75)"},
		{"POLYGON((0 0,3 0,3 3,0 3,0 0),(1.5 1.5,2 2,1.5 2.5,1 2,1.5 1.5))", "POINT(0.625 1.75)"},
		{"MULTIPOINT EMPTY", "POINT EMPTY"},
		{"MULTIPOINT(EMPTY)", "POINT EMPTY"},
		{"MULTIPOINT(1 2,3 4)", "POINT(1 2)"},
		{"MULTIPOINT(1 2,3 4,5 6)", "POINT(3 4)"},
		{"MULTIPOINT(3 4,1 2,5 6)", "POINT(3 4)"},
		{"MULTIPOINT(1 2,3 4,5 6,7 8)", "POINT(3 4)"},
		{"MULTIPOINT(1 2,3 4,5 6,7 8,9 10)", "POINT(5 6)"},
		{"MULTILINESTRING EMPTY", "POINT EMPTY"},
		{"MULTILINESTRING(EMPTY)", "POINT EMPTY"},
		{"MULTILINESTRING((0 0,1 1),(2 2,3 3))", "POINT(1 1)"},
		{"MULTILINESTRING((0 0,1 1,2 2,3 3),(4 4,5 5))", "POINT(2 2)"},
		{"MULTILINESTRING((0 0,1 1,2 2,3 3),(4 4,5 5),(2.2 2.2,2.3 2.3))", "POINT(2 2)"},
		{"MULTILINESTRING((0 0,1 1,2 2,3 3),(4 4,5 5),(6 6,7 7))", "POINT(2 2)"},
		{"MULTILINESTRING((0 0,1 1,2 2,3 3),(4 4,5 5),(6 6,7 7,8 8,9 9))", "POINT(2 2)"},
		{"MULTILINESTRING((0 0,1 1,2 2,3 3),(4 4,5 5),(6 6,7 7,8 8,9 9,10 10))", "POINT(7 7)"},
		{"MULTILINESTRING((0 0,1 1,2 2,3 3),(4 4,5 5),(6 6,7 7,8 8,9 9,10 10,1 1))", "POINT(7 7)"},
		{"MULTILINESTRING((0 0,1 1,2 2,3 3,2 2),(4 4,5 5),(6 6,7 7,8 8,9 9,10 10,1 1))", "POINT(7 7)"},
		{"MULTILINESTRING((0 0,1 1),(0.4 0.4,0.6 0.6))", "POINT(0.4 0.4)"},
		{"MULTILINESTRING((0 0,1 1),(0.3 0.3,0.6 0.6))", "POINT(0.6 0.6)"},
		{"MULTILINESTRING((0 0,1 1,3 3),(1.4 1.4,1.6 1.6))", "POINT(1 1)"},
		{"MULTIPOLYGON EMPTY", "POINT EMPTY"},
		{"MULTIPOLYGON(((0 0,2 0,1 1,2 2,0 2,0 0)),((2 0,4 0,4 2,2 2,3 1,2 0)))", "POINT(0.75 1.5)"},
		{"MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0),(0.5 0.5,1 0.5,1 1.5,0.5 1.5,0.5 0.5)),((2 0,5 0,5 2,4 2,2 0)))", "POINT(4 1)"},
		{"GEOMETRYCOLLECTION(POINT(1 2),POINT(3 4))", "POINT(1 2)"},
		{"GEOMETRYCOLLECTION(POLYGON((0 0,0 1,1 1,1 0,0 0)),POINT(0.5 0.5))", "POINT(0.5 0.5)"},
		{"GEOMETRYCOLLECTION(POLYGON((0 0,0 2,2 2,2 0,0 0)),POLYGON((1 1,1 3,3 3,3 1,1 1)))", "POINT(1 1)"},
		{"GEOMETRYCOLLECTION(POLYGON((0 0,0 2,2 2,2 0,0 0)),POLYGON((1 1,1 4,4 4,4 1,1 1)))", "POINT(2.5 2.5)"},
		{"GEOMETRYCOLLECTION(POLYGON((0 0,0 3,3 3,3 0,0 0),(1 1,1 2,2 2,2 1,1 1)),POINT(1.5 1.5))", "POINT(0.5 1.5)"},
		{"GEOMETRYCOLLECTION(POLYGON((0 0,0 3,3 3,3 0,0 0),(1 1,1 2,2 2,2 1,1 1)),LINESTRING(1.4 1.4,1.6 1.6))", "POINT(0.5 1.5)"},
		{"GEOMETRYCOLLECTION(LINESTRING(0 0,1 1),POINT(0.5 0.5))", "POINT(0 0)"},
		{"POINT EMPTY", "POINT EMPTY"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			input := geomFromWKT(t, tt.inputWKT)
			got := input.PointOnSurface()
			t.Logf("input: %v", tt.inputWKT)
			expectGeomEq(t, got.AsGeometry(), geomFromWKT(t, tt.outputWKT))
		})
	}
}
