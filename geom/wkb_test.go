package geom_test

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func hexStringToBytes(t *testing.T, s string) []byte {
	t.Helper()
	s = strings.ReplaceAll(s, " ", "")
	if len(s)%2 != 0 {
		t.Fatal("hex string must have even length")
	}
	var buf []byte
	for i := 0; i < len(s); i += 2 {
		x, err := strconv.ParseUint(s[i:i+2], 16, 8)
		if err != nil {
			t.Fatal(err)
		}
		buf = append(buf, byte(x))
	}
	return buf
}

func TestWKBParseValid(t *testing.T) {
	// Test cases generated from:
	/*
		SELECT
			ST_AsText(ST_GeomFromText(wkt)),
			ST_AsBinary(ST_GeomFromText(wkt)) AS wkb
		FROM (
			VALUES
			('POINT EMPTY'),
			('POINTZ EMPTY'),
			('POINTM EMPTY'),
			('POINTZM EMPTY'),
			('POINT(1 2)'),
			('POINTZ(1 2 3)'),
			('POINTM(1 2 3)'),
			('POINTZM(1 2 3 4)'),
			('LINESTRING EMPTY'),
			('LINESTRINGZ EMPTY'),
			('LINESTRINGM EMPTY'),
			('LINESTRINGZM EMPTY'),
			('LINESTRING(1 2,3 4)'),
			('LINESTRINGZ(1 2 3,4 5 6)'),
			('LINESTRINGM(1 2 3,4 5 6)'),
			('LINESTRINGZM(1 2 3 4,5 6 7 8)'),
			('LINESTRING(1 2,3 4,5 6)'),
			('LINESTRINGZ(1 2 3,3 4 5,5 6 7)'),
			('LINESTRINGM(1 2 3,3 4 5,5 6 7)'),
			('LINESTRINGZM(1 2 3 4,3 4 5 6,5 6 7 8)'),
			('POLYGON EMPTY'),
			('POLYGONZ EMPTY'),
			('POLYGONM EMPTY'),
			('POLYGONZM EMPTY'),
			('POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))'),
			('POLYGONZ((0 0 9,4 0 9,0 4 9,0 0 9),(1 1 9,2 1 9,1 2 9,1 1 9))'),
			('POLYGONM((0 0 9,4 0 9,0 4 9,0 0 9),(1 1 9,2 1 9,1 2 9,1 1 9))'),
			('POLYGONZM((0 0 9 9,4 0 9 9,0 4 9 9,0 0 9 9),(1 1 9 9,2 1 9 9,1 2 9 9,1 1 9 9))'),
			('MULTIPOINT EMPTY'),
			('MULTIPOINTZ EMPTY'),
			('MULTIPOINTM EMPTY'),
			('MULTIPOINTZM EMPTY'),
			('MULTIPOINT(1 2)'),
			('MULTIPOINTZ(1 2 3)'),
			('MULTIPOINTM(1 2 3)'),
			('MULTIPOINTZM(1 2 3 4)'),
			('MULTIPOINT(1 2,3 4)'),
			('MULTIPOINTZ(1 2 3,3 4 5)'),
			('MULTIPOINTM(1 2 3,3 4 5)'),
			('MULTIPOINTZM(1 2 3 4,3 4 5 6)'),
			('MULTILINESTRING EMPTY'),
			('MULTILINESTRINGZ EMPTY'),
			('MULTILINESTRINGM EMPTY'),
			('MULTILINESTRINGZM EMPTY'),
			('MULTILINESTRING((0 1,2 3,4 5))'),
			('MULTILINESTRINGZ((0 1 8,2 3 8,4 5 8))'),
			('MULTILINESTRINGM((0 1 8,2 3 8,4 5 8))'),
			('MULTILINESTRINGZM((0 1 8 9,2 3 8 9,4 5 8 9))'),
			('MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))'),
			('MULTILINESTRINGZ((0 1 9,2 3 9),(4 5 9,6 7 9,8 9 9))'),
			('MULTILINESTRINGM((0 1 9,2 3 9),(4 5 9,6 7 9,8 9 9))'),
			('MULTILINESTRINGZM((0 1 9 9,2 3 9 9),(4 5 9 9,6 7 9 9,8 9 9 9))'),
			('MULTIPOLYGON EMPTY'),
			('MULTIPOLYGONZ EMPTY'),
			('MULTIPOLYGONM EMPTY'),
			('MULTIPOLYGONZM EMPTY'),
			('MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,1 1,1 0)))'),
			('MULTIPOLYGONZ(((0 0 9,1 0 9,0 1 9,0 0 9)),((1 0 9,2 0 9,1 1 9,1 0 9)))'),
			('MULTIPOLYGONM(((0 0 9,1 0 9,0 1 9,0 0 9)),((1 0 9,2 0 9,1 1 9,1 0 9)))'),
			('MULTIPOLYGONZM(((0 0 8 9,1 0 8 9,0 1 8 9,0 0 8 9)),((1 0 8 9,2 0 8 9,1 1 8 9,1 0 8 9)))'),
			('GEOMETRYCOLLECTION EMPTY'),
			('GEOMETRYCOLLECTIONZ EMPTY'),
			('GEOMETRYCOLLECTIONM EMPTY'),
			('GEOMETRYCOLLECTIONZM EMPTY'),
			('GEOMETRYCOLLECTION(POINT(1 2),POINT(3 4))'),
			('GEOMETRYCOLLECTIONZ(POINTZ(1 2 3),POINTZ(3 4 5))'),
			('GEOMETRYCOLLECTIONM(POINTM(1 2 3),POINTM(3 4 5))'),
			('GEOMETRYCOLLECTIONZM(POINTZM(1 2 3 4),POINTZM(3 4 5 5))')
		) AS q (wkt);
	*/
	for i, tt := range []struct {
		wkb string
		wkt string
	}{
		{
			wkb: "0101000000000000000000f87f000000000000f87f",
			wkt: "POINT EMPTY",
		},
		{
			wkb: "01e9030000000000000000f87f000000000000f87f000000000000f87f",
			wkt: "POINT Z EMPTY",
		},
		{
			wkb: "01d1070000000000000000f87f000000000000f87f000000000000f87f",
			wkt: "POINT M EMPTY",
		},
		{
			wkb: "01b90b0000000000000000f87f000000000000f87f000000000000f87f000000000000f87f",
			wkt: "POINT ZM EMPTY",
		},
		{
			wkb: "0101000000000000000000f03f0000000000000040",
			wkt: "POINT(1 2)",
		},
		{
			wkb: "01e9030000000000000000f03f00000000000000400000000000000840",
			wkt: "POINT Z (1 2 3)",
		},
		{
			wkb: "01d1070000000000000000f03f00000000000000400000000000000840",
			wkt: "POINT M (1 2 3)",
		},
		{
			wkb: "01b90b0000000000000000f03f000000000000004000000000000008400000000000001040",
			wkt: "POINT ZM (1 2 3 4)",
		},
		{
			wkb: "010200000000000000",
			wkt: "LINESTRING EMPTY",
		},
		{
			wkb: "01ea03000000000000",
			wkt: "LINESTRING Z EMPTY",
		},
		{
			wkb: "01d207000000000000",
			wkt: "LINESTRING M EMPTY",
		},
		{
			wkb: "01ba0b000000000000",
			wkt: "LINESTRING ZM EMPTY",
		},
		{
			wkb: "010200000002000000000000000000f03f000000000000004000000000000008400000000000001040",
			wkt: "LINESTRING(1 2,3 4)",
		},
		{
			wkb: "01ea03000002000000000000000000f03f00000000000000400000000000000840000000000000104000000000000014400000000000001840",
			wkt: "LINESTRING Z (1 2 3,4 5 6)",
		},
		{
			wkb: "01d207000002000000000000000000f03f00000000000000400000000000000840000000000000104000000000000014400000000000001840",
			wkt: "LINESTRING M (1 2 3,4 5 6)",
		},
		{
			wkb: "01ba0b000002000000000000000000f03f000000000000004000000000000008400000000000001040000000000000144000000000000018400000000000001c400000000000002040",
			wkt: "LINESTRING ZM (1 2 3 4,5 6 7 8)",
		},
		{
			wkb: "010200000003000000000000000000f03f00000000000000400000000000000840000000000000104000000000000014400000000000001840",
			wkt: "LINESTRING(1 2,3 4,5 6)",
		},
		{
			wkb: "01ea03000003000000000000000000f03f00000000000000400000000000000840000000000000084000000000000010400000000000001440000000000000144000000000000018400000000000001c40",
			wkt: "LINESTRING Z (1 2 3,3 4 5,5 6 7)",
		},
		{
			wkb: "01d207000003000000000000000000f03f00000000000000400000000000000840000000000000084000000000000010400000000000001440000000000000144000000000000018400000000000001c40",
			wkt: "LINESTRING M (1 2 3,3 4 5,5 6 7)",
		},
		{
			wkb: "01ba0b000003000000000000000000f03f0000000000000040000000000000084000000000000010400000000000000840000000000000104000000000000014400000000000001840000000000000144000000000000018400000000000001c400000000000002040",
			wkt: "LINESTRING ZM (1 2 3 4,3 4 5 6,5 6 7 8)",
		},
		{
			wkb: "010300000000000000",
			wkt: "POLYGON EMPTY",
		},
		{
			wkb: "01eb03000000000000",
			wkt: "POLYGON Z EMPTY",
		},
		{
			wkb: "01d307000000000000",
			wkt: "POLYGON M EMPTY",
		},
		{
			wkb: "01bb0b000000000000",
			wkt: "POLYGON ZM EMPTY",
		},
		{
			wkb: "010300000002000000040000000000000000000000000000000000000000000000000010400000000000000000000000000000000000000000000010400000000000000000000000000000000004000000000000000000f03f000000000000f03f0000000000000040000000000000f03f000000000000f03f0000000000000040000000000000f03f000000000000f03f",
			wkt: "POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))",
		},
		{
			wkb: "01eb030000020000000400000000000000000000000000000000000000000000000000224000000000000010400000000000000000000000000000224000000000000000000000000000001040000000000000224000000000000000000000000000000000000000000000224004000000000000000000f03f000000000000f03f00000000000022400000000000000040000000000000f03f0000000000002240000000000000f03f00000000000000400000000000002240000000000000f03f000000000000f03f0000000000002240",
			wkt: "POLYGON Z ((0 0 9,4 0 9,0 4 9,0 0 9),(1 1 9,2 1 9,1 2 9,1 1 9))",
		},
		{
			wkb: "01d3070000020000000400000000000000000000000000000000000000000000000000224000000000000010400000000000000000000000000000224000000000000000000000000000001040000000000000224000000000000000000000000000000000000000000000224004000000000000000000f03f000000000000f03f00000000000022400000000000000040000000000000f03f0000000000002240000000000000f03f00000000000000400000000000002240000000000000f03f000000000000f03f0000000000002240",
			wkt: "POLYGON M ((0 0 9,4 0 9,0 4 9,0 0 9),(1 1 9,2 1 9,1 2 9,1 1 9))",
		},
		{
			wkb: "01bb0b00000200000004000000000000000000000000000000000000000000000000002240000000000000224000000000000010400000000000000000000000000000224000000000000022400000000000000000000000000000104000000000000022400000000000002240000000000000000000000000000000000000000000002240000000000000224004000000000000000000f03f000000000000f03f000000000000224000000000000022400000000000000040000000000000f03f00000000000022400000000000002240000000000000f03f000000000000004000000000000022400000000000002240000000000000f03f000000000000f03f00000000000022400000000000002240",
			wkt: "POLYGON ZM ((0 0 9 9,4 0 9 9,0 4 9 9,0 0 9 9),(1 1 9 9,2 1 9 9,1 2 9 9,1 1 9 9))",
		},
		{
			wkb: "010400000000000000",
			wkt: "MULTIPOINT EMPTY",
		},
		{
			wkb: "01ec03000000000000",
			wkt: "MULTIPOINT Z EMPTY",
		},
		{
			wkb: "01d407000000000000",
			wkt: "MULTIPOINT M EMPTY",
		},
		{
			wkb: "01bc0b000000000000",
			wkt: "MULTIPOINT ZM EMPTY",
		},
		{
			wkb: "0104000000010000000101000000000000000000f03f0000000000000040",
			wkt: "MULTIPOINT(1 2)",
		},
		{
			wkb: "01ec0300000100000001e9030000000000000000f03f00000000000000400000000000000840",
			wkt: "MULTIPOINT Z (1 2 3)",
		},
		{
			wkb: "01d40700000100000001d1070000000000000000f03f00000000000000400000000000000840",
			wkt: "MULTIPOINT M (1 2 3)",
		},
		{
			wkb: "01bc0b00000100000001b90b0000000000000000f03f000000000000004000000000000008400000000000001040",
			wkt: "MULTIPOINT ZM (1 2 3 4)",
		},
		{
			wkb: "0104000000020000000101000000000000000000f03f0000000000000040010100000000000000000008400000000000001040",
			wkt: "MULTIPOINT(1 2,3 4)",
		},
		{
			wkb: "01ec0300000200000001e9030000000000000000f03f0000000000000040000000000000084001e9030000000000000000084000000000000010400000000000001440",
			wkt: "MULTIPOINT Z (1 2 3,3 4 5)",
		},
		{
			wkb: "01d40700000200000001d1070000000000000000f03f0000000000000040000000000000084001d1070000000000000000084000000000000010400000000000001440",
			wkt: "MULTIPOINT M (1 2 3,3 4 5)",
		},
		{
			wkb: "01bc0b00000200000001b90b0000000000000000f03f00000000000000400000000000000840000000000000104001b90b00000000000000000840000000000000104000000000000014400000000000001840",
			wkt: "MULTIPOINT ZM (1 2 3 4,3 4 5 6)",
		},
		{
			wkb: "010500000000000000",
			wkt: "MULTILINESTRING EMPTY",
		},
		{
			wkb: "01ed03000000000000",
			wkt: "MULTILINESTRING Z EMPTY",
		},
		{
			wkb: "01d507000000000000",
			wkt: "MULTILINESTRING M EMPTY",
		},
		{
			wkb: "01bd0b000000000000",
			wkt: "MULTILINESTRING ZM EMPTY",
		},
		{
			wkb: "0105000000010000000102000000030000000000000000000000000000000000f03f0000000000000040000000000000084000000000000010400000000000001440",
			wkt: "MULTILINESTRING((0 1,2 3,4 5))",
		},
		{
			wkb: "01ed0300000100000001ea030000030000000000000000000000000000000000f03f0000000000002040000000000000004000000000000008400000000000002040000000000000104000000000000014400000000000002040",
			wkt: "MULTILINESTRING Z ((0 1 8,2 3 8,4 5 8))",
		},
		{
			wkb: "01d50700000100000001d2070000030000000000000000000000000000000000f03f0000000000002040000000000000004000000000000008400000000000002040000000000000104000000000000014400000000000002040",
			wkt: "MULTILINESTRING M ((0 1 8,2 3 8,4 5 8))",
		},
		{
			wkb: "01bd0b00000100000001ba0b0000030000000000000000000000000000000000f03f0000000000002040000000000000224000000000000000400000000000000840000000000000204000000000000022400000000000001040000000000000144000000000000020400000000000002240",
			wkt: "MULTILINESTRING ZM ((0 1 8 9,2 3 8 9,4 5 8 9))",
		},
		{
			wkb: "0105000000020000000102000000020000000000000000000000000000000000f03f000000000000004000000000000008400102000000030000000000000000001040000000000000144000000000000018400000000000001c4000000000000020400000000000002240",
			wkt: "MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))",
		},
		{
			wkb: "01ed0300000200000001ea030000020000000000000000000000000000000000f03f000000000000224000000000000000400000000000000840000000000000224001ea0300000300000000000000000010400000000000001440000000000000224000000000000018400000000000001c400000000000002240000000000000204000000000000022400000000000002240",
			wkt: "MULTILINESTRING Z ((0 1 9,2 3 9),(4 5 9,6 7 9,8 9 9))",
		},
		{
			wkb: "01d50700000200000001d2070000020000000000000000000000000000000000f03f000000000000224000000000000000400000000000000840000000000000224001d20700000300000000000000000010400000000000001440000000000000224000000000000018400000000000001c400000000000002240000000000000204000000000000022400000000000002240",
			wkt: "MULTILINESTRING M ((0 1 9,2 3 9),(4 5 9,6 7 9,8 9 9))",
		},
		{
			wkb: "01bd0b00000200000001ba0b0000020000000000000000000000000000000000f03f00000000000022400000000000002240000000000000004000000000000008400000000000002240000000000000224001ba0b000003000000000000000000104000000000000014400000000000002240000000000000224000000000000018400000000000001c40000000000000224000000000000022400000000000002040000000000000224000000000000022400000000000002240",
			wkt: "MULTILINESTRING ZM ((0 1 9 9,2 3 9 9),(4 5 9 9,6 7 9 9,8 9 9 9))",
		},
		{
			wkb: "010600000000000000",
			wkt: "MULTIPOLYGON EMPTY",
		},
		{
			wkb: "01ee03000000000000",
			wkt: "MULTIPOLYGON Z EMPTY",
		},
		{
			wkb: "01d607000000000000",
			wkt: "MULTIPOLYGON M EMPTY",
		},
		{
			wkb: "01be0b000000000000",
			wkt: "MULTIPOLYGON ZM EMPTY",
		},
		{
			wkb: "0106000000020000000103000000010000000400000000000000000000000000000000000000000000000000f03f00000000000000000000000000000000000000000000f03f0000000000000000000000000000000001030000000100000004000000000000000000f03f000000000000000000000000000000400000000000000000000000000000f03f000000000000f03f000000000000f03f0000000000000000",
			wkt: "MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,1 1,1 0)))",
		},
		{
			wkb: "01ee0300000200000001eb0300000100000004000000000000000000000000000000000000000000000000002240000000000000f03f000000000000000000000000000022400000000000000000000000000000f03f000000000000224000000000000000000000000000000000000000000000224001eb0300000100000004000000000000000000f03f00000000000000000000000000002240000000000000004000000000000000000000000000002240000000000000f03f000000000000f03f0000000000002240000000000000f03f00000000000000000000000000002240",
			wkt: "MULTIPOLYGON Z (((0 0 9,1 0 9,0 1 9,0 0 9)),((1 0 9,2 0 9,1 1 9,1 0 9)))",
		},
		{
			wkb: "01d60700000200000001d30700000100000004000000000000000000000000000000000000000000000000002240000000000000f03f000000000000000000000000000022400000000000000000000000000000f03f000000000000224000000000000000000000000000000000000000000000224001d30700000100000004000000000000000000f03f00000000000000000000000000002240000000000000004000000000000000000000000000002240000000000000f03f000000000000f03f0000000000002240000000000000f03f00000000000000000000000000002240",
			wkt: "MULTIPOLYGON M (((0 0 9,1 0 9,0 1 9,0 0 9)),((1 0 9,2 0 9,1 1 9,1 0 9)))",
		},
		{
			wkb: "01be0b00000200000001bb0b000001000000040000000000000000000000000000000000000000000000000020400000000000002240000000000000f03f0000000000000000000000000000204000000000000022400000000000000000000000000000f03f00000000000020400000000000002240000000000000000000000000000000000000000000002040000000000000224001bb0b00000100000004000000000000000000f03f0000000000000000000000000000204000000000000022400000000000000040000000000000000000000000000020400000000000002240000000000000f03f000000000000f03f00000000000020400000000000002240000000000000f03f000000000000000000000000000020400000000000002240",
			wkt: "MULTIPOLYGON ZM (((0 0 8 9,1 0 8 9,0 1 8 9,0 0 8 9)),((1 0 8 9,2 0 8 9,1 1 8 9,1 0 8 9)))",
		},
		{
			wkb: "010700000000000000",
			wkt: "GEOMETRYCOLLECTION EMPTY",
		},
		{
			wkb: "01ef03000000000000",
			wkt: "GEOMETRYCOLLECTION Z EMPTY",
		},
		{
			wkb: "01d707000000000000",
			wkt: "GEOMETRYCOLLECTION M EMPTY",
		},
		{
			wkb: "01bf0b000000000000",
			wkt: "GEOMETRYCOLLECTION ZM EMPTY",
		},
		{
			wkb: "0107000000020000000101000000000000000000f03f0000000000000040010100000000000000000008400000000000001040",
			wkt: "GEOMETRYCOLLECTION(POINT(1 2),POINT(3 4))",
		},
		{
			wkb: "01ef0300000200000001e9030000000000000000f03f0000000000000040000000000000084001e9030000000000000000084000000000000010400000000000001440",
			wkt: "GEOMETRYCOLLECTION Z (POINT Z (1 2 3),POINT Z (3 4 5))",
		},
		{
			wkb: "01d70700000200000001d1070000000000000000f03f0000000000000040000000000000084001d1070000000000000000084000000000000010400000000000001440",
			wkt: "GEOMETRYCOLLECTION M (POINT M (1 2 3),POINT M (3 4 5))",
		},
		{
			wkb: "01bf0b00000200000001b90b0000000000000000f03f00000000000000400000000000000840000000000000104001b90b00000000000000000840000000000000104000000000000014400000000000001440",
			wkt: "GEOMETRYCOLLECTION ZM (POINT ZM (1 2 3 4),POINT ZM (3 4 5 5))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			geom, err := UnmarshalWKB(bytes.NewReader(hexStringToBytes(t, tt.wkb)))
			expectNoErr(t, err)
			expectGeomEq(t, geom, geomFromWKT(t, tt.wkt))
		})
	}
}

