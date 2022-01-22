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
	z.MustAsGeometryCollection() // Doesn't crash.
	expectStringEq(t, z.AsText(), "GEOMETRYCOLLECTION EMPTY")
	gc, ok := z.AsGeometryCollection()
	expectTrue(t, ok)
	expectIntEq(t, gc.NumGeometries(), 0)

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(z)
	expectNoErr(t, err)
	expectStringEq(t, strings.TrimSpace(buf.String()), `{"type":"GeometryCollection","geometries":[]}`)

	pt, err := XY{1, 2}.AsPoint()
	expectNoErr(t, err)
	z = pt.AsGeometry() // Set away from zero value
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

func TestAsConcreteType(t *testing.T) {
	for _, wkt := range []string{
		"GEOMETRYCOLLECTION(POINT(1 2))",
		"POINT(1 2)",
		"LINESTRING(1 2,3 4)",
		"POLYGON((0 0,0 1,1 0,0 0))",
		"MULTIPOINT((1 2))",
		"MULTILINESTRING((1 2,3 4))",
		"MULTIPOLYGON(((0 0,0 1,1 0,0 0)))",
	} {
		t.Run(wkt, func(t *testing.T) {
			g := geomFromWKT(t, wkt)

			if g.IsGeometryCollection() {
				concrete, ok := g.AsGeometryCollection()
				expectTrue(t, ok)
				expectFalse(t, concrete.IsEmpty())
			} else {
				_, ok := g.AsGeometryCollection()
				expectFalse(t, ok)
			}

			if g.IsPoint() {
				concrete, ok := g.AsPoint()
				expectTrue(t, ok)
				expectFalse(t, concrete.IsEmpty())
			} else {
				_, ok := g.AsPoint()
				expectFalse(t, ok)
			}

			if g.IsLineString() {
				concrete, ok := g.AsLineString()
				expectTrue(t, ok)
				expectFalse(t, concrete.IsEmpty())
			} else {
				_, ok := g.AsLineString()
				expectFalse(t, ok)
			}

			if g.IsPolygon() {
				concrete, ok := g.AsPolygon()
				expectTrue(t, ok)
				expectFalse(t, concrete.IsEmpty())
			} else {
				_, ok := g.AsPolygon()
				expectFalse(t, ok)
			}

			if g.IsMultiPoint() {
				concrete, ok := g.AsMultiPoint()
				expectTrue(t, ok)
				expectFalse(t, concrete.IsEmpty())
			} else {
				_, ok := g.AsMultiPoint()
				expectFalse(t, ok)
			}

			if g.IsMultiLineString() {
				concrete, ok := g.AsMultiLineString()
				expectTrue(t, ok)
				expectFalse(t, concrete.IsEmpty())
			} else {
				_, ok := g.AsMultiLineString()
				expectFalse(t, ok)
			}

			if g.IsMultiPolygon() {
				concrete, ok := g.AsMultiPolygon()
				expectTrue(t, ok)
				expectFalse(t, concrete.IsEmpty())
			} else {
				_, ok := g.AsMultiPolygon()
				expectFalse(t, ok)
			}
		})
	}
}

