package geom_test

import (
	"math"
	"strconv"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestIsEmptyDimension(t *testing.T) {
	for _, tt := range []struct {
		wkt       string
		wantEmpty bool
		wantDim   int
	}{
		{"POINT EMPTY", true, 0},
		{"POINT(1 1)", false, 0},
		{"LINESTRING EMPTY", true, 1},
		{"LINESTRING(0 0,1 1)", false, 1},
		{"LINESTRING(0 0,1 1,2 2)", false, 1},
		{"LINESTRING(0 0,1 1,1 0,0 0)", false, 1},
		{"POLYGON EMPTY", true, 2},
		{"POLYGON((0 0,1 1,1 0,0 0))", false, 2},
		{"MULTIPOINT EMPTY", true, 0},
		{"MULTIPOINT((0 0))", false, 0},
		{"MULTIPOINT((0 0),(1 1))", false, 0},
		{"MULTILINESTRING EMPTY", true, 1},
		{"MULTILINESTRING((0 0,1 1,2 2))", false, 1},
		{"MULTILINESTRING(EMPTY)", true, 1},
		{"MULTIPOLYGON EMPTY", true, 2},
		{"MULTIPOLYGON(((0 0,1 0,1 1,0 0)))", false, 2},
		{"MULTIPOLYGON(((0 0,1 0,1 1,0 0)))", false, 2},
		{"MULTIPOLYGON(EMPTY)", true, 2},
		{"GEOMETRYCOLLECTION EMPTY", true, 0},
		{"GEOMETRYCOLLECTION(POINT EMPTY)", true, 0},
		{"GEOMETRYCOLLECTION(POLYGON EMPTY)", true, 2},
		{"GEOMETRYCOLLECTION(POINT(1 1))", false, 0},
		{"GEOMETRYCOLLECTION(POINT(1 1),LINESTRING(0 0,1 1))", false, 1},
		{"GEOMETRYCOLLECTION(POLYGON((0 0,1 1,1 0,0 0)),POINT(1 1),LINESTRING(0 0,1 1))", false, 2},
	} {
		t.Run(tt.wkt, func(t *testing.T) {
			geom, err := UnmarshalWKT(strings.NewReader(tt.wkt))
			if err != nil {
				t.Fatal(err)
			}
			t.Run("IsEmpty_"+tt.wkt, func(t *testing.T) {
				gotEmpty := geom.IsEmpty()
				if gotEmpty != tt.wantEmpty {
					t.Errorf("want=%v got=%v", tt.wantEmpty, gotEmpty)
				}
			})
			t.Run("Dimension_"+tt.wkt, func(t *testing.T) {
				gotDim := geom.Dimension()
				if gotDim != tt.wantDim {
					t.Errorf("want=%v got=%v", tt.wantDim, gotDim)
				}
			})
		})
	}
}

func TestEnvelope(t *testing.T) {
	xy := func(x, y float64) XY {
		return XY{x, y}
	}
	for i, tt := range []struct {
		wkt string
		min XY
		max XY
	}{
		{"POINT(1 1)", xy(1, 1), xy(1, 1)},
		{"LINESTRING(1 2,3 4)", xy(1, 2), xy(3, 4)},
		{"LINESTRING(4 1,2 3)", xy(2, 1), xy(4, 3)},
		{"LINESTRING(1 1,3 1,2 2,2 4)", xy(1, 1), xy(3, 4)},
		{"POLYGON((1 1,3 1,2 2,2 4,1 1))", xy(1, 1), xy(3, 4)},
		{"MULTIPOINT(1 1,3 1,2 2,2 4,1 1)", xy(1, 1), xy(3, 4)},
		{"MULTILINESTRING((1 1,3 1,2 2,2 4,1 1),(4 1,4 2))", xy(1, 1), xy(4, 4)},
		{"MULTILINESTRING((4 1,4 2),(1 1,3 1,2 2,2 4,1 1))", xy(1, 1), xy(4, 4)},
		{"MULTIPOLYGON(((4 1,4 2,3 2,4 1)),((1 1,3 1,2 2,2 4,1 1)))", xy(1, 1), xy(4, 4)},
		{"GEOMETRYCOLLECTION(POINT(4 1),POINT(2 3))", xy(2, 1), xy(4, 3)},
		{"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT(4 1),POINT(2 3)))", xy(2, 1), xy(4, 3)},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log("wkt:", tt.wkt)
			g := geomFromWKT(t, tt.wkt)
			env, have := g.Envelope()
			if !have {
				t.Fatalf("expected to have envelope but didn't")
			}
			if !env.Min().Equals(tt.min) {
				t.Errorf("min: got=%v want=%v", env.Min(), tt.min)
			}
			if !env.Max().Equals(tt.max) {
				t.Errorf("max: got=%v want=%v", env.Max(), tt.max)
			}
		})
	}
}

