package geom_test

import (
	"bytes"
	"strconv"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
	. "github.com/peterstace/simplefeatures/internal/geomtest"
)

func TestValuerAny(t *testing.T) {
	g := GeomFromWKT(t, "POINT(1 2)")
	val, err := g.Value()
	if err != nil {
		t.Fatal(err)
	}
	geom, err := UnmarshalWKB(bytes.NewReader(val.([]byte)))
	if err != nil {
		t.Fatal(err)
	}
	ExpectGeomEq(t, g, geom)
}

func TestScanner(t *testing.T) {
	const wkt = "POINT(2 3)"
	var wkb bytes.Buffer
	ExpectNoErr(t, GeomFromWKT(t, wkt).AsBinary(&wkb))
	var g Geometry
	check := func(t *testing.T, err error) {
		if err != nil {
			t.Fatal(err)
		}
		ExpectGeomEq(t, g, GeomFromWKT(t, wkt))
	}
	t.Run("string", func(t *testing.T) {
		g = Geometry{}
		check(t, g.Scan(string(wkb.Bytes())))
	})
	t.Run("byte", func(t *testing.T) {
		g = Geometry{}
		check(t, g.Scan([]byte(wkb.Bytes())))
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
			geom := GeomFromWKT(t, wkt)
			val, err := geom.Value()
			ExpectNoErr(t, err)
			g, err := UnmarshalWKB(bytes.NewReader(val.([]byte)))
			ExpectNoErr(t, err)
			ExpectGeomEq(t, g, geom)
		})
	}
}
