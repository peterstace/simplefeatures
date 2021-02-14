package geom_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

// These tests aren't exhaustive, because we are leveraging the Relate
// function. The testing is just enough to make sure the pattern matching is
// correct and there are no obvious problems.
func TestRelate(t *testing.T) {
	for i, tt := range []struct {
		wkt1, wkt2 string
		equals     bool
		disjoint   bool
		touches    bool
		contains   bool
		covers     bool
		within     bool
		coveredBy  bool
	}{
		{
			wkt1:      "POINT EMPTY",
			wkt2:      "POINT EMPTY",
			equals:    true,
			disjoint:  true,
			touches:   false,
			contains:  false,
			covers:    false,
			within:    false,
			coveredBy: false,
		},
		{
			wkt1:      "POINT EMPTY",
			wkt2:      "POINT(1 2)",
			equals:    false,
			disjoint:  true,
			touches:   false,
			contains:  false,
			covers:    false,
			within:    false,
			coveredBy: false,
		},
		{
			wkt1:      "POINT(1 2)",
			wkt2:      "POINT(1 2)",
			equals:    true,
			disjoint:  false,
			touches:   false,
			contains:  true,
			covers:    true,
			within:    true,
			coveredBy: true,
		},
		{
			wkt1:      "POINT(1 2)",
			wkt2:      "POINT(1 3)",
			equals:    false,
			disjoint:  true,
			touches:   false,
			contains:  false,
			covers:    false,
			within:    false,
			coveredBy: false,
		},
		{
			wkt1:      "POINT Z(1 2 3)",
			wkt2:      "POINT(1 2)",
			equals:    true,
			disjoint:  false,
			touches:   false,
			contains:  true,
			covers:    true,
			within:    true,
			coveredBy: true,
		},
		{
			wkt1:      "POINT M(1 2 3)",
			wkt2:      "POINT(1 2)",
			equals:    true,
			disjoint:  false,
			touches:   false,
			contains:  true,
			covers:    true,
			within:    true,
			coveredBy: true,
		},
		{
			wkt1:      "POINT Z(1 2 3)",
			wkt2:      "POINT M(1 2 3)",
			equals:    true,
			disjoint:  false,
			touches:   false,
			contains:  true,
			covers:    true,
			within:    true,
			coveredBy: true,
		},
		{
			wkt1:      "LINESTRING EMPTY",
			wkt2:      "LINESTRING EMPTY",
			equals:    true,
			disjoint:  true,
			touches:   false,
			contains:  false,
			covers:    false,
			within:    false,
			coveredBy: false,
		},
		{
			wkt1:      "LINESTRING(0 0,2 2)",
			wkt2:      "LINESTRING(0 0,1 1,2 2)",
			equals:    true,
			disjoint:  false,
			touches:   false,
			contains:  true,
			covers:    true,
			within:    true,
			coveredBy: true,
		},
		{
			wkt1:      "LINESTRING(0 0,3 3)",
			wkt2:      "LINESTRING(0 0,1 1,2 2)",
			equals:    false,
			disjoint:  false,
			touches:   false,
			contains:  true,
			covers:    true,
			within:    false,
			coveredBy: false,
		},
		{
			wkt1:      "LINESTRING(1 0,1 1)",
			wkt2:      "LINESTRING(2 1,2 2)",
			equals:    false,
			disjoint:  true,
			touches:   false,
			contains:  false,
			covers:    false,
			within:    false,
			coveredBy: false,
		},
		{
			wkt1:      "LINESTRING(0 0,1 1)",
			wkt2:      "LINESTRING(2 2,1 1)",
			equals:    false,
			disjoint:  false,
			touches:   true,
			contains:  false,
			covers:    false,
			within:    false,
			coveredBy: false,
		},
		{
			wkt1:      "POLYGON EMPTY",
			wkt2:      "POLYGON EMPTY",
			equals:    true,
			disjoint:  true,
			touches:   false,
			contains:  false,
			covers:    false,
			within:    false,
			coveredBy: false,
		},
		{
			wkt1:      "POLYGON EMPTY",
			wkt2:      "POLYGON((0 0,0 1,1 0,0 0))",
			equals:    false,
			disjoint:  true,
			touches:   false,
			contains:  false,
			covers:    false,
			within:    false,
			coveredBy: false,
		},
		{
			wkt1:      "POLYGON((1 0,0 1,0 0,1 0))",
			wkt2:      "POLYGON((0 0,0 1,1 0,0 0))",
			equals:    true,
			disjoint:  false,
			touches:   false,
			contains:  true,
			covers:    true,
			within:    true,
			coveredBy: true,
		},
		{
			wkt1:      "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			wkt2:      "POLYGON((2 2,2 3,3 3,3 2,2 2))",
			equals:    false,
			disjoint:  true,
			touches:   false,
			contains:  false,
			covers:    false,
			within:    false,
			coveredBy: false,
		},
		{
			wkt1:      "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			wkt2:      "POLYGON((2 2,2 3,3 3,3 2,2 2))",
			equals:    false,
			disjoint:  true,
			touches:   false,
			contains:  false,
			covers:    false,
			within:    false,
			coveredBy: false,
		},
		{
			wkt1:      "POLYGON((0 0,0 2,2 2,2 0,0 0))",
			wkt2:      "POLYGON((1 1,1 3,3 3,3 1,1 1))",
			equals:    false,
			disjoint:  false,
			touches:   false,
			contains:  false,
			covers:    false,
			within:    false,
			coveredBy: false,
		},
		{
			wkt1:      "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			wkt2:      "POLYGON((0 1,0 2,1 2,1 1,0 1))",
			equals:    false,
			disjoint:  false,
			touches:   true,
			contains:  false,
			covers:    false,
			within:    false,
			coveredBy: false,
		},
		{
			wkt1:      "POLYGON((0 0,0 3,3 3,3 0,0 0))",
			wkt2:      "POLYGON((1 1,1 2,2 2,2 1,1 1))",
			equals:    false,
			disjoint:  false,
			touches:   false,
			contains:  true,
			covers:    true,
			within:    false,
			coveredBy: false,
		},
		{
			wkt1:      "POLYGON((1 1,1 2,2 2,2 1,1 1))",
			wkt2:      "POLYGON((0 0,0 3,3 3,3 0,0 0))",
			equals:    false,
			disjoint:  false,
			touches:   false,
			contains:  false,
			covers:    false,
			within:    true,
			coveredBy: true,
		},
		{
			wkt1:      "MULTILINESTRING((0 0,1 1))",
			wkt2:      "LINESTRING(0 0,1 1)",
			equals:    true,
			disjoint:  false,
			touches:   false,
			contains:  true,
			covers:    true,
			within:    true,
			coveredBy: true,
		},
		{
			wkt1:      "MULTILINESTRING((0 0,1 1),(1 1,2 2))",
			wkt2:      "LINESTRING(0 0,1 1,2 2)",
			equals:    true,
			disjoint:  false,
			touches:   false,
			contains:  true,
			covers:    true,
			within:    true,
			coveredBy: true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g1 := geomFromWKT(t, tt.wkt1)
			g2 := geomFromWKT(t, tt.wkt2)
			t.Run("Equals", func(t *testing.T) {
				got, err := geom.Equals(g1, g2)
				expectNoErr(t, err)
				if got != tt.equals {
					t.Logf("WKT1: %v", tt.wkt1)
					t.Logf("WKT2: %v", tt.wkt2)
					t.Errorf("got: %v want: %v", got, tt.equals)
				}
			})
			t.Run("Disjoint", func(t *testing.T) {
				got, err := geom.Disjoint(g1, g2)
				expectNoErr(t, err)
				if got != tt.disjoint {
					t.Logf("WKT1: %v", tt.wkt1)
					t.Logf("WKT2: %v", tt.wkt2)
					t.Errorf("got: %v want: %v", got, tt.disjoint)
				}
			})
			t.Run("Touches", func(t *testing.T) {
				got, err := geom.Touches(g1, g2)
				expectNoErr(t, err)
				if got != tt.touches {
					t.Logf("WKT1: %v", tt.wkt1)
					t.Logf("WKT2: %v", tt.wkt2)
					t.Errorf("got: %v want: %v", got, tt.touches)
				}
			})
			t.Run("Contains", func(t *testing.T) {
				got, err := geom.Contains(g1, g2)
				expectNoErr(t, err)
				if got != tt.contains {
					t.Logf("WKT1: %v", tt.wkt1)
					t.Logf("WKT2: %v", tt.wkt2)
					t.Errorf("got: %v want: %v", got, tt.contains)
				}
			})
			t.Run("Covers", func(t *testing.T) {
				got, err := geom.Covers(g1, g2)
				expectNoErr(t, err)
				if got != tt.covers {
					t.Logf("WKT1: %v", tt.wkt1)
					t.Logf("WKT2: %v", tt.wkt2)
					t.Errorf("got: %v want: %v", got, tt.covers)
				}
			})
			t.Run("Within", func(t *testing.T) {
				got, err := geom.Within(g1, g2)
				expectNoErr(t, err)
				if got != tt.within {
					t.Logf("WKT1: %v", tt.wkt1)
					t.Logf("WKT2: %v", tt.wkt2)
					t.Errorf("got: %v want: %v", got, tt.within)
				}
			})
			t.Run("CoveredBy", func(t *testing.T) {
				got, err := geom.CoveredBy(g1, g2)
				expectNoErr(t, err)
				if got != tt.coveredBy {
					t.Logf("WKT1: %v", tt.wkt1)
					t.Logf("WKT2: %v", tt.wkt2)
					t.Errorf("got: %v want: %v", got, tt.coveredBy)
				}
			})
		})
	}
}

