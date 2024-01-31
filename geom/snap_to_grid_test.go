package geom_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestSnapPiFloat64ToGrid(t *testing.T) {
	for _, tc := range []struct {
		dp   int
		want float64
	}{
		{-999, 0},
		{-99, 0},
		{-9, 0},
		{-2, 0},
		{-1, 0},
		{0, 3},
		{1, 3.1},
		{2, 3.14},
		{3, 3.142},
		{4, 3.1416},
		{5, 3.14159},
		{6, 3.141593},
		{7, 3.1415927},
		{8, 3.14159265},
		{9, 3.141592654},
		{10, 3.1415926536},
		{11, 3.14159265359},
		{12, 3.141592653590},
		{13, 3.1415926535898},
		{14, 3.14159265358979},
		{15, 3.141592653589793},
		{16, 3.141592653589793},
		{17, 3.141592653589793},
		{99, 3.141592653589793},
		{999, 3.141592653589793},
	} {
		t.Run(strconv.Itoa(tc.dp), func(t *testing.T) {
			pt := geom.XY{0, math.Pi}.AsPoint()
			pt = pt.SnapToGrid(tc.dp)
			xy, ok := pt.XY()
			expectTrue(t, ok)
			expectFloat64Eq(t, xy.Y, tc.want)
		})
	}
}

func TestSnapToGrid(t *testing.T) {
	for i, tc := range []struct {
		input  string
		output string
	}{
		{"GEOMETRYCOLLECTION EMPTY", "GEOMETRYCOLLECTION EMPTY"},
		{"POINT EMPTY", "POINT EMPTY"},
		{"LINESTRING EMPTY", "LINESTRING EMPTY"},
		{"POLYGON EMPTY", "POLYGON EMPTY"},
		{"MULTIPOINT EMPTY", "MULTIPOINT EMPTY"},
		{"MULTILINESTRING EMPTY", "MULTILINESTRING EMPTY"},
		{"MULTIPOLYGON EMPTY", "MULTIPOLYGON EMPTY"},

		{
			"GEOMETRYCOLLECTION(POINT(1.11 2.22))",
			"GEOMETRYCOLLECTION(POINT(1.1 2.2))",
		},
		{
			"POINT(1.11 2.22)",
			"POINT(1.1 2.2)",
		},
		{
			"LINESTRING(1.11 2.22,3.33 4.44)",
			"LINESTRING(1.1 2.2,3.3 4.4)",
		},
		{
			"POLYGON((0.00 0.00,0.00 1.11,1.11 0.00,0.00 0.00))",
			"POLYGON((0.0 0.0,0.0 1.1,1.1 0.0,0.0 0.0))",
		},
		{
			"MULTIPOINT(1.11 2.22,3.33 4.44)",
			"MULTIPOINT(1.1 2.2,3.3 4.4)",
		},
		{
			"MULTILINESTRING((1.11 2.22,3.33 4.44),(5.55 6.66,7.77 8.88))",
			"MULTILINESTRING((1.1 2.2,3.3 4.4),(5.6 6.7,7.8 8.9))",
		},
		{
			"MULTIPOLYGON(((0.00 0.00,0.00 1.11,1.11 0.00,0.00 0.00)),((2.22 3.33,2.22 4.44,3.33 3.33,2.22 3.33)))",
			"MULTIPOLYGON(((0.0 0.0,0.0 1.1,1.1 0.0,0.0 0.0)),((2.2 3.3,2.2 4.4,3.3 3.3,2.2 3.3)))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			in := geomFromWKT(t, tc.input)
			got := in.SnapToGrid(1)
			expectGeomEqWKT(t, got, tc.output)
		})
	}
}
