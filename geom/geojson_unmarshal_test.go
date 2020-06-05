package geom_test

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/peterstace/simplefeatures/geom"
	. "github.com/peterstace/simplefeatures/geom"
)

func TestGeoJSONUnmarshalValid(t *testing.T) {
	// Test data from the following query:
	/*
		SELECT
			ST_AsText(ST_GeomFromText(wkt)),
			ST_AsGeoJSON(ST_GeomFromText(wkt)) AS geojson
		FROM (
			VALUES
			('POINT EMPTY'),
			('POINT(1 2)'),
			('POINTZ(1 2 3)'),
			('LINESTRING EMPTY'),
			('LINESTRING(1 2,3 4)'),
			('LINESTRINGZ(1 2 3,4 5 6)'),
			('LINESTRING(1 2,3 4,5 6)'),
			('LINESTRINGZ(1 2 3,3 4 5,5 6 7)'),
			('POLYGON EMPTY'),
			('POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))'),
			('POLYGONZ((0 0 9,4 0 9,0 4 9,0 0 9),(1 1 9,2 1 9,1 2 9,1 1 9))'),
			('MULTIPOINT EMPTY'),
			('MULTIPOINT(1 2)'),
			('MULTIPOINTZ(1 2 3)'),
			('MULTIPOINT(1 2,3 4)'),
			('MULTIPOINTZ(1 2 3,3 4 5)'),
			('MULTILINESTRING EMPTY'),
			('MULTILINESTRINGZ EMPTY'),
			('MULTILINESTRING((0 1,2 3,4 5))'),
			('MULTILINESTRINGZ((0 1 8,2 3 8,4 5 8))'),
			('MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))'),
			('MULTILINESTRINGZ((0 1 9,2 3 9),(4 5 9,6 7 9,8 9 9))'),
			('MULTIPOLYGON EMPTY'),
			('MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,1 1,1 0)))'),
			('MULTIPOLYGONZ(((0 0 9,1 0 9,0 1 9,0 0 9)),((1 0 9,2 0 9,1 1 9,1 0 9)))'),
			('MULTIPOLYGON(EMPTY,((1 0,2 0,1 1,1 0)))'),
			('MULTIPOLYGONZ(EMPTY,((1 0 9,2 0 9,1 1 9,1 0 9)))'),
			('GEOMETRYCOLLECTION EMPTY'),
			('GEOMETRYCOLLECTION(POINT(1 2),POINT(3 4))'),
			('GEOMETRYCOLLECTIONZ(POINTZ(1 2 3),POINTZ(3 4 5))')
		) AS q (wkt);
	*/
	for i, tt := range []struct {
		geojson string
		wkt     string
	}{
		{
			geojson: `{"type":"Point","coordinates":[]}`,
			wkt:     "POINT EMPTY",
		},
		{
			geojson: `{"type":"Point","coordinates":[1,2]}`,
			wkt:     "POINT(1 2)",
		},
		{
			geojson: `{"type":"Point","coordinates":[1,2,3]}`,
			wkt:     "POINT Z (1 2 3)",
		},
		{
			geojson: `{"type":"LineString","coordinates":[]}`,
			wkt:     "LINESTRING EMPTY",
		},
		{
			geojson: `{"type":"LineString","coordinates":[[1,2],[3,4]]}`,
			wkt:     "LINESTRING(1 2,3 4)",
		},
		{
			geojson: `{"type":"LineString","coordinates":[[1,2,3],[4,5,6]]}`,
			wkt:     "LINESTRING Z (1 2 3,4 5 6)",
		},
		{
			geojson: `{"type":"LineString","coordinates":[[1,2],[3,4],[5,6]]}`,
			wkt:     "LINESTRING(1 2,3 4,5 6)",
		},
		{
			geojson: `{"type":"LineString","coordinates":[[1,2,3],[3,4,5],[5,6,7]]}`,
			wkt:     "LINESTRING Z (1 2 3,3 4 5,5 6 7)",
		},
		{
			geojson: `{"type":"Polygon","coordinates":[]}`,
			wkt:     "POLYGON EMPTY",
		},
		{
			geojson: `{"type":"Polygon","coordinates":[[[0,0],[4,0],[0,4],[0,0]],[[1,1],[2,1],[1,2],[1,1]]]}`,
			wkt:     "POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))",
		},
		{
			geojson: `{"type":"Polygon","coordinates":[[[0,0,9],[4,0,9],[0,4,9],[0,0,9]],[[1,1,9],[2,1,9],[1,2,9],[1,1,9]]]}`,
			wkt:     "POLYGON Z ((0 0 9,4 0 9,0 4 9,0 0 9),(1 1 9,2 1 9,1 2 9,1 1 9))",
		},
		{
			geojson: `{"type":"MultiPoint","coordinates":[]}`,
			wkt:     "MULTIPOINT EMPTY",
		},
		{
			geojson: `{"type":"MultiPoint","coordinates":[[1,2]]}`,
			wkt:     "MULTIPOINT(1 2)",
		},
		{
			geojson: `{"type":"MultiPoint","coordinates":[[1,2,3]]}`,
			wkt:     "MULTIPOINT Z (1 2 3)",
		},
		{
			geojson: `{"type":"MultiPoint","coordinates":[[1,2],[3,4]]}`,
			wkt:     "MULTIPOINT(1 2,3 4)",
		},
		{
			geojson: `{"type":"MultiPoint","coordinates":[[1,2,3],[3,4,5]]}`,
			wkt:     "MULTIPOINT Z (1 2 3,3 4 5)",
		},
		{
			geojson: `{"type":"MultiLineString","coordinates":[]}`,
			wkt:     "MULTILINESTRING EMPTY",
		},
		{
			geojson: `{"type":"MultiLineString","coordinates":[[[0,1],[2,3],[4,5]]]}`,
			wkt:     "MULTILINESTRING((0 1,2 3,4 5))",
		},
		{
			geojson: `{"type":"MultiLineString","coordinates":[[[0,1,8],[2,3,8],[4,5,8]]]}`,
			wkt:     "MULTILINESTRING Z ((0 1 8,2 3 8,4 5 8))",
		},
		{
			geojson: `{"type":"MultiLineString","coordinates":[[[0,1],[2,3]],[[4,5],[6,7],[8,9]]]}`,
			wkt:     "MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))",
		},
		{
			geojson: `{"type":"MultiLineString","coordinates":[[[0,1,9],[2,3,9]],[[4,5,9],[6,7,9],[8,9,9]]]}`,
			wkt:     "MULTILINESTRING Z ((0 1 9,2 3 9),(4 5 9,6 7 9,8 9 9))",
		},
		{
			geojson: `{"type":"MultiPolygon","coordinates":[]}`,
			wkt:     "MULTIPOLYGON EMPTY",
		},
		{
			geojson: `{"type":"MultiPolygon","coordinates":[[[[0,0],[1,0],[0,1],[0,0]]],[[[1,0],[2,0],[1,1],[1,0]]]]}`,
			wkt:     "MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,1 1,1 0)))",
		},
		{
			geojson: `{"type":"MultiPolygon","coordinates":[[[[0,0,9],[1,0,9],[0,1,9],[0,0,9]]],[[[1,0,9],[2,0,9],[1,1,9],[1,0,9]]]]}`,
			wkt:     "MULTIPOLYGON Z (((0 0 9,1 0 9,0 1 9,0 0 9)),((1 0 9,2 0 9,1 1 9,1 0 9)))",
		},
		{
			geojson: `{"type":"MultiPolygon","coordinates":[[],[[[1,0],[2,0],[1,1],[1,0]]]]}`,
			wkt:     "MULTIPOLYGON(EMPTY,((1 0,2 0,1 1,1 0)))",
		},
		{
			geojson: `{"type":"MultiPolygon","coordinates":[[],[[[1,0,9],[2,0,9],[1,1,9],[1,0,9]]]]}`,
			wkt:     "MULTIPOLYGON Z (EMPTY,((1 0 9,2 0 9,1 1 9,1 0 9)))",
		},
		{
			geojson: `{"type":"GeometryCollection","geometries":[]}`,
			wkt:     "GEOMETRYCOLLECTION EMPTY",
		},
		{
			geojson: `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2]},{"type":"Point","coordinates":[3,4]}]}`,
			wkt:     "GEOMETRYCOLLECTION(POINT(1 2),POINT(3 4))",
		},
		{
			geojson: `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2,3]},{"type":"Point","coordinates":[3,4,5]}]}`,
			wkt:     "GEOMETRYCOLLECTION Z (POINT Z (1 2 3),POINT Z (3 4 5))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := UnmarshalGeoJSON([]byte(tt.geojson))
			expectNoErr(t, err)
			want := geomFromWKT(t, tt.wkt)
			expectGeomEq(t, got, want)
		})
	}
}

