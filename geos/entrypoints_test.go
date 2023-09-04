package geos

import (
	"math"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func geomFromWKT(t *testing.T, wkt string) geom.Geometry {
	t.Helper()
	geom, err := geom.UnmarshalWKT(wkt)
	if err != nil {
		t.Fatalf("could not unmarshal WKT:\n  wkt: %s\n  err: %v", wkt, err)
	}
	return geom
}

func geomFromWKTWithoutValidation(t *testing.T, wkt string) geom.Geometry {
	t.Helper()
	geom, err := geom.UnmarshalWKTWithoutValidation(wkt)
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

func expectErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("unexpected error but got nil")
	}
}

func expectGeomEq(t *testing.T, got, want geom.Geometry, opts ...geom.ExactEqualsOption) {
	t.Helper()
	if !geom.ExactEquals(got, want, opts...) {
		t.Errorf("\ngot:  %v\nwant: %v\n", got.AsText(), want.AsText())
	}
}

// These tests aren't exhaustive, because we are leveraging GEOS.  The
// testing is just enough to make use confident that we're invoking GEOS
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

func TestRelateCode(t *testing.T) {
	g1 := geomFromWKT(t, "POLYGON((0 0,0 2,2 2,2 0,0 0))")
	g2 := geomFromWKT(t, "POLYGON((1 1,1 3,3 3,3 1,1 1))")
	got, err := Relate(g1, g2)
	expectNoErr(t, err)
	const want = "212101212"
	if got != want {
		t.Errorf("got: %v want: %v", got, want)
	}
}

type BinaryOperationTestCase struct {
	In1, In2 string
	Out      string
}

