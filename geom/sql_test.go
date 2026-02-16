package geom_test

import (
	"strconv"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/test"
)

func TestSQLValueGeometry(t *testing.T) {
	g := test.FromWKT(t, "POINT(1 2)")
	val, err := g.Value()
	if err != nil {
		t.Fatal(err)
	}
	geom, err := geom.UnmarshalWKB(val.([]byte))
	if err != nil {
		t.Fatal(err)
	}
	test.ExactEquals(t, g, geom)
}

func TestSQLScanGeometry(t *testing.T) {
	const wkt = "POINT(2 3)"
	wkb := test.FromWKT(t, wkt).AsBinary()
	var g geom.Geometry
	check := func(t *testing.T, err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		test.ExactEquals(t, g, test.FromWKT(t, wkt))
	}
	t.Run("string", func(t *testing.T) {
		g = geom.Geometry{}
		check(t, g.Scan(string(wkb)))
	})
	t.Run("byte", func(t *testing.T) {
		g = geom.Geometry{}
		check(t, g.Scan(wkb))
	})
}

func TestSQLValueConcrete(t *testing.T) {
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
			in := test.FromWKT(t, wkt)
			val, err := in.Value()
			test.NoErr(t, err)
			out, err := geom.UnmarshalWKB(val.([]byte))
			test.NoErr(t, err)
			test.ExactEquals(t, out, in)
		})
	}
}

func TestSQLScanConcrete(t *testing.T) {
	for i, tc := range []struct {
		wkt      string
		concrete interface {
			AsText() string
			Scan(any) error
		}
	}{
		{"POINT(0 1)", new(geom.Point)},
		{"MULTIPOINT((0 1))", new(geom.MultiPoint)},
		{"LINESTRING(0 1,1 0)", new(geom.LineString)},
		{"MULTILINESTRING((0 1,1 0))", new(geom.MultiLineString)},
		{"POLYGON((0 0,1 0,0 1,0 0))", new(geom.Polygon)},
		{"MULTIPOLYGON(((0 0,1 0,0 1,0 0)))", new(geom.MultiPolygon)},
		{"GEOMETRYCOLLECTION(MULTIPOLYGON(((0 0,1 0,0 1,0 0))))", new(geom.GeometryCollection)},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			wkb := test.FromWKT(t, tc.wkt).AsBinary()
			err := tc.concrete.Scan(wkb)
			test.NoErr(t, err)
			test.Eq(t, tc.concrete.AsText(), tc.wkt)
		})
	}
}