func TestNoEnvelope(t *testing.T) {
	for _, wkt := range []string{
		"POINT EMPTY",
		"MULTIPOINT EMPTY",
		"MULTILINESTRING EMPTY",
		"MULTIPOLYGON EMPTY",
		"GEOMETRYCOLLECTION EMPTY",
		"GEOMETRYCOLLECTION(POINT EMPTY)",
	} {
		t.Run(wkt, func(t *testing.T) {
			g := geomFromWKT(t, wkt)
			if _, have := g.Envelope(); have {
				t.Errorf("have envelope but expected not to")
			}
		})
	}
}

func TestIsSimple(t *testing.T) {
	for i, tt := range []struct {
		wkt        string
		wantSimple bool
	}{
		{"POINT EMPTY", true},
		{"POINT(1 2)", true},

		{"LINESTRING EMPTY", true},
		{"LINESTRING(0 0,1 2)", true},
		{"LINESTRING(0 0,1 1,1 1)", true},
		{"LINESTRING(0 0,0 0,1 1)", true},
		{"LINESTRING(0 0,1 1,0 0)", false},
		{"LINESTRING(0 0,1 1,0 1)", true},
		{"LINESTRING(0 0,1 1,0 1,0 0)", true},
		{"LINESTRING(0 0,1 1,0 1,1 0)", false},
		{"LINESTRING(0 0,1 1,0 1,1 0,0 0)", false},
		{"LINESTRING(0 0,1 1,0 1,1 0,2 0)", false},
		{"LINESTRING(0 0,1 1,0 1,0 0,1 1)", false},
		{"LINESTRING(0 0,1 1,0 1,0 0,2 2)", false},
		{"LINESTRING(1 1,2 2,0 0)", false},
		{"LINESTRING(1 1,2 2,3 2,3 3,0 0)", false},
		{"LINESTRING(0 0,1 1,2 2)", true},

		{"POLYGON((0 0,0 1,1 0,0 0))", true},

		{"MULTIPOINT((1 2),(3 4),(5 6))", true},
		{"MULTIPOINT((1 2),(3 4),(1 2))", false},
		{"MULTIPOINT EMPTY", true},

		{"POLYGON EMPTY", true},

		{"MULTILINESTRING EMPTY", true},
		{"MULTILINESTRING((0 0,1 0))", true},
		{"MULTILINESTRING((0 0,1 0,0 1,0 0))", true},
		{"MULTILINESTRING((0 0,1 1,2 2),(0 2,1 1,2 0))", false},
		{"MULTILINESTRING((0 0,2 1,4 2),(4 2,2 3,0 4))", true},
		{"MULTILINESTRING((0 0,2 0,4 0),(2 0,2 1))", false},

		// Cases for reproducing bugs.
		{"MULTILINESTRING((0 0,0 1,1 1),(0 1,0 0,1 0))", false},
		{"MULTILINESTRING((0 0,1 0),(0 1,1 1))", true},
		{"MULTILINESTRING((0 0,3 0,3 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1))", true},
		{"MULTILINESTRING((1 1,1 0,0 0),(1 1,0 1,0 0))", true},

		// Cases for behaviour around duplicated lines. These cases are to
		// match PostGIS and libgeos behaviour (the OGC spec is unclear about
		// what the behaviour should be).
		{"MULTILINESTRING((1 1,2 2),(1 1,2 2))", true},
		{"MULTILINESTRING((1 1,2 2),(2 2,1 1))", true},
		{"MULTILINESTRING((1 1,2 2),(1 1,2 2,3 3))", false},
		{"MULTILINESTRING((1 1,2 2),(2 2,1 1,3 1))", false},

		{"MULTIPOLYGON(((0 0,1 0,0 1,0 0)))", true},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, defined := geomFromWKT(t, tt.wkt).IsSimple()
			if !defined {
				t.Fatal("not defined")
			}
			if got != tt.wantSimple {
				t.Logf("wkt: %s", tt.wkt)
				t.Errorf("got=%v want=%v", got, tt.wantSimple)
			}
		})
	}
}

