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

func expectGeomEq(t *testing.T, got, want geom.Geometry, opts ...geom.EqualsExactOption) {
	t.Helper()
	if !got.EqualsExact(want, opts...) {
		t.Errorf("\ngot:  %v\nwant: %v\n", got.AsText(), want.AsText())
	}
}

func TestRelease(t *testing.T) {
	h, err := NewHandle()
	expectNoErr(t, err)
	h.Release()
}

// These tests aren't exhaustive, because we are leveraging libgeos.  The
// testing is just enough to make use confident that we're invoking libgeos
// correctly.

func TestRelate(t *testing.T) {
	for i, tt := range []struct {
		wkt1, wkt2 string
		equals     bool
		disjoint   bool
		touches    bool
		contains   bool
		covers     bool
		intersects bool
		within     bool
		coveredBy  bool
	}{
		{
			wkt1:       "POINT EMPTY",
			wkt2:       "POINT EMPTY",
			equals:     true,
			disjoint:   true,
			touches:    false,
			contains:   false,
			covers:     false,
			intersects: false,
			within:     false,
			coveredBy:  false,
		},
		{
			wkt1:       "POINT EMPTY",
			wkt2:       "POINT(1 2)",
			equals:     false,
			disjoint:   true,
			touches:    false,
			contains:   false,
			covers:     false,
			intersects: false,
			within:     false,
			coveredBy:  false,
		},
		{
			wkt1:       "POINT(1 2)",
			wkt2:       "POINT(1 2)",
			equals:     true,
			disjoint:   false,
			touches:    false,
			contains:   true,
			covers:     true,
			intersects: true,
			within:     true,
			coveredBy:  true,
		},
		{
			wkt1:       "POINT(1 2)",
			wkt2:       "POINT(1 3)",
			equals:     false,
			disjoint:   true,
			touches:    false,
			contains:   false,
			covers:     false,
			intersects: false,
			within:     false,
			coveredBy:  false,
		},
		{
			wkt1:       "POINT Z(1 2 3)",
			wkt2:       "POINT(1 2)",
			equals:     true,
			disjoint:   false,
			touches:    false,
			contains:   true,
			covers:     true,
			intersects: true,
			within:     true,
			coveredBy:  true,
		},
		{
			wkt1:       "POINT M(1 2 3)",
			wkt2:       "POINT(1 2)",
			equals:     true,
			disjoint:   false,
			touches:    false,
			contains:   true,
			covers:     true,
			intersects: true,
			within:     true,
			coveredBy:  true,
		},
		{
			wkt1:       "POINT Z(1 2 3)",
			wkt2:       "POINT M(1 2 3)",
			equals:     true,
			disjoint:   false,
			touches:    false,
			contains:   true,
			covers:     true,
			intersects: true,
			within:     true,
			coveredBy:  true,
		},
		{
			wkt1:       "LINESTRING EMPTY",
			wkt2:       "LINESTRING EMPTY",
			equals:     true,
			disjoint:   true,
			touches:    false,
			contains:   false,
			covers:     false,
			intersects: false,
			within:     false,
			coveredBy:  false,
		},
		{
			wkt1:       "LINESTRING(0 0,2 2)",
			wkt2:       "LINESTRING(0 0,1 1,2 2)",
			equals:     true,
			disjoint:   false,
			touches:    false,
			contains:   true,
			covers:     true,
			intersects: true,
			within:     true,
			coveredBy:  true,
		},
		{
			wkt1:       "LINESTRING(0 0,3 3)",
			wkt2:       "LINESTRING(0 0,1 1,2 2)",
			equals:     false,
			disjoint:   false,
			touches:    false,
			contains:   true,
			covers:     true,
			intersects: true,
			within:     false,
			coveredBy:  false,
		},
		{
			wkt1:       "LINESTRING(1 0,1 1)",
			wkt2:       "LINESTRING(2 1,2 2)",
			equals:     false,
			disjoint:   true,
			touches:    false,
			contains:   false,
			covers:     false,
			intersects: false,
			within:     false,
			coveredBy:  false,
		},
		{
			wkt1:       "LINESTRING(0 0,1 1)",
			wkt2:       "LINESTRING(2 2,1 1)",
			equals:     false,
			disjoint:   false,
			touches:    true,
			contains:   false,
			covers:     false,
			intersects: true,
			within:     false,
			coveredBy:  false,
		},
		{
			wkt1:       "POLYGON EMPTY",
			wkt2:       "POLYGON EMPTY",
			equals:     true,
			disjoint:   true,
			touches:    false,
			contains:   false,
			covers:     false,
			intersects: false,
			within:     false,
			coveredBy:  false,
		},
		{
			wkt1:       "POLYGON EMPTY",
			wkt2:       "POLYGON((0 0,0 1,1 0,0 0))",
			equals:     false,
			disjoint:   true,
			touches:    false,
			contains:   false,
			covers:     false,
			intersects: false,
			within:     false,
			coveredBy:  false,
		},
		{
			wkt1:       "POLYGON((1 0,0 1,0 0,1 0))",
			wkt2:       "POLYGON((0 0,0 1,1 0,0 0))",
			equals:     true,
			disjoint:   false,
			touches:    false,
			contains:   true,
			covers:     true,
			intersects: true,
			within:     true,
			coveredBy:  true,
		},
		{
			wkt1:       "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			wkt2:       "POLYGON((2 2,2 3,3 3,3 2,2 2))",
			equals:     false,
			disjoint:   true,
			touches:    false,
			contains:   false,
			covers:     false,
			intersects: false,
			within:     false,
			coveredBy:  false,
		},
		{
			wkt1:       "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			wkt2:       "POLYGON((2 2,2 3,3 3,3 2,2 2))",
			equals:     false,
			disjoint:   true,
			touches:    false,
			contains:   false,
			covers:     false,
			intersects: false,
			within:     false,
			coveredBy:  false,
		},
		{
			wkt1:       "POLYGON((0 0,0 2,2 2,2 0,0 0))",
			wkt2:       "POLYGON((1 1,1 3,3 3,3 1,1 1))",
			equals:     false,
			disjoint:   false,
			touches:    false,
			contains:   false,
			covers:     false,
			intersects: true,
			within:     false,
			coveredBy:  false,
		},
		{
			wkt1:       "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			wkt2:       "POLYGON((0 1,0 2,1 2,1 1,0 1))",
			equals:     false,
			disjoint:   false,
			touches:    true,
			contains:   false,
			covers:     false,
			intersects: true,
			within:     false,
			coveredBy:  false,
		},
		{
			wkt1:       "POLYGON((0 0,0 3,3 3,3 0,0 0))",
			wkt2:       "POLYGON((1 1,1 2,2 2,2 1,1 1))",
			equals:     false,
			disjoint:   false,
			touches:    false,
			contains:   true,
			covers:     true,
			intersects: true,
			within:     false,
			coveredBy:  false,
		},
		{
			wkt1:       "POLYGON((1 1,1 2,2 2,2 1,1 1))",
			wkt2:       "POLYGON((0 0,0 3,3 3,3 0,0 0))",
			equals:     false,
			disjoint:   false,
			touches:    false,
			contains:   false,
			covers:     false,
			intersects: true,
			within:     true,
			coveredBy:  true,
		},
		{
			wkt1:       "MULTILINESTRING((0 0,1 1))",
			wkt2:       "LINESTRING(0 0,1 1)",
			equals:     true,
			disjoint:   false,
			touches:    false,
			contains:   true,
			covers:     true,
			intersects: true,
			within:     true,
			coveredBy:  true,
		},
		{
			wkt1:       "MULTILINESTRING((0 0,1 1),(1 1,2 2))",
			wkt2:       "LINESTRING(0 0,1 1,2 2)",
			equals:     true,
			disjoint:   false,
			touches:    false,
			contains:   true,
			covers:     true,
			intersects: true,
			within:     true,
			coveredBy:  true,
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
			t.Run("Contains", func(t *testing.T) {
				got, err := Contains(g1, g2)
				expectNoErr(t, err)
				if got != tt.contains {
					t.Logf("WKT1: %v", tt.wkt1)
					t.Logf("WKT2: %v", tt.wkt2)
					t.Errorf("got: %v want: %v", got, tt.contains)
				}
			})
			t.Run("Covers", func(t *testing.T) {
				got, err := Covers(g1, g2)
				expectNoErr(t, err)
				if got != tt.covers {
					t.Logf("WKT1: %v", tt.wkt1)
					t.Logf("WKT2: %v", tt.wkt2)
					t.Errorf("got: %v want: %v", got, tt.covers)
				}
			})
			t.Run("Intersects", func(t *testing.T) {
				got, err := Intersects(g1, g2)
				expectNoErr(t, err)
				if got != tt.intersects {
					t.Logf("WKT1: %v", tt.wkt1)
					t.Logf("WKT2: %v", tt.wkt2)
					t.Errorf("got: %v want: %v", got, tt.intersects)
				}
			})
			t.Run("Within", func(t *testing.T) {
				got, err := Within(g1, g2)
				expectNoErr(t, err)
				if got != tt.within {
					t.Logf("WKT1: %v", tt.wkt1)
					t.Logf("WKT2: %v", tt.wkt2)
					t.Errorf("got: %v want: %v", got, tt.within)
				}
			})
			t.Run("CoveredBy", func(t *testing.T) {
				got, err := CoveredBy(g1, g2)
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
					got, err := Crosses(g1, g2)
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
					got, err := Overlaps(g1, g2)
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

type BinaryOperationTestCase struct {
	In1, In2 string
	Out      string
}

func RunBinaryOperationTest(t *testing.T, fn func(a, b geom.Geometry) (geom.Geometry, error), cases []BinaryOperationTestCase) {
	for i, c := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			run := func(rev bool) func(t *testing.T) {
				return func(t *testing.T) {
					g1 := geomFromWKT(t, c.In1)
					g2 := geomFromWKT(t, c.In2)
					if rev {
						g1, g2 = g2, g1
					}
					t.Logf("WKT1: %v", g1.AsText())
					t.Logf("WKT2: %v", g2.AsText())
					got, err := fn(g1, g2)
					expectNoErr(t, err)
					expectGeomEq(t, got, geomFromWKT(t, c.Out), geom.IgnoreOrder)
				}
			}
			t.Run("Forward", run(false))
			t.Run("Reverse", run(true))
		})
	}
}

func TestUnion(t *testing.T) {
	RunBinaryOperationTest(t, Union, []BinaryOperationTestCase{
		{
			"POINT(1 2)",
			"POINT(3 4)",
			"MULTIPOINT(1 2,3 4)",
		},
		{
			"POINT EMPTY",
			"POINT(3 4)",
			"POINT(3 4)",
		},
		{
			"POINT EMPTY",
			"POINT EMPTY",
			"POINT EMPTY",
		},
		{
			"MULTIPOINT(EMPTY)",
			"MULTIPOINT(EMPTY)",
			"MULTIPOINT EMPTY",
		},
		{
			"GEOMETRYCOLLECTION(POINT EMPTY)",
			"GEOMETRYCOLLECTION(POINT EMPTY)",
			"GEOMETRYCOLLECTION EMPTY",
		},
		{
			"POLYGON((0 0,0 2,2 2,2 0,0 0))",
			"POLYGON((1 1,1 3,3 3,3 1,1 1))",
			"POLYGON((0 0,2 0,2 1,3 1,3 3,1 3,1 2,0 2,0 0))",
		},
		{
			"POLYGON((0 0,0 3,3 3,3 0,0 0),(1 1,1 2,2 2,2 1,1 1))",
			"POLYGON((1 1,1 2,2 2,2 1,1 1))",
			"POLYGON((0 0,0 3,3 3,3 0,0 0))",
		},
		{
			"GEOMETRYCOLLECTION(POINT(0 0),POLYGON((0 1,1 1,1 2,0 2,0 1)))",
			"LINESTRING(1 0,1 1,0 2)",
			"GEOMETRYCOLLECTION(POINT(0 0),LINESTRING(1 0,1 1),POLYGON((0 1,1 1,1 2,0 2,0 1)))",
		},
	})
}

func TestIntersection(t *testing.T) {
	RunBinaryOperationTest(t, Intersection, []BinaryOperationTestCase{
		{"POINT EMPTY", "POINT EMPTY", "GEOMETRYCOLLECTION EMPTY"},
		{"POINT(1 2)", "POINT EMPTY", "GEOMETRYCOLLECTION EMPTY"},
		{"POINT(1 2)", "POINT(1 2)", "POINT(1 2)"},
		{
			"POLYGON((0 0,3 0,3 3,2 3,2 1,0 1,0 0))",
			"POLYGON((0 0,0 3,3 3,3 2,1 2,1 0,0 0))",
			"MULTIPOLYGON(((0 0,1 0,1 1,0 1,0 0)),((2 2,2 3,3 3,3 2,2 2)))",
		},
	})
}

func TestBuffer(t *testing.T) {
	for i, tt := range []struct {
		wkt    string
		radius float64
		want   string
	}{
		// The following test cases were generated by taking the input, and
		// seeing what the output from Buffer is. It was then verified visually
		// to be correct.
		{
			"POINT(0 0)", 1,
			"POLYGON((1 0,0.9807852804032305 -0.19509032201612808,0.923879532511287 -0.3826834323650894,0.8314696123025456 -0.5555702330196017,0.7071067811865481 -0.7071067811865469,0.5555702330196031 -0.8314696123025447,0.38268343236509084 -0.9238795325112863,0.19509032201612964 -0.9807852804032302,0.0000000000000016155445744325867 -1,-0.19509032201612647 -0.9807852804032308,-0.38268343236508784 -0.9238795325112875,-0.5555702330196005 -0.8314696123025463,-0.7071067811865459 -0.7071067811865491,-0.8314696123025438 -0.5555702330196043,-0.9238795325112857 -0.38268343236509234,-0.9807852804032299 -0.19509032201613122,-1 -0.0000000000000032310891488651735,-0.9807852804032311 0.19509032201612486,-0.9238795325112882 0.38268343236508634,-0.8314696123025475 0.555570233019599,-0.7071067811865505 0.7071067811865446,-0.5555702330196058 0.8314696123025428,-0.3826834323650936 0.9238795325112852,-0.19509032201613213 0.9807852804032297,-0.000000000000003736410698672604 1,0.1950903220161248 0.9807852804032311,0.38268343236508673 0.9238795325112881,0.5555702330195996 0.8314696123025469,0.7071067811865455 0.7071067811865496,0.8314696123025438 0.5555702330196044,0.9238795325112859 0.38268343236509206,0.98078528040323 0.19509032201613047,1 0))",
		},
		{
			"LINESTRING(10 10,10 0,0 0)", 2,
			"POLYGON((12 0,11.96157056080646 -0.3901806440322565,11.847759065022574 -0.7653668647301796,11.66293922460509 -1.1111404660392044,11.414213562373096 -1.414213562373095,11.111140466039204 -1.6629392246050905,10.76536686473018 -1.8477590650225735,10.390180644032256 -1.9615705608064609,10 -2,0 -2,-0.39018064403225905 -1.9615705608064604,-0.7653668647301823 -1.8477590650225724,-1.1111404660392072 -1.6629392246050885,-1.4142135623730965 -1.4142135623730936,-1.6629392246050914 -1.111140466039203,-1.847759065022574 -0.7653668647301785,-1.961570560806461 -0.39018064403225583,-2 0.00000000000000024492935982947064,-1.9615705608064609 0.39018064403225633,-1.8477590650225737 0.7653668647301789,-1.6629392246050911 1.1111404660392035,-1.4142135623730963 1.4142135623730938,-1.1111404660392061 1.6629392246050894,-0.7653668647301819 1.8477590650225726,-0.39018064403225944 1.9615705608064604,0 2,8 2,8 10,8.03842943919354 10.39018064403226,8.152240934977428 10.765366864730183,8.33706077539491 11.111140466039206,8.585786437626906 11.414213562373096,8.888859533960797 11.662939224605092,9.23463313526982 11.847759065022574,9.609819355967744 11.96157056080646,10 12,10.390180644032256 11.96157056080646,10.76536686473018 11.847759065022574,11.111140466039203 11.662939224605092,11.414213562373094 11.414213562373096,11.66293922460509 11.111140466039206,11.847759065022572 10.765366864730183,11.96157056080646 10.39018064403226,12 10,12 0))",
		},
		{
			"POLYGON((0 0,10 0,10 10,0 10,0 0),(1 1,9 1,9 9,1 9,1 1))", 0.5,
			"POLYGON((-0.5 0,-0.5 10,-0.4903926402016152 10.097545161008064,-0.46193976625564337 10.191341716182546,-0.4157348061512727 10.2777851165098,-0.35355339059327373 10.353553390593273,-0.277785116509801 10.415734806151272,-0.19134171618254486 10.461939766255643,-0.0975451610080641 10.490392640201616,0 10.5,10 10.5,10.097545161008064 10.490392640201616,10.191341716182546 10.461939766255643,10.2777851165098 10.415734806151272,10.353553390593273 10.353553390593273,10.415734806151272 10.2777851165098,10.461939766255643 10.191341716182546,10.490392640201616 10.097545161008064,10.5 10,10.5 0,10.490392640201616 -0.09754516100806412,10.461939766255643 -0.1913417161825449,10.415734806151272 -0.2777851165098011,10.353553390593273 -0.35355339059327373,10.2777851165098 -0.4157348061512726,10.191341716182546 -0.46193976625564337,10.097545161008064 -0.4903926402016152,10 -0.5,0 -0.5,-0.09754516100806378 -0.49039264020161527,-0.19134171618254414 -0.4619397662556437,-0.2777851165098001 -0.41573480615127334,-0.3535533905932726 -0.3535533905932749,-0.41573480615127156 -0.2777851165098027,-0.4619397662556424 -0.1913417161825472,-0.49039264020161466 -0.09754516100806691,-0.5 0),(1.5 1.5,8.5 1.5,8.5 8.5,1.5 8.5,1.5 1.5))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := geomFromWKT(t, tt.wkt)
			t.Logf("WKT: %v", g.AsText())
			got, err := Buffer(g, tt.radius)
			expectNoErr(t, err)
			expectGeomEq(t, got, geomFromWKT(t, tt.want), geom.IgnoreOrder)
		})
	}
}
