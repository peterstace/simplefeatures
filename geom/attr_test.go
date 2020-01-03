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
			g := gFromWKT(t, tt.wkt)
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
			g := gFromWKT(t, wkt)
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
		{"MULTILINESTRING((0 0,0 1,1 1),(0 1,0 0,1 0))", false}, // reproduced a bug
		{"MULTILINESTRING((0 0,1 0),(0 1,1 1))", true},          // reproduced a bug
		{"MULTILINESTRING((1 1,2 2),(1 1,2 2))", false},

		{"MULTIPOLYGON(((0 0,1 0,0 1,0 0)))", true},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := geomFromWKT(t, tt.wkt).(interface{ IsSimple() bool })
			got := g.IsSimple()
			if got != tt.wantSimple {
				t.Logf("wkt: %s", tt.wkt)
				t.Errorf("got=%v want=%v", got, tt.wantSimple)
			}
		})
	}
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
			want := gFromWKT(t, tt.boundary)
			got := gFromWKT(t, tt.wkt).Boundary()
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
			geomFromWKT(t, "POINT(1 2)").(Point).Coordinates(),
			[2]float64{1, 2},
		)
	})
	t.Run("Line-LineString-LinearRing-MultiPoint", func(t *testing.T) {
		for _, tt := range []struct {
			wkt  string
			want [][2]float64
		}{
			{"LINESTRING(0 1,2 3)", [][2]float64{{0, 1}, {2, 3}}},
			{"LINESTRING(0 1,2 3,4 5)", [][2]float64{{0, 1}, {2, 3}, {4, 5}}},
			{"MULTIPOINT(0 1,2 3,4 5)", [][2]float64{{0, 1}, {2, 3}, {4, 5}}},
		} {
			cmp1d(t,
				geomFromWKT(t, tt.wkt).(interface{ Coordinates() []Coordinates }).Coordinates(),
				tt.want,
			)
		}
	})
	t.Run("Polygon-MultiLineString", func(t *testing.T) {
		for _, tt := range []struct {
			wkt  string
			want [][][2]float64
		}{
			{
				"POLYGON((0 0,0 10,10 0,0 0),(2 2,2 7,7 2,2 2))",
				[][][2]float64{
					{{0, 0}, {0, 10}, {10, 0}, {0, 0}},
					{{2, 2}, {2, 7}, {7, 2}, {2, 2}},
				},
			},
			{
				"MULTILINESTRING((0 0,0 10,10 0,0 0),(2 2,2 8,8 2,2 2))",
				[][][2]float64{
					{{0, 0}, {0, 10}, {10, 0}, {0, 0}},
					{{2, 2}, {2, 8}, {8, 2}, {2, 2}},
				},
			},
		} {
			cmp2d(t,
				geomFromWKT(t, tt.wkt).(interface{ Coordinates() [][]Coordinates }).Coordinates(),
				tt.want,
			)
		}

	})
	t.Run("MultiPolygon", func(t *testing.T) {
		const wkt = `MULTIPOLYGON(
			((0 0,0 10,10 0,0 0),(2 2,2 7,7 2,2 2)),
			((100 100,100 110,110 100,100 100),(102 102,102 107,107 102,102 102))
		)`
		cmp3d(t,
			geomFromWKT(t, wkt).(interface{ Coordinates() [][][]Coordinates }).Coordinates(),
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
			want := gFromWKT(t, tt.wktOut)
			expectGeomEq(t, ToGeometry(got), want)
		})
	}
}

func TestIsRing(t *testing.T) {
	for i, tt := range []struct {
		wkt  string
		want bool
	}{
		{"LINESTRING(0 1,2 3)", false},
		{"LINESTRING(0 0,0 1,1 0,0 0)", true},
		{"LINESTRING(0 0,1 1,1 0,0 1,0 0)", false}, // not simple
		{"LINESTRING(0 0,1 0,1 1,0 1)", false},     // not closed
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := geomFromWKT(t, tt.wkt)
			got := g.(interface{ IsRing() bool }).IsRing()
			if got != tt.want {
				t.Logf("WKT: %v", ToGeometry(g).AsText())
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
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := geomFromWKT(t, tt.wkt)
			got := g.(interface{ Length() float64 }).Length()
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
		{"POLYGON((0 0,1 1,0 1,0 0))", 0.5},
		{"POLYGON((0 0,0 1,1 1,0 0))", 0.5},
		{"POLYGON((0 0,0 1,1 1,1 0,0 0))", 1.0},
		{"POLYGON((0 0,0 3,3 3,3 0,0 0),(1 1,1 2,2 2,2 1,1 1))", 8.0},
		{"MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((2 1,1 1,2 0,2 1)))", 1.0},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := geomFromWKT(t, tt.wkt).(interface{ Area() float64 })
			got := g.Area()
			if got != tt.want {
				t.Errorf("got=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestCentroidPolygon(t *testing.T) {
	for i, tt := range []struct {
		wkt  string
		want XY
	}{
		{"POLYGON((0 0,1 1,0 1,0 0))", XY{1.0 / 3, 2.0 / 3}},
		{"POLYGON((0 0,0 1,1 1,0 0))", XY{1.0 / 3, 2.0 / 3}},
		{"POLYGON((0 0,1 0,1 1,0 1,0 0))", XY{0.5, 0.5}},
		{"POLYGON((0 0,0 1,1 1,1 0,0 0))", XY{0.5, 0.5}},
		{"POLYGON((0 0,2 0,2 1,0 1,0 0))", XY{1, 0.5}},
		{"POLYGON((0 0,4 0,4 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1))", XY{2 + 1.0/22, 1.5}},
		{"POLYGON((0 0,0 3,3 3,3 0,0 0),(1 1,1 2,2 2,2 1,1 1))", XY{1.5, 1.5}},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := geomFromWKT(t, tt.wkt).(interface{ Centroid() Point })
			got := g.Centroid()
			wantPt := NewPointXY(tt.want)
			if !wantPt.EqualsExact(got, Tolerance(0.00000001)) {
				t.Log(tt.wkt)
				t.Errorf("got=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestCentroidMultiPolygon(t *testing.T) {
	for i, tt := range []struct {
		wkt  string
		want XY
	}{
		{"MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((2 0,4 0,4 2,2 2,2 0)))", XY{2.7037037037037, 0.925925925925926}},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := geomFromWKT(t, tt.wkt).(interface{ Centroid() (Point, bool) })
			got, ok := g.Centroid()
			if !ok {
				t.Fatal("could not generate centroid")
			}
			wantPt := NewPointXY(tt.want)
			if !wantPt.EqualsExact(got, Tolerance(0.00000001)) {
				t.Log(tt.wkt)
				t.Errorf("got=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestLineStringToMultiLineString(t *testing.T) {
	ls := geomFromWKT(t, "LINESTRING(1 2,3 4,5 6)").(LineString)
	got := ls.AsMultiLineString()
	want := geomFromWKT(t, "MULTILINESTRING((1 2,3 4,5 6))")
	if !got.EqualsExact(want) {
		t.Errorf("want=%v got=%v", want, got)
	}
}

func TestLineToLineString(t *testing.T) {
	ln := geomFromWKT(t, "LINESTRING(1 2,3 4)").(Line)
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
	p := geomFromWKT(t, "POLYGON((0 0,0 1,1 0,0 0))").(Polygon)
	mp := p.AsMultiPolygon()
	if mp.AsText() != "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))" {
		t.Errorf("got %v", mp.AsText())
	}
}