func TestGeoJSONUnmarshalValidAllowAdditionalCoordDimensions(t *testing.T) {
	for i, tt := range []struct {
		geojson string
		wkt     string
	}{
		{
			geojson: `{"type":"Point","coordinates":[1,2,3,4]}`,
			wkt:     "POINT Z (1 2 3)",
		},
		{
			geojson: `{"type":"LineString","coordinates":[[1,2,3,4],[2,3,4,5]]}`,
			wkt:     "LINESTRING Z (1 2 3,2 3 4)",
		},
		{
			geojson: `{"type":"LineString","coordinates":[[1,2,3,4],[2,3,4,5],[3,4,5,6,7]]}`,
			wkt:     "LINESTRING Z (1 2 3,2 3 4,3 4 5)",
		},
		{
			geojson: `{"type":"Polygon","coordinates":[[[0,0,0,0],[0,1,0,0],[1,0,0,0],[0,0,0,0]]]}`,
			wkt:     "POLYGON Z ((0 0 0,0 1 0,1 0 0,0 0 0))",
		},
		{
			geojson: `{"type":"MultiPoint","coordinates":[[1,2,3,4]]}`,
			wkt:     "MULTIPOINT Z (1 2 3)",
		},
		{
			geojson: `{"type":"MultiLineString","coordinates":[[[1,2,3,4],[2,3,4,5]]]}`,
			wkt:     "MULTILINESTRING Z ((1 2 3,2 3 4))",
		},
		{
			geojson: `{"type":"MultiLineString","coordinates":[[[1,2,3,4],[2,3,4,5],[3,4,5,6,7]]]}`,
			wkt:     "MULTILINESTRING Z ((1 2 3,2 3 4,3 4 5))",
		},
		{
			geojson: `{"type":"MultiPolygon","coordinates":[[[[0,0,0,0],[0,1,0,0],[1,0,0,0],[0,0,0,0]]]]}`,
			wkt:     "MULTIPOLYGON Z (((0 0 0,0 1 0,1 0 0,0 0 0)))",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := UnmarshalGeoJSON([]byte(tt.geojson))
			expectNoErr(t, err)
			want := geomFromWKT(t, tt.wkt)
			expectGeomEq(t, got, want)
		})
	}
}