func RunBinaryOperationTest(t *testing.T, fn func(a, b geom.Geometry, opts ...geom.ConstructorOption) (geom.Geometry, error), cases []BinaryOperationTestCase) {
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

					if got.IsEmpty() {
						// Normalise the result to a geometry collection when
						// it's empty. This is needed because different
						// versions of GEOS will give different geometry types
						// for empty geometries.
						got = geom.GeometryCollection{}.AsGeometry()
					}
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
			"GEOMETRYCOLLECTION EMPTY",
		},
		{
			"MULTIPOINT(EMPTY)",
			"MULTIPOINT(EMPTY)",
			"GEOMETRYCOLLECTION EMPTY",
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
		opts   []BufferOption
		want   string
	}{
		// The following test cases were generated by taking the input, and
		// seeing what the output from Buffer is. It was then verified visually
		// to be correct.
		{
			"POINT(0 0)", 1, nil,
			"POLYGON((1 0,0.9807852804032304 -0.19509032201612825,0.9238795325112867 -0.3826834323650898,0.8314696123025452 -0.5555702330196022,0.7071067811865476 -0.7071067811865475,0.5555702330196023 -0.8314696123025452,0.38268343236508984 -0.9238795325112867,0.19509032201612833 -0.9807852804032304,0.00000000000000006123233995736766 -1,-0.1950903220161282 -0.9807852804032304,-0.3826834323650897 -0.9238795325112867,-0.555570233019602 -0.8314696123025455,-0.7071067811865475 -0.7071067811865476,-0.8314696123025453 -0.5555702330196022,-0.9238795325112867 -0.3826834323650899,-0.9807852804032304 -0.1950903220161286,-1 -0.00000000000000012246467991473532,-0.9807852804032304 0.19509032201612836,-0.9238795325112868 0.38268343236508967,-0.8314696123025455 0.555570233019602,-0.7071067811865477 0.7071067811865475,-0.5555702330196022 0.8314696123025452,-0.38268343236509034 0.9238795325112865,-0.19509032201612866 0.9807852804032303,-0.00000000000000018369701987210297 1,0.1950903220161283 0.9807852804032304,0.38268343236509 0.9238795325112866,0.5555702330196018 0.8314696123025455,0.7071067811865474 0.7071067811865477,0.8314696123025452 0.5555702330196022,0.9238795325112865 0.3826834323650904,0.9807852804032303 0.19509032201612872,1 0))",
		},
		{
			"LINESTRING(10 10,10 0,0 0)", 2, nil,
			"POLYGON((12 0,11.96157056080646 -0.3901806440322565,11.847759065022574 -0.7653668647301796,11.66293922460509 -1.1111404660392044,11.414213562373096 -1.414213562373095,11.111140466039204 -1.6629392246050905,10.76536686473018 -1.8477590650225735,10.390180644032256 -1.9615705608064609,10 -2,0 -2,-0.39018064403225733 -1.9615705608064606,-0.7653668647301807 -1.847759065022573,-1.1111404660392044 -1.6629392246050905,-1.4142135623730954 -1.414213562373095,-1.662939224605091 -1.111140466039204,-1.8477590650225737 -0.7653668647301793,-1.9615705608064609 -0.3901806440322567,-2 0.00000000000000024492935982947064,-1.9615705608064609 0.3901806440322572,-1.8477590650225735 0.7653668647301798,-1.6629392246050907 1.1111404660392044,-1.414213562373095 1.4142135623730951,-1.111140466039204 1.662939224605091,-0.7653668647301795 1.8477590650225735,-0.39018064403225683 1.9615705608064609,0 2,8 2,8 10,8.03842943919354 10.390180644032258,8.152240934977426 10.76536686473018,8.33706077539491 11.111140466039204,8.585786437626904 11.414213562373096,8.888859533960796 11.662939224605092,9.23463313526982 11.847759065022574,9.609819355967744 11.96157056080646,10 12,10.390180644032256 11.96157056080646,10.76536686473018 11.847759065022574,11.111140466039204 11.66293922460509,11.414213562373096 11.414213562373096,11.66293922460509 11.111140466039204,11.847759065022574 10.76536686473018,11.96157056080646 10.390180644032258,12 10,12 0))",
		},
		{
			"POLYGON((0 0,10 0,10 10,0 10,0 0),(1 1,9 1,9 9,1 9,1 1))", 0.5, nil,
			"POLYGON((-0.5 0,-0.5 10,-0.4903926402016152 10.097545161008064,-0.46193976625564337 10.191341716182546,-0.4157348061512727 10.2777851165098,-0.35355339059327373 10.353553390593273,-0.277785116509801 10.415734806151272,-0.19134171618254486 10.461939766255643,-0.0975451610080641 10.490392640201616,0 10.5,10 10.5,10.097545161008064 10.490392640201616,10.191341716182546 10.461939766255643,10.2777851165098 10.415734806151272,10.353553390593273 10.353553390593273,10.415734806151272 10.2777851165098,10.461939766255643 10.191341716182546,10.490392640201616 10.097545161008064,10.5 10,10.5 0,10.490392640201616 -0.09754516100806412,10.461939766255643 -0.1913417161825449,10.415734806151272 -0.2777851165098011,10.353553390593273 -0.35355339059327373,10.2777851165098 -0.4157348061512726,10.191341716182546 -0.46193976625564337,10.097545161008064 -0.4903926402016152,10 -0.5,0 -0.5,-0.0975451610080641 -0.4903926402016152,-0.19134171618254486 -0.46193976625564337,-0.277785116509801 -0.41573480615127273,-0.35355339059327373 -0.3535533905932738,-0.4157348061512727 -0.2777851165098011,-0.46193976625564337 -0.19134171618254495,-0.4903926402016152 -0.0975451610080643,-0.5 0),(1.5 1.5,8.5 1.5,8.5 8.5,1.5 8.5,1.5 1.5))",
		},
		{
			"POINT(1 1)", 0.5,
			[]BufferOption{BufferQuadSegments(2)},
			"POLYGON((1.5 1,1.3535533905932737 0.6464466094067263,1 0.5,0.6464466094067263 0.6464466094067263,0.5 0.9999999999999999,0.6464466094067262 1.3535533905932737,0.9999999999999999 1.5,1.3535533905932737 1.353553390593274,1.5 1))",
		},
		{
			"LINESTRING(0 0,1 1)", 0.1,
			[]BufferOption{ /*BufferEndCapRound()*/ },
			"POLYGON((0.9292893218813453 1.0707106781186548,0.9444429766980399 1.0831469612302547,0.9617316567634911 1.0923879532511287,0.9804909677983873 1.0980785280403231,1 1.1,1.019509032201613 1.0980785280403231,1.038268343236509 1.0923879532511287,1.0555570233019602 1.0831469612302544,1.0707106781186548 1.0707106781186548,1.0831469612302544 1.0555570233019602,1.0923879532511287 1.038268343236509,1.0980785280403231 1.019509032201613,1.1 1,1.0980785280403231 0.9804909677983873,1.0923879532511287 0.9617316567634911,1.0831469612302547 0.9444429766980399,1.0707106781186548 0.9292893218813453,0.07071067811865475 -0.07071067811865475,0.05555702330196012 -0.0831469612302546,0.03826834323650888 -0.09238795325112872,0.019509032201612746 -0.09807852804032308,-0.00000000000000006049014748177263 -0.1,-0.019509032201612864 -0.09807852804032305,-0.03826834323650899 -0.09238795325112868,-0.0555570233019602 -0.08314696123025456,-0.07071067811865475 -0.07071067811865477,-0.08314696123025449 -0.05555702330196029,-0.09238795325112865 -0.03826834323650903,-0.09807852804032302 -0.019509032201612948,-0.1 -0.00000000000000010106430996148606,-0.09807852804032308 0.019509032201612663,-0.09238795325112874 0.038268343236508844,-0.08314696123025465 0.055557023301960044,-0.07071067811865475 0.07071067811865475,0.9292893218813453 1.0707106781186548))",
		},
		{
			"LINESTRING(0 0,1 1)", 0.1,
			[]BufferOption{BufferEndCapRound()},
			"POLYGON((0.9292893218813453 1.0707106781186548,0.9444429766980399 1.0831469612302547,0.9617316567634911 1.0923879532511287,0.9804909677983873 1.0980785280403231,1 1.1,1.019509032201613 1.0980785280403231,1.038268343236509 1.0923879532511287,1.0555570233019602 1.0831469612302544,1.0707106781186548 1.0707106781186548,1.0831469612302544 1.0555570233019602,1.0923879532511287 1.038268343236509,1.0980785280403231 1.019509032201613,1.1 1,1.0980785280403231 0.9804909677983873,1.0923879532511287 0.9617316567634911,1.0831469612302547 0.9444429766980399,1.0707106781186548 0.9292893218813453,0.07071067811865475 -0.07071067811865475,0.05555702330196012 -0.0831469612302546,0.03826834323650888 -0.09238795325112872,0.019509032201612746 -0.09807852804032308,-0.00000000000000006049014748177263 -0.1,-0.019509032201612864 -0.09807852804032305,-0.03826834323650899 -0.09238795325112868,-0.0555570233019602 -0.08314696123025456,-0.07071067811865475 -0.07071067811865477,-0.08314696123025449 -0.05555702330196029,-0.09238795325112865 -0.03826834323650903,-0.09807852804032302 -0.019509032201612948,-0.1 -0.00000000000000010106430996148606,-0.09807852804032308 0.019509032201612663,-0.09238795325112874 0.038268343236508844,-0.08314696123025465 0.055557023301960044,-0.07071067811865475 0.07071067811865475,0.9292893218813453 1.0707106781186548))",
		},
		{
			"LINESTRING(0 0,1 1)", 0.1,
			[]BufferOption{BufferEndCapFlat()},
			"POLYGON((0.9292893218813453 1.0707106781186548,1.0707106781186548 0.9292893218813453,0.07071067811865475 -0.07071067811865475,-0.07071067811865475 0.07071067811865475,0.9292893218813453 1.0707106781186548))",
		},
		{
			"LINESTRING(0 0,1 1)", 0.1,
			[]BufferOption{BufferEndCapSquare()},
			"POLYGON((0.9292893218813453 1.0707106781186548,1 1.1414213562373097,1.1414213562373097 1,0.07071067811865475 -0.07071067811865475,0 -0.1414213562373095,-0.1414213562373095 -0.000000000000000013877787807814457,0.9292893218813453 1.0707106781186548))",
		},
		{
			"LINESTRING(0 0,1 0,1 1)", 0.1,
			[]BufferOption{ /*BufferJoinStyleRound()*/ },
			"POLYGON((0.9 0.1,0.9 1,0.901921471959677 1.019509032201613,0.9076120467488714 1.0382683432365092,0.9168530387697456 1.0555570233019602,0.9292893218813453 1.0707106781186548,0.9444429766980398 1.0831469612302547,0.961731656763491 1.0923879532511287,0.9804909677983872 1.0980785280403231,1 1.1,1.0195090322016127 1.0980785280403231,1.038268343236509 1.0923879532511287,1.0555570233019602 1.0831469612302547,1.0707106781186546 1.0707106781186548,1.0831469612302544 1.0555570233019602,1.0923879532511287 1.0382683432365092,1.0980785280403231 1.019509032201613,1.1 1,1.1 0,1.0980785280403231 -0.019509032201612826,1.0923879532511287 -0.03826834323650898,1.0831469612302544 -0.05555702330196022,1.0707106781186548 -0.07071067811865475,1.0555570233019602 -0.08314696123025453,1.038268343236509 -0.09238795325112868,1.019509032201613 -0.09807852804032305,1 -0.1,0 -0.1,-0.019509032201612955 -0.09807852804032302,-0.03826834323650912 -0.09238795325112863,-0.055557023301960363 -0.08314696123025443,-0.07071067811865482 -0.07071067811865468,-0.08314696123025457 -0.055557023301960155,-0.0923879532511287 -0.03826834323650893,-0.09807852804032306 -0.01950903220161279,-0.1 0.000000000000000012246467991473533,-0.09807852804032305 0.01950903220161282,-0.0923879532511287 0.03826834323650895,-0.08314696123025456 0.055557023301960176,-0.07071067811865482 0.0707106781186547,-0.05555702330196031 0.08314696123025447,-0.0382683432365091 0.09238795325112864,-0.019509032201612972 0.09807852804032302,0 0.1,0.9 0.1))",
		},
		{
			"LINESTRING(0 0,1 0,1 1)", 0.1,
			[]BufferOption{BufferJoinStyleRound()},
			"POLYGON((0.9 0.1,0.9 1,0.901921471959677 1.019509032201613,0.9076120467488714 1.0382683432365092,0.9168530387697456 1.0555570233019602,0.9292893218813453 1.0707106781186548,0.9444429766980398 1.0831469612302547,0.961731656763491 1.0923879532511287,0.9804909677983872 1.0980785280403231,1 1.1,1.0195090322016127 1.0980785280403231,1.038268343236509 1.0923879532511287,1.0555570233019602 1.0831469612302547,1.0707106781186546 1.0707106781186548,1.0831469612302544 1.0555570233019602,1.0923879532511287 1.0382683432365092,1.0980785280403231 1.019509032201613,1.1 1,1.1 0,1.0980785280403231 -0.019509032201612826,1.0923879532511287 -0.03826834323650898,1.0831469612302544 -0.05555702330196022,1.0707106781186548 -0.07071067811865475,1.0555570233019602 -0.08314696123025453,1.038268343236509 -0.09238795325112868,1.019509032201613 -0.09807852804032305,1 -0.1,0 -0.1,-0.019509032201612955 -0.09807852804032302,-0.03826834323650912 -0.09238795325112863,-0.055557023301960363 -0.08314696123025443,-0.07071067811865482 -0.07071067811865468,-0.08314696123025457 -0.055557023301960155,-0.0923879532511287 -0.03826834323650893,-0.09807852804032306 -0.01950903220161279,-0.1 0.000000000000000012246467991473533,-0.09807852804032305 0.01950903220161282,-0.0923879532511287 0.03826834323650895,-0.08314696123025456 0.055557023301960176,-0.07071067811865482 0.0707106781186547,-0.05555702330196031 0.08314696123025447,-0.0382683432365091 0.09238795325112864,-0.019509032201612972 0.09807852804032302,0 0.1,0.9 0.1))",
		},
		{
			"LINESTRING(0 0,5 0,0 1)", 0.1,
			[]BufferOption{BufferJoinStyleMitre(3.5)},
			"POLYGON((3.9900980486407214 0.1,-0.019611613513818404 0.9019419324309079,-0.038364961837643596 0.9076521266991159,-0.05564396619336664 0.9169111979489927,-0.07078460413377427 0.9293633252649527,-0.08320502943378437 0.944529980377477,-0.0924279321145722 0.9618283172361496,-0.09809888119837851 0.9805935704565105,-0.09999994529221752 1.0001046018809767,-0.09805806756909202 1.0196116135138185,-0.09234787330088415 1.0383649618376436,-0.08308880205100731 1.0556439661933665,-0.0706366747350473 1.0707846041337743,-0.05547001962252293 1.0832050294337843,-0.03817168276385041 1.0924279321145722,-0.01940642954348947 1.0980988811983785,0.00010460188097669381 1.0999999452922176,0.019611613513818404 1.098058067569092,5.3547836142425265 0.03102366742335057,5.341809714425123 -0.1,0 -0.1,-0.019509032201612868 -0.09807852804032303,-0.03826834323650904 -0.09238795325112865,-0.05555702330196022 -0.08314696123025453,-0.07071067811865477 -0.07071067811865475,-0.08314696123025456 -0.0555570233019602,-0.0923879532511287 -0.03826834323650897,-0.09807852804032305 -0.019509032201612837,-0.1 0.000000000000000012246467991473533,-0.09807852804032305 0.01950903220161286,-0.09238795325112868 0.03826834323650899,-0.08314696123025454 0.05555702330196022,-0.07071067811865475 0.07071067811865477,-0.0555570233019602 0.08314696123025456,-0.038268343236508975 0.09238795325112868,-0.019509032201612844 0.09807852804032305,0 0.1,3.9900980486407214 0.1))",
		},
		{
			"LINESTRING(0 0,1 0,1 1)", 0.1,
			[]BufferOption{BufferJoinStyleBevel()},
			"POLYGON((0.9 0.1,0.9 1,0.901921471959677 1.019509032201613,0.9076120467488714 1.0382683432365092,0.9168530387697456 1.0555570233019602,0.9292893218813453 1.0707106781186548,0.9444429766980398 1.0831469612302547,0.961731656763491 1.0923879532511287,0.9804909677983872 1.0980785280403231,1 1.1,1.0195090322016127 1.0980785280403231,1.038268343236509 1.0923879532511287,1.0555570233019602 1.0831469612302547,1.0707106781186546 1.0707106781186548,1.0831469612302544 1.0555570233019602,1.0923879532511287 1.0382683432365092,1.0980785280403231 1.019509032201613,1.1 1,1.1 0,1 -0.1,0 -0.1,-0.019509032201612955 -0.09807852804032302,-0.03826834323650912 -0.09238795325112863,-0.055557023301960363 -0.08314696123025443,-0.07071067811865482 -0.07071067811865468,-0.08314696123025457 -0.055557023301960155,-0.0923879532511287 -0.03826834323650893,-0.09807852804032306 -0.01950903220161279,-0.1 0.000000000000000012246467991473533,-0.09807852804032305 0.01950903220161282,-0.0923879532511287 0.03826834323650895,-0.08314696123025456 0.055557023301960176,-0.07071067811865482 0.0707106781186547,-0.05555702330196031 0.08314696123025447,-0.0382683432365091 0.09238795325112864,-0.019509032201612972 0.09807852804032302,0 0.1,0.9 0.1))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := geomFromWKT(t, tt.wkt)
			t.Logf("WKT: %v", g.AsText())

			got, err := Buffer(g, tt.radius, tt.opts...)
			expectNoErr(t, err)

			gWant := geomFromWKT(t, tt.want)
			symDiff, err := geom.SymmetricDifference(gWant, got)
			expectNoErr(t, err)
			const threshold = 1e-3
			relativeAreaDiff := symDiff.Area() / math.Min(gWant.Area(), got.Area())
			if relativeAreaDiff > threshold {
				t.Errorf(
					"expected relativeAreaDiff <= %f but was %f",
					threshold, relativeAreaDiff,
				)
			}
		})
	}
}

