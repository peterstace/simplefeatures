package simplefeatures_test

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures"
)

func hexStringToBytes(t *testing.T, s string) []byte {
	t.Helper()
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

func TestWKBParser(t *testing.T) {
	for i, tt := range []struct {
		wkb string
		wkt string
	}{
		{
			// POINT(1 2)
			wkb: "0101000000000000000000f03f0000000000000040",
			wkt: "POINT(1 2)",
		},
		{
			// POINTZ(1 2 3)
			wkb: "01e9030000000000000000f03f00000000000000400000000000000840",
			wkt: "POINT(1 2)",
		},
		{
			// POINTM(1 2 3)
			wkb: "01d1070000000000000000f03f00000000000000400000000000000840",
			wkt: "POINT(1 2)",
		},
		{
			// POINTZM(1 2 3 4)
			wkb: "01b90b0000000000000000f03f000000000000004000000000000008400000000000001040",
			wkt: "POINT(1 2)",
		},
		{
			// LINESTRING EMPTY
			wkb: "010200000000000000",
			wkt: "LINESTRING EMPTY",
		},
		{
			// LINESTRINGZ EMPTY
			wkb: "01ea03000000000000",
			wkt: "LINESTRING EMPTY",
		},
		{
			// LINESTRINGM EMPTY
			wkb: "01d207000000000000",
			wkt: "LINESTRING EMPTY",
		},
		{
			// LINESTRINGZM EMPTY
			wkb: "01ba0b000000000000",
			wkt: "LINESTRING EMPTY",
		},
		{
			// LINESTRING(1 2,3 4)
			wkb: "010200000002000000000000000000f03f000000000000004000000000000008400000000000001040",
			wkt: "LINESTRING(1 2,3 4)",
		},

		{
			// LINESTRINGZ(1 2 3,4 5 6)
			wkb: "01ea03000002000000000000000000f03f00000000000000400000000000000840000000000000104000000000000014400000000000001840",
			wkt: "LINESTRING(1 2,4 5)",
		},
		{
			// LINESTRINGM(1 2 3,4 5 6)
			wkb: "01d207000002000000000000000000f03f00000000000000400000000000000840000000000000104000000000000014400000000000001840",
			wkt: "LINESTRING(1 2,4 5)",
		},
		{
			// LINESTRINGZM(1 2 3 4,5 6 7 8)
			wkb: "01ba0b000002000000000000000000f03f000000000000004000000000000008400000000000001040000000000000144000000000000018400000000000001c400000000000002040",
			wkt: "LINESTRING(1 2,5 6)",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			geom, err := UnmarshalWKB(bytes.NewReader(hexStringToBytes(t, tt.wkb)))
			expectNoErr(t, err)
			expectDeepEqual(t, geom, geomFromWKT(t, tt.wkt))
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
	if !strings.Contains(err.Error(), "unknown geometry type") {
		t.Errorf("expected to be an error about unknown geometry type, but got: %v", err)
	}
}
