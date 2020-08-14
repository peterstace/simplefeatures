package geom_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestDistance(t *testing.T) {
	for i, tt := range []struct {
		wkt1, wkt2 string
		wantOK     bool
		wantDist   float64
	}{
		{"POINT EMPTY", "POINT EMPTY", false, 0},
		{"POINT (0 0)", "POINT EMPTY", false, 0},
		{"POINT EMPTY", "POINT (0 0)", false, 0},
		{"POINT(1 2)", "POINT(3 2)", true, 2.0},
		{"POINT(1 2)", "POINT(2 3)", true, math.Sqrt(2)},

		{"POINT EMPTY", "LINESTRING EMPTY", false, 0},
		{"POINT EMPTY", "LINESTRING(0 0,1 1)", false, 0},
		{"POINT(1 1)", "LINESTRING EMPTY", false, 0},
		{"LINESTRING(1 0,2 1)", "POINT(1 0)", true, 0},
		{"LINESTRING(1 0,2 1)", "POINT(2 1)", true, 0},
		{"LINESTRING(1 0,2 1)", "POINT(1.5 0.5)", true, 0},
		{"LINESTRING(1 0,2 1)", "POINT(1 1)", true, math.Sqrt(2) / 2},
		{"LINESTRING(1 0,2 1)", "POINT(0 1)", true, math.Sqrt(2)},
		{"LINESTRING(1 0,2 1)", "POINT(2 -1)", true, math.Sqrt(2)},
		{"LINESTRING(1 0,2 1)", "POINT(3 0)", true, math.Sqrt(2)},
		{"LINESTRING(1 0,2 1)", "POINT(0 0)", true, 1},
		{"LINESTRING(1 0,2 1)", "POINT(2 2)", true, 1},
		{"LINESTRING(0 0,1 2,2 2)", "POINT(0 0)", true, 0},
		{"LINESTRING(0 0,1 2,2 2)", "POINT(0 -1)", true, 1},
		{"LINESTRING(0 0,1 2,2 2)", "POINT(0 3)", true, math.Sqrt(2)},
		{"LINESTRING(0 0,1 2,2 2)", "POINT(1 0)", true, 2 / math.Sqrt(5)},
		{"LINESTRING(0 0,1 2,2 2)", "POINT(1.5 1.5)", true, 0.5},
		{"LINESTRING(0 0,1 2,2 2)", "POINT(3 2)", true, 1},
		{"LINESTRING(0 0,4 0)", "POINT(3 1)", true, 1},

		{"POINT EMPTY", "POLYGON EMPTY", false, 0},
		{"POINT EMPTY", "POLYGON((0 0,0 1,1 0,0 0))", false, 0},
		{"POINT(0 0)", "POLYGON EMPTY", false, 0},
		{"POLYGON((0 0,0 1,1 0,0 0))", "POINT(0 1)", true, 0},
		{"POLYGON((0 0,0 1,1 0,0 0))", "POINT(1 1)", true, math.Sqrt(2) / 2},
		{"POLYGON((0 0,0 1,1 0,0 0))", "POINT(2 0)", true, 1},
		{"POLYGON((0 0,0 1,1 0,0 0))", "POINT(0.1 0.1)", true, 0},
		{"POLYGON((0 0,0 3,3 3,3 0,0 0),(1 1,1 2,2 2,2 1,1 1))", "POINT(1.5 1.5)", true, 0.5},

		{"POINT EMPTY", "MULTIPOINT EMPTY", false, 0},
		{"POINT EMPTY", "MULTIPOINT(EMPTY)", false, 0},
		{"POINT EMPTY", "MULTIPOINT(EMPTY,EMPTY)", false, 0},
		{"POINT EMPTY", "MULTIPOINT(0 0)", false, 0},
		{"POINT EMPTY", "MULTIPOINT(0 0,EMPTY)", false, 0},
		{"POINT(0 0)", "MULTIPOINT EMPTY", false, 0},
		{"POINT(0 0)", "MULTIPOINT(EMPTY)", false, 0},
		{"POINT(0 0)", "MULTIPOINT(0 0)", true, 0},
		{"POINT(0 0)", "MULTIPOINT(0 1)", true, 1},
		{"POINT(0 0)", "MULTIPOINT(0 1,1 0)", true, 1},
		{"POINT(0 0)", "MULTIPOINT(0 2,3 0)", true, 2},
		{"POINT(0 0)", "MULTIPOINT(0 2,1 1,3 0)", true, math.Sqrt(2)},

		{"POINT EMPTY", "MULTILINESTRING EMPTY", false, 0},
		{"POINT EMPTY", "MULTILINESTRING(EMPTY)", false, 0},
		{"POINT EMPTY", "MULTILINESTRING(EMPTY,EMPTY)", false, 0},
		{"POINT(0 0)", "MULTILINESTRING EMPTY", false, 0},
		{"POINT(0 0)", "MULTILINESTRING(EMPTY)", false, 0},
		{"POINT(0 0)", "MULTILINESTRING(EMPTY,EMPTY)", false, 0},
		{"POINT EMPTY", "MULTILINESTRING((0 0,1 1))", false, 0},
		{"POINT EMPTY", "MULTILINESTRING((0 0,1 1),EMPTY)", false, 0},
		{"POINT(1 1)", "MULTILINESTRING((0 2,1 2),(10 0,10 1))", true, 1},
		{"POINT(1 1)", "MULTILINESTRING((10 0,10 1),(0 2,1 2))", true, 1},

		{"POINT EMPTY", "MULTIPOLYGON EMPTY", false, 0},
		{"POINT EMPTY", "MULTIPOLYGON(EMPTY)", false, 0},
		{"POINT(0 0)", "MULTIPOLYGON EMPTY", false, 0},
		{"POINT(0 0)", "MULTIPOLYGON(EMPTY)", false, 0},
		{"POINT EMPTY", "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))", false, 0},
		{"POINT EMPTY", "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))", false, 0},
		{"POINT(0 0)", "MULTIPOLYGON(((10 0,11 0,10 1,10 0)),((0 3,1 3,0 4,0 3)))", true, 3},
		{"POINT(0 0)", "MULTIPOLYGON(((0 3,1 3,0 4,0 3)),((10 0,11 0,10 1,10 0)))", true, 3},

		{"POINT EMPTY", "GEOMETRYCOLLECTION EMPTY", false, 0},
		{"POINT EMPTY", "GEOMETRYCOLLECTION(POINT EMPTY)", false, 0},
		{"POINT EMPTY", "GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT EMPTY))", false, 0},
		{"POINT EMPTY", "GEOMETRYCOLLECTION(POINT(0 0))", false, 0},
		{"POINT EMPTY", "GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT(0 0)))", false, 0},
		{"POINT(0 0)", "GEOMETRYCOLLECTION EMPTY", false, 0},
		{"POINT(0 0)", "GEOMETRYCOLLECTION(POINT EMPTY)", false, 0},
		{"POINT(0 0)", "GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT EMPTY))", false, 0},
		{"POINT(0 0)", "GEOMETRYCOLLECTION(POINT(1 1))", true, math.Sqrt(2)},
		{"POINT(0 0)", "GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT(1 1)))", true, math.Sqrt(2)},

		{"LINESTRING EMPTY", "LINESTRING EMPTY", false, 0},
		{"LINESTRING EMPTY", "LINESTRING(0 0,1 1)", false, 0},
		{"LINESTRING(0 0,1 1)", "LINESTRING EMPTY", false, 0},
		{"LINESTRING(0 0,4 0)", "LINESTRING(1 1,3 1)", true, 1},
		{"LINESTRING(0 0,4 0)", "LINESTRING(1 1,2 2)", true, 1},
		{"LINESTRING(0 0,4 0)", "LINESTRING(3 1,2 2)", true, 1},
		{"LINESTRING(0 0,4 0)", "LINESTRING(-1 1,-1 2)", true, math.Sqrt(2)},
		{"LINESTRING(0 0,4 0)", "LINESTRING(5 1,5 2)", true, math.Sqrt(2)},

		{"POLYGON((1 2,0 1,1 0,2 1,1 2))", "POLYGON((4 2,3 1,4 0,5 1,4 2))", true, 1},
		{"MULTIPOLYGON(((1 2,0 1,1 0,2 1,1 2)),((1 5,0 4,1 3,2 4,1 5)))", "POLYGON((4 2,3 1,4 0,5 1,4 2))", true, 1},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			for _, flip := range []bool{false, true} {
				desc := "fwd"
				if flip {
					desc = "rev"
				}
				t.Run(desc, func(t *testing.T) {
					g1 := geomFromWKT(t, tt.wkt1)
					g2 := geomFromWKT(t, tt.wkt2)
					if flip {
						g1, g2 = g2, g1
					}
					gotDist, gotOK := geom.Distance(g1, g2)
					if gotOK != tt.wantOK {
						t.Logf("WKT1: %s", tt.wkt1)
						t.Logf("WKT2: %s", tt.wkt2)
						t.Logf("want ok: %v", tt.wantOK)
						t.Logf("got ok:  %v", gotOK)
						t.Fatal("mismatch")
					}
					if !gotOK {
						return
					}
					if math.Abs(gotDist-tt.wantDist) > 1e-15 {
						t.Logf("WKT1: %s", tt.wkt1)
						t.Logf("WKT2: %s", tt.wkt2)
						t.Logf("want distance: %f", tt.wantDist)
						t.Logf("got distance:  %f", gotDist)
						t.Error("mismatch")
					}
				})
			}
		})
	}
}
