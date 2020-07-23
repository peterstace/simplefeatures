package geom_test

import (
	"math"
	"strconv"
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
		{"MULTIPOINT(EMPTY)", true, 0},
		{"MULTIPOINT(EMPTY,EMPTY)", true, 0},
		{"MULTIPOINT((1 2),EMPTY)", false, 0},
		{"MULTIPOINT(EMPTY,(1 2))", false, 0},
		{"MULTILINESTRING EMPTY", true, 1},
		{"MULTILINESTRING((0 0,1 1,2 2))", false, 1},
		{"MULTILINESTRING((0 0,1 1,2 2),EMPTY)", false, 1},
		{"MULTILINESTRING(EMPTY,(0 0,1 1,2 2))", false, 1},
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
			geom, err := UnmarshalWKT(tt.wkt)
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
		{"MULTIPOINT(1 1,EMPTY,3 4)", xy(1, 1), xy(3, 4)},
		{"MULTIPOINT(EMPTY,1 1,EMPTY,3 4)", xy(1, 1), xy(3, 4)},
		{"MULTILINESTRING((1 1,3 1,2 2,2 4,1 1),(4 1,4 2))", xy(1, 1), xy(4, 4)},
		{"MULTILINESTRING((4 1,4 2),(1 1,3 1,2 2,2 4,1 1))", xy(1, 1), xy(4, 4)},
		{"MULTILINESTRING((4 1,4 2),EMPTY,(1 1,3 1,2 2,2 4,1 1))", xy(1, 1), xy(4, 4)},
		{"MULTIPOLYGON(((4 1,4 2,3 2,4 1)),((1 1,3 1,2 2,2 4,1 1)))", xy(1, 1), xy(4, 4)},
		{"MULTIPOLYGON(EMPTY,((0 0,0 1,1 0,0 0)))", xy(0, 0), xy(1, 1)},
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
			if env.Min() != tt.min {
				t.Errorf("min: got=%v want=%v", env.Min(), tt.min)
			}
			if env.Max() != tt.max {
				t.Errorf("max: got=%v want=%v", env.Max(), tt.max)
			}
		})
	}
}

func TestNoEnvelope(t *testing.T) {
	for _, wkt := range []string{
		"POINT EMPTY",
		"LINESTRING EMPTY",
		"POLYGON EMPTY",
		"MULTIPOINT EMPTY",
		"MULTIPOINT(EMPTY)",
		"MULTIPOINT(EMPTY,EMPTY)",
		"MULTILINESTRING EMPTY",
		"MULTILINESTRING(EMPTY)",
		"MULTILINESTRING(EMPTY,EMPTY)",
		"MULTIPOLYGON EMPTY",
		"MULTIPOLYGON(EMPTY)",
		"MULTIPOLYGON(EMPTY,EMPTY)",
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
		{"LINESTRING(0 0,0 0,0 1,1 0,0 0)", true},
		{"LINESTRING(0 0,0 1,1 0,0 0,0 0)", true},
		{"LINESTRING(1 2,1 2,3 4,3 4,3 4,5 6,5 6)", true},

		{"POLYGON((0 0,0 1,1 0,0 0))", true},

		{"MULTIPOINT((1 2),(3 4),(5 6))", true},
		{"MULTIPOINT((1 2),(3 4),(1 2))", false},
		{"MULTIPOINT EMPTY", true},
		{"MULTIPOINT((1 2),EMPTY)", true},
		{"MULTIPOINT(EMPTY,(1 2),EMPTY)", true},

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
			t.Logf("WKT: %v", tt.wkt)
			want := geomFromWKT(t, tt.boundary)
			got := geomFromWKT(t, tt.wkt).Boundary()
			expectGeomEq(t, got, want)
		})
	}
}