func TestSimplify(t *testing.T) {
	for i, tt := range []struct {
		input     string
		tolerance float64
		output    string
	}{
		// The following test cases were generated by taking the input, and
		// seeing what the output from Simplify is. It was then verified
		// visually to be correct.
		{
			input:     "POLYGON((1 0,0.998026728428272 -0.062790519529313,0.992114701314478 -0.125333233564304,0.982287250728689 -0.187381314585724,0.968583161128631 -0.248689887164855,0.951056516295154 -0.309016994374947,0.929776485888252 -0.368124552684678,0.90482705246602 -0.425779291565072,0.876306680043864 -0.481753674101715,0.844327925502015 -0.535826794978996,0.809016994374948 -0.587785252292473,0.77051324277579 -0.637423989748689,0.728968627421412 -0.684547105928688,0.684547105928689 -0.728968627421411,0.63742398974869 -0.770513242775789,0.587785252292474 -0.809016994374947,0.535826794978997 -0.844327925502015,0.481753674101716 -0.876306680043863,0.425779291565074 -0.904827052466019,0.368124552684679 -0.929776485888251,0.309016994374949 -0.951056516295153,0.248689887164856 -0.968583161128631,0.187381314585726 -0.982287250728688,0.125333233564306 -0.992114701314478,0.062790519529315 -0.998026728428271,0 -1,-0.062790519529311 -0.998026728428272,-0.125333233564302 -0.992114701314478,-0.187381314585722 -0.982287250728689,-0.248689887164852 -0.968583161128632,-0.309016994374945 -0.951056516295154,-0.368124552684675 -0.929776485888252,-0.42577929156507 -0.904827052466021,-0.481753674101713 -0.876306680043865,-0.535826794978994 -0.844327925502017,-0.587785252292471 -0.809016994374949,-0.637423989748688 -0.770513242775791,-0.684547105928687 -0.728968627421413,-0.72896862742141 -0.684547105928691,-0.770513242775788 -0.637423989748692,-0.809016994374946 -0.587785252292475,-0.844327925502014 -0.535826794978999,-0.876306680043863 -0.481753674101717,-0.904827052466019 -0.425779291565074,-0.929776485888251 -0.36812455268468,-0.951056516295153 -0.309016994374949,-0.968583161128631 -0.248689887164857,-0.982287250728688 -0.187381314585726,-0.992114701314478 -0.125333233564306,-0.998026728428271 -0.062790519529315,-1 0,-0.998026728428272 0.062790519529312,-0.992114701314478 0.125333233564303,-0.982287250728689 0.187381314585723,-0.968583161128631 0.248689887164854,-0.951056516295154 0.309016994374946,-0.929776485888252 0.368124552684677,-0.90482705246602 0.425779291565072,-0.876306680043864 0.481753674101715,-0.844327925502015 0.535826794978996,-0.809016994374948 0.587785252292473,-0.77051324277579 0.637423989748689,-0.728968627421412 0.684547105928688,-0.684547105928689 0.728968627421411,-0.63742398974869 0.770513242775789,-0.587785252292474 0.809016994374947,-0.535826794978998 0.844327925502014,-0.481753674101717 0.876306680043863,-0.425779291565075 0.904827052466019,-0.36812455268468 0.92977648588825,-0.30901699437495 0.951056516295153,-0.248689887164858 0.96858316112863,-0.187381314585728 0.982287250728688,-0.125333233564308 0.992114701314477,-0.062790519529318 0.998026728428271,0 1,0.062790519529308 0.998026728428272,0.125333233564299 0.992114701314479,0.187381314585719 0.98228725072869,0.248689887164849 0.968583161128633,0.309016994374941 0.951056516295156,0.368124552684672 0.929776485888254,0.425779291565066 0.904827052466023,0.481753674101709 0.876306680043867,0.53582679497899 0.844327925502019,0.587785252292466 0.809016994374952,0.637423989748683 0.770513242775795,0.684547105928682 0.728968627421418,0.728968627421405 0.684547105928695,0.770513242775783 0.637423989748697,0.809016994374942 0.587785252292481,0.84432792550201 0.535826794979005,0.876306680043858 0.481753674101725,0.904827052466015 0.425779291565083,0.929776485888247 0.368124552684689,0.95105651629515 0.309016994374959,0.968583161128628 0.248689887164867,0.982287250728686 0.187381314585737,0.992114701314476 0.125333233564317,0.998026728428271 0.062790519529327,1 0,1 0))",
			tolerance: 0.5,
			output:    "POLYGON((1 0,0 -1,-1 0,0 1,1 0))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := geomFromWKT(t, tt.input)
			t.Logf("WKT: %v", g.AsText())
			got, err := Simplify(g, tt.tolerance)
			expectNoErr(t, err)
			expectGeomEq(t, got, geomFromWKT(t, tt.output), geom.IgnoreOrder)
		})
	}
}

