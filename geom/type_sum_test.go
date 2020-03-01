package geom_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
	. "github.com/peterstace/simplefeatures/internal/geomtest"
)

func TestZeroGeometry(t *testing.T) {
	var z Geometry
	ExpectBoolEq(t, z.IsGeometryCollection(), true)
	z.AsGeometryCollection() // Doesn't crash.
	ExpectStringEq(t, z.AsText(), "GEOMETRYCOLLECTION EMPTY")

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(z)
	ExpectNoErr(t, err)
	ExpectStringEq(t, strings.TrimSpace(buf.String()), `{"type":"GeometryCollection","geometries":[]}`)

	z = NewPointF(1, 2).AsGeometry() // Set away from zero value
	ExpectBoolEq(t, z.IsPoint(), true)
	err = json.NewDecoder(&buf).Decode(&z)
	ExpectNoErr(t, err)
	ExpectBoolEq(t, z.IsPoint(), false)
	ExpectBoolEq(t, z.IsGeometryCollection(), true)
	ExpectBoolEq(t, z.IsEmpty(), true)
	z = Geometry{}

	z.AsBinary(ioutil.Discard) // Doesn't crash

	_, err = z.Value()
	ExpectNoErr(t, err)

	ExpectIntEq(t, z.Dimension(), 0)
}

func TestGeometryType(t *testing.T) {
	for i, tt := range []struct {
		wkt     string
		geoType string
	}{
		{"POINT(1 1)", "Point"},
		{"POINT EMPTY", "Point"},
		{"MULTIPOINT EMPTY", "MultiPoint"},
		{"MULTIPOINT ((10 40), (40 30), (20 20), (30 10))", "MultiPoint"},
		{"LINESTRING(1 2,3 4)", "LineString"},
		{"LINESTRING(1 2,3 4,5 6)", "LineString"},
		{"LINESTRING EMPTY", "LineString"},
		{"MULTILINESTRING ((10 10, 20 20, 10 40),(40 40, 30 30, 40 20, 30 10))", "MultiLineString"},
		{"MULTILINESTRING EMPTY", "MultiLineString"},
		{"MULTILINESTRING(EMPTY)", "MultiLineString"},
		{"POLYGON((1 1,3 1,2 2,2 4,1 1))", "Polygon"},
		{"POLYGON EMPTY", "Polygon"},
		{"MULTIPOLYGON (((40 40, 20 45, 45 30, 40 40)),((20 35, 10 30, 10 10, 30 5, 45 20, 20 35),(30 20, 20 15, 20 25, 30 20)))", "MultiPolygon"},
		{"MULTIPOLYGON EMPTY", "MultiPolygon"},
		{"MULTIPOLYGON(EMPTY)", "MultiPolygon"},
		{"GEOMETRYCOLLECTION EMPTY", "GeometryCollection"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log("wkt:", tt.wkt)
			g := GeomFromWKT(t, tt.wkt)
			if tt.geoType != g.Type() {
				t.Errorf("expect: %s, got %s", tt.geoType, g.Type())
			}
		})
	}
}