func TestIsSimpleGeometryCollection(t *testing.T) {
	_, defined := geomFromWKT(t, "GEOMETRYCOLLECTION(POINT(1 2))").IsSimple()
	expectBoolEq(t, defined, false)
}

func TestBoundary(t *testing.T) {
	for i, tt := range []struct {
		wkt, boundary string
	}{
		{"POINT EMPTY", "POINT EMPTY"},
		{"LINESTRING EMPTY", "LINESTRING EMPTY"},
		{"POLYGON EMPTY", "POLYGON EMPTY"},
		{"MULTIPOINT EMPTY", "MULTIPOINT EMPTY"},
		{"MULTILINESTRING EMPTY", "MULTILINESTRING EMPTY"},
		{"MULTIPOLYGON EMPTY", "MULTIPOLYGON EMPTY"},

		{"POINT(1 2)", "GEOMETRYCOLLECTION EMPTY"},
		{"LINESTRING(1 2,3 4)", "MULTIPOINT(1 2,3 4)"},
		{"LINESTRING(1 2,3 4,5 6)", "MULTIPOINT(1 2,5 6)"},
		{"LINESTRING(1 2,3 4,5 6,7 8)", "MULTIPOINT(1 2,7 8)"},
		{"LINESTRING(0 0,1 0,0 1,0 0)", "MULTIPOINT EMPTY"},

		{"POLYGON((0 0,1 0,1 1,0 1,0 0))", "LINESTRING(0 0,1 0,1 1,0 1,0 0)"},
		{"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1))", "MULTILINESTRING((0 0,3 0,3 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1))"},

		{"MULTIPOINT((1 2))", "GEOMETRYCOLLECTION EMPTY"},
		{"MULTIPOINT((1 2),(3 4))", "GEOMETRYCOLLECTION EMPTY"},

		{
			"MULTILINESTRING((0 0,1 1))",
			"MULTIPOINT(0 0,1 1)",
		},
		{
			"MULTILINESTRING((0 0,1 0),(0 1,1 1))",
			"MULTIPOINT(0 0,1 0,0 1,1 1)",
		},
		{
			"MULTILINESTRING((0 0,1 1),(1 1,1 0))",
			"MULTIPOINT(0 0,1 0)",
		},
		{
			"MULTILINESTRING((0 0,1 0,1 1),(0 0,0 1,1 1))",
			"MULTIPOINT EMPTY",
		},
		{
			"MULTILINESTRING((0 0,1 1),(0 1,1 1),(1 0,1 1))",
			"MULTIPOINT(0 0,1 1,0 1,1 0)",
		},
		{
			"MULTILINESTRING((0 0,0 1,1 1),(0 1,0 0,1 0))",
			"MULTIPOINT(0 0,1 1,0 1,1 0)",
		},
		{
			"MULTILINESTRING((0 1,1 1),(1 1,1 0),(1 1,2 1),(1 2,1 1))",
			"MULTIPOINT(0 1,1 0,2 1,1 2)",
		},
		{
			"MULTILINESTRING((1 1,2 2),(1 1,2 2))",
			"MULTIPOINT EMPTY",
		},

		{
			"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1)),((4 0,5 0,5 1,4 1,4 0)))",
			"MULTILINESTRING((0 0,3 0,3 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1),(4 0,5 0,5 1,4 1,4 0))",
		},
		{
			"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)))",
			"MULTILINESTRING((0 0,3 0,3 3,0 3,0 0))",
		},

		{
			"GEOMETRYCOLLECTION EMPTY",
			"GEOMETRYCOLLECTION EMPTY",
		},
		{
			"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION EMPTY)",
			"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION EMPTY)",
		},
		{
			"GEOMETRYCOLLECTION(POINT EMPTY, GEOMETRYCOLLECTION EMPTY)",
			"GEOMETRYCOLLECTION(POINT EMPTY, GEOMETRYCOLLECTION EMPTY)",
		},
		{
			"GEOMETRYCOLLECTION(POINT(1 1))",
			"GEOMETRYCOLLECTION EMPTY",
		},
		{
			`GEOMETRYCOLLECTION(
				LINESTRING(1 0,0 5,5 2),
				POINT(2 3),
				POLYGON((0 0,1 0,0 1,0 0))
			)`,
			`GEOMETRYCOLLECTION(
				MULTIPOINT(1 0,5 2),
				LINESTRING(0 0,1 0,0 1,0 0)
			)`,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			want := geomFromWKT(t, tt.boundary)
			got := geomFromWKT(t, tt.wkt).Boundary()
			expectGeomEq(t, got, want)
		})
	}
}

