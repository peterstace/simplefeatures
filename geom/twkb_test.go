package geom_test

import (
	"fmt"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
)

func TestUnmarshalTWKBValid(t *testing.T) {
	for _, tc := range []struct {
		description, twkbHex, wkt string
		hasZ, hasM                bool
		precXY, precZ, precM      int
		hasSize                   bool
		hasBBox                   bool
		listedBBox                []int64
		hasIDList                 bool
		listedIDs                 []int64
		closeRings                bool
		skipDecode, skipEncode    bool
	}{
		// Several test cases adapted from https://github.com/TWKB/twkb.js/blob/master/test/twkb.spec.js
		{
			description: "point lacking data",
			twkbHex:     "0110",
			wkt:         "POINT EMPTY",
		},
		{
			description: "point",
			twkbHex:     "01000204",
			wkt:         "POINT(1 2)",
		},
		{
			description: "point z",
			twkbHex:     "010801020406",
			wkt:         "POINT Z (1 2 3)",
			hasZ:        true,
		},
		{
			description: "point m",
			twkbHex:     "010802020408",
			wkt:         "POINT M (1 2 4)",
			hasM:        true,
		},
		{
			description: "point zm",
			twkbHex:     "01080302040608",
			wkt:         "POINT ZM (1 2 3 4)",
			hasZ:        true,
			hasM:        true,
		},
		{
			description: "point with prec -1",
			twkbHex:     "11000204",
			wkt:         "POINT(10 20)",
			precXY:      -1,
		},
		{
			description: "point with prec 1",
			twkbHex:     "21000204",
			wkt:         "POINT(0.1 0.2)",
			precXY:      1,
		},
		{
			description: "point with prec -2",
			twkbHex:     "31000204",
			wkt:         "POINT(100 200)",
			precXY:      -2,
		},
		{
			description: "point with prec 2",
			twkbHex:     "41000204",
			wkt:         "POINT(0.01 0.02)",
			precXY:      2,
		},
		{
			description: "line string lacking data ",
			twkbHex:     "0210",
			wkt:         "LINESTRING EMPTY",
		},
		{
			description: "line string no points",
			twkbHex:     "020000",
			wkt:         "LINESTRING EMPTY",
			skipEncode:  true,
		},
		{
			description: "line string",
			twkbHex:     "02000202020808",
			wkt:         "LINESTRING(1 1,5 5)",
		},
		{
			description: "line string z",
			twkbHex:     "02080102020202080808",
			wkt:         "LINESTRING Z(1 1 1,5 5 5)",
			hasZ:        true,
		},
		{
			description: "line string z with prec -1 & prec z 1",
			twkbHex:     "12080902020202080808",
			wkt:         "LINESTRING Z(10 10 0.1,50 50 0.5)",
			hasZ:        true,
			precXY:      -1,
			precZ:       1,
		},
		{
			description: "line string z with prec 1 & prec z -2",
			twkbHex:     "22080d02020202080808",
			wkt:         "LINESTRING Z(0.1 0.1 100,0.5 0.5 500)",
			hasZ:        true,
			precXY:      1,
			precZ:       -2,
		},
		{
			description: "line string m with prec 2 & prec m -3",
			twkbHex:     "4208a202020202080808",
			wkt:         "LINESTRING M(0.01 0.01 1000,0.05 0.05 5000)",
			hasM:        true,
			precXY:      2,
			precM:       -3,
		},
		{
			description: "polygon lacking data",
			twkbHex:     "0310",
			wkt:         "POLYGON EMPTY",
		},
		{
			description: "polygon no rings",
			twkbHex:     "030000",
			wkt:         "POLYGON EMPTY",
			skipEncode:  true,
		},
		{
			description: "polygon unclosed rings",
			twkbHex:     "030002040000060000060500040203000202000001",
			wkt:         "POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1))",
		},
		{
			description: "polygon closed rings",
			twkbHex:     "03000205000006000006050000050502020002020000010100",
			wkt:         "POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1))",
			closeRings:  true,
		},
		{
			description: "polygon with size & bbox",
			twkbHex:     "0303170006000602040000060000060500040203000202000001",
			wkt:         "POLYGON((0 0,3 0,3 3,0 3,0 0),(1 1,1 2,2 2,2 1,1 1))",
			hasSize:     true,
			hasBBox:     true,
			listedBBox:  []int64{0, +3, 0, +3},
		},
		{
			description: "multipoint lacking data",
			twkbHex:     "0410",
			wkt:         "MULTIPOINT EMPTY",
		},
		{
			description: "multipoint no contents",
			twkbHex:     "040000",
			wkt:         "MULTIPOINT EMPTY",
			skipEncode:  true,
		},
		{
			description: "multipoint with size & bbox & ids",
			twkbHex:     "04070b0004020402000200020404",
			wkt:         "MULTIPOINT(0 1,2 3)",
			hasSize:     true,
			hasBBox:     true,
			listedBBox:  []int64{0, +2, 1, +2},
			hasIDList:   true,
			listedIDs:   []int64{0, 1},
		},
		{
			description: "multilinestring lacking data",
			twkbHex:     "0510",
			wkt:         "MULTILINESTRING EMPTY",
		},
		{
			description: "multilinestring no contents",
			twkbHex:     "050000",
			wkt:         "MULTILINESTRING EMPTY",
			skipEncode:  true,
		},
		{
			description: "multilinestring",
			twkbHex:     "050002020000020203020202020202",
			wkt:         "MULTILINESTRING((0 0,1 1),(2 2,3 3,4 4))",
		},
		{
			description: "multipolygon lacking data",
			twkbHex:     "0610",
			wkt:         "MULTIPOLYGON EMPTY",
		},
		{
			description: "multipolygon no contents",
			twkbHex:     "060000",
			wkt:         "MULTIPOLYGON EMPTY",
			skipEncode:  true,
		},
		{
			description: "multipolygon with polygon lacking data",
			twkbHex:     "06000100",
			wkt:         "MULTIPOLYGON(EMPTY)",
			skipEncode:  true,
		},
		{
			description: "multipolygon with two polygons lacking data",
			twkbHex:     "0600020000",
			wkt:         "MULTIPOLYGON(EMPTY,EMPTY)",
			skipEncode:  true,
		},
		{
			description: "multipolygon with various contents",
			twkbHex:     "0600020001040000060000060500",
			wkt:         "MULTIPOLYGON(EMPTY,((0 0,3 0,3 3,0 3,0 0)))",
		},
		{
			description: "multipolygon",
			twkbHex:     "0600020104000006000006050001040802000202000001",
			wkt:         "MULTIPOLYGON(((0 0,3 0,3 3,0 3,0 0)),((4 4,4 5,5 5,5 4,4 4)))",
		},
		{
			description: "geometry collection lacking data",
			twkbHex:     "0710",
			wkt:         "GEOMETRYCOLLECTION EMPTY",
		},
		{
			description: "geometry collection no contents",
			twkbHex:     "070000",
			wkt:         "GEOMETRYCOLLECTION EMPTY",
			skipEncode:  true,
		},
		{
			description: "geometry collection with point and empty",
			twkbHex:     "070002010000020310",
			wkt:         "GEOMETRYCOLLECTION(POINT(0 1),POLYGON EMPTY)",
		},
		{
			description: "geometry collection",
			twkbHex:     "07000201000002020002080a0404",
			wkt:         "GEOMETRYCOLLECTION(POINT(0 1),LINESTRING(4 5,6 7))",
		},
		{
			description: "geometry collection with ids",
			twkbHex:     "070402000201000002020002080a0404",
			wkt:         "GEOMETRYCOLLECTION(POINT(0 1),LINESTRING(4 5,6 7))",
			hasIDList:   true,
			listedIDs:   []int64{0, 1},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
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