func TestGeoJSONUnmarshalInvalid(t *testing.T) {
	for i, geojson := range []string{
		// GeoJSON cannot represent empty points in MultiPoints. When parsing,
		// we should complain that the dimensionality cannot be zero.
		`{"type":"MultiPoint","coordinates":[[0,1],[]]}`,
		`{"type":"MultiPoint","coordinates":[[],[0,1]]}`,
		`{"type":"MultiPoint","coordinates":[[]]}`,

		// Coordinates (other than the first level) must have either 2 or 3 (or more) parts.
		`{"type":"LineString","coordinates":[[0,1],[]]}`,
		`{"type":"LineString","coordinates":[[0,1],[3]]}`,
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var g geom.Geometry
			err := json.NewDecoder(strings.NewReader(geojson)).Decode(&g)
			if err == nil {
				t.Error("expected error but got nil")
			}
		})
	}
}

func TestGeoJSONUnmarshalDisableAllValidations(t *testing.T) {
	for i, geojson := range []string{
		`{"type":"LineString","coordinates":[[0,0],[0,0]]}`,
		`{"type":"Polygon","coordinates":[[[0,0],[1,1],[1,0],[0,1],[0,0]]]}`,
		`{"type":"Polygon","coordinates":[[[0,0],[0,0]]]}`,
		`{"type":"MultiLineString","coordinates":[[[0,0],[0,0]]]}`,
		`{"type":"MultiPolygon","coordinates":[[[[0,0],[1,1],[1,0],[0,1],[0,0]]]]}`,
		`{"type":"MultiPolygon","coordinates":[[[[0,0],[0,0]]]]}`,
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if _, err := UnmarshalGeoJSON([]byte(geojson)); err == nil {
				t.Fatal("invalid test case -- geometry should be invalid")
			}
			if _, err := UnmarshalGeoJSON([]byte(geojson), DisableAllValidations); err != nil {
				t.Errorf("got error but didn't expect one because validations are disabled")
			}
		})
	}
}