func TestCoordinates(t *testing.T) {
	cmp0d := func(t *testing.T, got Coordinates, want [2]float64) {
		if got.XY.X != want[0] {
			t.Errorf("coordinate mismatch: got=%v want=%v", got, want)
		}
		if got.XY.Y != want[1] {
			t.Errorf("coordinate mismatch: got=%v want=%v", got, want)
		}
	}
	cmp1d := func(t *testing.T, got []Coordinates, want [][2]float64) {
		if len(got) != len(want) {
			t.Errorf("length mismatch: got=%v want=%v", len(got), len(want))
		}
		for i := range got {
			cmp0d(t, got[i], want[i])
		}
	}
	cmp2d := func(t *testing.T, got [][]Coordinates, want [][][2]float64) {
		if len(got) != len(want) {
			t.Errorf("length mismatch: got=%v want=%v", len(got), len(want))
		}
		for i := range got {
			cmp1d(t, got[i], want[i])
		}
	}
	cmp3d := func(t *testing.T, got [][][]Coordinates, want [][][][2]float64) {
		if len(got) != len(want) {
			t.Errorf("length mismatch: got=%v want=%v", len(got), len(want))
		}
		for i := range got {
			cmp2d(t, got[i], want[i])
		}
	}
	t.Run("Point", func(t *testing.T) {
		cmp0d(t,
			geomFromWKT(t, "POINT(1 2)").AsPoint().Coordinates(),
			[2]float64{1, 2},
		)
	})
	t.Run("Line-LineString-MultiPoint", func(t *testing.T) {
		cmp1d(t,
			geomFromWKT(t, "LINESTRING(0 1,2 3)").AsLine().Coordinates(),
			[][2]float64{{0, 1}, {2, 3}},
		)
		cmp1d(t,
			geomFromWKT(t, "LINESTRING(0 1,2 3,4 5)").AsLineString().Coordinates(),
			[][2]float64{{0, 1}, {2, 3}, {4, 5}},
		)
		cmp1d(t,
			geomFromWKT(t, "MULTIPOINT(0 1,2 3,4 5)").AsMultiPoint().Coordinates(),
			[][2]float64{{0, 1}, {2, 3}, {4, 5}},
		)
	})
	t.Run("Polygon-MultiLineString", func(t *testing.T) {
		cmp2d(t,
			geomFromWKT(t, "POLYGON((0 0,0 10,10 0,0 0),(2 2,2 7,7 2,2 2))").AsPolygon().Coordinates(),
			[][][2]float64{
				{{0, 0}, {0, 10}, {10, 0}, {0, 0}},
				{{2, 2}, {2, 7}, {7, 2}, {2, 2}},
			},
		)
		cmp2d(t,
			geomFromWKT(t, "MULTILINESTRING((0 0,0 10,10 0,0 0),(2 2,2 8,8 2,2 2))").AsMultiLineString().Coordinates(),
			[][][2]float64{
				{{0, 0}, {0, 10}, {10, 0}, {0, 0}},
				{{2, 2}, {2, 8}, {8, 2}, {2, 2}},
			},
		)
	})
	t.Run("MultiPolygon", func(t *testing.T) {
		cmp3d(t,
			geomFromWKT(t, `
				MULTIPOLYGON(
					(
						(0 0,0 10,10 0,0 0),
						(2 2,2 7,7 2,2 2)
					),
					(
						(100 100,100 110,110 100,100 100),
						(102 102,102 107,107 102,102 102)
					)
				)`,
			).AsMultiPolygon().Coordinates(),
			[][][][2]float64{
				{
					{{0, 0}, {0, 10}, {10, 0}, {0, 0}},
					{{2, 2}, {2, 7}, {7, 2}, {2, 2}},
				},
				{
					{{100, 100}, {100, 110}, {110, 100}, {100, 100}},
					{{102, 102}, {102, 107}, {107, 102}, {102, 102}},
				},
			},
		)
	})
}

