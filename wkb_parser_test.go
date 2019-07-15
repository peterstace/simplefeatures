package simplefeatures_test

import (
	"bytes"
	"strconv"
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
	for _, tt := range []struct {
		wkt string
		wkb string
	}{
		{
			wkt: "POINT(1 2)",
			wkb: "0101000000000000000000F03F0000000000000040",
		},
	} {
		_, err := UnmarshalWKB(bytes.NewReader(hexStringToBytes(t, tt.wkb)))
		expectNoErr(t, err)
		//expectDeepEqual(t, geom, geomFromWKT(t, tt.wkt))
	}
}