func TestCrosses(t *testing.T) {
	for i, tt := range []struct {
		wkt1, wkt2 string
		want       bool
	}{
		// Point/Line
		{"POINT EMPTY", "LINESTRING(1 2,3 4)", false},
		{"POINT(1 2)", "LINESTRING EMPTY", false},
		{"POINT EMPTY", "LINESTRING EMPTY", false},
		{"POINT(1 2)", "LINESTRING(1 2,3 4)", false},
		{"POINT(1 2)", "LINESTRING(1 2,3 4)", false},
		{"POINT(1 2)", "LINESTRING(0 0,2 4)", false},
		{"MULTIPOINT(1 2,3 3)", "LINESTRING(0 0,2 4)", true},

		// Point/Area
		{"POINT EMPTY", "POLYGON((0 0,0 1,1 1,1 0,0 0))", false},
		{"POINT(2 2)", "POLYGON EMPTY", false},
		{"POINT EMPTY", "POLYGON EMPTY", false},
		{"POINT(2 2)", "POLYGON((0 0,0 1,1 1,1 0,0 0))", false},
		{"POINT(0.5 0.5)", "POLYGON((0 0,0 1,1 1,1 0,0 0))", false},
		{"MULTIPOINT(2 2,0.5 0.5)", "POLYGON((0 0,0 1,1 1,1 0,0 0))", true},

		// Line/Area
		{"LINESTRING EMPTY", "POLYGON((1 1,1 4,4 4,4 1,1 1))", false},
		{"LINESTRING(0 3,2 5)", "POLYGON EMPTY", false},
		{"LINESTRING EMPTY", "POLYGON EMPTY", false},
		{"LINESTRING(0 3,2 5)", "POLYGON((1 1,1 4,4 4,4 1,1 1))", false},
		{"LINESTRING(0 4,5 4)", "POLYGON((1 1,1 4,4 4,4 1,1 1))", false},
		{"LINESTRING(0 4,2 6)", "POLYGON((1 1,1 4,4 4,4 1,1 1))", false},
		{"LINESTRING(0 2,3 5)", "POLYGON((1 1,1 4,4 4,4 1,1 1))", true},
		{"LINESTRING(2 3,2 7)", "POLYGON((1 1,1 4,4 4,4 1,1 1))", true},
		{"LINESTRING(2 2,3 3)", "POLYGON((1 1,1 4,4 4,4 1,1 1))", false},
		{"LINESTRING(2 2,4 4)", "POLYGON((1 1,1 4,4 4,4 1,1 1))", false},

		// Area/Point, Area/Line, Line/Point: these are just the reverse of the above cases.
		{"LINESTRING(1 2,3 4)", "POINT(1 2)", false},
		{"LINESTRING(0 0,2 4)", "POINT(1 2)", false},
		{"LINESTRING(0 0,2 4)", "MULTIPOINT(1 2,3 3)", true},
		{"POLYGON((0 0,0 1,1 1,1 0,0 0))", "POINT(2 2)", false},
		{"POLYGON((0 0,0 1,1 1,1 0,0 0))", "POINT(0.5 0.5)", false},
		{"POLYGON((0 0,0 1,1 1,1 0,0 0))", "MULTIPOINT(2 2,0.5 0.5)", true},
		{"POLYGON((1 1,1 4,4 4,4 1,1 1))", "LINESTRING(0 3,2 5)", false},
		{"POLYGON((1 1,1 4,4 4,4 1,1 1))", "LINESTRING(0 4,5 4)", false},
		{"POLYGON((1 1,1 4,4 4,4 1,1 1))", "LINESTRING(0 4,2 6)", false},
		{"POLYGON((1 1,1 4,4 4,4 1,1 1))", "LINESTRING(0 2,3 5)", true},
		{"POLYGON((1 1,1 4,4 4,4 1,1 1))", "LINESTRING(2 3,2 7)", true},
		{"POLYGON((1 1,1 4,4 4,4 1,1 1))", "LINESTRING(2 2,3 3)", false},
		{"POLYGON((1 1,1 4,4 4,4 1,1 1))", "LINESTRING(2 2,4 4)", false},

		{"POLYGON((1 1,1 4,4 4,4 1,1 1))", "LINESTRING EMPTY", false},
		{"POLYGON EMPTY", "LINESTRING(0 3,2 5)", false},
		{"POLYGON EMPTY", "LINESTRING EMPTY", false},
		{"POLYGON((0 0,0 1,1 1,1 0,0 0))", "POINT EMPTY", false},
		{"POLYGON EMPTY", "POINT(2 2)", false},
		{"POLYGON EMPTY", "POINT EMPTY", false},
		{"LINESTRING(1 2,3 4)", "POINT EMPTY", false},
		{"LINESTRING EMPTY", "POINT(1 2)", false},
		{"LINESTRING EMPTY", "POINT EMPTY", false},

		// Line/Line
		{"LINESTRING(0 0,0 1)", "LINESTRING EMPTY", false},
		{"LINESTRING EMPTY", "LINESTRING EMPTY", false},
		{"LINESTRING EMPTY", "LINESTRING(0 0,0 1)", false},
		{"LINESTRING(0 0,0 1)", "LINESTRING(1 0,1 1)", false},
		{"LINESTRING(0 0,0 1)", "LINESTRING(0 0,1 0)", false},
		{"LINESTRING(0 0,0 2)", "LINESTRING(0 1,1 1)", false},
		{"LINESTRING(0 0,1 1)", "LINESTRING(0 1,1 0)", true},
		{"LINESTRING(0 0,2 2)", "LINESTRING(1 1,3 3)", false},

		// Other (Point/Point, Area/Area)
		{"POINT(0 0)", "POINT(0 0)", false},
		{"POINT(2 1)", "POINT(0 0)", false},
		{"POLYGON((0 5,10 5,10 6,0 6,0 5))", "POLYGON((5 0,5 10,6 10,6 0,5 0))", false},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			run := func(rev bool) func(*testing.T) {
				return func(t *testing.T) {
					g1 := geomFromWKT(t, tt.wkt1)
					g2 := geomFromWKT(t, tt.wkt2)
					if rev {
						g1, g2 = g2, g1
					}
					got, err := geom.Crosses(g1, g2)
					expectNoErr(t, err)
					if got != tt.want {
						t.Logf("WKT1: %v", tt.wkt1)
						t.Logf("WKT2: %v", tt.wkt2)
						t.Errorf("got: %v want: %v", got, tt.want)
					}
				}
			}
			t.Run("Forward", run(false))
			t.Run("Reverse", run(true))
		})
	}
}

