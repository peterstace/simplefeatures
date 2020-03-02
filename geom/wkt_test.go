package geom_test

import (
	"strconv"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestUnmarshalWKTValidGrammar(t *testing.T) {
	for _, tt := range []struct {
		name, wkt string
	}{
		{"empty point", "POINT EMPTY"},
		{"mixed case", "PoInT (1 1)"},
		{"upper case", "POINT (1 1)"},
		{"lower case", "point (1 1)"},
		{"no space between tag and coord", "point(1 1)"},
		{"exponent", "point (1e3 1.5e2)"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := UnmarshalWKT(strings.NewReader(tt.wkt))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestUnmarshalWKTInvalidGrammar(t *testing.T) {
	for _, tt := range []struct {
		name, wkt string
	}{
		{"NaN coord", "point(NaN NaN)"},
		{"+Inf coord", "point(+Inf +Inf)"},
		{"-Inf coord", "point(-Inf -Inf)"},
		{"left unbalanced point", "point ( 1 2"},
		{"right unbalanced point", "point 1 2 )"},
		{"point no parens", "point 1 1"},
		{"point no coords", "point ( )"},

		{"mixed empty", "LINESTRING(0 0, EMPTY, 2 2)"},
		{"foo internal point", "LINESTRING(0 0, foo, 2 2)"},
		{"line string no coords", "LINESTRING()"},

		{"polygon no coords", "POLYGON()"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := UnmarshalWKT(strings.NewReader(tt.wkt))
			if err == nil {
				t.Fatalf("expected error but got nil")
			} else {
				t.Logf("got error: %v", err)
			}
		})
	}
}

func TestMarshalUnmarshalWKT(t *testing.T) {
	for i, wkt := range []string{
		"POINT(30 10)",
		"POINT(-30 -10)",
		"POINT EMPTY",

		"POINT Z (30 10 20)",
		"POINT Z (-30 -10 -20)",
		"POINT Z EMPTY",

		"POINT M (30 10 20)",
		"POINT M (-30 -10 -20)",
		"POINT M EMPTY",

		"POINT ZM (30 10 20 40)",
		"POINT ZM (-30 -10 -20 -40)",
		"POINT ZM EMPTY",

		"LINESTRING(30 10,10 30)",
		"LINESTRING(30 10,10 30,40 40)",
		"LINESTRING EMPTY",

		"LINESTRING Z (30 10 20,10 30 50)",
		"LINESTRING Z (30 10 20,10 30 50,40 40 70)",
		"LINESTRING Z EMPTY",

		"LINESTRING M (30 10 20,10 30 50)",
		"LINESTRING M (30 10 20,10 30 50,40 40 70)",
		"LINESTRING M EMPTY",

		"LINESTRING ZM (30 10 20 10,10 30 50 40)",
		"LINESTRING ZM (30 10 20 40,10 30 50 85,40 40 70 32)",
		"LINESTRING ZM EMPTY",

		"POLYGON((30 10,40 40,20 40,10 20,30 10))",
		"POLYGON((35 10,45 45,15 40,10 20,35 10),(20 30,35 35,30 20,20 30))",
		"POLYGON EMPTY",

		"POLYGON Z ((30 10 1,40 40 2,20 40 3,10 20 4,30 10 5))",
		"POLYGON Z ((35 10 6,45 45 7,15 40 8,10 20 9,35 10 8),(20 30 7,35 35 6,30 20 5,20 30 4))",
		"POLYGON Z EMPTY",

		"POLYGON M ((30 10 1,40 40 2,20 40 3,10 20 4,30 10 5))",
		"POLYGON M ((35 10 6,45 45 7,15 40 8,10 20 9,35 10 8),(20 30 7,35 35 6,30 20 5,20 30 4))",
		"POLYGON M EMPTY",

		"POLYGON ZM ((30 10 1 2,40 40 2 3,20 40 3 4,10 20 4 5,30 10 5 6))",
		"POLYGON ZM ((35 10 6 7,45 45 7 8,15 40 8 9,10 20 9 8,35 10 8 7),(20 30 7 6,35 35 6 5,30 20 5 4,20 30 4 3))",
		"POLYGON ZM EMPTY",

		"MULTIPOINT((10 40),(40 30),(20 20),(30 10))",
		"MULTIPOINT((10 40),(40 30),EMPTY)",
		"MULTIPOINT EMPTY",
		"MULTIPOINT(EMPTY)",

		"MULTIPOINT Z ((10 40 1),(40 30 2),(20 20 3),(30 10 4))",
		"MULTIPOINT Z ((10 40 5),(40 30 6),EMPTY)",
		"MULTIPOINT Z EMPTY",
		"MULTIPOINT Z (EMPTY)",

		"MULTIPOINT M ((10 40 1),(40 30 2),(20 20 3),(30 10 4))",
		"MULTIPOINT M ((10 40 5),(40 30 6),EMPTY)",
		"MULTIPOINT M EMPTY",
		"MULTIPOINT M (EMPTY)",

		"MULTIPOINT ZM ((10 40 1 2),(40 30 2 3),(20 20 3 4),(30 10 4 5))",
		"MULTIPOINT ZM ((10 40 5 6),(40 30 6 7),EMPTY)",
		"MULTIPOINT ZM EMPTY",
		"MULTIPOINT ZM (EMPTY)",

		"MULTILINESTRING((10 10,20 20,10 40),(40 40,30 30,40 20,30 10))",
		"MULTILINESTRING((1 2,3 4,5 6),EMPTY)",
		"MULTILINESTRING EMPTY",
		"MULTILINESTRING(EMPTY)",

		"MULTILINESTRING Z ((10 10 1,20 20 2,10 40 3),(40 40 4,30 30 5,40 20 6,30 10 7))",
		"MULTILINESTRING Z ((1 2 8,3 4 9,5 6 10),EMPTY)",
		"MULTILINESTRING Z EMPTY",
		"MULTILINESTRING Z (EMPTY)",

		"MULTILINESTRING M ((10 10 1,20 20 2,10 40 3),(40 40 4,30 30 5,40 20 6,30 10 7))",
		"MULTILINESTRING M ((1 2 8,3 4 9,5 6 10),EMPTY)",
		"MULTILINESTRING M EMPTY",
		"MULTILINESTRING M (EMPTY)",

		"MULTILINESTRING ZM ((10 10 1 2,20 20 2 3,10 40 3 4),(40 40 4 5,30 30 5 6,40 20 6 7,30 10 7 8))",
		"MULTILINESTRING ZM ((1 2 8 9,3 4 9 8,5 6 10 11),EMPTY)",
		"MULTILINESTRING ZM EMPTY",
		"MULTILINESTRING ZM (EMPTY)",

		"MULTIPOLYGON EMPTY",
		"MULTIPOLYGON(((30 20,45 40,10 40,30 20)),((15 5,40 10,10 20,5 10,15 5)))",
		"MULTIPOLYGON(((40 40,20 45,45 30,40 40)),((20 35,10 30,10 10,30 5,45 20,20 35),(30 20,20 15,20 25,30 20)))",
		"MULTIPOLYGON(EMPTY,((20 35,10 30,10 10,30 5,45 20,20 35),(30 20,20 15,20 25,30 20)))",
		`MULTIPOLYGON(EMPTY)`,

		"MULTIPOLYGON Z EMPTY",
		"MULTIPOLYGON Z (((30 20 1,45 40 2,10 40 3,30 20 4)),((15 5 5,40 10 6,10 20 7,5 10 8,15 5 9)))",
		"MULTIPOLYGON Z (((40 40 9,20 45 8,45 30 7,40 40 6)),((20 35 5,10 30 4,10 10 3,30 5 2,45 20 1,20 35 2),(30 20 3,20 15 4,20 25 5,30 20 6)))",
		"MULTIPOLYGON Z (EMPTY,((20 35 5,10 30 6,10 10 7,30 5 8,45 20 9,20 35 10),(30 20 11,20 15 12,20 25 13,30 20 14)))",
		`MULTIPOLYGON Z (EMPTY)`,

		"MULTIPOLYGON M EMPTY",
		"MULTIPOLYGON M (((30 20 1,45 40 2,10 40 3,30 20 4)),((15 5 5,40 10 6,10 20 7,5 10 8,15 5 9)))",
		"MULTIPOLYGON M (((40 40 9,20 45 8,45 30 7,40 40 6)),((20 35 5,10 30 4,10 10 3,30 5 2,45 20 1,20 35 2),(30 20 3,20 15 4,20 25 5,30 20 6)))",
		"MULTIPOLYGON M (EMPTY,((20 35 5,10 30 6,10 10 7,30 5 8,45 20 9,20 35 10),(30 20 11,20 15 12,20 25 13,30 20 14)))",
		`MULTIPOLYGON M (EMPTY)`,

		"MULTIPOLYGON ZM EMPTY",
		"MULTIPOLYGON ZM (((30 20 1 2,45 40 2 3,10 40 3 4,30 20 4 5)),((15 5 5 6,40 10 6 7,10 20 7 8,5 10 8 9,15 5 9 10)))",
		"MULTIPOLYGON ZM (((40 40 9 8,20 45 8 7,45 30 7 6,40 40 6 5)),((20 35 5 4,10 30 4 3,10 10 3 2,30 5 2 1,45 20 1 0,20 35 2 -1),(30 20 3 -2,20 15 4 -3,20 25 5 -4,30 20 6 -5)))",
		"MULTIPOLYGON ZM (EMPTY,((20 35 5 0,10 30 6 1,10 10 7 2,30 5 8 3,45 20 9 4,20 35 10 5),(30 20 11 6,20 15 12 7,20 25 13 8,30 20 14 9)))",
		`MULTIPOLYGON ZM (EMPTY)`,

		"GEOMETRYCOLLECTION EMPTY",
		"GEOMETRYCOLLECTION(GEOMETRYCOLLECTION EMPTY)",
		"GEOMETRYCOLLECTION(POINT EMPTY)",
		"GEOMETRYCOLLECTION(LINESTRING EMPTY)",
		"GEOMETRYCOLLECTION(POLYGON EMPTY)",
		"GEOMETRYCOLLECTION(MULTIPOINT EMPTY)",
		"GEOMETRYCOLLECTION(MULTILINESTRING EMPTY)",
		"GEOMETRYCOLLECTION(MULTIPOLYGON EMPTY)",
		"GEOMETRYCOLLECTION(LINESTRING(0 0,1 1))",
		"GEOMETRYCOLLECTION(POINT(4 6),LINESTRING(4 6,7 10))",

		"GEOMETRYCOLLECTION Z EMPTY",
		"GEOMETRYCOLLECTION Z (GEOMETRYCOLLECTION Z EMPTY)",
		"GEOMETRYCOLLECTION Z (POINT Z EMPTY)",
		"GEOMETRYCOLLECTION Z (LINESTRING Z EMPTY)",
		"GEOMETRYCOLLECTION Z (POLYGON Z EMPTY)",
		"GEOMETRYCOLLECTION Z (MULTIPOINT Z EMPTY)",
		"GEOMETRYCOLLECTION Z (MULTILINESTRING Z EMPTY)",
		"GEOMETRYCOLLECTION Z (MULTIPOLYGON Z EMPTY)",
		"GEOMETRYCOLLECTION Z (LINESTRING Z (0 0 3,1 1 4))",
		"GEOMETRYCOLLECTION Z (POINT Z (4 6 1),LINESTRING Z (4 6 5,7 10 11))",

		"GEOMETRYCOLLECTION M EMPTY",
		"GEOMETRYCOLLECTION M (GEOMETRYCOLLECTION M EMPTY)",
		"GEOMETRYCOLLECTION M (POINT M EMPTY)",
		"GEOMETRYCOLLECTION M (LINESTRING M EMPTY)",
		"GEOMETRYCOLLECTION M (POLYGON M EMPTY)",
		"GEOMETRYCOLLECTION M (MULTIPOINT M EMPTY)",
		"GEOMETRYCOLLECTION M (MULTILINESTRING M EMPTY)",
		"GEOMETRYCOLLECTION M (MULTIPOLYGON M EMPTY)",
		"GEOMETRYCOLLECTION M (LINESTRING M (0 0 3,1 1 4))",
		"GEOMETRYCOLLECTION M (POINT M (4 6 1),LINESTRING M (4 6 5,7 10 11))",

		"GEOMETRYCOLLECTION ZM EMPTY",
		"GEOMETRYCOLLECTION ZM (GEOMETRYCOLLECTION ZM EMPTY)",
		"GEOMETRYCOLLECTION ZM (POINT ZM EMPTY)",
		"GEOMETRYCOLLECTION ZM (LINESTRING ZM EMPTY)",
		"GEOMETRYCOLLECTION ZM (POLYGON ZM EMPTY)",
		"GEOMETRYCOLLECTION ZM (MULTIPOINT ZM EMPTY)",
		"GEOMETRYCOLLECTION ZM (MULTILINESTRING ZM EMPTY)",
		"GEOMETRYCOLLECTION ZM (MULTIPOLYGON ZM EMPTY)",
		"GEOMETRYCOLLECTION ZM (LINESTRING ZM (0 0 3 1,1 1 4 2))",
		"GEOMETRYCOLLECTION ZM (POINT ZM (4 6 1 8),LINESTRING ZM (4 6 5 7,7 10 11 0))",
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := geomFromWKT(t, wkt).AsText()
			expectStringEq(t, got, wkt)
		})
	}
}

func TestUnmarshalWKT(t *testing.T) {
	t.Run("multi line string containing an empty line string", func(t *testing.T) {
		g := geomFromWKT(t, "MULTILINESTRING((1 2,3 4),EMPTY,(5 6,7 8))")
		mls := g.AsMultiLineString()
		expectIntEq(t, mls.NumLineStrings(), 3)
		expectGeomEq(t,
			mls.LineStringN(0).AsGeometry(),
			geomFromWKT(t, "LINESTRING(1 2,3 4)"),
		)
		expectGeomEq(t,
			mls.LineStringN(1).AsGeometry(),
			geomFromWKT(t, "LINESTRING EMPTY"),
		)
		expectGeomEq(t,
			mls.LineStringN(2).AsGeometry(),
			geomFromWKT(t, "LINESTRING(5 6,7 8)"),
		)
	})
	t.Run("multipoints with and without parenthesised points", func(t *testing.T) {
		g1 := geomFromWKT(t, "MULTIPOINT((10 40),(40 30),(20 20),(30 10))")
		g2 := geomFromWKT(t, "MULTIPOINT(10 40,40 30,20 20,30 10)")
		expectGeomEq(t, g1, g2)
	})
}

func TestAsTextEmpty(t *testing.T) {
	for i, tt := range []struct {
		want string
		g    Geometry
	}{
		{"POINT EMPTY", NewEmptyPoint(XYOnly).AsGeometry()},
		{"LINESTRING EMPTY", NewEmptyLineString(XYOnly).AsGeometry()},
		{"POLYGON EMPTY", NewEmptyPolygon(XYOnly).AsGeometry()},
		{"MULTIPOINT EMPTY", NewEmptyMultiPoint(XYOnly).AsGeometry()},
		{"MULTILINESTRING EMPTY", NewEmptyMultiLineString(XYOnly).AsGeometry()},
		{"MULTIPOLYGON EMPTY", NewEmptyMultiPolygon(XYOnly).AsGeometry()},
		{"GEOMETRYCOLLECTION EMPTY", NewEmptyGeometryCollection(XYOnly).AsGeometry()},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got := tt.g.AsText()
			expectStringEq(t, got, tt.want)
		})
	}
}