func TestWKBParserInvalidGeometryType(t *testing.T) {
	// Same as POINT(1 2), but with the geometry type byte set to 0xff.
	const wkb = "01ff000000000000000000f03f0000000000000040"
	_, err := UnmarshalWKB(bytes.NewReader(hexStringToBytes(t, wkb)))
	if err == nil {
		t.Errorf("expected an error but got nil")
	}
	if !strings.Contains(err.Error(), "invalid geometry type") {
		t.Errorf("expected to be an error about unknown geometry type, but got: %v", err)
	}
}

func TestWKBMarshalValid(t *testing.T) {
	for i, wkt := range []string{
		"POINT EMPTY",
		"POINT(1 2)",
		"LINESTRING EMPTY",
		"LINESTRING(1 2,3 4)",
		"LINESTRING(1 2,3 4,5 6)",
		"POLYGON EMPTY",
		"POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))",
		"MULTIPOINT EMPTY",
		"MULTIPOINT(1 2)",
		"MULTIPOINT(1 2,3 4)",
		"MULTIPOINT(EMPTY,1 2)",
		"MULTIPOINT(1 2,EMPTY)",
		"MULTILINESTRING EMPTY",
		"MULTILINESTRING((0 1,2 3,4 5))",
		"MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))",
		"MULTILINESTRING(EMPTY,(0 1,2 3,4 5))",
		"MULTILINESTRING((0 1,2 3,4 5),EMPTY)",
		"MULTIPOLYGON EMPTY",
		"MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,1 1,1 0)))",
		"GEOMETRYCOLLECTION EMPTY",
		"GEOMETRYCOLLECTION(POINT(1 2),LINESTRING(1 2,3 4))",
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			geom := geomFromWKT(t, wkt)
			buf := geom.AsBinary()
			readBackGeom, err := UnmarshalWKB(bytes.NewReader(buf))
			expectNoErr(t, err)
			expectGeomEq(t, readBackGeom, geom)
		})
	}
}