func TestGeometryIsCWandIsCCW(t *testing.T) {
	for i, tt := range []struct {
		wkt     string
		geoType GeometryType
		isCW    bool
		isCCW   bool
		note    string
	}{
		{"POINT(1 1)", TypePoint, true, true, "undefined"},
		{"POINT EMPTY", TypePoint, true, true, "undefined"},
		{"MULTIPOINT EMPTY", TypeMultiPoint, true, true, "undefined"},
		{"MULTIPOINT ((10 40), (40 30), (20 20), (30 10))", TypeMultiPoint, true, true, "undefined"},
		{"LINESTRING(1 2,3 4)", TypeLineString, true, true, "undefined"},
		{"LINESTRING(1 2,3 4,5 6)", TypeLineString, true, true, "undefined"},
		{"LINESTRING EMPTY", TypeLineString, true, true, "undefined"},
		{"MULTILINESTRING ((10 10, 20 20, 10 40),(40 40, 30 30, 40 20, 30 10))", TypeMultiLineString, true, true, "undefined"},
		{"MULTILINESTRING EMPTY", TypeMultiLineString, true, true, "undefined"},
		{"MULTILINESTRING(EMPTY)", TypeMultiLineString, true, true, "undefined"},
		{"POLYGON((0 0,0 5,5 5,5 0,0 0))", TypePolygon, true, false, "CW"},
		{"POLYGON((1 1,3 1,2 2,2 4,1 1))", TypePolygon, false, true, "CCW"},
		{"POLYGON((0 0,0 5,5 5,5 0,0 0), (1 1,3 1,2 2,2 4,1 1))", TypePolygon, true, false, "outer CW inner CCW"},
		{"POLYGON((0 0,5 0,5 5,0 5,0 0), (1 1,1 2,2 2,2 1,1 1))", TypePolygon, false, true, "outer CCW inner CW"},
		{"POLYGON((0 0,5 0,5 5,0 5,0 0), (1 1,2 1,2 2,1 2,1 1))", TypePolygon, false, false, "outer CCW, inner CCW, inconsistent"},
		{"POLYGON EMPTY", TypePolygon, true, true, "empty"},
		{"MULTIPOLYGON(((40 40, 45 30, 20 45, 40 40)),((20 35, 45 20, 30 5, 10 10, 10 30, 20 35),(30 20, 20 25, 20 15, 30 20)))", TypeMultiPolygon, true, false, "all CW"},
		{"MULTIPOLYGON(((40 40, 20 45, 45 30, 40 40)),((20 35, 10 30, 10 10, 30 5, 45 20, 20 35),(30 20, 20 15, 20 25, 30 20)))", TypeMultiPolygon, false, true, "all CCW"},
		{"MULTIPOLYGON(((40 40, 45 30, 20 45, 40 40)),((20 35, 10 30, 10 10, 30 5, 45 20, 20 35),(30 20, 20 15, 20 25, 30 20)))", TypeMultiPolygon, false, false, "first CW, next CCW, inconsistent"},
		{"MULTIPOLYGON EMPTY", TypeMultiPolygon, true, true, "empty"},
		{"MULTIPOLYGON(EMPTY)", TypeMultiPolygon, true, true, "empty"},
		{"GEOMETRYCOLLECTION(POLYGON((0 0,0 5,5 5,5 0,0 0)), MULTIPOLYGON(((40 40, 45 30, 20 45, 40 40)),((20 35, 45 20, 30 5, 10 10, 10 30, 20 35),(30 20, 20 25, 20 15, 30 20))))", TypeGeometryCollection, true, false, "all CW"},
		{"GEOMETRYCOLLECTION(POLYGON((0 0,0 5,5 5,5 0,0 0)), MULTIPOLYGON(((40 40, 20 45, 45 30, 40 40)),((20 35, 10 30, 10 10, 30 5, 45 20, 20 35),(30 20, 20 15, 20 25, 30 20))))", TypeGeometryCollection, false, false, "first CW, next CCW, inconsistent"},
		{"GEOMETRYCOLLECTION EMPTY", TypeGeometryCollection, true, true, "empty"},
		{"GEOMETRYCOLLECTION(POINT(1 1))", TypeGeometryCollection, true, true, "first undefined"},
		{"GEOMETRYCOLLECTION(POLYGON((0 0,0 5,5 5,5 0,0 0)), POINT(1 1))", TypeGeometryCollection, true, false, "first CW, second undefined"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log("wkt:", tt.wkt)
			g := geomFromWKT(t, tt.wkt)
			if tt.geoType != g.Type() {
				t.Errorf("expect: type %s, got %s", tt.geoType, g.Type())
			}
			if tt.isCW != g.IsCW() {
				t.Errorf("expect: isCW %v, got %v, should be %s", tt.isCW, g.IsCW(), tt.note)
			}
			if tt.isCCW != g.IsCCW() {
				t.Errorf("expect: isCCW %v, got %v, should be %s", tt.isCCW, g.IsCCW(), tt.note)
			}
		})
	}
}