func TestTransformXY(t *testing.T) {
	transform := func(in XY) XY {
		return XY{in.X * 1.5, in.Y}
	}
	for i, tt := range []struct {
		wktIn, wktOut string
	}{
		{"POINT EMPTY", "POINT EMPTY"},
		{"LINESTRING EMPTY", "LINESTRING EMPTY"},
		{"POLYGON EMPTY", "POLYGON EMPTY"},

		{"POINT(1 3)", "POINT(1.5 3)"},
		{"LINESTRING(1 2,3 4)", "LINESTRING(1.5 2,4.5 4)"},
		{"LINESTRING(1 2,3 4,5 6)", "LINESTRING(1.5 2,4.5 4,7.5 6)"},
		{"POLYGON((0 0,0 1,1 0,0 0))", "POLYGON((0 0,0 1,1.5 0,0 0))"},
		{"MULTIPOINT(0 0,0 1,1 0,0 0)", "MULTIPOINT(0 0,0 1,1.5 0,0 0)"},
		{"MULTILINESTRING((1 2,3 4,5 6))", "MULTILINESTRING((1.5 2,4.5 4,7.5 6))"},
		{"MULTIPOLYGON(((0 0,0 1,1 0,0 0)))", "MULTIPOLYGON(((0 0,0 1,1.5 0,0 0)))"},

		{"GEOMETRYCOLLECTION EMPTY", "GEOMETRYCOLLECTION EMPTY"},
		{"GEOMETRYCOLLECTION(POINT EMPTY)", "GEOMETRYCOLLECTION(POINT EMPTY)"},
		{"GEOMETRYCOLLECTION(POINT(1 2))", "GEOMETRYCOLLECTION(POINT(1.5 2))"},
		{"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT(1 2)))", "GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT(1.5 2)))"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := geomFromWKT(t, tt.wktIn)
			got, err := g.TransformXY(transform)
			expectNoErr(t, err)
			want := geomFromWKT(t, tt.wktOut)
			expectGeomEq(t, got, want)
		})
	}
}

