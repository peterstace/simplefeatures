package geom_test

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestZeroGeometry(t *testing.T) {
	var z Geometry
	expectBoolEq(t, z.IsGeometryCollection(), true)
	z.AsGeometryCollection() // Doesn't crash.
	expectStringEq(t, z.AsText(), "GEOMETRYCOLLECTION EMPTY")

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(z)
	expectNoErr(t, err)
	expectStringEq(t, strings.TrimSpace(buf.String()), `{"type":"GeometryCollection","geometries":[]}`)

	pt, err := NewPointFromXY(XY{1, 2}) // Set away from zero value
	z = pt.AsGeometry()
	expectNoErr(t, err)
	expectBoolEq(t, z.IsPoint(), true)
	err = json.NewDecoder(&buf).Decode(&z)
	expectNoErr(t, err)
	expectBoolEq(t, z.IsPoint(), false)
	expectBoolEq(t, z.IsGeometryCollection(), true)
	expectBoolEq(t, z.IsEmpty(), true)
	z = Geometry{}

	_ = z.AsBinary() // Doesn't crash

	_, err = z.Value()
	expectNoErr(t, err)

	expectIntEq(t, z.Dimension(), 0)
}

func TestGeometryType(t *testing.T) {
	for i, tt := range []struct {
		wkt     string
		geoType GeometryType
	}{
		{"POINT(1 1)", TypePoint},
		{"POINT EMPTY", TypePoint},
		{"MULTIPOINT EMPTY", TypeMultiPoint},
		{"MULTIPOINT ((10 40), (40 30), (20 20), (30 10))", TypeMultiPoint},
		{"LINESTRING(1 2,3 4)", TypeLineString},
		{"LINESTRING(1 2,3 4,5 6)", TypeLineString},
		{"LINESTRING EMPTY", TypeLineString},
		{"MULTILINESTRING ((10 10, 20 20, 10 40),(40 40, 30 30, 40 20, 30 10))", TypeMultiLineString},
		{"MULTILINESTRING EMPTY", TypeMultiLineString},
		{"MULTILINESTRING(EMPTY)", TypeMultiLineString},
		{"POLYGON((1 1,3 1,2 2,2 4,1 1))", TypePolygon},
		{"POLYGON EMPTY", TypePolygon},
		{"MULTIPOLYGON (((40 40, 20 45, 45 30, 40 40)),((20 35, 10 30, 10 10, 30 5, 45 20, 20 35),(30 20, 20 15, 20 25, 30 20)))", TypeMultiPolygon},
		{"MULTIPOLYGON EMPTY", TypeMultiPolygon},
		{"MULTIPOLYGON(EMPTY)", TypeMultiPolygon},
		{"GEOMETRYCOLLECTION EMPTY", TypeGeometryCollection},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log("wkt:", tt.wkt)
			g := geomFromWKT(t, tt.wkt)
			if tt.geoType != g.Type() {
				t.Errorf("expect: %s, got %s", tt.geoType, g.Type())
			}
		})
	}
}

func TestGeometryTypeString(t *testing.T) {
	for _, tc := range []struct {
		typ  GeometryType
		want string
	}{
		{TypeGeometryCollection, "GeometryCollection"},
		{TypePoint, "Point"},
		{TypeMultiPoint, "MultiPoint"},
		{TypeLineString, "LineString"},
		{TypeMultiLineString, "MultiLineString"},
		{TypePolygon, "Polygon"},
		{TypeMultiPolygon, "MultiPolygon"},
		{99, "invalid"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			got := tc.typ.String()
			expectStringEq(t, got, tc.want)
		})
	}
}
