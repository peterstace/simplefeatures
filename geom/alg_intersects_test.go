package geom_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestIntersects(t *testing.T) {
	for i, tt := range []struct {
		in1, in2 string
		want     bool
	}{
		// Empty/ANY
		{"POINT EMPTY", "POINT(2 3)", false},
		{"POLYGON EMPTY", "POINT(2 3)", false},
		{"LINESTRING EMPTY", "POINT(2 3)", false},

		// Empty/Empty
		{"POINT EMPTY", "LINESTRING EMPTY", false},
		{"POLYGON EMPTY", "GEOMETRYCOLLECTION EMPTY", false},
		{"POLYGON EMPTY", "GEOMETRYCOLLECTION(POLYGON EMPTY)", false},

		// Point/Point
		{"POINT(1 2)", "POINT(1 2)", true},
		{"POINT(1 2)", "POINT(2 1)", false},

		// Point/Line
		{"POINT(0 0)", "LINESTRING(0 0,2 2)", true},
		{"POINT(1 1)", "LINESTRING(0 0,2 2)", true},
		{"POINT(2 2)", "LINESTRING(0 0,2 2)", true},
		{"POINT(3 3)", "LINESTRING(0 0,2 2)", false},
		{"POINT(-1 -1)", "LINESTRING(0 0,2 2)", false},
		{"POINT(0 2)", "LINESTRING(0 0,2 2)", false},
		{"POINT(2 0)", "LINESTRING(0 0,2 2)", false},
		{"POINT(0 3.14)", "LINESTRING(0 0,0 4)", true},
		{"POINT(1 0.25)", "LINESTRING(0 0,4 1)", true},
		{"POINT(2 0.5)", "LINESTRING(0 0,4 1)", true},

		// Point/LineString
		{"POINT(0 0)", "LINESTRING(1 0,2 1,3 0)", false},
		{"POINT(1 0)", "LINESTRING(1 0,2 1,3 0)", true},
		{"POINT(2 1)", "LINESTRING(1 0,2 1,3 0)", true},
		{"POINT(1.5 0.5)", "LINESTRING(1 0,2 1,3 0)", true},
		{"POINT(1 2)", "LINESTRING(0 0,0 4)", false},

		// Point/Polygon
		{`POINT(1 2)`, `POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, true},
		{`POINT(2.5 1.5)`, `POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, true},
		{`POINT(4 1)`, `POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, true},
		{`POINT(5 3)`, `POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, true},
		{`POINT(1.5 1.5)`, `POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, false},
		{`POINT(3.5 1.5)`, `POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, false},
		{`POINT(6 2)`, `POLYGON(
			(0 0,5 0,5 3,0 3,0 0),
			(1 1,2 1,2 2,1 2,1 1),
			(3 1,4 1,4 2,3 2,3 1)
		)`, false},

		// Point/MultiLineString
		{"POINT(0 0)", "MULTILINESTRING((1 0,2 1,3 0))", false},
		{"POINT(1 0)", "MULTILINESTRING((1 0,2 1,3 0))", true},
		{"POINT(0 0)", "MULTILINESTRING((0 0,1 1),(1 1,2 1))", true},
		{"POINT(1 1)", "MULTILINESTRING((0 0,1 1),(1 1,2 1))", true},
		{"POINT(2 1)", "MULTILINESTRING((0 0,1 1),(1 1,2 1))", true},
		{"POINT(3 1)", "MULTILINESTRING((0 0,1 1),(1 1,2 1))", false},

		// Point/MultiPolygon
		{"POINT(0 0)", "MULTIPOLYGON(((0 0,1 0,1 1,0 0)))", true},
		{"POINT(1 1)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)))", true},
		{"POINT(1 1)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((3 0,5 0,5 2,3 2,3 0)))", true},
		{"POINT(4 1)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((3 0,5 0,5 2,3 2,3 0)))", true},
		{"POINT(6 1)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((3 0,5 0,5 2,3 2,3 0)))", false},

		// Line/Line
		{"LINESTRING(0 0,0 1)", "LINESTRING(0 0,1 0)", true},
		{"LINESTRING(0 1,1 1)", "LINESTRING(1 0,1 1)", true},
		{"LINESTRING(0 1,0 0)", "LINESTRING(0 0,1 0)", true},
		{"LINESTRING(0 0,0 1)", "LINESTRING(1 0,0 0)", true},
		{"LINESTRING(0 0,1 0)", "LINESTRING(1 0,2 0)", true},
		{"LINESTRING(0 0,1 0)", "LINESTRING(2 0,3 0)", false},
		{"LINESTRING(1 0,2 0)", "LINESTRING(0 0,3 0)", true},
		{"LINESTRING(0 0,0 1)", "LINESTRING(1 0,1 1)", false},
		{"LINESTRING(0 0,1 1)", "LINESTRING(1 0,0 1)", true},
		{"LINESTRING(1 0,0 1)", "LINESTRING(0 1,1 0)", true},
		{"LINESTRING(1 0,0 1)", "LINESTRING(1 0,0 1)", true},
		{"LINESTRING(0 0,1 1)", "LINESTRING(1 1,0 0)", true},
		{"LINESTRING(0 0,1 1)", "LINESTRING(0 0,1 1)", true},
		{"LINESTRING(0 0,0 1)", "LINESTRING(0 1,0 0)", true},
		{"LINESTRING(0 0,0 1)", "LINESTRING(0 0,0 1)", true},
		{"LINESTRING(0 0,1 0)", "LINESTRING(1 0,0 0)", true},
		{"LINESTRING(0 0,1 0)", "LINESTRING(0 0,1 0)", true},
		{"LINESTRING(1 1,2 2)", "LINESTRING(0 0,3 3)", true},
		{"LINESTRING(3 1,2 2)", "LINESTRING(1 3,2 2)", true},

		// Line/LineString
		{"LINESTRING(0 0,1 1)", "LINESTRING(0 0,1 1,0 0)", true},
		{"LINESTRING(0 0,1 1)", "LINESTRING(0 0,1 1,0 1,1 0)", true},

		// Line/Polygon
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(3 0,3 2)", false},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(1 2.1,2.1 1)", true},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(1 -1,1 1)", true},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(0.25 0.25,0.75 0.75)", true},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(2 0,3 -1)", true},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(-1 1,1 -1)", true},

		// Line/MultiPoint
		{"LINESTRING(0 0,1 1)", "MULTIPOINT EMPTY", false},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(1 0)", false},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(1 0,0 1)", false},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(0.5 0.5)", true},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(0 0)", true},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(0.5 0.5,1 0)", true},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(1 1,0 1)", true},
		{"LINESTRING(1 2,4 5)", "MULTIPOINT((7 6),(3 3),(3 3))", false},
		{"LINESTRING(2 1,3 6)", "MULTIPOINT((1 2))", false},

		// Line/MultiLineString
		{"LINESTRING(0 0,1 1)", "MULTILINESTRING((0 0.5,1 0.5,1 -0.5),(2 0.5,2 -0.5))", true},
		{"LINESTRING(0 1,1 2)", "MULTILINESTRING((0 0.5,1 0.5,1 -0.5),(2 0.5,2 -0.5))", false},

		// Line/MultiPolygon
		{"LINESTRING(5 2,5 4)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", false},
		{"LINESTRING(3 3,3 5)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},
		{"LINESTRING(1 1,3 1)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},
		{"LINESTRING(0 2,2 4)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},

		// LineString/LineString
		{"LINESTRING(0 0,1 0,1 1,0 1)", "LINESTRING(1 1,2 1,2 2,1 2)", true},
		{"LINESTRING(0 0,0 1,1 0,0 0)", "LINESTRING(0 0,1 1,0 1,0 0,1 1)", true},
		{"LINESTRING(0 0,1 0,0 1,0 0)", "LINESTRING(0 0,1 0,1 1,0 1)", true},
		{"LINESTRING(0 0,1 0,1 1,0 1)", "LINESTRING(1 1,2 1,2 2,1 2,1 1)", true},
		{"LINESTRING(0 0,1 0,1 1,0 1,0 0)", "LINESTRING(2 2,3 2,3 3,2 3,2 2)", false},
		{"LINESTRING(0 0,1 0,1 1,0 1,0 0)", "LINESTRING(1 1,2 1,2 2,1 2,1 1)", true},
		{"LINESTRING(0 0,1 0,1 1,0 1,0 0)", "LINESTRING(1 0,2 0,2 1,1 1,1 0)", true},
		{"LINESTRING(0 0,1 0,0 1,0 0)", "LINESTRING(1 0,1 1,0 1,1 0)", true},
		{"LINESTRING(0 0,1 0,1 1,0 1,0 0)", "LINESTRING(0.5 0.5,1.5 0.5,1.5 1.5,0.5 1.5,0.5 0.5)", true},
		{"LINESTRING(0 0,1 0,1 1,0 1,0 0)", "LINESTRING(1 0,2 0,2 1,1 1,1.5 0.5,1 0.5,1 0)", true},
		{"LINESTRING(-1 1,1 -1)", "LINESTRING(0 0,2 0,2 2,0 2,0 0)", true},

		// LineString/Polygon
		{"LINESTRING(3 0,3 1,3 2)", "POLYGON((0 0,2 0,2 2,0 2,0 0))", false},
		{"LINESTRING(1 1,2 1, 3 1)", "POLYGON((0 0,2 0,2 2,0 2,0 0))", true},

		// LineString/MultiPoint
		{"LINESTRING(1 0,2 1,3 0)", "MULTIPOINT((0 0))", false},
		{"LINESTRING(1 0,2 1,3 0)", "MULTIPOINT((1 0))", true},

		// LineString/MultiLineString
		{"LINESTRING(0 0,1 0,0 1,0 0)", "MULTILINESTRING((0 0,0 1,1 1),(0 1,0 0,1 0))", true},
		{"LINESTRING(1 1,2 1,2 2,1 2,1 1)", "MULTILINESTRING((0 0,1 0,1 1,0 1))", true},
		{"LINESTRING(1 2,3 4,5 6)", "MULTILINESTRING((0 1,2 3,4 5))", true},

		// LineString/MultiPolygon
		{"LINESTRING(3 0,3 1,3 2)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)))", false},
		{"LINESTRING(1 1,2 1, 3 1)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)))", true},

		// Polygon/Polygon
		{"POLYGON((0 0,1 0,1 1,0 1,0 0))", "POLYGON((2 0,3 0,3 1,2 1,2 0))", false},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "POLYGON((1 1,3 1,3 3,1 3,1 1))", true},
		{"POLYGON((0 0,4 0,4 4,0 4,0 0))", "POLYGON((1 1,3 1,3 3,1 3,1 1))", true},

		// Polygon/MultiPoint
		{
			`POLYGON(
				(0 0,5 0,5 3,0 3,0 0),
				(1 1,2 1,2 2,1 2,1 1),
				(3 1,4 1,4 2,3 2,3 1)
			)`,
			`MULTIPOINT(1 2,10 10)`,
			true,
		},
		{
			`POLYGON(
				(0 0,5 0,5 3,0 3,0 0),
				(1 1,2 1,2 2,1 2,1 1),
				(3 1,4 1,4 2,3 2,3 1)
			)`,
			`MULTIPOINT(1 2)`,
			true,
		},
		{
			"POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))",
			"MULTIPOINT((2 1),(1 2),(2 1))",
			true,
		},
		{
			"POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))",
			"MULTIPOINT((2 1),(3 6),(2 1))",
			true,
		},

		// Polygon/MultiLineString
		{"POLYGON((0 0,0 2,2 2,2 0,0 0))", "MULTILINESTRING((-1 1,-1 3),(1 -1,1 3))", true},
		{"POLYGON((0 0,0 2,2 2,2 0,0 0))", "MULTILINESTRING((-1 1,-1 3),(3 -1,3 3))", false},

		// Polygon/MultiPolygon
		{"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)),((4 0,7 0,7 3,4 3,4 0),(4.1 0.1,6.9 0.1,6.9 2.9,4.1 2.9,4.1 0.1)))", "POLYGON((8 1,9 1,9 2,8 2,8 1))", false},
		{"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)),((4 0,7 0,7 3,4 3,4 0),(4.1 0.1,6.9 0.1,6.9 2.9,4.1 2.9,4.1 0.1)))", "POLYGON((6 1,7.5 1,7.5 -1,6 -1,6 1))", true},

		// MultiPoint/MultiPoint
		{"MULTIPOINT EMPTY", "MULTIPOINT EMPTY", false},
		{"MULTIPOINT EMPTY", "MULTIPOINT((1 2))", false},
		{"MULTIPOINT((1 2))", "MULTIPOINT((1 2))", true},
		{"MULTIPOINT((1 2))", "MULTIPOINT((1 2),(1 2))", true},
		{"MULTIPOINT((1 2))", "MULTIPOINT((1 2),(3 4))", true},
		{"MULTIPOINT((3 4),(1 2))", "MULTIPOINT((1 2),(3 4))", true},
		{"MULTIPOINT((3 4),(1 2))", "MULTIPOINT((1 4),(2 2))", false},
		{"MULTIPOINT((1 2))", "MULTIPOINT((4 8))", false},
		{"MULTIPOINT((1 2))", "MULTIPOINT((7 6),(3 3),(3 3))", false},

		// MultiPoint/Point
		{"MULTIPOINT EMPTY", "POINT(1 2)", false},
		{"MULTIPOINT((2 1))", "POINT(1 2)", false},
		{"MULTIPOINT((1 2))", "POINT(1 2)", true},
		{"MULTIPOINT((1 2),(1 2))", "POINT(1 2)", true},
		{"MULTIPOINT((1 2),(3 4))", "POINT(1 2)", true},
		{"MULTIPOINT((3 4),(1 2))", "POINT(1 2)", true},
		{"MULTIPOINT((5 6),(7 8))", "POINT(1 2)", false},

		// MultiPoint/MultiLineString
		{"MULTIPOINT(0 0,1 0)", "MULTILINESTRING((0 1,1 1),(1 0,2 -1))", true},
		{"MULTIPOINT(0 0,1 0)", "MULTILINESTRING((0 1,1 1),(1 0.5,2 -0.5))", false},

		// MultiPoint/MultiPolygon
		{"MULTIPOINT((1 1))", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)))", true},
		{"MULTIPOINT((3 3))", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)))", false},

		// MultiLineString/MultiLineString
		{"MULTILINESTRING((0 0,1 0,1 1,0 1))", "MULTILINESTRING((1 1,2 1,2 2,1 2,1 1))", true},
		{"MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))", "MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))", true},

		// MultiLineString/MultiPolygon
		{"MULTILINESTRING((5 2,5 4))", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", false},
		{"MULTILINESTRING((3 3,3 5))", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},
		{"MULTILINESTRING((1 1,3 1))", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},
		{"MULTILINESTRING((0 2,2 4))", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},

		// MultiPolygon/MultiPolygon
		{"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)),((4 0,7 0,7 3,4 3,4 0),(4.1 0.1,6.9 0.1,6.9 2.9,4.1 2.9,4.1 0.1)))", "MULTIPOLYGON(((8 1,9 1,9 2,8 2,8 1)))", false},
		{"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)),((4 0,7 0,7 3,4 3,4 0),(4.1 0.1,6.9 0.1,6.9 2.9,4.1 2.9,4.1 0.1)))", "MULTIPOLYGON(((6 1,7.5 1,7.5 -1,6 -1,6 1)))", true},
		{"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)),((4 0,7 0,7 3,4 3,4 0),(4.1 0.1,6.9 0.1,6.9 2.9,4.1 2.9,4.1 0.1)))", "MULTIPOLYGON(((5 1,6 1,6 2,5 2,5 1)))", false},
		{"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)),((4 0,7 0,7 3,4 3,4 0),(4.1 0.1,6.9 0.1,6.9 2.9,4.1 2.9,4.1 0.1)))", "MULTIPOLYGON(((1 1,1 2,2 2,2 1,1 1)))", true},
		{"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)),((4 0,7 0,7 3,4 3,4 0),(4.1 0.1,6.9 0.1,6.9 2.9,4.1 2.9,4.1 0.1)))", "MULTIPOLYGON(((1 1,1 -1,2 -1,2 1,1 1)))", true},

		// GeometryCollection/OtherTypes
		{"GEOMETRYCOLLECTION(POINT(1 2))", "POINT(1 2)", true},
		{"GEOMETRYCOLLECTION(POINT(1 2))", "POINT(1 3)", false},
		{"GEOMETRYCOLLECTION(POINT(1 2))", "LINESTRING(0 2,2 2)", true},
		{"GEOMETRYCOLLECTION(POINT(1 2))", "LINESTRING(0 3,2 3)", false},
		{"GEOMETRYCOLLECTION(POINT(1 2))", "LINESTRING(0 2,2 2,3 3)", true},
		{"GEOMETRYCOLLECTION(POINT(1 2))", "LINESTRING(0 3,2 3,3 3)", false},
		{"GEOMETRYCOLLECTION(POINT(1 2))", "POLYGON((0.5 1.5,1.5 1.5,1.5 2.5,0.5 2.5, 0.5 1.5))", true},
		{"GEOMETRYCOLLECTION(POINT(5 5))", "POLYGON((0.5 1.5,1.5 1.5,1.5 2.5,0.5 2.5, 0.5 1.5))", false},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			runTest := func(g1, g2 geom.Geometry) func(t *testing.T) {
				return func(t *testing.T) {
					got := g1.Intersects(g2)
					if got != tt.want {
						t.Errorf(
							"\ninput1: %s\ninput2: %s\ngot:  %v\nwant: %v\n",
							g1.AsText(), g2.AsText(), got, tt.want,
						)
					}
				}
			}
			g1 := geomFromWKT(t, tt.in1)
			g2 := geomFromWKT(t, tt.in2)
			t.Run("fwd", runTest(g1, g2))
			t.Run("rev", runTest(g2, g1))
		})
	}
}

func BenchmarkIntersectsLineStringWithLineString(b *testing.B) {
	for _, sz := range []int{10, 100, 1000, 10000} {
		b.Run(fmt.Sprintf("n=%d", sz), func(b *testing.B) {
			xys1 := make([]geom.XY, sz)
			xys2 := make([]geom.XY, sz)
			for i := 0; i < sz; i++ {
				x := float64(i) / float64(sz)
				xys1[i] = geom.XY{X: x, Y: 1}
				xys2[i] = geom.XY{X: x, Y: 2}
			}
			ls1, err := geom.NewLineStringXY(xys1)
			if err != nil {
				b.Fatal(err)
			}
			ls2, err := geom.NewLineStringXY(xys2)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if ls1.Intersects(ls2) {
					b.Fatal("should not intersect")
				}
			}
		})
	}
}
