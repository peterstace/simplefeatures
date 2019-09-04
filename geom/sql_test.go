package geom_test

import (
	"bytes"
	"database/sql/driver"
	"strconv"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestValuerAny(t *testing.T) {
	any := AnyGeometry{Geom: geomFromWKT(t, "POINT(1 2)")}
	val, err := any.Value()
	if err != nil {
		t.Fatal(err)
	}
	geom, err := UnmarshalWKB(bytes.NewReader(val.([]byte)))
	if err != nil {
		t.Fatal(err)
	}
	expectDeepEqual(t, any.Geom, geom)
}

func TestValuerAnyZero(t *testing.T) {
	var any AnyGeometry
	if _, err := any.Value(); err == nil {
		t.Fatal("expected an error")
	}
}

func TestScanner(t *testing.T) {
	const wkt = "POINT(2 3)"
	var wkb bytes.Buffer
	expectNoErr(t, geomFromWKT(t, wkt).AsBinary(&wkb))
	var any AnyGeometry
	check := func(t *testing.T, err error) {
		if err != nil {
			t.Fatal(err)
		}
		expectDeepEqual(t, any.Geom, geomFromWKT(t, wkt))
	}
	t.Run("string", func(t *testing.T) {
		any = AnyGeometry{}
		check(t, any.Scan(string(wkb.Bytes())))
	})
	t.Run("byte", func(t *testing.T) {
		any = AnyGeometry{}
		check(t, any.Scan([]byte(wkb.Bytes())))
	})
}

func TestValuerConcrete(t *testing.T) {
	for i, wkt := range []string{
		"POINT EMPTY",
		"POINT(1 2)",
		"LINESTRING(1 2,3 4)",
		"LINESTRING(1 2,3 4,5 6)",
		"POLYGON((0 0,1 0,0 1,0 0))",
		"MULTIPOINT((1 2))",
		"MULTILINESTRING((1 2,3 4,5 6))",
		"MULTIPOLYGON(((0 0,1 0,0 1,0 0)))",
		"GEOMETRYCOLLECTION(POINT(1 2))",
		"GEOMETRYCOLLECTION EMPTY",
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(wkt)
			geom := geomFromWKT(t, wkt).(driver.Valuer)
			val, err := geom.Value()
			expectNoErr(t, err)
			g, err := UnmarshalWKB(bytes.NewReader(val.([]byte)))
			expectNoErr(t, err)
			expectDeepEqual(t, geom, g)
		})
	}
}