func TestIsRing(t *testing.T) {
	for i, tt := range []struct {
		wkt  string
		want bool
	}{
		{"LINESTRING(0 0,0 1,1 0,0 0)", true},
		{"LINESTRING(0 0,1 1,1 0,0 1,0 0)", false}, // not simple
		{"LINESTRING(0 0,1 0,1 1,0 1)", false},     // not closed
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := geomFromWKT(t, tt.wkt).AsLineString().IsRing()
			if got != tt.want {
				t.Logf("WKT: %v", tt.wkt)
				t.Errorf("got=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestLength(t *testing.T) {
	for i, tt := range []struct {
		wkt  string
		want float64
	}{
		{"LINESTRING(0 0,1 0)", 1},
		{"LINESTRING(5 8,4 9)", math.Sqrt(2)},
		{"LINESTRING(0 0,0 1,1 3)", 1 + math.Sqrt(5)},
		{"MULTILINESTRING((4 2,5 1),(9 2,7 1))", math.Sqrt(2) + math.Sqrt(5)},
		{"MULTILINESTRING((0 0,2 0),(1 0,3 0))", 4},
		{"POINT(1 3)", 0},
		{"MULTIPOINT(0 0,0 1,1 0,0 0)", 0},
		{"POLYGON((0 0,1 1,0 1,0 0))", 0},
		{"POLYGON((0 0,0 3,3 3,3 0,0 0),(1 1,1 2,2 2,2 1,1 1))", 0},
		{"MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((2 1,1 1,2 0,2 1)))", 0},
		{"GEOMETRYCOLLECTION EMPTY", 0},
		{"GEOMETRYCOLLECTION(POINT EMPTY)", 0},
		{"GEOMETRYCOLLECTION(POINT(1 2))", 0},
		{"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT(1 2)))", 0},
		{`GEOMETRYCOLLECTION(
			LINESTRING(0 0,0 1,1 3),
			POINT(2 3),
			MULTILINESTRING((4 2,5 1),(9 2,7 1)),
			POLYGON((0 0,0 3,3 3,3 0,0 0),(1 1,1 2,2 2,2 1,1 1))
		)`, 1 + math.Sqrt(5) + math.Sqrt(2) + math.Sqrt(5)},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := geomFromWKT(t, tt.wkt).Length()
			if math.Abs(tt.want-got) > 1e-6 {
				t.Errorf("got=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestArea(t *testing.T) {
	for i, tt := range []struct {
		wkt  string
		want float64
	}{
		{"GEOMETRYCOLLECTION EMPTY", 0},
		{"LINESTRING(1 1,5 5)", 0},
		{"LINESTRING(5 8,4 9)", 0},
		{"LINESTRING(0 0,0 1,1 3)", 0},
		{"MULTILINESTRING((4 2,5 1),(9 2,7 1))", 0},
		{"MULTILINESTRING((0 0,2 0),(1 0,3 0))", 0},
		{"POINT(1 3)", 0},
		{"MULTIPOINT(0 0,0 1,1 0,0 0)", 0},
		{"POLYGON((0 0,1 1,0 1,0 0))", 0.5},
		{"POLYGON((0 0,0 1,1 1,0 0))", 0.5},
		{"POLYGON((0 0,0 1,1 1,1 0,0 0))", 1.0},
		{"POLYGON((0 0,0 3,3 3,3 0,0 0),(1 1,1 2,2 2,2 1,1 1))", 8.0},
		{"MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((2 1,1 1,2 0,2 1)))", 1.0},
		{"GEOMETRYCOLLECTION(POINT EMPTY)", 0},
		{"GEOMETRYCOLLECTION(POINT(1 2))", 0},
		{"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT(1 2)))", 0},
		{`GEOMETRYCOLLECTION(
			LINESTRING(1 0,0 5,5 2),
			POINT(2 3),
			POLYGON((0 0,0 3,3 3,3 0,0 0),(1 1,1 2,2 2,2 1,1 1))
		)`, 8.0},
		{`GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(
			LINESTRING(1 0,0 5,5 2),
			POINT(2 3),
			MULTIPOLYGON(((0 0,0 3,3 3,3 0,0 0),(1 1,1 2,2 2,2 1,1 1)))
		))`, 8.0},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := geomFromWKT(t, tt.wkt).Area()
			if got != tt.want {
				t.Errorf("got=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestSignedArea(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "when a polygon is the unit square",
			input:    "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			expected: 1,
		},
		{
			name:     "when a polygon is the unit square wound clockwise",
			input:    "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			expected: -1,
		},
		{
			name: "when a polygon has holes",
			input: `POLYGON(
						(0 0,5 0,5 3,0 3,0 0),
						(1 1,1 2,2 2,2 1,1 1),
						(3 1,3 2,4 2,4 1,3 1)
					)`,
			expected: 13,
		},
		{
			name:     "when a polygon is angular",
			input:    `POLYGON((3 4,5 6,9 5,12 8,5 11,3 4))`,
			expected: 30,
		},
		{
			name:     "when a multipolygon has two polygons",
			input:    `MULTIPOLYGON(((0 0,1 0,1 1,0 1,0 0)),((3 4,5 6,9 5,12 8,5 11,3 4)))`,
			expected: 31,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("input: %s", tc.input)
			geom := geomFromWKT(t, tc.input)
			var got float64
			switch {
			case geom.IsPolygon():
				got = geom.AsPolygon().SignedArea()
			case geom.IsMultiPolygon():
				got = geom.AsMultiPolygon().SignedArea()
			default:
				t.Errorf("expected: Polygon or MultiPolygon but got a different type")
			}
			if got != tc.expected {
				t.Errorf("expected: %f, got: %f", tc.expected, got)
			}
		})
	}

}

func TestCentroid(t *testing.T) {
	for i, tt := range []struct {
		wkt  string
		want XY
	}{
		{"GEOMETRYCOLLECTION(LINESTRING(1 0,0 5,5 2),POINT(2 3),POLYGON((0 0,1 0,0 1,0 0)))", XY{1.0 / 3, 1.0 / 3}},
		{"GEOMETRYCOLLECTION(POLYGON((0 0,1 0,0 1,0 0)),POLYGON((2 0,4 0,4 2,2 2,2 0)))", XY{2.7037037037037, 0.925925925925926}},
		{"GEOMETRYCOLLECTION(LINESTRING(1 0,0 5,5 2),POINT(2 3),MULTIPOLYGON EMPTY)", XY{1.5669656263407472, 3.033482813170374}},
		{"GEOMETRYCOLLECTION(POINT(1 3),MULTIPOINT(1 1,2 2,3 3))", XY{7.0 / 4, 9.0 / 4}},
		{"GEOMETRYCOLLECTION(LINESTRING(0 0,1 1))", XY{1.0 / 2, 1.0 / 2}},
		{"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(LINESTRING(1 2,3 4),POINT(1 5)))", XY{4.0 / 2, 6.0 / 2}},
		{"POINT(1 2)", XY{1.0, 2.0}},
		{"LINESTRING(1 2,3 4)", XY{4.0 / 2, 6.0 / 2}},
		{"LINESTRING(4 3,2 7)", XY{6.0 / 2, 10.0 / 2}},
		{"LINESTRING(0 0,0 1,1 0,0 0)", XY{0.35355339059327373, 0.35355339059327373}},
		{"POLYGON((0 0,1 1,0 1,0 0))", XY{1.0 / 3, 2.0 / 3}},
		{"POLYGON((0 0,0 1,1 1,0 0))", XY{1.0 / 3, 2.0 / 3}},
		{"POLYGON((0 0,1 0,1 1,0 1,0 0))", XY{0.5, 0.5}},
		{"POLYGON((0 0,0 1,1 1,1 0,0 0))", XY{0.5, 0.5}},
		{"POLYGON((0 0,2 0,2 1,0 1,0 0))", XY{1, 0.5}},
		{"POLYGON((0 0,4 0,4 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1))", XY{2 + 1.0/22, 1.5}},
		{"POLYGON((0 0,0 3,3 3,3 0,0 0),(1 1,1 2,2 2,2 1,1 1))", XY{1.5, 1.5}},
		{"MULTIPOINT(-1 0,-1 2,-1 3,-1 4,-1 7,0 1,0 3,1 1,2 0,6 0,7 8,9 8,10 6)", XY{2.30769230769231, 3.30769230769231}},
		{"MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((2 0,4 0,4 2,2 2,2 0)))", XY{2.7037037037037, 0.925925925925926}},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, ok := geomFromWKT(t, tt.wkt).Centroid()
			if !ok {
				t.Fatal("could not generate centroid")
			}
			wantPt := NewPointXY(tt.want)
			if !wantPt.EqualsExact(got.AsGeometry(), Tolerance(0.00000001)) {
				t.Log(tt.wkt)
				t.Errorf("got=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestNoCentroid(t *testing.T) {
	for i, wkt := range []string{
		"GEOMETRYCOLLECTION EMPTY",
		"LINESTRING EMPTY",
		"MULTILINESTRING EMPTY",
		"MULTIPOINT EMPTY",
		"MULTIPOLYGON EMPTY",
		"POLYGON EMPTY",
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, defined := geomFromWKT(t, wkt).Centroid()
			if defined {
				t.Errorf("expected centroid not to be defined, but was")
			}
		})
	}
}

func TestLineStringToMultiLineString(t *testing.T) {
	ls := geomFromWKT(t, "LINESTRING(1 2,3 4,5 6)").AsLineString()
	got := ls.AsMultiLineString()
	want := geomFromWKT(t, "MULTILINESTRING((1 2,3 4,5 6))")
	if !got.EqualsExact(want) {
		t.Errorf("want=%v got=%v", want, got)
	}
}

func TestLineToLineString(t *testing.T) {
	ln := geomFromWKT(t, "LINESTRING(1 2,3 4)").AsLine()
	got := ln.AsLineString()
	if got.NumPoints() != 2 {
		t.Errorf("want num points 2 but got %v", got.NumPoints())
	}
	if got.StartPoint().XY() != (XY{1, 2}) {
		t.Errorf("want start point 1,2 but got %v", got.StartPoint().XY())
	}
	if got.EndPoint().XY() != (XY{3, 4}) {
		t.Errorf("want start point 3,4 but got %v", got.EndPoint().XY())
	}
}

func TestPolygonToMultiPolygon(t *testing.T) {
	p := geomFromWKT(t, "POLYGON((0 0,0 1,1 0,0 0))").AsPolygon()
	mp := p.AsMultiPolygon()
	if mp.AsText() != "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))" {
		t.Errorf("got %v", mp.AsText())
	}
}

func TestReverse(t *testing.T) {
	for i, tt := range []struct {
		wkt, boundary string
	}{
		{"POINT EMPTY", "POINT EMPTY"},
		{"LINESTRING EMPTY", "LINESTRING EMPTY"},
		{"POLYGON EMPTY", "POLYGON EMPTY"},
		{"MULTIPOINT EMPTY", "MULTIPOINT EMPTY"},
		{"MULTILINESTRING EMPTY", "MULTILINESTRING EMPTY"},
		{"MULTIPOLYGON EMPTY", "MULTIPOLYGON EMPTY"},

		{"POINT(1 2)", "POINT(1 2)"},
		{"LINESTRING(1 2,3 4)", "LINESTRING(3 4,1 2)"},
		{"LINESTRING(1 2,3 4,5 6)", "LINESTRING(5 6,3 4,1 2)"},
		{"LINESTRING(1 2,3 4,5 6,7 8)", "LINESTRING(7 8,5 6,3 4,1 2)"},
		{"LINESTRING(0 0,1 0,0 1,0 0)", "LINESTRING(0 0,0 1,1 0,0 0)"},

		{"POLYGON((0 0,1 0,1 1,0 1,0 0))", "POLYGON((0 0,0 1,1 1,1 0,0 0))"},
		{"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1))", "POLYGON((0 0,0 3,3 3,3 0,0 0),(1 1,1 2,2 2,2 1,1 1))"},

		{"MULTIPOINT((1 2))", "MULTIPOINT((1 2))"},
		{"MULTIPOINT((1 2),(3 4))", "MULTIPOINT((1 2),(3 4))"},

		{
			"MULTILINESTRING((0 0,1 1))",
			"MULTILINESTRING((1 1,0 0))",
		},
		{
			"MULTILINESTRING((0 0,1 0),(0 1,1 1))",
			"MULTILINESTRING((1 0,0 0),(1 1,0 1))",
		},
		{
			"MULTILINESTRING((0 0,1 0,1 1),(0 0,0 1,1 1))",
			"MULTILINESTRING((1 1,1 0,0 0),(1 1,0 1,0 0))",
		},

		{
			"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1)),((4 0,5 0,5 1,4 1,4 0)))",
			"MULTIPOLYGON(((0 0,0 3,3 3,3 0,0 0),(1 1,1 2,2 2,2 1,1 1)),((4 0,4 1,5 1,5 0,4 0)))",
		},

		{
			"GEOMETRYCOLLECTION EMPTY",
			"GEOMETRYCOLLECTION EMPTY",
		},
		{
			"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION EMPTY)",
			"GEOMETRYCOLLECTION EMPTY",
		},
		{
			"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION EMPTY,MULTIPOLYGON EMPTY,GEOMETRYCOLLECTION EMPTY)",
			"GEOMETRYCOLLECTION EMPTY",
		},
		{
			"GEOMETRYCOLLECTION(POINT(1 1))",
			"GEOMETRYCOLLECTION(POINT(1 1))",
		},
		{
			`GEOMETRYCOLLECTION(
				LINESTRING(1 0,0 5,5 2),
				POINT(2 3),
				POLYGON((0 0,1 0,0 1,0 0))
			)`,
			`GEOMETRYCOLLECTION(
				LINESTRING(5 2,0 5,1 0),
				POINT(2 3),
				POLYGON((0 0,0 1,1 0,0 0))
			)`,
		},
		{
			`GEOMETRYCOLLECTION(
				LINESTRING(1 0,0 5,5 2),
				POINT(2 3),
				MULTIPOLYGON EMPTY
			)`,
			`GEOMETRYCOLLECTION(
				LINESTRING(5 2,0 5,1 0),
				POINT(2 3),
				MULTIPOLYGON EMPTY
			)`,
		},
		{
			`GEOMETRYCOLLECTION(
				POINT(1 2),
				POINT EMPTY,
				LINESTRING(1 2,3 4),
				GEOMETRYCOLLECTION(
					POINT EMPTY,
					LINESTRING(5 6,7 8),
					GEOMETRYCOLLECTION(POINT EMPTY, POINT(5 6)),
					POINT(3 4)
				),
				GEOMETRYCOLLECTION(
					GEOMETRYCOLLECTION(POINT(9 10), POINT EMPTY),
					POINT(11 12),
					POINT EMPTY
				)
			) `,
			`GEOMETRYCOLLECTION(
				POINT(1 2),
				POINT EMPTY,
				LINESTRING(3 4,1 2),
				GEOMETRYCOLLECTION(
					POINT EMPTY,
					LINESTRING(7 8,5 6),
					GEOMETRYCOLLECTION(POINT EMPTY,POINT(5 6)),
					POINT(3 4)
				),
				GEOMETRYCOLLECTION(
					GEOMETRYCOLLECTION(POINT(9 10), POINT EMPTY),
					POINT(11 12),
					POINT EMPTY
				)
			)`,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			want := geomFromWKT(t, tt.boundary)
			got := geomFromWKT(t, tt.wkt).Reverse()
			expectGeomEq(t, got, want)
		})
	}
}