func TestDifference(t *testing.T) {
	a := geomFromWKT(t, "POLYGON((0 0,0 2,2 2,2 0,0 0))")
	b := geomFromWKT(t, "POLYGON((1 1,1 3,3 3,3 1,1 1))")

	got, err := Difference(a, b)
	expectNoErr(t, err)
	want := geomFromWKT(t, "POLYGON((0 0,0 2,1 2,1 1,2 1,2 0,0 0))")

	expectGeomEq(t, got, want, geom.IgnoreOrder)
}

func TestSymmetricDifference(t *testing.T) {
	a := geomFromWKT(t, "POLYGON((0 0,0 2,2 2,2 0,0 0))")
	b := geomFromWKT(t, "POLYGON((1 1,1 3,3 3,3 1,1 1))")

	got, err := SymmetricDifference(a, b)
	expectNoErr(t, err)
	want := geomFromWKT(t, "MULTIPOLYGON(((0 0,0 2,1 2,1 1,2 1,2 0,0 0)),((2 1,3 1,3 3,1 3,1 2,2 2,2 1)))")

	expectGeomEq(t, got, want, geom.IgnoreOrder)
}

func TestMakeValid(t *testing.T) {
	for i, tt := range []struct {
		input      string
		wantOutput string
	}{
		{
			"POLYGON((0 0,2 2,2 0,0 2,0 0))",
			"MULTIPOLYGON(((0 2,1 1,0 0,0 2)),((2 0,1 1,2 2,2 0)))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := geom.UnmarshalWKT(tt.input)
			expectErr(t, err)
			in := geomFromWKTWithoutValidation(t, tt.input)
			gotGeom, err := MakeValid(in)
			if _, ok := err.(unsupportedGEOSVersionError); ok {
				t.Skip(err)
			}
			expectNoErr(t, err)
			wantGeom := geomFromWKT(t, tt.wantOutput)
			expectGeomEq(t, gotGeom, wantGeom, geom.IgnoreOrder)
		})
	}
}