func TestWKBMarshalEmptyPoint(t *testing.T) {
	for i, tt := range []struct {
		wkt string
		hex string
	}{
		{
			"POINT EMPTY",
			"01" +
				"01000000" +
				"010000000000f87f" +
				"010000000000f87f",
		},
		{
			"POINT Z EMPTY",
			"01" +
				"e9030000" +
				"010000000000f87f" +
				"010000000000f87f" +
				"010000000000f87f",
		},
		{
			"POINT M EMPTY",
			"01" +
				"d1070000" +
				"010000000000f87f" +
				"010000000000f87f" +
				"010000000000f87f",
		},
		{
			"POINT ZM EMPTY",
			"01" +
				"b90b0000" +
				"010000000000f87f" +
				"010000000000f87f" +
				"010000000000f87f" +
				"010000000000f87f",
		},
		{
			"MULTIPOINT(EMPTY)",
			"01" +
				"04000000" +
				"01000000" +
				"01" +
				"01000000" +
				"010000000000f87f" +
				"010000000000f87f",
		},
		{
			"MULTIPOINT Z (EMPTY)",
			"01" +
				"ec030000" +
				"01000000" +
				"01" +
				"e9030000" +
				"010000000000f87f" +
				"010000000000f87f" +
				"010000000000f87f",
		},
		{
			"MULTIPOINT M (EMPTY)",
			"01" +
				"d4070000" +
				"01000000" +
				"01" +
				"d1070000" +
				"010000000000f87f" +
				"010000000000f87f" +
				"010000000000f87f",
		},
		{
			"MULTIPOINT ZM (EMPTY)",
			"01" +
				"bc0b0000" +
				"01000000" +
				"01" +
				"b90b0000" +
				"010000000000f87f" +
				"010000000000f87f" +
				"010000000000f87f" +
				"010000000000f87f",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			g := geomFromWKT(t, tt.wkt)
			buf := g.AsBinary()
			want := hexStringToBytes(t, tt.hex)
			if !bytes.Equal(want, buf) {
				t.Logf("wkt: %v", tt.wkt)
				t.Logf("want:\n%v", hex.Dump(want))
				t.Logf("got:\n%v", hex.Dump(buf))
				t.Errorf("mismatch")
			}
		})
	}
}
