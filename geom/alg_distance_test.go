package geom_test

import (
	"math"
	"strconv"
	"testing"
)

func TestDistanceNonEmpty(t *testing.T) {
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
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g1 := geomFromWKT(t, tt.wkt1)
			g2 := geomFromWKT(t, tt.wkt2)
			gotDist, gotOK := g1.Distance(g2)
			if gotOK != tt.wantOK {
				t.Logf("want ok: %v", tt.wantOK)
				t.Logf("got ok:  %v", gotOK)
				t.Fatal("mismatch")
			}
			if !gotOK {
				return
			}
			if gotDist != tt.wantDist {
				t.Logf("WKT1: %s", tt.wkt1)
				t.Logf("WKT2: %s", tt.wkt2)
				t.Logf("want distance: %f", tt.wantDist)
				t.Logf("got distance:  %f", gotDist)
				t.Error("mismatch")
			}
		})
	}
}