func TestOverlaps(t *testing.T) {
	for i, tt := range []struct {
		wkt1, wkt2 string
		want       bool
	}{
		// Point/Point
		{"POINT EMPTY", "POINT(1 2)", false},
		{"POINT(1 2)", "POINT(1 2)", false},
		{"POINT(1 2)", "MULTIPOINT(1 2,2 3)", false},
		{"MULTIPOINT(1 2,4 5)", "MULTIPOINT(1 2,2 3)", true},

		// Line/Line
		{"LINESTRING EMPTY", "LINESTRING EMPTY", false},
		{"LINESTRING EMPTY", "LINESTRING(0 1,1 1)", false},
		{"LINESTRING(0 0,1 0)", "LINESTRING(0 1,1 1)", false},
		{"LINESTRING(0 0,1 0)", "LINESTRING(0 0,0 1)", false},
		{"LINESTRING(0 0,1 0)", "LINESTRING(0.5 0,0.5 1)", false},
		{"LINESTRING(0 0,1 1)", "LINESTRING(0 1,1 0)", false},
		{"LINESTRING(0 0,2 2)", "LINESTRING(1 1,3 3)", true},

		// Area/Area
		{"POLYGON((0 0,0 1,1 1,1 0,0 0))", "POLYGON((2 2,2 3,3 3,3 2,2 2))", false},
		{"POLYGON((0 0,0 1,1 1,1 0,0 0))", "POLYGON((1 1,1 2,2 2,2 1,1 1))", false},
		{"POLYGON((0 0,0 1,1 1,1 0,0 0))", "POLYGON((0 1,0 2,1 2,1 1,0 1))", false},
		{"POLYGON((0 0,0 2,2 2,2 0,0 0))", "POLYGON((1 1,1 3,3 3,3 1,1 1))", true},
		{"POLYGON((0 0,0 2,2 2,2 0,0 0))", "POLYGON((0 0,0 2,2 2,2 0,0 0))", false},
		{"POLYGON((0 0,0 3,3 3,3 0,0 0))", "POLYGON((1 1,1 2,2 2,2 1,1 1))", false},

		// Mixed dimension
		{"POINT(0.5 0.5)", "LINESTRING(0 0,1 1)", false},
		{"POINT(0.5 0.5)", "POLYGON((0 0,0 1,1 1,1 0,0 0))", false},
		{"LINESTRING(0 0,1 1)", "POLYGON((0 0,0 1,1 1,1 0,0 0))", false},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			run := func(rev bool) func(t *testing.T) {
				return func(t *testing.T) {
					g1 := geomFromWKT(t, tt.wkt1)
					g2 := geomFromWKT(t, tt.wkt2)
					if rev {
						g1, g2 = g2, g1
					}
					got, err := geom.Overlaps(g1, g2)
					expectNoErr(t, err)
					if got != tt.want {
						t.Logf("WKT1: %v", tt.wkt1)
						t.Logf("WKT2: %v", tt.wkt2)
						t.Errorf("got: %v want: %v", got, tt.want)
					}
				}
			}
			t.Run("Forward", run(false))
			t.Run("Reverse", run(true))
		})
	}
}
