package geom_test

import (
	"fmt"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestUnmarshalTWKBValid(t *testing.T) {
	for _, tc := range []struct {
		name, twkbHex, wkt string
	}{
		// Several test cases adapted from https://github.com/TWKB/twkb.js/blob/master/test/twkb.spec.js
		{"point lacking data", "0110", "POINT EMPTY"},
		{"point", "01000204", "POINT(1 2)"},
		{"point z", "010801020406", "POINT Z (1 2 3)"},
		{"point m", "010802020408", "POINT M (1 2 4)"},
		{"point zm", "01080302040608", "POINT ZM (1 2 3 4)"},
		{"point with prec -1", "11000204", "POINT(10 20)"},
		{"point with prec 1", "21000204", "POINT(0.1 0.2)"},
		{"point with prec -2", "31000204", "POINT(100 200)"},
		{"point with prec 2", "41000204", "POINT(0.01 0.02)"},
		{"line string lacking data ", "0210", "LINESTRING EMPTY"},
		{"line string no points", "020000", "LINESTRING EMPTY"},
		{"line string", "02000202020808", "LINESTRING(1 1,5 5)"},
		{"line string z", "02080102020202080808", "LINESTRING Z(1 1 1,5 5 5)"},
		{"line string z with prec -1 & prec z 1", "12080902020202080808", "LINESTRING Z(10 10 0.1,50 50 0.5)"},
		{"line string z with prec 1 & prec z -2", "22080d02020202080808", "LINESTRING Z(0.1 0.1 100,0.5 0.5 500)"},
		{"line string m with prec 2 & prec m -3", "4208a202020202080808", "LINESTRING M(0.01 0.01 1000,0.05 0.05 5000)"},
		{"polygon lacking data", "0310", "POLYGON EMPTY"},
		{"polygon no rings", "030000", "POLYGON EMPTY"},
		{"polygon unclosed rings", "030002040000060000060500040203000202000001", "POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1))"},
		{"polygon closed rings", "03000205000006000006050000050502020002020000010100", "POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1))"},
		{"polygon with size & bbox", "0303170006000602040000060000060500040203000202000001", "POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1))"},
		{"multipoint lacking data", "0410", "MULTIPOINT EMPTY"},
		{"multipoint no contents", "040000", "MULTIPOINT EMPTY"},
		{"multipoint with size & bbox & ids", "04070b0004020402000200020404", "MULTIPOINT(0 1,2 3)"},
		{"multilinestring lacking data", "0510", "MULTILINESTRING EMPTY"},
		{"multilinestring no contents", "050000", "MULTILINESTRING EMPTY"},
		{"multilinestring", "050002020000020203020202020202", "MULTILINESTRING((0 0,1 1),(2 2,3 3,4 4))"},
		{"multipolygon lacking data", "0610", "MULTIPOLYGON EMPTY"},
		{"multipolygon no contents", "060000", "MULTIPOLYGON EMPTY"},
		{"multipolygon with polygon lacking data", "06000100", "MULTIPOLYGON(EMPTY)"},
		{"multipolygon with two polygons lacking data", "0600020000", "MULTIPOLYGON(EMPTY,EMPTY)"},
		{"multipolygon with various contents", "0600020001040000060000060500", "MULTIPOLYGON(EMPTY,((0 0,3 0,3 3,0 3,0 0)))"},
		{"multipolygon", "0600020104000006000006050001040802000202000001", "MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)),((4 4,4 5,5 5,5 4,4 4)))"},
		{"geometry collection lacking data", "0710", "GEOMETRYCOLLECTION EMPTY"},
		{"geometry collection no contents", "070000", "GEOMETRYCOLLECTION EMPTY"},
		{"geometry collection with point and empty", "070002010000020310", "GEOMETRYCOLLECTION(POINT(0 1),POLYGON EMPTY)"},
		{"geometry collection", "07000201000002020002080a0404", "GEOMETRYCOLLECTION(POINT(0 1),LINESTRING(4 5,6 7))"},
		{"geometry collection with ids", "070402000201000002020002080a0404", "GEOMETRYCOLLECTION(POINT(0 1),LINESTRING(4 5,6 7))"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			twkb := hexStringToBytes(t, tc.twkbHex)
			t.Logf("TWKB (hex): %v", tc.twkbHex)
			g, err := geom.UnmarshalTWKB(twkb)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			expectGeomEqWKT(t, g, tc.wkt)
		})
	}
}

func TestZigZagInt(t *testing.T) {
	for _, tc := range []struct {
		n int32
		z uint32
	}{
		{0, 0},
		{-1, 1},
		{1, 2},
		{-2, 3},
		{2, 4},
		{-3, 5},
		{3, 6},
		{-4, 7},
		{4, 8},
		{-128, 255},
		{128, 256},
		{-32768, 65535},
		{32768, 65536},
	} {
		t.Run(fmt.Sprintf("%v", tc.n), func(t *testing.T) {
			t.Logf("ZigZag encode int32: %v", tc.n)
			z := geom.EncodeZigZagInt32(tc.n)
			if tc.z != z {
				t.Fatalf("expected: %v, got: %v", tc.z, z)
			}
			t.Logf("ZigZag decode int32: %v", tc.z)
			n := geom.DecodeZigZagInt32(tc.z)
			if tc.n != n {
				t.Fatalf("expected: %v, got: %v", tc.n, n)
			}

			t.Logf("ZigZag encode int64: %v", tc.n)
			z = uint32(geom.EncodeZigZagInt64(int64(tc.n)))
			if tc.z != z {
				t.Fatalf("expected: %v, got: %v", tc.z, z)
			}
			t.Logf("ZigZag decode int64: %v", tc.z)
			n = int32(geom.DecodeZigZagInt64(uint64(tc.z)))
			if tc.n != n {
				t.Fatalf("expected: %v, got: %v", tc.n, n)
			}
		})
	}
}