func TestCoordinatesSequence(t *testing.T) {
	t.Run("point", func(t *testing.T) {
		t.Run("populated", func(t *testing.T) {
			c, ok := geomFromWKT(t, "POINT(1 2)").AsPoint().Coordinates()
			expectBoolEq(t, ok, true)
			expectXYEq(t, c.XY, XY{1, 2})
		})
		t.Run("empty", func(t *testing.T) {
			_, ok := geomFromWKT(t, "POINT EMPTY").AsPoint().Coordinates()
			expectBoolEq(t, ok, false)
		})
	})
	t.Run("linestring", func(t *testing.T) {
		seq := geomFromWKT(t, "LINESTRING(0 1,2 3,4 5)").AsLineString().Coordinates()
		expectIntEq(t, seq.Length(), 3)
		expectXYEq(t, seq.GetXY(0), XY{0, 1})
		expectXYEq(t, seq.GetXY(1), XY{2, 3})
		expectXYEq(t, seq.GetXY(2), XY{4, 5})
	})
	t.Run("linestring with dupe", func(t *testing.T) {
		seq := geomFromWKT(t, "LINESTRING(1 5,5 2,5 2,4 9)").AsLineString().Coordinates()
		expectIntEq(t, seq.Length(), 4)
		expectXYEq(t, seq.GetXY(0), XY{1, 5})
		expectXYEq(t, seq.GetXY(1), XY{5, 2})
		expectXYEq(t, seq.GetXY(2), XY{5, 2})
		expectXYEq(t, seq.GetXY(3), XY{4, 9})
	})
	t.Run("polygon", func(t *testing.T) {
		seq := geomFromWKT(t, "POLYGON((0 0,0 10,10 0,0 0),(2 2,2 7,7 2,2 2))").AsPolygon().Coordinates()
		expectIntEq(t, len(seq), 2)
		expectIntEq(t, seq[0].Length(), 4)
		expectXYEq(t, seq[0].GetXY(0), XY{0, 0})
		expectXYEq(t, seq[0].GetXY(1), XY{0, 10})
		expectXYEq(t, seq[0].GetXY(2), XY{10, 0})
		expectXYEq(t, seq[0].GetXY(3), XY{0, 0})
		expectIntEq(t, seq[1].Length(), 4)
		expectXYEq(t, seq[1].GetXY(0), XY{2, 2})
		expectXYEq(t, seq[1].GetXY(1), XY{2, 7})
		expectXYEq(t, seq[1].GetXY(2), XY{7, 2})
		expectXYEq(t, seq[1].GetXY(3), XY{2, 2})

	})
	t.Run("multipoint", func(t *testing.T) {
		seq, empty := geomFromWKT(t, "MULTIPOINT(0 1,2 3,EMPTY,4 5)").AsMultiPoint().Coordinates()
		expectIntEq(t, seq.Length(), 4)
		expectXYEq(t, seq.GetXY(0), XY{0, 1})
		expectXYEq(t, seq.GetXY(1), XY{2, 3})
		expectXYEq(t, seq.GetXY(2), XY{0, 0})
		expectXYEq(t, seq.GetXY(3), XY{4, 5})
		expectBoolEq(t, empty.Get(0), false)
		expectBoolEq(t, empty.Get(1), false)
		expectBoolEq(t, empty.Get(2), true)
		expectBoolEq(t, empty.Get(3), false)
	})
	t.Run("multilinestring", func(t *testing.T) {
		seq := geomFromWKT(t, "MULTILINESTRING((0 0,0 10,10 0,0 0),(2 2,2 8,8 2,2 2))").AsMultiLineString().Coordinates()
		expectIntEq(t, len(seq), 2)
		expectIntEq(t, seq[0].Length(), 4)
		expectXYEq(t, seq[0].GetXY(0), XY{0, 0})
		expectXYEq(t, seq[0].GetXY(1), XY{0, 10})
		expectXYEq(t, seq[0].GetXY(2), XY{10, 0})
		expectXYEq(t, seq[0].GetXY(3), XY{0, 0})
		expectIntEq(t, seq[1].Length(), 4)
		expectXYEq(t, seq[1].GetXY(0), XY{2, 2})
		expectXYEq(t, seq[1].GetXY(1), XY{2, 8})
		expectXYEq(t, seq[1].GetXY(2), XY{8, 2})
		expectXYEq(t, seq[1].GetXY(3), XY{2, 2})
	})
	t.Run("multipolygon", func(t *testing.T) {
		seq := geomFromWKT(t, `
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
		).AsMultiPolygon().Coordinates()
		expectIntEq(t, len(seq), 2)

		expectIntEq(t, len(seq[0]), 2)
		expectIntEq(t, seq[0][0].Length(), 4)
		expectXYEq(t, seq[0][0].GetXY(0), XY{0, 0})
		expectXYEq(t, seq[0][0].GetXY(1), XY{0, 10})
		expectXYEq(t, seq[0][0].GetXY(2), XY{10, 0})
		expectXYEq(t, seq[0][0].GetXY(3), XY{0, 0})
		expectIntEq(t, seq[0][1].Length(), 4)
		expectXYEq(t, seq[0][1].GetXY(0), XY{2, 2})
		expectXYEq(t, seq[0][1].GetXY(1), XY{2, 7})
		expectXYEq(t, seq[0][1].GetXY(2), XY{7, 2})
		expectXYEq(t, seq[0][1].GetXY(3), XY{2, 2})

		expectIntEq(t, len(seq[1]), 2)
		expectIntEq(t, seq[1][0].Length(), 4)
		expectXYEq(t, seq[1][0].GetXY(0), XY{100, 100})
		expectXYEq(t, seq[1][0].GetXY(1), XY{100, 110})
		expectXYEq(t, seq[1][0].GetXY(2), XY{110, 100})
		expectXYEq(t, seq[1][0].GetXY(3), XY{100, 100})
		expectIntEq(t, seq[1][1].Length(), 4)
		expectXYEq(t, seq[1][1].GetXY(0), XY{102, 102})
		expectXYEq(t, seq[1][1].GetXY(1), XY{102, 107})
		expectXYEq(t, seq[1][1].GetXY(2), XY{107, 102})
		expectXYEq(t, seq[1][1].GetXY(3), XY{102, 102})
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
		{"LINESTRING EMPTY", false},
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

func TestIsClosed(t *testing.T) {
	for i, tt := range []struct {
		wkt  string
		want bool
	}{
		{"LINESTRING EMPTY", false},
		{"LINESTRING(0 0,0 1,1 0,0 0)", true},
		{"LINESTRING(0 0,1 0,1 1,0 1)", false},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := geomFromWKT(t, tt.wkt).AsLineString().IsClosed()
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
			got := geom.Area(SignedArea)
			if got != tc.expected {
				t.Errorf("expected: %f, got: %f", tc.expected, got)
			}
		})
	}

}

func TestTransformedArea(t *testing.T) {
	for i, tt := range []struct {
		wkt  string
		want float64
	}{
		{"GEOMETRYCOLLECTION(POLYGON((0 0,2 0,2 1,0 1,0 0)))", 0.25},
		{"POLYGON((0 0,2 0,2 1,0 1,0 0))", 0.25},
		{"POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1))", 1},
		{"MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((2 2,3 2,2 3,2 2)))", 0.125},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := geomFromWKT(t, tt.wkt).Area(WithTransform(func(xy XY) XY {
				xy.X *= 0.5
				xy.Y *= 0.25
				return xy
			}))
			if got != tt.want {
				t.Errorf("got=%v want=%v", got, tt.want)
			}
		})
	}
}

func TestCentroid(t *testing.T) {
	for i, tt := range []struct {
		input  string
		output string
	}{
		{"POINT EMPTY", "POINT EMPTY"},
		{"POINT Z EMPTY", "POINT EMPTY"},
		{"POINT M EMPTY", "POINT EMPTY"},
		{"POINT ZM EMPTY", "POINT EMPTY"},
		{"POINT(1 2)", "POINT(1 2)"},
		{"POINT Z (1 2 3)", "POINT(1 2)"},
		{"POINT M (1 2 3)", "POINT(1 2)"},
		{"POINT ZM (1 2 3 4)", "POINT(1 2)"},

		{"LINESTRING EMPTY", "POINT EMPTY"},
		{"LINESTRING(1 2,3 4)", "POINT(2 3)"},
		{"LINESTRING(4 3,2 7)", "POINT(3 5)"},
		{"LINESTRING(0 0,0 1,1 0,0 0)", "POINT(0.35355339059327373 0.35355339059327373)"},

		{"POLYGON EMPTY", "POINT EMPTY"},
		{"POLYGON((0 0,1 1,0 1,0 0))", "POINT(0.3333333333 0.6666666666)"},
		{"POLYGON((0 0,0 1,1 1,0 0))", "POINT(0.3333333333 0.6666666666)"},
		{"POLYGON((0 0,1 0,1 1,0 1,0 0))", "POINT(0.5 0.5)"},
		{"POLYGON((0 0,0 1,1 1,1 0,0 0))", "POINT(0.5 0.5)"},
		{"POLYGON((0 0,2 0,2 1,0 1,0 0))", "POINT(1 0.5)"},
		{"POLYGON((0 0,4 0,4 3,0 3,0 0),(1 1,2 1,2 2,1 2,1 1))", "POINT(2.045454545 1.5)"},
		{"POLYGON((0 0,0 3,3 3,3 0,0 0),(1 1,1 2,2 2,2 1,1 1))", "POINT(1.5 1.5)"},
		{"POLYGON((0 0,1 0,1 3,4 3,4 4,0 4,0 0))", "POINT(1.35714285714286 2.64285714285714)"},
		{"POLYGON((151 -33,151.00001 -33,151.00001 -33.00001,151 -33.00001,151 -33))", "POINT(151.000005 -33.000005)"},

		{"MULTIPOINT EMPTY", "POINT EMPTY"},
		{"MULTIPOINT(-1 0,-1 2,-1 3,-1 4,-1 7,0 1,0 3,1 1,2 0,6 0,7 8,9 8,10 6)", "POINT(2.30769230769231 3.30769230769231)"},
		{"MULTIPOINT(EMPTY)", "POINT EMPTY"},
		{"MULTIPOINT(EMPTY,EMPTY)", "POINT EMPTY"},
		{"MULTIPOINT(EMPTY,(1 2),EMPTY)", "POINT(1 2)"},

		{"MULTILINESTRING EMPTY", "POINT EMPTY"},

		{"MULTIPOLYGON EMPTY", "POINT EMPTY"},
		{"MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((2 0,4 0,4 2,2 2,2 0)))", "POINT(2.7037037037037 0.925925925925926)"},
		{"MULTIPOLYGON(((151 -33,151.00001 -33,151.00001 -33.00001,151 -33.00001,151 -33)))", "POINT(151.000005 -33.000005)"},
		{"MULTIPOLYGON(((0 0,0 1,1 1,1 0,0 0)),EMPTY)", "POINT(0.5 0.5)"},

		{"GEOMETRYCOLLECTION EMPTY", "POINT EMPTY"},
		{"GEOMETRYCOLLECTION(POINT EMPTY)", "POINT EMPTY"},
		{"GEOMETRYCOLLECTION(LINESTRING EMPTY)", "POINT EMPTY"},
		{"GEOMETRYCOLLECTION(POLYGON EMPTY)", "POINT EMPTY"},
		{"GEOMETRYCOLLECTION(LINESTRING(1 0,0 5,5 2),POINT(2 3),POLYGON((0 0,1 0,0 1,0 0)))", "POINT(0.3333333333 0.3333333333)"},
		{"GEOMETRYCOLLECTION(POLYGON((0 0,1 0,0 1,0 0)),POLYGON((2 0,4 0,4 2,2 2,2 0)))", "POINT(2.7037037037037 0.925925925925926)"},
		{"GEOMETRYCOLLECTION(LINESTRING(1 0,0 5,5 2),POINT(2 3),MULTIPOLYGON EMPTY)", "POINT(1.5669656263407472 3.033482813170374)"},
		{"GEOMETRYCOLLECTION(POINT(1 3),MULTIPOINT(1 1,2 2,3 3))", "POINT(1.75 2.25)"},
		{"GEOMETRYCOLLECTION(LINESTRING(0 0,1 1))", "POINT(0.5 0.5)"},
		{"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(LINESTRING(1 2,3 4),POINT(1 5)))", "POINT(2 3)"},
		{"GEOMETRYCOLLECTION(POINT EMPTY,POINT(5 5))", "POINT(5 5)"},
		{"GEOMETRYCOLLECTION(POLYGON((151 -33,151.00001 -33,151.00001 -33.00001,151 -33.00001,151 -33)))", "POINT(151.000005 -33.000005)"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := geomFromWKT(t, tt.input).Centroid()
			want := geomFromWKT(t, tt.output)
			if !want.EqualsExact(got.AsGeometry(), ToleranceXY(0.00000001)) {
				t.Log(tt.input)
				t.Errorf("got=%v want=%v", got.AsText(), tt.output)
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
		{"GEOMETRYCOLLECTION EMPTY", "GEOMETRYCOLLECTION EMPTY"},

		{"POINT ZM EMPTY", "POINT ZM EMPTY"},
		{"LINESTRING ZM EMPTY", "LINESTRING ZM EMPTY"},
		{"POLYGON ZM EMPTY", "POLYGON ZM EMPTY"},
		{"MULTIPOINT ZM EMPTY", "MULTIPOINT ZM EMPTY"},
		{"MULTILINESTRING ZM EMPTY", "MULTILINESTRING ZM EMPTY"},
		{"MULTIPOLYGON ZM EMPTY", "MULTIPOLYGON ZM EMPTY"},
		{"GEOMETRYCOLLECTION ZM EMPTY", "GEOMETRYCOLLECTION ZM EMPTY"},

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
			"MULTIPOLYGON M (((0 0 4,1 0 2,0 1 8,0 0 9)),EMPTY)",
			"MULTIPOLYGON M (((0 0 9,0 1 8,1 0 2,0 0 4)),EMPTY)",
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
			"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION EMPTY,MULTIPOLYGON EMPTY,GEOMETRYCOLLECTION EMPTY)",
			"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION EMPTY,MULTIPOLYGON EMPTY,GEOMETRYCOLLECTION EMPTY)",
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
			t.Log("Input", tt.wkt)
			want := geomFromWKT(t, tt.boundary)
			got := geomFromWKT(t, tt.wkt).Reverse()
			expectGeomEq(t, got, want)
		})
	}
}

func TestForceCoordinatesType(t *testing.T) {
	for i, tt := range []struct {
		input  string
		ct     CoordinatesType
		output string
	}{
		{"POINT EMPTY", DimXY, "POINT EMPTY"},
		{"POINT EMPTY", DimXYZ, "POINT Z EMPTY"},
		{"POINT EMPTY", DimXYM, "POINT M EMPTY"},
		{"POINT EMPTY", DimXYZM, "POINT ZM EMPTY"},
		{"POINT Z EMPTY", DimXY, "POINT EMPTY"},
		{"POINT Z EMPTY", DimXYZ, "POINT Z EMPTY"},
		{"POINT Z EMPTY", DimXYM, "POINT M EMPTY"},
		{"POINT Z EMPTY", DimXYZM, "POINT ZM EMPTY"},
		{"POINT M EMPTY", DimXY, "POINT EMPTY"},
		{"POINT M EMPTY", DimXYZ, "POINT Z EMPTY"},
		{"POINT M EMPTY", DimXYM, "POINT M EMPTY"},
		{"POINT M EMPTY", DimXYZM, "POINT ZM EMPTY"},
		{"POINT ZM EMPTY", DimXY, "POINT EMPTY"},
		{"POINT ZM EMPTY", DimXYZ, "POINT Z EMPTY"},
		{"POINT ZM EMPTY", DimXYM, "POINT M EMPTY"},
		{"POINT ZM EMPTY", DimXYZM, "POINT ZM EMPTY"},

		{"POINT(1 2)", DimXY, "POINT(1 2)"},
		{"POINT(1 2)", DimXYZ, "POINT Z (1 2 0)"},
		{"POINT(1 2)", DimXYM, "POINT M (1 2 0)"},
		{"POINT(1 2)", DimXYZM, "POINT ZM (1 2 0 0)"},
		{"POINT Z (1 2 3)", DimXY, "POINT(1 2)"},
		{"POINT Z (1 2 3)", DimXYZ, "POINT Z (1 2 3)"},
		{"POINT Z (1 2 3)", DimXYM, "POINT M (1 2 0)"},
		{"POINT Z (1 2 3)", DimXYZM, "POINT ZM (1 2 3 0)"},
		{"POINT M (1 2 4)", DimXY, "POINT(1 2)"},
		{"POINT M (1 2 4)", DimXYZ, "POINT Z (1 2 0)"},
		{"POINT M (1 2 4)", DimXYM, "POINT M (1 2 4)"},
		{"POINT M (1 2 4)", DimXYZM, "POINT ZM (1 2 0 4)"},
		{"POINT ZM (1 2 3 4)", DimXY, "POINT(1 2)"},
		{"POINT ZM (1 2 3 4)", DimXYZ, "POINT Z (1 2 3)"},
		{"POINT ZM (1 2 3 4)", DimXYM, "POINT M (1 2 4)"},
		{"POINT ZM (1 2 3 4)", DimXYZM, "POINT ZM (1 2 3 4)"},

		{"LINESTRING(1 2,3 4)", DimXY, "LINESTRING(1 2,3 4)"},
		{"LINESTRING(1 2,3 4)", DimXYZ, "LINESTRING Z (1 2 0,3 4 0)"},
		{"LINESTRING(1 2,3 4)", DimXYM, "LINESTRING M (1 2 0,3 4 0)"},
		{"LINESTRING(1 2,3 4)", DimXYZM, "LINESTRING ZM (1 2 0 0,3 4 0 0)"},
		{"LINESTRING Z (1 2 3,4 5 6)", DimXY, "LINESTRING(1 2,4 5)"},
		{"LINESTRING Z (1 2 3,4 5 6)", DimXYZ, "LINESTRING Z (1 2 3,4 5 6)"},
		{"LINESTRING Z (1 2 3,4 5 6)", DimXYM, "LINESTRING M (1 2 0,4 5 0)"},
		{"LINESTRING Z (1 2 3,4 5 6)", DimXYZM, "LINESTRING ZM (1 2 3 0,4 5 6 0)"},
		{"LINESTRING M (1 2 3,4 5 6)", DimXY, "LINESTRING(1 2,4 5)"},
		{"LINESTRING M (1 2 3,4 5 6)", DimXYZ, "LINESTRING Z (1 2 0,4 5 0)"},
		{"LINESTRING M (1 2 3,4 5 6)", DimXYM, "LINESTRING M (1 2 3,4 5 6)"},
		{"LINESTRING M (1 2 3,4 5 6)", DimXYZM, "LINESTRING ZM (1 2 0 3,4 5 0 6)"},
		{"LINESTRING ZM (1 2 3 4,5 6 7 8)", DimXY, "LINESTRING(1 2,5 6)"},
		{"LINESTRING ZM (1 2 3 4,5 6 7 8)", DimXYZ, "LINESTRING Z (1 2 3,5 6 7)"},
		{"LINESTRING ZM (1 2 3 4,5 6 7 8)", DimXYM, "LINESTRING M (1 2 4,5 6 8)"},
		{"LINESTRING ZM (1 2 3 4,5 6 7 8)", DimXYZM, "LINESTRING ZM (1 2 3 4,5 6 7 8)"},

		{"LINESTRING(1 2,3 4,5 6)", DimXY, "LINESTRING(1 2,3 4,5 6)"},
		{"LINESTRING(1 2,3 4,5 6)", DimXYZ, "LINESTRING Z (1 2 0,3 4 0,5 6 0)"},
		{"LINESTRING(1 2,3 4,5 6)", DimXYM, "LINESTRING M (1 2 0,3 4 0,5 6 0)"},
		{"LINESTRING(1 2,3 4,5 6)", DimXYZM, "LINESTRING ZM (1 2 0 0,3 4 0 0,5 6 0 0)"},
		{"LINESTRING Z (1 2 3,4 5 6,7 8 9)", DimXY, "LINESTRING(1 2,4 5,7 8)"},
		{"LINESTRING Z (1 2 3,4 5 6,7 8 9)", DimXYZ, "LINESTRING Z (1 2 3,4 5 6,7 8 9)"},
		{"LINESTRING Z (1 2 3,4 5 6,7 8 9)", DimXYM, "LINESTRING M (1 2 0,4 5 0,7 8 0)"},
		{"LINESTRING Z (1 2 3,4 5 6,7 8 9)", DimXYZM, "LINESTRING ZM (1 2 3 0,4 5 6 0,7 8 9 0)"},
		{"LINESTRING M (1 2 3,4 5 6,7 8 9)", DimXY, "LINESTRING(1 2,4 5,7 8)"},
		{"LINESTRING M (1 2 3,4 5 6,7 8 9)", DimXYZ, "LINESTRING Z (1 2 0,4 5 0,7 8 0)"},
		{"LINESTRING M (1 2 3,4 5 6,7 8 9)", DimXYM, "LINESTRING M (1 2 3,4 5 6,7 8 9)"},
		{"LINESTRING M (1 2 3,4 5 6,7 8 9)", DimXYZM, "LINESTRING ZM (1 2 0 3,4 5 0 6,7 8 0 9)"},
		{"LINESTRING ZM (1 2 3 4,5 6 7 8,9 10 11 12)", DimXY, "LINESTRING(1 2,5 6,9 10)"},
		{"LINESTRING ZM (1 2 3 4,5 6 7 8,9 10 11 12)", DimXYZ, "LINESTRING Z (1 2 3,5 6 7,9 10 11)"},
		{"LINESTRING ZM (1 2 3 4,5 6 7 8,9 10 11 12)", DimXYM, "LINESTRING M (1 2 4,5 6 8,9 10 12)"},
		{"LINESTRING ZM (1 2 3 4,5 6 7 8,9 10 11 12)", DimXYZM, "LINESTRING ZM (1 2 3 4,5 6 7 8,9 10 11 12)"},

		{"POLYGON((0 0,0 1,1 0,0 0))", DimXY, "POLYGON((0 0,0 1,1 0,0 0))"},
		{"POLYGON((0 0,0 1,1 0,0 0))", DimXYZ, "POLYGON Z ((0 0 0,0 1 0,1 0 0,0 0 0))"},
		{"POLYGON((0 0,0 1,1 0,0 0))", DimXYM, "POLYGON M ((0 0 0,0 1 0,1 0 0,0 0 0))"},
		{"POLYGON((0 0,0 1,1 0,0 0))", DimXYZM, "POLYGON ZM ((0 0 0 0,0 1 0 0,1 0 0 0,0 0 0 0))"},
		{"POLYGON Z ((0 0 3,1 0 6,0 1 9,0 0 3))", DimXY, "POLYGON((0 0,1 0,0 1,0 0))"},
		{"POLYGON Z ((0 0 3,1 0 6,0 1 9,0 0 3))", DimXYZ, "POLYGON Z ((0 0 3,1 0 6,0 1 9,0 0 3))"},
		{"POLYGON Z ((0 0 3,1 0 6,0 1 9,0 0 3))", DimXYM, "POLYGON M ((0 0 0,1 0 0,0 1 0,0 0 0))"},
		{"POLYGON Z ((0 0 3,1 0 6,0 1 9,0 0 3))", DimXYZM, "POLYGON ZM ((0 0 3 0,1 0 6 0,0 1 9 0,0 0 3 0))"},
		{"POLYGON M ((0 0 3,1 0 6,0 1 9,0 0 3))", DimXY, "POLYGON((0 0,1 0,0 1,0 0))"},
		{"POLYGON M ((0 0 3,1 0 6,0 1 9,0 0 3))", DimXYZ, "POLYGON Z ((0 0 0,1 0 0,0 1 0,0 0 0))"},
		{"POLYGON M ((0 0 3,1 0 6,0 1 9,0 0 3))", DimXYM, "POLYGON M ((0 0 3,1 0 6,0 1 9,0 0 3))"},
		{"POLYGON M ((0 0 3,1 0 6,0 1 9,0 0 3))", DimXYZM, "POLYGON ZM ((0 0 0 3,1 0 0 6,0 1 0 9,0 0 0 3))"},
		{"POLYGON ZM ((0 0 3 4,1 0 7 8,0 1 11 12,0 0 3 4))", DimXY, "POLYGON((0 0,1 0,0 1,0 0))"},
		{"POLYGON ZM ((0 0 3 4,1 0 7 8,0 1 11 12,0 0 3 4))", DimXYZ, "POLYGON Z ((0 0 3,1 0 7,0 1 11,0 0 3))"},
		{"POLYGON ZM ((0 0 3 4,1 0 7 8,0 1 11 12,0 0 3 4))", DimXYM, "POLYGON M ((0 0 4,1 0 8,0 1 12,0 0 4))"},
		{"POLYGON ZM ((0 0 3 4,1 0 7 8,0 1 11 12,0 0 3 4))", DimXYZM, "POLYGON ZM ((0 0 3 4,1 0 7 8,0 1 11 12,0 0 3 4))"},

		{"MULTIPOINT(1 2,3 4,5 6)", DimXY, "MULTIPOINT(1 2,3 4,5 6)"},
		{"MULTIPOINT(1 2,3 4,5 6)", DimXYZ, "MULTIPOINT Z (1 2 0,3 4 0,5 6 0)"},
		{"MULTIPOINT(1 2,3 4,5 6)", DimXYM, "MULTIPOINT M (1 2 0,3 4 0,5 6 0)"},
		{"MULTIPOINT(1 2,3 4,5 6)", DimXYZM, "MULTIPOINT ZM (1 2 0 0,3 4 0 0,5 6 0 0)"},
		{"MULTIPOINT Z (1 2 3,4 5 6,7 8 9)", DimXY, "MULTIPOINT(1 2,4 5,7 8)"},
		{"MULTIPOINT Z (1 2 3,4 5 6,7 8 9)", DimXYZ, "MULTIPOINT Z (1 2 3,4 5 6,7 8 9)"},
		{"MULTIPOINT Z (1 2 3,4 5 6,7 8 9)", DimXYM, "MULTIPOINT M (1 2 0,4 5 0,7 8 0)"},
		{"MULTIPOINT Z (1 2 3,4 5 6,7 8 9)", DimXYZM, "MULTIPOINT ZM (1 2 3 0,4 5 6 0,7 8 9 0)"},
		{"MULTIPOINT M (1 2 3,4 5 6,7 8 9)", DimXY, "MULTIPOINT(1 2,4 5,7 8)"},
		{"MULTIPOINT M (1 2 3,4 5 6,7 8 9)", DimXYZ, "MULTIPOINT Z (1 2 0,4 5 0,7 8 0)"},
		{"MULTIPOINT M (1 2 3,4 5 6,7 8 9)", DimXYM, "MULTIPOINT M (1 2 3,4 5 6,7 8 9)"},
		{"MULTIPOINT M (1 2 3,4 5 6,7 8 9)", DimXYZM, "MULTIPOINT ZM (1 2 0 3,4 5 0 6,7 8 0 9)"},
		{"MULTIPOINT ZM (1 2 3 4,5 6 7 8,9 10 11 12)", DimXY, "MULTIPOINT(1 2,5 6,9 10)"},
		{"MULTIPOINT ZM (1 2 3 4,5 6 7 8,9 10 11 12)", DimXYZ, "MULTIPOINT Z (1 2 3,5 6 7,9 10 11)"},
		{"MULTIPOINT ZM (1 2 3 4,5 6 7 8,9 10 11 12)", DimXYM, "MULTIPOINT M (1 2 4,5 6 8,9 10 12)"},
		{"MULTIPOINT ZM (1 2 3 4,5 6 7 8,9 10 11 12)", DimXYZM, "MULTIPOINT ZM (1 2 3 4,5 6 7 8,9 10 11 12)"},

		{"MULTILINESTRING((1 2,3 4,5 6))", DimXY, "MULTILINESTRING((1 2,3 4,5 6))"},
		{"MULTILINESTRING((1 2,3 4,5 6))", DimXYZ, "MULTILINESTRING Z ((1 2 0,3 4 0,5 6 0))"},
		{"MULTILINESTRING((1 2,3 4,5 6))", DimXYM, "MULTILINESTRING M ((1 2 0,3 4 0,5 6 0))"},
		{"MULTILINESTRING((1 2,3 4,5 6))", DimXYZM, "MULTILINESTRING ZM ((1 2 0 0,3 4 0 0,5 6 0 0))"},
		{"MULTILINESTRING Z ((1 2 3,4 5 6,7 8 9))", DimXY, "MULTILINESTRING((1 2,4 5,7 8))"},
		{"MULTILINESTRING Z ((1 2 3,4 5 6,7 8 9))", DimXYZ, "MULTILINESTRING Z ((1 2 3,4 5 6,7 8 9))"},
		{"MULTILINESTRING Z ((1 2 3,4 5 6,7 8 9))", DimXYM, "MULTILINESTRING M ((1 2 0,4 5 0,7 8 0))"},
		{"MULTILINESTRING Z ((1 2 3,4 5 6,7 8 9))", DimXYZM, "MULTILINESTRING ZM ((1 2 3 0,4 5 6 0,7 8 9 0))"},
		{"MULTILINESTRING M ((1 2 3,4 5 6,7 8 9))", DimXY, "MULTILINESTRING((1 2,4 5,7 8))"},
		{"MULTILINESTRING M ((1 2 3,4 5 6,7 8 9))", DimXYZ, "MULTILINESTRING Z ((1 2 0,4 5 0,7 8 0))"},
		{"MULTILINESTRING M ((1 2 3,4 5 6,7 8 9))", DimXYM, "MULTILINESTRING M ((1 2 3,4 5 6,7 8 9))"},
		{"MULTILINESTRING M ((1 2 3,4 5 6,7 8 9))", DimXYZM, "MULTILINESTRING ZM ((1 2 0 3,4 5 0 6,7 8 0 9))"},
		{"MULTILINESTRING ZM ((1 2 3 4,5 6 7 8,9 10 11 12))", DimXY, "MULTILINESTRING((1 2,5 6,9 10))"},
		{"MULTILINESTRING ZM ((1 2 3 4,5 6 7 8,9 10 11 12))", DimXYZ, "MULTILINESTRING Z ((1 2 3,5 6 7,9 10 11))"},
		{"MULTILINESTRING ZM ((1 2 3 4,5 6 7 8,9 10 11 12))", DimXYM, "MULTILINESTRING M ((1 2 4,5 6 8,9 10 12))"},
		{"MULTILINESTRING ZM ((1 2 3 4,5 6 7 8,9 10 11 12))", DimXYZM, "MULTILINESTRING ZM ((1 2 3 4,5 6 7 8,9 10 11 12))"},

		{"MULTIPOLYGON(((0 0,0 1,1 0,0 0)))", DimXY, "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))"},
		{"MULTIPOLYGON(((0 0,0 1,1 0,0 0)))", DimXYZ, "MULTIPOLYGON Z (((0 0 0,0 1 0,1 0 0,0 0 0)))"},
		{"MULTIPOLYGON(((0 0,0 1,1 0,0 0)))", DimXYM, "MULTIPOLYGON M (((0 0 0,0 1 0,1 0 0,0 0 0)))"},
		{"MULTIPOLYGON(((0 0,0 1,1 0,0 0)))", DimXYZM, "MULTIPOLYGON ZM (((0 0 0 0,0 1 0 0,1 0 0 0,0 0 0 0)))"},
		{"MULTIPOLYGON Z (((0 0 3,1 0 6,0 1 9,0 0 3)))", DimXY, "MULTIPOLYGON(((0 0,1 0,0 1,0 0)))"},
		{"MULTIPOLYGON Z (((0 0 3,1 0 6,0 1 9,0 0 3)))", DimXYZ, "MULTIPOLYGON Z (((0 0 3,1 0 6,0 1 9,0 0 3)))"},
		{"MULTIPOLYGON Z (((0 0 3,1 0 6,0 1 9,0 0 3)))", DimXYM, "MULTIPOLYGON M (((0 0 0,1 0 0,0 1 0,0 0 0)))"},
		{"MULTIPOLYGON Z (((0 0 3,1 0 6,0 1 9,0 0 3)))", DimXYZM, "MULTIPOLYGON ZM (((0 0 3 0,1 0 6 0,0 1 9 0,0 0 3 0)))"},
		{"MULTIPOLYGON M (((0 0 3,1 0 6,0 1 9,0 0 3)))", DimXY, "MULTIPOLYGON(((0 0,1 0,0 1,0 0)))"},
		{"MULTIPOLYGON M (((0 0 3,1 0 6,0 1 9,0 0 3)))", DimXYZ, "MULTIPOLYGON Z (((0 0 0,1 0 0,0 1 0,0 0 0)))"},
		{"MULTIPOLYGON M (((0 0 3,1 0 6,0 1 9,0 0 3)))", DimXYM, "MULTIPOLYGON M (((0 0 3,1 0 6,0 1 9,0 0 3)))"},
		{"MULTIPOLYGON M (((0 0 3,1 0 6,0 1 9,0 0 3)))", DimXYZM, "MULTIPOLYGON ZM (((0 0 0 3,1 0 0 6,0 1 0 9,0 0 0 3)))"},
		{"MULTIPOLYGON ZM (((0 0 3 4,1 0 7 8,0 1 11 12,0 0 3 4)))", DimXY, "MULTIPOLYGON(((0 0,1 0,0 1,0 0)))"},
		{"MULTIPOLYGON ZM (((0 0 3 4,1 0 7 8,0 1 11 12,0 0 3 4)))", DimXYZ, "MULTIPOLYGON Z (((0 0 3,1 0 7,0 1 11,0 0 3)))"},
		{"MULTIPOLYGON ZM (((0 0 3 4,1 0 7 8,0 1 11 12,0 0 3 4)))", DimXYM, "MULTIPOLYGON M (((0 0 4,1 0 8,0 1 12,0 0 4)))"},
		{"MULTIPOLYGON ZM (((0 0 3 4,1 0 7 8,0 1 11 12,0 0 3 4)))", DimXYZM, "MULTIPOLYGON ZM (((0 0 3 4,1 0 7 8,0 1 11 12,0 0 3 4)))"},

		{"GEOMETRYCOLLECTION(POINT(1 2))", DimXY, "GEOMETRYCOLLECTION(POINT(1 2))"},
		{"GEOMETRYCOLLECTION(POINT(1 2))", DimXYZ, "GEOMETRYCOLLECTION Z (POINT Z (1 2 0))"},
		{"GEOMETRYCOLLECTION(POINT(1 2))", DimXYM, "GEOMETRYCOLLECTION M (POINT M (1 2 0))"},
		{"GEOMETRYCOLLECTION(POINT(1 2))", DimXYZM, "GEOMETRYCOLLECTION ZM (POINT ZM (1 2 0 0))"},
		{"GEOMETRYCOLLECTION Z (POINT Z (1 2 3))", DimXY, "GEOMETRYCOLLECTION(POINT(1 2))"},
		{"GEOMETRYCOLLECTION Z (POINT Z (1 2 3))", DimXYZ, "GEOMETRYCOLLECTION Z (POINT Z (1 2 3))"},
		{"GEOMETRYCOLLECTION Z (POINT Z (1 2 3))", DimXYM, "GEOMETRYCOLLECTION M (POINT M (1 2 0))"},
		{"GEOMETRYCOLLECTION Z (POINT Z (1 2 3))", DimXYZM, "GEOMETRYCOLLECTION ZM (POINT ZM (1 2 3 0))"},
		{"GEOMETRYCOLLECTION M (POINT M (1 2 4))", DimXY, "GEOMETRYCOLLECTION(POINT(1 2))"},
		{"GEOMETRYCOLLECTION M (POINT M (1 2 4))", DimXYZ, "GEOMETRYCOLLECTION Z (POINT Z (1 2 0))"},
		{"GEOMETRYCOLLECTION M (POINT M (1 2 4))", DimXYM, "GEOMETRYCOLLECTION M (POINT M (1 2 4))"},
		{"GEOMETRYCOLLECTION M (POINT M (1 2 4))", DimXYZM, "GEOMETRYCOLLECTION ZM (POINT ZM (1 2 0 4))"},
		{"GEOMETRYCOLLECTION ZM (POINT ZM (1 2 3 4))", DimXY, "GEOMETRYCOLLECTION(POINT(1 2))"},
		{"GEOMETRYCOLLECTION ZM (POINT ZM (1 2 3 4))", DimXYZ, "GEOMETRYCOLLECTION Z (POINT Z (1 2 3))"},
		{"GEOMETRYCOLLECTION ZM (POINT ZM (1 2 3 4))", DimXYM, "GEOMETRYCOLLECTION M (POINT M (1 2 4))"},
		{"GEOMETRYCOLLECTION ZM (POINT ZM (1 2 3 4))", DimXYZM, "GEOMETRYCOLLECTION ZM (POINT ZM (1 2 3 4))"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log("input", tt.input)
			t.Log("ct", tt.ct)
			got := geomFromWKT(t, tt.input).ForceCoordinatesType(tt.ct)
			want := geomFromWKT(t, tt.output)
			expectGeomEq(t, got, want)
		})
	}
}

func TestForceWindingDirection(t *testing.T) {
	for _, tt := range []struct {
		desc     string
		input    string
		forceCW  string
		forceCCW string
	}{
		{
			desc:     "point",
			input:    "POINT(4 5)",
			forceCW:  "POINT(4 5)",
			forceCCW: "POINT(4 5)",
		},
		{
			desc:     "polygon with outer ring wound CW",
			input:    "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			forceCW:  "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			forceCCW: "POLYGON((0 0,1 0,1 1,0 1,0 0))",
		},
		{
			desc:     "polygon with outer ring wound CCW",
			input:    "POLYGON((0 0,1 0,1 1,0 1,0 0))",
			forceCW:  "POLYGON((0 0,0 1,1 1,1 0,0 0))",
			forceCCW: "POLYGON((0 0,1 0,1 1,0 1,0 0))",
		},
		{
			desc:     "polygon with outer ring wound CW and inner ring wound CW",
			input:    "POLYGON((0 0,0 4,4 0,0 0),(1 1,1 2,2 1,1 1))",
			forceCW:  "POLYGON((0 0,0 4,4 0,0 0),(1 1,2 1,1 2,1 1))",
			forceCCW: "POLYGON((0 0,4 0,0 4,0 0),(1 1,1 2,2 1,1 1))",
		},
		{
			desc:     "polygon with outer ring wound CW and inner ring wound CCW",
			input:    "POLYGON((0 0,0 4,4 0,0 0),(1 1,2 1,1 2,1 1))",
			forceCW:  "POLYGON((0 0,0 4,4 0,0 0),(1 1,2 1,1 2,1 1))",
			forceCCW: "POLYGON((0 0,4 0,0 4,0 0),(1 1,1 2,2 1,1 1))",
		},
		{
			desc:     "polygon with outer ring wound CCW and inner ring wound CW",
			input:    "POLYGON((0 0,4 0,0 4,0 0),(1 1,1 2,2 1,1 1))",
			forceCW:  "POLYGON((0 0,0 4,4 0,0 0),(1 1,2 1,1 2,1 1))",
			forceCCW: "POLYGON((0 0,4 0,0 4,0 0),(1 1,1 2,2 1,1 1))",
		},
		{
			desc:     "polygon with outer ring wound CCW and inner ring wound CCW",
			input:    "POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))",
			forceCW:  "POLYGON((0 0,0 4,4 0,0 0),(1 1,2 1,1 2,1 1))",
			forceCCW: "POLYGON((0 0,4 0,0 4,0 0),(1 1,1 2,2 1,1 1))",
		},
		//{
		//	desc:     "polygon with outer ring wound CW and inner rings mixed",
		//	input:    "",
		//	forceCW:  "",
		//	forceCCW: "",
		//},
		//{
		//	desc:     "polygon with outer ring wound CCW and inner rings mixed",
		//	input:    "",
		//	forceCW:  "",
		//	forceCCW: "",
		//},
		//{
		//	desc:     "multipolygon with single poly wound CW",
		//	input:    "",
		//	forceCW:  "",
		//	forceCCW: "",
		//},
		//{
		//	desc:     "multipolygon with single poly wound CCW",
		//	input:    "",
		//	forceCW:  "",
		//	forceCCW: "",
		//},
		//{
		//	desc:     "multipolygon with two polys of mixed winding",
		//	input:    "",
		//	forceCW:  "",
		//	forceCCW: "",
		//},
		//{
		//	desc:     "geometry collection containing poly and multipoly",
		//	input:    "",
		//	forceCW:  "",
		//	forceCCW: "",
		//},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			t.Run("ForceCW", func(t *testing.T) {
				got := geomFromWKT(t, tt.input).ForceCW()
				want := geomFromWKT(t, tt.forceCW)
				expectGeomEq(t, got, want)
			})
			t.Run("ForceCCW", func(t *testing.T) {
				got := geomFromWKT(t, tt.input).ForceCCW()
				want := geomFromWKT(t, tt.forceCCW)
				expectGeomEq(t, got, want)
			})
		})
	}
}
