package libgeos

import (
	"strconv"
	"strings"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func geomFromWKT(t *testing.T, wkt string) geom.Geometry {
	t.Helper()
	geom, err := geom.UnmarshalWKT(strings.NewReader(wkt))
	if err != nil {
		t.Fatalf("could not unmarshal WKT:\n  wkt: %s\n  err: %v", wkt, err)
	}
	return geom
}

func expectNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// These tests aren't exhaustive, because we are leaveraging libgeos.  The
// testing is just enough to make use confident that we're invoking libgeos
// correctly.

func TestRelate(t *testing.T) {
	for i, tt := range []struct {
		wkt1, wkt2 string
		equals     bool
		disjoint   bool
		touches    bool
	}{
		{
			wkt1:     "POINT EMPTY",
			wkt2:     "POINT EMPTY",
			equals:   true,
			disjoint: true,
			touches:  false,
		},
		{
			wkt1:     "POINT EMPTY",
			wkt2:     "POINT(1 2)",
			equals:   false,
			disjoint: true,
			touches:  false,
		},
		{
			wkt1:     "POINT(1 2)",
			wkt2:     "POINT(1 2)",
			equals:   true,
			disjoint: false,
			touches:  false,
		},
		{
			wkt1:     "POINT(1 2)",
			wkt2:     "POINT(1 3)",
			equals:   false,
			disjoint: true,
			touches:  false,
		},
		{
			wkt1:     "POINT Z(1 2 3)",
			wkt2:     "POINT(1 2)",
			equals:   true,
			disjoint: false,
			touches:  false,
		},
		{
			wkt1:     "POINT M(1 2 3)",
			wkt2:     "POINT(1 2)",
			equals:   true,
			disjoint: false,
			touches:  false,
		},
		{
			wkt1:     "POINT Z(1 2 3)",
			wkt2:     "POINT M(1 2 3)",
			equals:   true,
			disjoint: false,
			touches:  false,
		},
		{
			wkt1:     "LINESTRING EMPTY",
			wkt2:     "LINESTRING EMPTY",
			equals:   true,
			disjoint: true,
			touches:  false,
		},
		{
			wkt1:     "LINESTRING(0 0,2 2)",
			wkt2:     "LINESTRING(0 0,1 1,2 2)",
			equals:   true,
			disjoint: false,
			touches:  false,
		},
		{
			wkt1:     "LINESTRING(0 0,3 3)",
			wkt2:     "LINESTRING(0 0,1 1,2 2)",
			equals:   false,
			disjoint: false,
			touches:  false,
		},
		{
			wkt1:     "LINESTRING(1 0,1 1)",
			wkt2:     "LINESTRING(2 1,2 2)",
			equals:   false,
			disjoint: true,
			touches:  false,
		},
		{
			wkt1:     "LINESTRING(0 0,1 1)",
			wkt2:     "LINESTRING(2 2,1 1)",
			equals:   false,
			disjoint: false,
			touches:  true,
		},
		{
			wkt1:     "POLYGON EMPTY",
			wkt2:     "POLYGON EMPTY",
			equals:   true,
			disjoint: true,
			touches:  false,
		},
		{
			wkt1:     "POLYGON EMPTY",
			wkt2:     "POLYGON((0 0,0 1,1 0,0 0))",
			equals:   false,
			disjoint: true,
			touches:  false,
		},
		{
			wkt1:     "POLYGON((1 0,0 1,0 0,1 0))",
			wkt2:     "POLYGON((0 0,0 1,1 0,0 0))",
			equals:   true,
			disjoint: false,
			touches:  false,
		},
		{
			wkt1:     "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			wkt2:     "POLYGON((2 2,2 3,3 3,3 2,2 2))",
			equals:   false,
			disjoint: true,
			touches:  false,
		},
		{
			wkt1:     "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			wkt2:     "POLYGON((2 2,2 3,3 3,3 2,2 2))",
			equals:   false,
			disjoint: true,
			touches:  false,
		},
		{
			wkt1:     "POLYGON((0 0,0 2,2 2,2 0,0 0))",
			wkt2:     "POLYGON((1 1,1 3,3 3,3 1,1 1))",
			equals:   false,
			disjoint: false,
			touches:  false,
		},
		{
			wkt1:     "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			wkt2:     "POLYGON((0 1,0 2,1 2,1 1,0 1))",
			equals:   false,
			disjoint: false,
			touches:  true,
		},
		{
			wkt1:     "MULTILINESTRING((0 0,1 1))",
			wkt2:     "LINESTRING(0 0,1 1)",
			equals:   true,
			disjoint: false,
			touches:  false,
		},
		{
			wkt1:     "MULTILINESTRING((0 0,1 1),(1 1,2 2))",
			wkt2:     "LINESTRING(0 0,1 1,2 2)",
			equals:   true,
			disjoint: false,
			touches:  false,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g1 := geomFromWKT(t, tt.wkt1)
			g2 := geomFromWKT(t, tt.wkt2)
			t.Run("Equals", func(t *testing.T) {
				got, err := Equals(g1, g2)
				expectNoErr(t, err)
				if got != tt.equals {
					t.Logf("WKT1: %v", tt.wkt1)
					t.Logf("WKT2: %v", tt.wkt2)
					t.Errorf("got: %v want: %v", got, tt.equals)
				}
			})
			t.Run("Disjoint", func(t *testing.T) {
				got, err := Disjoint(g1, g2)
				expectNoErr(t, err)
				if got != tt.disjoint {
					t.Logf("WKT1: %v", tt.wkt1)
					t.Logf("WKT2: %v", tt.wkt2)
					t.Errorf("got: %v want: %v", got, tt.disjoint)
				}
			})
			t.Run("Touches", func(t *testing.T) {
				got, err := Touches(g1, g2)
				expectNoErr(t, err)
				if got != tt.touches {
					t.Logf("WKT1: %v", tt.wkt1)
					t.Logf("WKT2: %v", tt.wkt2)
					t.Errorf("got: %v want: %v", got, tt.touches)
				}
			})
		})
	}
}
