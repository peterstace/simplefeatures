package geom_test

import (
	"encoding/json"
	"strconv"
	"testing"

	. "github.com/peterstace/simplefeatures/geom"
)

func TestGeoJSONUnmarshalValid(t *testing.T) {
	// Test data from the following query:
	/*
		SELECT
		        wkt,
		        ST_AsText(ST_Force2D(ST_GeomFromText(wkt))) AS flat,
		        ST_AsGeoJSON(ST_GeomFromText(wkt)) AS geojson
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
		geojson string
		wkt     string
	}{
		{
			// POINT EMPTY
			geojson: `{"type":"Point","coordinates":[]}`,
			wkt:     "POINT EMPTY",
		},
		{
			// POINTZ EMPTY
			geojson: `{"type":"Point","coordinates":[]}`,
			wkt:     "POINT EMPTY",
		},
		{
			// POINTM EMPTY
			geojson: `{"type":"Point","coordinates":[]}`,
			wkt:     "POINT EMPTY",
		},
		{
			// POINTZM EMPTY
			geojson: `{"type":"Point","coordinates":[]}`,
			wkt:     "POINT EMPTY",
		},
		{
			// POINT(1 2)
			geojson: `{"type":"Point","coordinates":[1,2]}`,
			wkt:     "POINT(1 2)",
		},
		{
			// POINTZ(1 2 3)
			geojson: `{"type":"Point","coordinates":[1,2,3]}`,
			wkt:     "POINT(1 2)",
		},
		{
			// POINTM(1 2 3)
			geojson: `{"type":"Point","coordinates":[1,2]}`,
			wkt:     "POINT(1 2)",
		},
		{
			// POINTZM(1 2 3 4)
			geojson: `{"type":"Point","coordinates":[1,2,3]}`,
			wkt:     "POINT(1 2)",
		},
		{
			// LINESTRING EMPTY
			geojson: `{"type":"LineString","coordinates":[]}`,
			wkt:     "LINESTRING EMPTY",
		},
		{
			// LINESTRINGZ EMPTY
			geojson: `{"type":"LineString","coordinates":[]}`,
			wkt:     "LINESTRING EMPTY",
		},
		{
			// LINESTRINGM EMPTY
			geojson: `{"type":"LineString","coordinates":[]}`,
			wkt:     "LINESTRING EMPTY",
		},
		{
			// LINESTRINGZM EMPTY
			geojson: `{"type":"LineString","coordinates":[]}`,
			wkt:     "LINESTRING EMPTY",
		},
		{
			// LINESTRING(1 2,3 4)
			geojson: `{"type":"LineString","coordinates":[[1,2],[3,4]]}`,
			wkt:     "LINESTRING(1 2,3 4)",
		},
		{
			// LINESTRINGZ(1 2 3,4 5 6)
			geojson: `{"type":"LineString","coordinates":[[1,2,3],[4,5,6]]}`,
			wkt:     "LINESTRING(1 2,4 5)",
		},
		{
			// LINESTRINGM(1 2 3,4 5 6)
			geojson: `{"type":"LineString","coordinates":[[1,2],[4,5]]}`,
			wkt:     "LINESTRING(1 2,4 5)",
		},
		{
			// LINESTRINGZM(1 2 3 4,5 6 7 8)
			geojson: `{"type":"LineString","coordinates":[[1,2,3],[5,6,7]]}`,
			wkt:     "LINESTRING(1 2,5 6)",
		},
		{
			// LINESTRING(1 2,3 4,5 6)
			geojson: `{"type":"LineString","coordinates":[[1,2],[3,4],[5,6]]}`,
			wkt:     "LINESTRING(1 2,3 4,5 6)",
		},
		{
			// LINESTRINGZ(1 2 3,3 4 5,5 6 7)
			geojson: `{"type":"LineString","coordinates":[[1,2,3],[3,4,5],[5,6,7]]}`,
			wkt:     "LINESTRING(1 2,3 4,5 6)",
		},
		{
			// LINESTRINGM(1 2 3,3 4 5,5 6 7)
			geojson: `{"type":"LineString","coordinates":[[1,2],[3,4],[5,6]]}`,
			wkt:     "LINESTRING(1 2,3 4,5 6)",
		},
		{
			// LINESTRINGZM(1 2 3 4,3 4 5 6,5 6 7 8)
			geojson: `{"type":"LineString","coordinates":[[1,2,3],[3,4,5],[5,6,7]]}`,
			wkt:     "LINESTRING(1 2,3 4,5 6)",
		},
		{
			// POLYGON EMPTY
			geojson: `{"type":"Polygon","coordinates":[]}`,
			wkt:     "POLYGON EMPTY",
		},
		{
			// POLYGONZ EMPTY
			geojson: `{"type":"Polygon","coordinates":[]}`,
			wkt:     "POLYGON EMPTY",
		},
		{
			// POLYGONM EMPTY
			geojson: `{"type":"Polygon","coordinates":[]}`,
			wkt:     "POLYGON EMPTY",
		},
		{
			// POLYGONZM EMPTY
			geojson: `{"type":"Polygon","coordinates":[]}`,
			wkt:     "POLYGON EMPTY",
		},
		{
			// POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))
			geojson: `{"type":"Polygon","coordinates":[[[0,0],[4,0],[0,4],[0,0]],[[1,1],[2,1],[1,2],[1,1]]]}`,
			wkt:     "POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))",
		},
		{
			// POLYGONZ((0 0 9,4 0 9,0 4 9,0 0 9),(1 1 9,2 1 9,1 2 9,1 1 9))
			geojson: `{"type":"Polygon","coordinates":[[[0,0,9],[4,0,9],[0,4,9],[0,0,9]],[[1,1,9],[2,1,9],[1,2,9],[1,1,9]]]}`,
			wkt:     "POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))",
		},
		{
			// POLYGONM((0 0 9,4 0 9,0 4 9,0 0 9),(1 1 9,2 1 9,1 2 9,1 1 9))
			geojson: `{"type":"Polygon","coordinates":[[[0,0],[4,0],[0,4],[0,0]],[[1,1],[2,1],[1,2],[1,1]]]}`,
			wkt:     "POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))",
		},
		{
			// POLYGONZM((0 0 9 9,4 0 9 9,0 4 9 9,0 0 9 9),(1 1 9 9,2 1 9 9,1 2 9 9,1 1 9 9))
			geojson: `{"type":"Polygon","coordinates":[[[0,0,9],[4,0,9],[0,4,9],[0,0,9]],[[1,1,9],[2,1,9],[1,2,9],[1,1,9]]]}`,
			wkt:     "POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))",
		},
		{
			// MULTIPOINT EMPTY
			geojson: `{"type":"MultiPoint","coordinates":[]}`,
			wkt:     "MULTIPOINT EMPTY",
		},
		{
			// MULTIPOINTZ EMPTY
			geojson: `{"type":"MultiPoint","coordinates":[]}`,
			wkt:     "MULTIPOINT EMPTY",
		},
		{
			// MULTIPOINTM EMPTY
			geojson: `{"type":"MultiPoint","coordinates":[]}`,
			wkt:     "MULTIPOINT EMPTY",
		},
		{
			// MULTIPOINTZM EMPTY
			geojson: `{"type":"MultiPoint","coordinates":[]}`,
			wkt:     "MULTIPOINT EMPTY",
		},
		{
			// MULTIPOINT(1 2)
			geojson: `{"type":"MultiPoint","coordinates":[[1,2]]}`,
			wkt:     "MULTIPOINT(1 2)",
		},
		{
			// MULTIPOINTZ(1 2 3)
			geojson: `{"type":"MultiPoint","coordinates":[[1,2,3]]}`,
			wkt:     "MULTIPOINT(1 2)",
		},
		{
			// MULTIPOINTM(1 2 3)
			geojson: `{"type":"MultiPoint","coordinates":[[1,2]]}`,
			wkt:     "MULTIPOINT(1 2)",
		},
		{
			// MULTIPOINTZM(1 2 3 4)
			geojson: `{"type":"MultiPoint","coordinates":[[1,2,3]]}`,
			wkt:     "MULTIPOINT(1 2)",
		},
		{
			// MULTIPOINT(1 2,3 4)
			geojson: `{"type":"MultiPoint","coordinates":[[1,2],[3,4]]}`,
			wkt:     "MULTIPOINT(1 2,3 4)",
		},
		{
			// MULTIPOINTZ(1 2 3,3 4 5)
			geojson: `{"type":"MultiPoint","coordinates":[[1,2,3],[3,4,5]]}`,
			wkt:     "MULTIPOINT(1 2,3 4)",
		},
		{
			// MULTIPOINTM(1 2 3,3 4 5)
			geojson: `{"type":"MultiPoint","coordinates":[[1,2],[3,4]]}`,
			wkt:     "MULTIPOINT(1 2,3 4)",
		},
		{
			// MULTIPOINTZM(1 2 3 4,3 4 5 6)
			geojson: `{"type":"MultiPoint","coordinates":[[1,2,3],[3,4,5]]}`,
			wkt:     "MULTIPOINT(1 2,3 4)",
		},
		{
			// MULTILINESTRING EMPTY
			geojson: `{"type":"MultiLineString","coordinates":[]}`,
			wkt:     "MULTILINESTRING EMPTY",
		},
		{
			// MULTILINESTRINGZ EMPTY
			geojson: `{"type":"MultiLineString","coordinates":[]}`,
			wkt:     "MULTILINESTRING EMPTY",
		},
		{
			// MULTILINESTRINGM EMPTY
			geojson: `{"type":"MultiLineString","coordinates":[]}`,
			wkt:     "MULTILINESTRING EMPTY",
		},
		{
			// MULTILINESTRINGZM EMPTY
			geojson: `{"type":"MultiLineString","coordinates":[]}`,
			wkt:     "MULTILINESTRING EMPTY",
		},
		{
			// MULTILINESTRING((0 1,2 3,4 5))
			geojson: `{"type":"MultiLineString","coordinates":[[[0,1],[2,3],[4,5]]]}`,
			wkt:     "MULTILINESTRING((0 1,2 3,4 5))",
		},
		{
			// MULTILINESTRINGZ((0 1 8,2 3 8,4 5 8))
			geojson: `{"type":"MultiLineString","coordinates":[[[0,1,8],[2,3,8],[4,5,8]]]}`,
			wkt:     "MULTILINESTRING((0 1,2 3,4 5))",
		},
		{
			// MULTILINESTRINGM((0 1 8,2 3 8,4 5 8))
			geojson: `{"type":"MultiLineString","coordinates":[[[0,1],[2,3],[4,5]]]}`,
			wkt:     "MULTILINESTRING((0 1,2 3,4 5))",
		},
		{
			// MULTILINESTRINGZM((0 1 8 9,2 3 8 9,4 5 8 9))
			geojson: `{"type":"MultiLineString","coordinates":[[[0,1,8],[2,3,8],[4,5,8]]]}`,
			wkt:     "MULTILINESTRING((0 1,2 3,4 5))",
		},
		{
			// MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))
			geojson: `{"type":"MultiLineString","coordinates":[[[0,1],[2,3]],[[4,5],[6,7],[8,9]]]}`,
			wkt:     "MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))",
		},
		{
			// MULTILINESTRINGZ((0 1 9,2 3 9),(4 5 9,6 7 9,8 9 9))
			geojson: `{"type":"MultiLineString","coordinates":[[[0,1,9],[2,3,9]],[[4,5,9],[6,7,9],[8,9,9]]]}`,
			wkt:     "MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))",
		},
		{
			// MULTILINESTRINGM((0 1 9,2 3 9),(4 5 9,6 7 9,8 9 9))
			geojson: `{"type":"MultiLineString","coordinates":[[[0,1],[2,3]],[[4,5],[6,7],[8,9]]]}`,
			wkt:     "MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))",
		},
		{
			// MULTILINESTRINGZM((0 1 9 9,2 3 9 9),(4 5 9 9,6 7 9 9,8 9 9 9))
			geojson: `{"type":"MultiLineString","coordinates":[[[0,1,9],[2,3,9]],[[4,5,9],[6,7,9],[8,9,9]]]}`,
			wkt:     "MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))",
		},
		{
			// MULTIPOLYGON EMPTY
			geojson: `{"type":"MultiPolygon","coordinates":[]}`,
			wkt:     "MULTIPOLYGON EMPTY",
		},
		{
			// MULTIPOLYGONZ EMPTY
			geojson: `{"type":"MultiPolygon","coordinates":[]}`,
			wkt:     "MULTIPOLYGON EMPTY",
		},
		{
			// MULTIPOLYGONM EMPTY
			geojson: `{"type":"MultiPolygon","coordinates":[]}`,
			wkt:     "MULTIPOLYGON EMPTY",
		},
		{
			// MULTIPOLYGONZM EMPTY
			geojson: `{"type":"MultiPolygon","coordinates":[]}`,
			wkt:     "MULTIPOLYGON EMPTY",
		},
		{
			// MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,1 1,1 0)))
			geojson: `{"type":"MultiPolygon","coordinates":[[[[0,0],[1,0],[0,1],[0,0]]],[[[1,0],[2,0],[1,1],[1,0]]]]}`,
			wkt:     "MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,1 1,1 0)))",
		},
		{
			// MULTIPOLYGONZ(((0 0 9,1 0 9,0 1 9,0 0 9)),((1 0 9,2 0 9,1 1 9,1 0 9)))
			geojson: `{"type":"MultiPolygon","coordinates":[[[[0,0,9],[1,0,9],[0,1,9],[0,0,9]]],[[[1,0,9],[2,0,9],[1,1,9],[1,0,9]]]]}`,
			wkt:     "MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,1 1,1 0)))",
		},
		{
			// MULTIPOLYGONM(((0 0 9,1 0 9,0 1 9,0 0 9)),((1 0 9,2 0 9,1 1 9,1 0 9)))
			geojson: `{"type":"MultiPolygon","coordinates":[[[[0,0],[1,0],[0,1],[0,0]]],[[[1,0],[2,0],[1,1],[1,0]]]]}`,
			wkt:     "MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,1 1,1 0)))",
		},
		{
			// MULTIPOLYGONZM(((0 0 8 9,1 0 8 9,0 1 8 9,0 0 8 9)),((1 0 8 9,2 0 8 9,1 1 8 9,1 0 8 9)))
			geojson: `{"type":"MultiPolygon","coordinates":[[[[0,0,8],[1,0,8],[0,1,8],[0,0,8]]],[[[1,0,8],[2,0,8],[1,1,8],[1,0,8]]]]}`,
			wkt:     "MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,1 1,1 0)))",
		},
		{
			// GEOMETRYCOLLECTION EMPTY
			geojson: `{"type":"GeometryCollection","geometries":[]}`,
			wkt:     "GEOMETRYCOLLECTION EMPTY",
		},
		{
			// GEOMETRYCOLLECTIONZ EMPTY
			geojson: `{"type":"GeometryCollection","geometries":[]}`,
			wkt:     "GEOMETRYCOLLECTION EMPTY",
		},
		{
			// GEOMETRYCOLLECTIONM EMPTY
			geojson: `{"type":"GeometryCollection","geometries":[]}`,
			wkt:     "GEOMETRYCOLLECTION EMPTY",
		},
		{
			// GEOMETRYCOLLECTIONZM EMPTY
			geojson: `{"type":"GeometryCollection","geometries":[]}`,
			wkt:     "GEOMETRYCOLLECTION EMPTY",
		},
		{
			// GEOMETRYCOLLECTION(POINT(1 2),POINT(3 4))
			geojson: `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2]},{"type":"Point","coordinates":[3,4]}]}`,
			wkt:     "GEOMETRYCOLLECTION(POINT(1 2),POINT(3 4))",
		},
		{
			// GEOMETRYCOLLECTIONZ(POINTZ(1 2 3),POINTZ(3 4 5))
			geojson: `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2,3]},{"type":"Point","coordinates":[3,4,5]}]}`,
			wkt:     "GEOMETRYCOLLECTION(POINT(1 2),POINT(3 4))",
		},
		{
			// GEOMETRYCOLLECTIONM(POINTM(1 2 3),POINTM(3 4 5))
			geojson: `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2]},{"type":"Point","coordinates":[3,4]}]}`,
			wkt:     "GEOMETRYCOLLECTION(POINT(1 2),POINT(3 4))",
		},
		{
			// GEOMETRYCOLLECTIONZM(POINTZM(1 2 3 4),POINTZM(3 4 5 5))
			geojson: `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2,3]},{"type":"Point","coordinates":[3,4,5]}]}`,
			wkt:     "GEOMETRYCOLLECTION(POINT(1 2),POINT(3 4))",
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

func TestGeoJSONMarshal(t *testing.T) {
	// Test cases are from:
	/*
	   SELECT wkt, ST_AsGeoJSON(ST_GeomFromText(wkt)) AS geojson
	   FROM (
	           VALUES
	           ('POINT EMPTY'),
	           ('POINT(1 2)'),
	           ('LINESTRING EMPTY'),
	           ('LINESTRING(1 2,3 4)'),
	           ('LINESTRING(1 2,3 4,5 6)'),
	           ('POLYGON EMPTY'),
	           ('POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))'),
	           ('MULTIPOINT EMPTY'),
	           ('MULTIPOINT(1 2)'),
	           ('MULTIPOINT(1 2,3 4)'),
	           ('MULTILINESTRING EMPTY'),
	           ('MULTILINESTRING((0 1,2 3,4 5))'),
	           ('MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))'),
	           ('MULTIPOLYGON EMPTY'),
	           ('MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,1 1,1 0)))'),
	           ('GEOMETRYCOLLECTION EMPTY'),
	           ('GEOMETRYCOLLECTION(POINT(1 2),POINT(3 4))')
	   ) AS q (wkt);
	*/
	for _, tt := range []struct {
		wkt  string
		want string
	}{
		{
			wkt:  "POINT EMPTY",
			want: `{"type":"Point","coordinates":[]}`,
		},
		{
			wkt:  "POINT(1 2)",
			want: `{"type":"Point","coordinates":[1,2]}`,
		},
		{
			wkt:  "LINESTRING EMPTY",
			want: `{"type":"LineString","coordinates":[]}`,
		},
		{
			wkt:  "LINESTRING(1 2,3 4)",
			want: `{"type":"LineString","coordinates":[[1,2],[3,4]]}`,
		},
		{
			wkt:  "LINESTRING(1 2,3 4,5 6)",
			want: `{"type":"LineString","coordinates":[[1,2],[3,4],[5,6]]}`,
		},
		{
			wkt:  "POLYGON EMPTY",
			want: `{"type":"Polygon","coordinates":[]}`,
		},
		{
			wkt:  "POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))",
			want: `{"type":"Polygon","coordinates":[[[0,0],[4,0],[0,4],[0,0]],[[1,1],[2,1],[1,2],[1,1]]]}`,
		},
		{
			wkt:  "MULTIPOINT EMPTY",
			want: `{"type":"MultiPoint","coordinates":[]}`,
		},
		{
			wkt:  "MULTIPOINT(1 2)",
			want: `{"type":"MultiPoint","coordinates":[[1,2]]}`,
		},
		{
			wkt:  "MULTIPOINT(1 2,3 4)",
			want: `{"type":"MultiPoint","coordinates":[[1,2],[3,4]]}`,
		},
		{
			wkt:  "MULTILINESTRING EMPTY",
			want: `{"type":"MultiLineString","coordinates":[]}`,
		},
		{
			wkt:  "MULTILINESTRING((0 1,2 3,4 5))",
			want: `{"type":"MultiLineString","coordinates":[[[0,1],[2,3],[4,5]]]}`,
		},
		{
			wkt:  "MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))",
			want: `{"type":"MultiLineString","coordinates":[[[0,1],[2,3]],[[4,5],[6,7],[8,9]]]}`,
		},
		{
			wkt:  "MULTIPOLYGON EMPTY",
			want: `{"type":"MultiPolygon","coordinates":[]}`,
		},
		{
			wkt:  "MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,1 1,1 0)))",
			want: `{"type":"MultiPolygon","coordinates":[[[[0,0],[1,0],[0,1],[0,0]]],[[[1,0],[2,0],[1,1],[1,0]]]]}`,
		},
		{
			wkt:  "GEOMETRYCOLLECTION EMPTY",
			want: `{"type":"GeometryCollection","geometries":[]}`,
		},
		{
			wkt:  "GEOMETRYCOLLECTION(POINT(1 2),POINT(3 4))",
			want: `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2]},{"type":"Point","coordinates":[3,4]}]}`,
		},
	} {
		t.Run(tt.wkt, func(t *testing.T) {
			geom := geomFromWKT(t, tt.wkt)
			gotJSON, err := json.Marshal(geom)
			expectNoErr(t, err)
			if string(gotJSON) != tt.want {
				t.Error("json doesn't match")
				t.Logf("got:  %v", string(gotJSON))
				t.Logf("want: %v", tt.want)
			}
		})
	}
}

func TestGeoJSONMarshalAnyGeometryPopulated(t *testing.T) {
	g := geomFromWKT(t, "POINT(1 2)")
	got, err := json.Marshal(g)
	expectNoErr(t, err)
	const want = `{"type":"Point","coordinates":[1,2]}`
	expectStringEq(t, string(got), want)
}
