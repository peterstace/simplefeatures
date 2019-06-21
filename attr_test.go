package simplefeatures_test

import (
	"strconv"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures"
)

func TestIsEmptyDimension(t *testing.T) {
	for _, tt := range []struct {
		wkt       string
		wantEmpty bool
		wantDim   int
	}{
		{"POINT EMPTY", true, 0},
		{"POINT(1 1)", false, 0},
		{"LINESTRING EMPTY", true, 0},
		{"LINESTRING(0 0,1 1)", false, 1},
		{"LINESTRING(0 0,1 1,2 2)", false, 1},
		{"LINESTRING(0 0,1 1,1 0,0 0)", false, 1},
		{"POLYGON EMPTY", true, 0},
		{"POLYGON((0 0,1 1,1 0,0 0))", false, 2},
		{"MULTIPOINT EMPTY", true, 0},
		{"MULTIPOINT((0 0))", false, 0},
		{"MULTIPOINT((0 0),(1 1))", false, 0},
		{"MULTILINESTRING EMPTY", true, 0},
		{"MULTILINESTRING((0 0,1 1,2 2))", false, 1},
		{"MULTILINESTRING(EMPTY)", true, 0},
		{"MULTIPOLYGON EMPTY", true, 0},
		{"MULTIPOLYGON(((0 0,1 0,1 1,0 0)))", false, 2},
		{"MULTIPOLYGON(((0 0,1 0,1 1,0 0)))", false, 2},
		{"MULTIPOLYGON(EMPTY)", true, 0},
		{"GEOMETRYCOLLECTION EMPTY", true, 0},
		{"GEOMETRYCOLLECTION(POINT EMPTY)", true, 0},
		{"GEOMETRYCOLLECTION(POLYGON EMPTY)", true, 0},
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

func TestIsEmptyAndDimensionLinearRing(t *testing.T) {
	// Tested on its own, since it cannot be constructed from WKT.
	// TODO: It now can be represented using WKT, so this can be merged into the other test.
	ring, err := NewLinearRing([]Coordinates{
		{XY: XY{0, 0}}, {XY: XY{1, 0}}, {XY: XY{1, 1}}, {XY: XY{0, 0}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if ring.IsEmpty() {
		t.Errorf("expected to not be empty")
	}
	if ring.Dimension() != 1 {
		t.Errorf("expected dimension 1")
	}
}

func TestFiniteNumberOfPoints(t *testing.T) {
	for i, tt := range []struct {
		wkt        string
		isFinite   bool
		pointCount int
	}{
		{"POINT EMPTY", true, 0},
		{"POINT(1 1) ", true, 1},
		{"MULTIPOINT(1 1,2 2,1 1)", true, 2},
		{"LINESTRING(0 0,1 1)", false, 0},
		{"LINESTRING(0 0,1 1,2 2)", false, 0},
		{"LINEARRING(0 0,0 1,2 2,0 0)", false, 0},
		{"POLYGON((0 0,0 1,2 2,0 0))", false, 0},
		{"MULTIPOLYGON(((0 0,0 1,2 2,0 0)))", false, 0},
		{"MULTIPOLYGON EMPTY", true, 0},
		{"MULTILINESTRING((0 0,1 1,2 2))", false, 0},
		{"MULTILINESTRING EMPTY", true, 0},
		{"GEOMETRYCOLLECTION EMPTY", true, 0},
		{"GEOMETRYCOLLECTION(POINT(0 1))", true, 1},
		{"GEOMETRYCOLLECTION(POINT(0 1),POINT(2 2))", true, 2},
		{"GEOMETRYCOLLECTION(POINT(0 1),POINT(0 1))", true, 1},
		{"GEOMETRYCOLLECTION(MULTIPOINT(0 1,1 1),POINT(0 1))", true, 2},
		{"GEOMETRYCOLLECTION(POINT(0 1),LINESTRING(0 0,1 1))", false, 0},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			geom := geomFromWKT(t, tt.wkt)
			gotPointCount, gotIsFinite := geom.FiniteNumberOfPoints()
			if tt.isFinite {
				if !gotIsFinite {
					t.Errorf("want finite but got not finite")
				}
				if gotPointCount != tt.pointCount {
					t.Errorf("point count: want=%d got=%d", tt.pointCount, gotPointCount)
				}
			} else {
				if gotIsFinite {
					t.Errorf("want not finite but got finite")
				}
			}
		})
	}
}

func TestEnvelope(t *testing.T) {
	for i, tt := range []struct {
		wkt string
		min XY
		max XY
	}{
		{"POINT(1 1)", XY{1, 1}, XY{1, 1}},
		{"LINESTRING(1 2,3 4)", XY{1, 2}, XY{3, 4}},
		{"LINESTRING(4 1,2 3)", XY{2, 1}, XY{4, 3}},
		{"LINESTRING(1 1,3 1,2 2,2 4)", XY{1, 1}, XY{3, 4}},
		{"LINEARRING(1 1,3 1,2 2,2 4,1 1)", XY{1, 1}, XY{3, 4}},
		{"POLYGON((1 1,3 1,2 2,2 4,1 1))", XY{1, 1}, XY{3, 4}},
		{"MULTIPOINT(1 1,3 1,2 2,2 4,1 1)", XY{1, 1}, XY{3, 4}},
		{"MULTILINESTRING((1 1,3 1,2 2,2 4,1 1),(4 1,4 2))", XY{1, 1}, XY{4, 4}},
		{"MULTILINESTRING((4 1,4 2),(1 1,3 1,2 2,2 4,1 1))", XY{1, 1}, XY{4, 4}},
		{"MULTIPOLYGON(((4 1,4 2,3 2,4 1)),((1 1,3 1,2 2,2 4,1 1)))", XY{1, 1}, XY{4, 4}},
		{"GEOMETRYCOLLECTION(POINT(4 1),POINT(2 3))", XY{2, 1}, XY{4, 3}},
		{"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION(POINT(4 1),POINT(2 3)))", XY{2, 1}, XY{4, 3}},
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