func TestUnaryUnion(t *testing.T) {
	for i, tt := range []struct {
		input      string
		wantOutput string
	}{
		{
			"GEOMETRYCOLLECTION(POLYGON((0 0,0 2,2 2,2 0,0 0)),POLYGON((1 1,1 3,3 3,3 1,1 1)))",
			"POLYGON((0 0,2 0,2 1,3 1,3 3,1 3,1 2,0 2,0 0))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := UnaryUnion(geomFromWKT(t, tt.input))
			expectNoErr(t, err)
			expectGeomEq(t, got, geomFromWKT(t, tt.wantOutput), geom.IgnoreOrder)
		})
	}
}

func TestCoverageUnion(t *testing.T) {
	for i, tc := range []struct {
		input   string
		output  string
		wantErr bool
	}{
		{
			// Noded correctly (shared edge).
			input: `GEOMETRYCOLLECTION(
				POLYGON((0 0,0 1,1 0,0 0)),
				POLYGON((1 1,0 1,1 0,1 1))
			)`,
			output: `POLYGON((0 0,0 1,1 1,1 0,0 0))`,
		},
		{
			// Noded correctly (shared vertex but no shared edge).
			input: `GEOMETRYCOLLECTION(
				POLYGON((0 0,0 1,1 1,1 0,0 0)),
				POLYGON((1 1,1 2,2 2,2 1,1 1))
			)`,
			output: `MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((1 1,1 2,2 2,2 1,1 1)))`,
		},
		{
			// Noded correctly (completely disjoint).
			input: `GEOMETRYCOLLECTION(
				POLYGON((0 0,0 1,1 1,1 0,0 0)),
				POLYGON((2 2,2 3,3 3,3 2,2 2))
			)`,
			output: `MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),((2 2,2 3,3 3,3 2,2 2)))`,
		},
		{
			// Input constraint violated: inputs overlap.
			input: `GEOMETRYCOLLECTION(
				POLYGON((0 0,0 1,1 0,0 0)),
				POLYGON((0 0,0 1,1 1,1 0,0 0))
			)`,
			wantErr: true,
		},
		{
			// Input constraint violated: inputs overlap and not noded correctly.
			input: `GEOMETRYCOLLECTION(
				POLYGON((0 0,0 1,1 0,0 0)),
				POLYGON((0 0,0 1,1 1,0 0))
			)`,
			wantErr: true,
		},
		{
			// Input constraint violated: not noded correctly.
			input: `GEOMETRYCOLLECTION(
				POLYGON((0 0,0 1,1 1,1 0,0 0)),
				POLYGON((0 1,2 1,2 2,0 2,0 1))
			)`,
			wantErr: true,
		},
		{
			// Input constraint violation: not everything is a polygon.
			input:   `GEOMETRYCOLLECTION(POINT(1 2),POLYGON((0 0,0 1,1 0,0 0)))`,
			wantErr: true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			in := geomFromWKT(t, tc.input)
			gotGeom, err := CoverageUnion(in)
			if _, ok := err.(unsupportedGEOSVersionError); ok {
				t.Skip(err)
			}
			if tc.wantErr {
				expectErr(t, err)
			} else {
				expectNoErr(t, err)
				wantGeom := geomFromWKT(t, tc.output)
				expectGeomEq(t, gotGeom, wantGeom, geom.IgnoreOrder)
			}
		})
	}
}
