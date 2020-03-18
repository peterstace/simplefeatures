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
		// Point/Point
		{"POINT EMPTY", "POINT EMPTY", false},
		{"POINT EMPTY", "POINT(10 10)", false},
		{"POINT(1 2)", "POINT(1 2)", true},
		{"POINT(1 2)", "POINT(2 1)", false},
		{"POINT Z (1 2 3)", "POINT M (1 2 3)", true},

		// Point/Line
		{"POINT EMPTY", "LINESTRING(0 0,1 1)", false},
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
		{"POINT EMPTY", "LINESTRING EMPTY", false},
		{"POINT EMPTY", "LINESTRING(0 0,1 1,2 2)", false},
		{"POINT(1 3)", "LINESTRING EMPTY", false},
		{"POINT(0 0)", "LINESTRING(1 0,2 1,3 0)", false},
		{"POINT(1 0)", "LINESTRING(1 0,2 1,3 0)", true},
		{"POINT(2 1)", "LINESTRING(1 0,2 1,3 0)", true},
		{"POINT(1.5 0.5)", "LINESTRING(1 0,2 1,3 0)", true},
		{"POINT(1 2)", "LINESTRING(0 0,0 4)", false},

		// Point/Polygon
		{"POLYGON EMPTY", "POINT EMPTY", false},
		{"POLYGON EMPTY", "POINT(2 3)", false},
		{"POINT EMPTY", "POLYGON((0 0,1 0,0 1,0 0))", false},
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
		{"POINT EMPTY", "MULTILINESTRING EMPTY", false},
		{"POINT EMPTY", "MULTILINESTRING(EMPTY)", false},
		{"POINT EMPTY", "MULTILINESTRING((0 0,1 1))", false},
		{"POINT(1 1)", "MULTILINESTRING EMPTY", false},
		{"POINT(1 1)", "MULTILINESTRING(EMPTY)", false},
		{"POINT(0 0)", "MULTILINESTRING((1 0,2 1,3 0))", false},
		{"POINT(1 0)", "MULTILINESTRING((1 0,2 1,3 0))", true},
		{"POINT(0 0)", "MULTILINESTRING((0 0,1 1),(1 1,2 1))", true},
		{"POINT(1 1)", "MULTILINESTRING((0 0,1 1),(1 1,2 1))", true},
		{"POINT(2 1)", "MULTILINESTRING((0 0,1 1),(1 1,2 1))", true},
		{"POINT(3 1)", "MULTILINESTRING((0 0,1 1),(1 1,2 1))", false},

		// Point/MultiPolygon
		{"POINT EMPTY", "MULTIPOLYGON EMPTY", false},
		{"POINT EMPTY", "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))", false},
		{"POINT(2 1)", "MULTIPOLYGON EMPTY", false},
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
		{"LINESTRING(0 0,1 1)", "LINESTRING EMPTY", false},
		{"LINESTRING(0 0,1 1)", "LINESTRING(0 0,1 1,0 0)", true},
		{"LINESTRING(0 0,1 1)", "LINESTRING(0 0,1 1,0 1,1 0)", true},

		// Line/Polygon
		{"POLYGON EMPTY", "LINESTRING(0 0,1 1)", false},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(3 0,3 2)", false},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(1 2.1,2.1 1)", true},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(1 -1,1 1)", true},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(0.25 0.25,0.75 0.75)", true},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(2 0,3 -1)", true},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "LINESTRING(-1 1,1 -1)", true},

		// Line/MultiPoint
		{"LINESTRING(0 0,1 1)", "MULTIPOINT EMPTY", false},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(EMPTY)", false},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(1 0)", false},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(1 0,0 1)", false},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(0.5 0.5)", true},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(0 0)", true},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(0.5 0.5,1 0)", true},
		{"LINESTRING(0 0,1 1)", "MULTIPOINT(1 1,0 1)", true},
		{"LINESTRING(1 2,4 5)", "MULTIPOINT((7 6),(3 3),(3 3))", false},
		{"LINESTRING(2 1,3 6)", "MULTIPOINT((1 2))", false},

		// Line/MultiLineString
		{"LINESTRING(0 0,1 1)", "MULTILINESTRING EMPTY", false},
		{"LINESTRING(0 0,1 1)", "MULTILINESTRING(EMPTY)", false},
		{"LINESTRING(0 0,1 1)", "MULTILINESTRING((0 0.5,1 0.5,1 -0.5),(2 0.5,2 -0.5))", true},
		{"LINESTRING(0 1,1 2)", "MULTILINESTRING((0 0.5,1 0.5,1 -0.5),(2 0.5,2 -0.5))", false},

		// Line/MultiPolygon
		{"LINESTRING(0 0,1 1)", "MULTIPOLYGON EMPTY", false},
		{"LINESTRING(5 2,5 4)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", false},
		{"LINESTRING(3 3,3 5)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},
		{"LINESTRING(1 1,3 1)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},
		{"LINESTRING(0 2,2 4)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},

		// LineString/LineString
		{"LINESTRING EMPTY", "LINESTRING EMPTY", false},
		{"LINESTRING EMPTY", "LINESTRING(0 0,1 1,2 2)", false},
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
		{"LINESTRING EMPTY", "POLYGON EMPTY", false},
		{"LINESTRING EMPTY", "POLYGON((0 0,0 1,1 0,0 0))", false},
		{"LINESTRING(0 0,1 1,2 2) ", "POLYGON EMPTY", false},
		{"LINESTRING(3 0,3 1,3 2)", "POLYGON((0 0,2 0,2 2,0 2,0 0))", false},
		{"LINESTRING(1 1,2 1, 3 1)", "POLYGON((0 0,2 0,2 2,0 2,0 0))", true},

		// LineString/MultiPoint
		{"LINESTRING EMPTY", "MULTIPOINT EMPTY", false},
		{"LINESTRING EMPTY", "MULTIPOINT(EMPTY)", false},
		{"LINESTRING(0 0,1 1,2 2)", "MULTIPOINT EMPTY", false},
		{"LINESTRING(0 0,1 1,2 2)", "MULTIPOINT(EMPTY)", false},
		{"LINESTRING EMPTY", "MULTIPOINT(1 1)", false},
		{"LINESTRING(1 0,2 1,3 0)", "MULTIPOINT((0 0))", false},
		{"LINESTRING(1 0,2 1,3 0)", "MULTIPOINT((1 0))", true},

		// LineString/MultiLineString
		{"LINESTRING EMPTY", "MULTILINESTRING EMPTY", false},
		{"LINESTRING EMPTY", "MULTILINESTRING(EMPTY)", false},
		{"LINESTRING EMPTY", "MULTILINESTRING((0 0,1 1,2 2))", false},
		{"LINESTRING(0 0,1 1,2 2)", "MULTILINESTRING EMPTY", false},
		{"LINESTRING(0 0,1 1,2 2)", "MULTILINESTRING(EMPTY)", false},
		{"LINESTRING(0 0,1 0,0 1,0 0)", "MULTILINESTRING((0 0,0 1,1 1),(0 1,0 0,1 0))", true},
		{"LINESTRING(1 1,2 1,2 2,1 2,1 1)", "MULTILINESTRING((0 0,1 0,1 1,0 1))", true},
		{"LINESTRING(1 2,3 4,5 6)", "MULTILINESTRING((0 1,2 3,4 5))", true},

		// LineString/MultiPolygon
		{"LINESTRING EMPTY", "MULTIPOLYGON EMPTY", false},
		{"LINESTRING EMPTY", "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))", false},
		{"LINESTRING(3 0,3 1,3 2)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)))", false},
		{"LINESTRING(1 1,2 1, 3 1)", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)))", true},

		// Polygon/Polygon
		{"POLYGON EMPTY", "POLYGON EMPTY", false},
		{"POLYGON EMPTY", "POLYGON((0 0,1 0,0 1,0 0))", false},
		{"POLYGON((0 0,1 0,1 1,0 1,0 0))", "POLYGON((2 0,3 0,3 1,2 1,2 0))", false},
		{"POLYGON((0 0,2 0,2 2,0 2,0 0))", "POLYGON((1 1,3 1,3 3,1 3,1 1))", true},
		{"POLYGON((0 0,4 0,4 4,0 4,0 0))", "POLYGON((1 1,3 1,3 3,1 3,1 1))", true},

		// Polygon/MultiPoint
		{"POLYGON EMPTY", "MULTIPOINT EMPTY", false},
		{"POLYGON EMPTY", "MULTIPOINT(EMPTY)", false},
		{"POLYGON EMPTY", "MULTIPOINT(1 1)", false},
		{"POLYGON((0 0,0 1,1 0,0 0))", "MULTIPOINT EMPTY", false},
		{"POLYGON((0 0,0 1,1 0,0 0))", "MULTIPOINT(EMPTY)", false},
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
		{"POLYGON EMPTY", "MULTILINESTRING EMPTY", false},
		{"POLYGON EMPTY", "MULTILINESTRING(EMPTY)", false},
		{"POLYGON EMPTY", "MULTILINESTRING((0 0,1 1,2 2))", false},
		{"POLYGON((0 0,0 1,1 0,0 0))", "MULTILINESTRING EMPTY", false},
		{"POLYGON((0 0,0 1,1 0,0 0))", "MULTILINESTRING(EMPTY)", false},
		{"POLYGON((0 0,0 2,2 2,2 0,0 0))", "MULTILINESTRING((-1 1,-1 3),(1 -1,1 3))", true},
		{"POLYGON((0 0,0 2,2 2,2 0,0 0))", "MULTILINESTRING((-1 1,-1 3),(3 -1,3 3))", false},
		{"POLYGON((0 0,0 2,2 2,2 0,0 0))", "MULTILINESTRING((0.5 0.5,1.5 1.5),(2.5 2.5,3.5 3.5))", true},
		{"POLYGON((0 0,0 2,2 2,2 0,0 0))", "MULTILINESTRING((2.5 2.5,3.5 3.5),(0.5 0.5,1.5 1.5))", true},

		// Polygon/MultiPolygon
		{"MULTIPOLYGON EMPTY", "POLYGON EMPTY", false},
		{"MULTIPOLYGON(EMPTY)", "POLYGON EMPTY", false},
		{"MULTIPOLYGON EMPTY", "POLYGON((0 0,1 0,0 1,0 0))", false},
		{"MULTIPOLYGON(EMPTY)", "POLYGON((0 0,1 0,0 1,0 0))", false},
		{"MULTIPOLYGON(((0 0,0 1,1 0,0 0)))", "POLYGON EMPTY", false},
		{"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)),((4 0,7 0,7 3,4 3,4 0),(4.1 0.1,6.9 0.1,6.9 2.9,4.1 2.9,4.1 0.1)))", "POLYGON((8 1,9 1,9 2,8 2,8 1))", false},
		{"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)),((4 0,7 0,7 3,4 3,4 0),(4.1 0.1,6.9 0.1,6.9 2.9,4.1 2.9,4.1 0.1)))", "POLYGON((6 1,7.5 1,7.5 -1,6 -1,6 1))", true},

		// MultiPoint/MultiPoint
		{"MULTIPOINT EMPTY", "MULTIPOINT EMPTY", false},
		{"MULTIPOINT EMPTY", "MULTIPOINT(EMPTY)", false},
		{"MULTIPOINT(EMPTY)", "MULTIPOINT EMPTY", false},
		{"MULTIPOINT EMPTY", "MULTIPOINT((1 2))", false},
		{"MULTIPOINT(EMPTY)", "MULTIPOINT((1 2))", false},
		{"MULTIPOINT((1 2))", "MULTIPOINT((1 2))", true},
		{"MULTIPOINT((1 2))", "MULTIPOINT((1 2),(1 2))", true},
		{"MULTIPOINT((1 2))", "MULTIPOINT((1 2),(3 4))", true},
		{"MULTIPOINT((3 4),(1 2))", "MULTIPOINT((1 2),(3 4))", true},
		{"MULTIPOINT((3 4),(1 2))", "MULTIPOINT((1 4),(2 2))", false},
		{"MULTIPOINT((1 2))", "MULTIPOINT((4 8))", false},
		{"MULTIPOINT((1 2))", "MULTIPOINT((7 6),(3 3),(3 3))", false},
		{"MULTIPOINT((10 40),(40 30),EMPTY)", "MULTIPOINT((1 2),(2 3),EMPTY)", false},

		// MultiPoint/Point
		{"MULTIPOINT EMPTY", "POINT EMPTY", false},
		{"MULTIPOINT(EMPTY)", "POINT EMPTY", false},
		{"MULTIPOINT EMPTY", "POINT(1 2)", false},
		{"MULTIPOINT(EMPTY)", "POINT(1 2)", false},
		{"MULTIPOINT((2 1))", "POINT EMPTY", false},
		{"MULTIPOINT((2 1))", "POINT(1 2)", false},
		{"MULTIPOINT((1 2))", "POINT(1 2)", true},
		{"MULTIPOINT((1 2),(1 2))", "POINT(1 2)", true},
		{"MULTIPOINT((1 2),(3 4))", "POINT(1 2)", true},
		{"MULTIPOINT((3 4),(1 2))", "POINT(1 2)", true},
		{"MULTIPOINT((5 6),(7 8))", "POINT(1 2)", false},

		// MultiPoint/MultiLineString
		{"MULTIPOINT EMPTY", "MULTILINESTRING EMPTY", false},
		{"MULTIPOINT(EMPTY)", "MULTILINESTRING EMPTY", false},
		{"MULTIPOINT EMPTY", "MULTILINESTRING(EMPTY)", false},
		{"MULTIPOINT(EMPTY)", "MULTILINESTRING(EMPTY)", false},
		{"MULTIPOINT EMPTY", "MULTILINESTRING((0 0,1 1,2 2))", false},
		{"MULTIPOINT(EMPTY)", "MULTILINESTRING((0 0,1 1,2 2))", false},
		{"MULTIPOINT((1 2))", "MULTILINESTRING EMPTY", false},
		{"MULTIPOINT((1 2))", "MULTILINESTRING(EMPTY)", false},
		{"MULTIPOINT(0 0,1 0)", "MULTILINESTRING((0 1,1 1),(1 0,2 -1))", true},
		{"MULTIPOINT(0 0,1 0)", "MULTILINESTRING((0 1,1 1),(1 0.5,2 -0.5))", false},
		{"MULTIPOINT(0.5 0.5)", "MULTILINESTRING((0 0,0 0,1 1))", true},

		// MultiPoint/MultiPolygon
		{"MULTIPOINT EMPTY", "MULTIPOLYGON EMPTY", false},
		{"MULTIPOINT(EMPTY)", "MULTIPOLYGON EMPTY", false},
		{"MULTIPOINT EMPTY", "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))", false},
		{"MULTIPOINT(EMPTY)", "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))", false},
		{"MULTIPOINT((1 1))", "MULTIPOLYGON EMPTY", false},
		{"MULTIPOINT((1 1))", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)))", true},
		{"MULTIPOINT((3 3))", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)))", false},
		{"MULTIPOINT((1 2),(2 3),EMPTY)", "MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((2 0,4 0,4 2,2 2,2 0)))", false},

		// MultiLineString/MultiLineString
		{"MULTILINESTRING EMPTY", "MULTILINESTRING EMPTY", false},
		{"MULTILINESTRING EMPTY", "MULTILINESTRING(EMPTY)", false},
		{"MULTILINESTRING EMPTY", "MULTILINESTRING((0 0,1 1,2 2))", false},
		{"MULTILINESTRING(EMPTY)", "MULTILINESTRING((0 0,1 1,2 2))", false},
		{"MULTILINESTRING((0 0,1 0,1 1,0 1))", "MULTILINESTRING((1 1,2 1,2 2,1 2,1 1))", true},
		{"MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))", "MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))", true},

		// MultiLineString/MultiPolygon
		{"MULTILINESTRING EMPTY", "MULTIPOLYGON EMPTY", false},
		{"MULTILINESTRING(EMPTY)", "MULTIPOLYGON EMPTY", false},
		{"MULTILINESTRING EMPTY", "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))", false},
		{"MULTILINESTRING(EMPTY)", "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))", false},
		{"MULTILINESTRING((5 2,5 4))", "MULTIPOLYGON EMPTY", false},
		{"MULTILINESTRING((5 2,5 4))", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", false},
		{"MULTILINESTRING((3 3,3 5))", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},
		{"MULTILINESTRING((1 1,3 1))", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},
		{"MULTILINESTRING((0 2,2 4))", "MULTIPOLYGON(((0 0,2 0,2 2,0 2,0 0)),((2 2,2 4,4 4,4 2,2 2)))", true},
		{"MULTILINESTRING((0.5 0.5,1.5 1.5),(2.5 2.5,3.5 3.5))", "MULTIPOLYGON(((0 0,0 2,2 2,2 0,0 0)))", true},
		{"MULTILINESTRING((2.5 2.5,3.5 3.5),(0.5 0.5,1.5 1.5))", "MULTIPOLYGON(((0 0,0 2,2 2,2 0,0 0)))", true},

		// MultiPolygon/MultiPolygon
		{"MULTIPOLYGON EMPTY", "MULTIPOLYGON EMPTY", false},
		{"MULTIPOLYGON EMPTY", "MULTIPOLYGON(((0 0,0 1,1 0,0 0)))", false},
		{"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)),((4 0,7 0,7 3,4 3,4 0),(4.1 0.1,6.9 0.1,6.9 2.9,4.1 2.9,4.1 0.1)))", "MULTIPOLYGON(((8 1,9 1,9 2,8 2,8 1)))", false},
		{"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)),((4 0,7 0,7 3,4 3,4 0),(4.1 0.1,6.9 0.1,6.9 2.9,4.1 2.9,4.1 0.1)))", "MULTIPOLYGON(((6 1,7.5 1,7.5 -1,6 -1,6 1)))", true},
		{"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)),((4 0,7 0,7 3,4 3,4 0),(4.1 0.1,6.9 0.1,6.9 2.9,4.1 2.9,4.1 0.1)))", "MULTIPOLYGON(((5 1,6 1,6 2,5 2,5 1)))", false},
		{"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)),((4 0,7 0,7 3,4 3,4 0),(4.1 0.1,6.9 0.1,6.9 2.9,4.1 2.9,4.1 0.1)))", "MULTIPOLYGON(((1 1,1 2,2 2,2 1,1 1)))", true},
		{"MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)),((4 0,7 0,7 3,4 3,4 0),(4.1 0.1,6.9 0.1,6.9 2.9,4.1 2.9,4.1 0.1)))", "MULTIPOLYGON(((1 1,1 -1,2 -1,2 1,1 1)))", true},

		// GeometryCollection/OtherTypes
		{"GEOMETRYCOLLECTION EMPTY", "POINT EMPTY", false},
		{"GEOMETRYCOLLECTION EMPTY", "POINT(1 2)", false},
		{"GEOMETRYCOLLECTION(POINT EMPTY)", "POINT EMPTY", false},
		{"GEOMETRYCOLLECTION(POINT EMPTY)", "POINT(1 2)", false},
		{"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION EMPTY)", "POINT EMPTY", false},
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
			var floats1, floats2 []float64
			for i := 0; i < sz; i++ {
				x := float64(i) / float64(sz)
				floats1 = append(floats1, x, 1)
				floats2 = append(floats2, x, 2)
			}
			seq1 := geom.NewSequence(floats1, geom.DimXY)
			seq2 := geom.NewSequence(floats2, geom.DimXY)
			ls1, err := geom.NewLineString(seq1)
			if err != nil {
				b.Fatal(err)
			}
			ls2, err := geom.NewLineString(seq2)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			ls2g := ls2.AsGeometry()

			for i := 0; i < b.N; i++ {
				if ls1.Intersects(ls2g) {
					b.Fatal("should not intersect")
				}
			}
		})
	}
}
