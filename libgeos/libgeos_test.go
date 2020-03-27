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

func TestEquals(t *testing.T) {
	for i, tt := range []struct {
		wkt1, wkt2 string
		want       bool
	}{

		{"POINT EMPTY", "POINT EMPTY", true},
		{"POINT EMPTY", "POINT(1 2)", false},
		{"POINT(1 2)", "POINT(1 2)", true},
		{"POINT(1 2)", "POINT(1 3)", false},

		{"POINT Z(1 2 3)", "POINT(1 2)", true},
		{"POINT M(1 2 3)", "POINT(1 2)", true},
		{"POINT Z(1 2 3)", "POINT M(1 2 3)", true},

		{"LINESTRING EMPTY", "LINESTRING EMPTY", true},
		{"LINESTRING(0 0,2 2)", "LINESTRING(0 0,1 1,2 2)", true},
		{"LINESTRING(0 0,3 3)", "LINESTRING(0 0,1 1,2 2)", false},

		{"POLYGON EMPTY", "POLYGON EMPTY", true},
		{"POLYGON EMPTY", "POLYGON((0 0,0 1,1 0,0 0))", false},
		{"POLYGON((1 0,0 1,0 0,1 0))", "POLYGON((0 0,0 1,1 0,0 0))", true},

		{"MULTILINESTRING((0 0,1 1))", "LINESTRING(0 0,1 1)", true},
		{"MULTILINESTRING((0 0,1 1),(1 1,2 2))", "LINESTRING(0 0,1 1,2 2)", true},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g1 := geomFromWKT(t, tt.wkt1)
			g2 := geomFromWKT(t, tt.wkt2)
			got, err := Equals(g1, g2)
			expectNoErr(t, err)
			if got != tt.want {
				t.Logf("WKT1: %v", tt.wkt1)
				t.Logf("WKT2: %v", tt.wkt2)
				t.Errorf("got: %v want: %v", got, tt.want)
			}
		})
	}
}
