package geom_test

import (
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

func TestGeoJSONMarshal(t *testing.T) {
	// Test cases are from (some minor tweaks around MultiPoints were needed):
	/*
		SELECT wkt, ST_AsGeoJSON(ST_GeomFromText(wkt)) AS geojson
		FROM (
				VALUES
				('POINT EMPTY'),
				('POINT Z EMPTY'),
				('POINT M EMPTY'),
				('POINT ZM EMPTY'),
				('POINT(1 2)'),
				('POINT Z (1 2 3)'),
				('POINT M (1 2 3)'),
				('POINT ZM (1 2 3 4)'),
				('LINESTRING EMPTY'),
				('LINESTRING(1 2,3 4)'),
				('LINESTRING(1 2,3 4,5 6)'),
				('LINESTRING Z EMPTY'),
				('LINESTRING Z (1 2 3,3 4 5)'),
				('LINESTRING Z (1 2 3,3 4 5,5 6 7)'),
				('LINESTRING M EMPTY'),
				('LINESTRING M (1 2 3,3 4 5)'),
				('LINESTRING M (1 2 3,3 4 5,5 6 7)'),
				('LINESTRING ZM EMPTY'),
				('LINESTRING ZM (1 2 3 4,3 4 5 6)'),
				('LINESTRING ZM (1 2 3 4,3 4 5 6,5 6 7 8)'),
				('POLYGON EMPTY'),
				('POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))'),
				('POLYGON Z EMPTY'),
				('POLYGON Z ((0 0 0,4 0 1,0 4 1,0 0 1),(1 1 2,2 1 3,1 2 4,1 1 5))'),
				('POLYGON M EMPTY'),
				('POLYGON M ((0 0 0,4 0 1,0 4 1,0 0 1),(1 1 2,2 1 3,1 2 4,1 1 5))'),
				('POLYGON ZM EMPTY'),
				('POLYGON ZM ((0 0 0 8,4 0 1 3,0 4 1 7,0 0 1 9),(1 1 2 3,2 1 3 7,1 2 4 8,1 1 5 4))'),
				('MULTIPOINT EMPTY'),
				('MULTIPOINT(1 2)'),
				('MULTIPOINT(1 2,3 4)'),
				('MULTIPOINT(1 2,EMPTY)'),
				('MULTIPOINT Z EMPTY'),
				('MULTIPOINT Z (1 2 3)'),
				('MULTIPOINT Z (1 2 3,3 4 5)'),
				('MULTIPOINT Z (1 2 3,EMPTY)'),
				('MULTIPOINT M EMPTY'),
				('MULTIPOINT M (1 2 3)'),
				('MULTIPOINT M (1 2 3,3 4 5)'),
				('MULTIPOINT M (1 2 3,EMPTY)'),
				('MULTIPOINT ZM EMPTY'),
				('MULTIPOINT ZM (1 2 3 4)'),
				('MULTIPOINT ZM (1 2 3 4,3 4 5 6)'),
				('MULTIPOINT ZM (1 2 3 4,EMPTY)'),
				('MULTILINESTRING EMPTY'),
				('MULTILINESTRING(EMPTY)'),
				('MULTILINESTRING((0 1,2 3,4 5))'),
				('MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))'),
				('MULTILINESTRING((0 1,2 3),EMPTY)'),
				('MULTILINESTRING Z EMPTY'),
				('MULTILINESTRING Z (EMPTY)'),
				('MULTILINESTRING Z ((0 1 6,2 3 8,4 5 6))'),
				('MULTILINESTRING Z ((0 1 2,2 3 4),(4 5 9,6 7 4,8 9 7))'),
				('MULTILINESTRING Z ((0 1 1,2 3 4),EMPTY)'),
				('MULTILINESTRING M EMPTY'),
				('MULTILINESTRING M (EMPTY)'),
				('MULTILINESTRING M ((0 1 6,2 3 8,4 5 6))'),
				('MULTILINESTRING M ((0 1 2,2 3 4),(4 5 9,6 7 4,8 9 7))'),
				('MULTILINESTRING M ((0 1 1,2 3 4),EMPTY)'),
				('MULTILINESTRING ZM EMPTY'),
				('MULTILINESTRING ZM (EMPTY)'),
				('MULTILINESTRING ZM ((0 1 6 8,2 3 8 5,4 5 6 8))'),
				('MULTILINESTRING ZM ((0 1 2 3,2 3 4 9),(4 5 9 8,6 7 4 6,8 9 7 9))'),
				('MULTILINESTRING ZM ((0 1 1 2,2 3 4 8),EMPTY)'),
				('MULTIPOLYGON EMPTY'),
				('MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,1 1,1 0)))'),
				('MULTIPOLYGON(((0 0,1 0,0 1,0 0)),EMPTY)'),
				('MULTIPOLYGON Z EMPTY'),
				('MULTIPOLYGON Z (((0 0 7,1 0 8,0 1 2,0 0 9)),((1 0 2,2 0 1,1 1 5,1 0 8)))'),
				('MULTIPOLYGON Z (((0 0 4,1 0 2,0 1 8,0 0 9)),EMPTY)'),
				('MULTIPOLYGON M EMPTY'),
				('MULTIPOLYGON M (((0 0 7,1 0 8,0 1 2,0 0 9)),((1 0 2,2 0 1,1 1 5,1 0 8)))'),
				('MULTIPOLYGON M (((0 0 4,1 0 2,0 1 8,0 0 9)),EMPTY)'),
				('MULTIPOLYGON ZM EMPTY'),
				('MULTIPOLYGON ZM (((0 0 7 8,1 0 8 6,0 1 2 3,0 0 9 9)),((1 0 2 3,2 0 1 1,1 1 5 7,1 0 8 9)))'),
				('MULTIPOLYGON ZM (((0 0 4 2,1 0 2 1,0 1 8 8,0 0 9 5)),EMPTY)'),
				('GEOMETRYCOLLECTION EMPTY'),
				('GEOMETRYCOLLECTION(POINT(1 2),POINT(3 4))'),
				('GEOMETRYCOLLECTION Z EMPTY'),
				('GEOMETRYCOLLECTION Z (POINT Z (1 2 3),POINT Z (3 4 5))'),
				('GEOMETRYCOLLECTION M EMPTY'),
				('GEOMETRYCOLLECTION M (POINT M (1 2 3),POINT M (3 4 5))'),
				('GEOMETRYCOLLECTION ZM EMPTY'),
				('GEOMETRYCOLLECTION ZM (POINT ZM (1 2 3 4),POINT ZM (3 4 5 6))')
		) AS q (wkt);
	*/
	for _, tt := range []struct {
		wkt  string
		want string
	}{
		{
			wkt:  `POINT EMPTY`,
			want: `{"type":"Point","coordinates":[]}`,
		},
		{
			wkt:  `POINT Z EMPTY`,
			want: `{"type":"Point","coordinates":[]}`,
		},
		{
			wkt:  `POINT M EMPTY`,
			want: `{"type":"Point","coordinates":[]}`,
		},
		{
			wkt:  `POINT ZM EMPTY`,
			want: `{"type":"Point","coordinates":[]}`,
		},
		{
			wkt:  `POINT(1 2)`,
			want: `{"type":"Point","coordinates":[1,2]}`,
		},
		{
			wkt:  `POINT Z (1 2 3)`,
			want: `{"type":"Point","coordinates":[1,2,3]}`,
		},
		{
			wkt:  `POINT M (1 2 3)`,
			want: `{"type":"Point","coordinates":[1,2]}`,
		},
		{
			wkt:  `POINT ZM (1 2 3 4)`,
			want: `{"type":"Point","coordinates":[1,2,3]}`,
		},
		{
			wkt:  `LINESTRING EMPTY`,
			want: `{"type":"LineString","coordinates":[]}`,
		},
		{
			wkt:  `LINESTRING(1 2,3 4)`,
			want: `{"type":"LineString","coordinates":[[1,2],[3,4]]}`,
		},
		{
			wkt:  `LINESTRING(1 2,3 4,5 6)`,
			want: `{"type":"LineString","coordinates":[[1,2],[3,4],[5,6]]}`,
		},
		{
			wkt:  `LINESTRING Z EMPTY`,
			want: `{"type":"LineString","coordinates":[]}`,
		},
		{
			wkt:  `LINESTRING Z (1 2 3,3 4 5)`,
			want: `{"type":"LineString","coordinates":[[1,2,3],[3,4,5]]}`,
		},
		{
			wkt:  `LINESTRING Z (1 2 3,3 4 5,5 6 7)`,
			want: `{"type":"LineString","coordinates":[[1,2,3],[3,4,5],[5,6,7]]}`,
		},
		{
			wkt:  `LINESTRING M EMPTY`,
			want: `{"type":"LineString","coordinates":[]}`,
		},
		{
			wkt:  `LINESTRING M (1 2 3,3 4 5)`,
			want: `{"type":"LineString","coordinates":[[1,2],[3,4]]}`,
		},
		{
			wkt:  `LINESTRING M (1 2 3,3 4 5,5 6 7)`,
			want: `{"type":"LineString","coordinates":[[1,2],[3,4],[5,6]]}`,
		},
		{
			wkt:  `LINESTRING ZM EMPTY`,
			want: `{"type":"LineString","coordinates":[]}`,
		},
		{
			wkt:  `LINESTRING ZM (1 2 3 4,3 4 5 6)`,
			want: `{"type":"LineString","coordinates":[[1,2,3],[3,4,5]]}`,
		},
		{
			wkt:  `LINESTRING ZM (1 2 3 4,3 4 5 6,5 6 7 8)`,
			want: `{"type":"LineString","coordinates":[[1,2,3],[3,4,5],[5,6,7]]}`,
		},

		// NOTE: Polygons oriented CCW in the output.
		{
			wkt:  `POLYGON EMPTY`,
			want: `{"type":"Polygon","coordinates":[]}`,
		},
		{
			wkt:  `POLYGON((0 0,4 0,0 4,0 0),(1 1,2 1,1 2,1 1))`,
			want: `{"type":"Polygon","coordinates":[[[0,0],[4,0],[0,4],[0,0]],[[1,1],[1,2],[2,1],[1,1]]]}`,
		},
		{
			wkt:  `POLYGON Z EMPTY`,
			want: `{"type":"Polygon","coordinates":[]}`,
		},
		{
			wkt:  `POLYGON Z ((0 0 0,4 0 1,0 4 1,0 0 1),(1 1 2,2 1 3,1 2 4,1 1 5))`,
			want: `{"type":"Polygon","coordinates":[[[0,0,0],[4,0,1],[0,4,1],[0,0,1]],[[1,1,5],[1,2,4],[2,1,3],[1,1,2]]]}`,
		},
		{
			wkt:  `POLYGON M EMPTY`,
			want: `{"type":"Polygon","coordinates":[]}`,
		},
		{
			wkt:  `POLYGON M ((0 0 0,4 0 1,0 4 1,0 0 1),(1 1 2,2 1 3,1 2 4,1 1 5))`,
			want: `{"type":"Polygon","coordinates":[[[0,0],[4,0],[0,4],[0,0]],[[1,1],[1,2],[2,1],[1,1]]]}`,
		},
		{
			wkt:  `POLYGON ZM EMPTY`,
			want: `{"type":"Polygon","coordinates":[]}`,
		},
		{
			wkt:  `POLYGON ZM ((0 0 0 8,4 0 1 3,0 4 1 7,0 0 1 9),(1 1 2 3,2 1 3 7,1 2 4 8,1 1 5 4))`,
			want: `{"type":"Polygon","coordinates":[[[0,0,0],[4,0,1],[0,4,1],[0,0,1]],[[1,1,5],[1,2,4],[2,1,3],[1,1,2]]]}`,
		},
		{
			wkt:  `MULTIPOINT EMPTY`,
			want: `{"type":"MultiPoint","coordinates":[]}`,
		},
		{
			wkt:  `MULTIPOINT(1 2)`,
			want: `{"type":"MultiPoint","coordinates":[[1,2]]}`,
		},
		{
			wkt:  `MULTIPOINT(1 2,3 4)`,
			want: `{"type":"MultiPoint","coordinates":[[1,2],[3,4]]}`,
		},
		{
			wkt:  `MULTIPOINT(1 2,EMPTY)`,
			want: `{"type":"MultiPoint","coordinates":[[1,2]]}`,
		},
		{
			wkt:  `MULTIPOINT Z EMPTY`,
			want: `{"type":"MultiPoint","coordinates":[]}`,
		},
		{
			wkt:  `MULTIPOINT Z (1 2 3)`,
			want: `{"type":"MultiPoint","coordinates":[[1,2,3]]}`,
		},
		{
			wkt:  `MULTIPOINT Z (1 2 3,3 4 5)`,
			want: `{"type":"MultiPoint","coordinates":[[1,2,3],[3,4,5]]}`,
		},
		{
			wkt:  `MULTIPOINT Z (1 2 3,EMPTY)`,
			want: `{"type":"MultiPoint","coordinates":[[1,2,3]]}`,
		},
		{
			wkt:  `MULTIPOINT M EMPTY`,
			want: `{"type":"MultiPoint","coordinates":[]}`,
		},
		{
			wkt:  `MULTIPOINT M (1 2 3)`,
			want: `{"type":"MultiPoint","coordinates":[[1,2]]}`,
		},
		{
			wkt:  `MULTIPOINT M (1 2 3,3 4 5)`,
			want: `{"type":"MultiPoint","coordinates":[[1,2],[3,4]]}`,
		},
		{
			wkt:  `MULTIPOINT M (1 2 3,EMPTY)`,
			want: `{"type":"MultiPoint","coordinates":[[1,2]]}`,
		},
		{
			wkt:  `MULTIPOINT ZM EMPTY`,
			want: `{"type":"MultiPoint","coordinates":[]}`,
		},
		{
			wkt:  `MULTIPOINT ZM (1 2 3 4)`,
			want: `{"type":"MultiPoint","coordinates":[[1,2,3]]}`,
		},
		{
			wkt:  `MULTIPOINT ZM (1 2 3 4,3 4 5 6)`,
			want: `{"type":"MultiPoint","coordinates":[[1,2,3],[3,4,5]]}`,
		},
		{
			wkt:  `MULTIPOINT ZM (1 2 3 4,EMPTY)`,
			want: `{"type":"MultiPoint","coordinates":[[1,2,3]]}`,
		},
		{
			wkt:  `MULTILINESTRING EMPTY`,
			want: `{"type":"MultiLineString","coordinates":[]}`,
		},
		{
			wkt:  `MULTILINESTRING(EMPTY)`,
			want: `{"type":"MultiLineString","coordinates":[[]]}`,
		},
		{
			wkt:  `MULTILINESTRING((0 1,2 3,4 5))`,
			want: `{"type":"MultiLineString","coordinates":[[[0,1],[2,3],[4,5]]]}`,
		},
		{
			wkt:  `MULTILINESTRING((0 1,2 3),(4 5,6 7,8 9))`,
			want: `{"type":"MultiLineString","coordinates":[[[0,1],[2,3]],[[4,5],[6,7],[8,9]]]}`,
		},
		{
			wkt:  `MULTILINESTRING((0 1,2 3),EMPTY)`,
			want: `{"type":"MultiLineString","coordinates":[[[0,1],[2,3]],[]]}`,
		},
		{
			wkt:  `MULTILINESTRING Z EMPTY`,
			want: `{"type":"MultiLineString","coordinates":[]}`,
		},
		{
			wkt:  `MULTILINESTRING Z (EMPTY)`,
			want: `{"type":"MultiLineString","coordinates":[[]]}`,
		},
		{
			wkt:  `MULTILINESTRING Z ((0 1 6,2 3 8,4 5 6))`,
			want: `{"type":"MultiLineString","coordinates":[[[0,1,6],[2,3,8],[4,5,6]]]}`,
		},
		{
			wkt:  `MULTILINESTRING Z ((0 1 2,2 3 4),(4 5 9,6 7 4,8 9 7))`,
			want: `{"type":"MultiLineString","coordinates":[[[0,1,2],[2,3,4]],[[4,5,9],[6,7,4],[8,9,7]]]}`,
		},
		{
			wkt:  `MULTILINESTRING Z ((0 1 1,2 3 4),EMPTY)`,
			want: `{"type":"MultiLineString","coordinates":[[[0,1,1],[2,3,4]],[]]}`,
		},
		{
			wkt:  `MULTILINESTRING M EMPTY`,
			want: `{"type":"MultiLineString","coordinates":[]}`,
		},
		{
			wkt:  `MULTILINESTRING M (EMPTY)`,
			want: `{"type":"MultiLineString","coordinates":[[]]}`,
		},
		{
			wkt:  `MULTILINESTRING M ((0 1 6,2 3 8,4 5 6))`,
			want: `{"type":"MultiLineString","coordinates":[[[0,1],[2,3],[4,5]]]}`,
		},
		{
			wkt:  `MULTILINESTRING M ((0 1 2,2 3 4),(4 5 9,6 7 4,8 9 7))`,
			want: `{"type":"MultiLineString","coordinates":[[[0,1],[2,3]],[[4,5],[6,7],[8,9]]]}`,
		},
		{
			wkt:  `MULTILINESTRING M ((0 1 1,2 3 4),EMPTY)`,
			want: `{"type":"MultiLineString","coordinates":[[[0,1],[2,3]],[]]}`,
		},
		{
			wkt:  `MULTILINESTRING ZM EMPTY`,
			want: `{"type":"MultiLineString","coordinates":[]}`,
		},
		{
			wkt:  `MULTILINESTRING ZM (EMPTY)`,
			want: `{"type":"MultiLineString","coordinates":[[]]}`,
		},
		{
			wkt:  `MULTILINESTRING ZM ((0 1 6 8,2 3 8 5,4 5 6 8))`,
			want: `{"type":"MultiLineString","coordinates":[[[0,1,6],[2,3,8],[4,5,6]]]}`,
		},
		{
			wkt:  `MULTILINESTRING ZM ((0 1 2 3,2 3 4 9),(4 5 9 8,6 7 4 6,8 9 7 9))`,
			want: `{"type":"MultiLineString","coordinates":[[[0,1,2],[2,3,4]],[[4,5,9],[6,7,4],[8,9,7]]]}`,
		},
		{
			wkt:  `MULTILINESTRING ZM ((0 1 1 2,2 3 4 8),EMPTY)`,
			want: `{"type":"MultiLineString","coordinates":[[[0,1,1],[2,3,4]],[]]}`,
		},

		// NOTE: MultiPolygons oriented CCW in the output.
		{
			wkt:  `MULTIPOLYGON EMPTY`,
			want: `{"type":"MultiPolygon","coordinates":[]}`,
		},
		{
			wkt:  `MULTIPOLYGON(((0 0,1 0,0 1,0 0)),((1 0,2 0,1 1,1 0)))`,
			want: `{"type":"MultiPolygon","coordinates":[[[[0,0],[1,0],[0,1],[0,0]]],[[[1,0],[2,0],[1,1],[1,0]]]]}`,
		},
		{
			wkt:  `MULTIPOLYGON(((0 0,1 0,0 1,0 0)),EMPTY)`,
			want: `{"type":"MultiPolygon","coordinates":[[[[0,0],[1,0],[0,1],[0,0]]],[]]}`,
		},
		{
			wkt:  `MULTIPOLYGON Z EMPTY`,
			want: `{"type":"MultiPolygon","coordinates":[]}`,
		},
		{
			wkt:  `MULTIPOLYGON Z (((0 0 7,1 0 8,0 1 2,0 0 9)),((1 0 2,2 0 1,1 1 5,1 0 8)))`,
			want: `{"type":"MultiPolygon","coordinates":[[[[0,0,7],[1,0,8],[0,1,2],[0,0,9]]],[[[1,0,2],[2,0,1],[1,1,5],[1,0,8]]]]}`,
		},
		{
			wkt:  `MULTIPOLYGON Z (((0 0 4,1 0 2,0 1 8,0 0 9)),EMPTY)`,
			want: `{"type":"MultiPolygon","coordinates":[[[[0,0,4],[1,0,2],[0,1,8],[0,0,9]]],[]]}`,
		},
		{
			wkt:  `MULTIPOLYGON M EMPTY`,
			want: `{"type":"MultiPolygon","coordinates":[]}`,
		},
		{
			wkt:  `MULTIPOLYGON M (((0 0 7,1 0 8,0 1 2,0 0 9)),((1 0 2,2 0 1,1 1 5,1 0 8)))`,
			want: `{"type":"MultiPolygon","coordinates":[[[[0,0],[1,0],[0,1],[0,0]]],[[[1,0],[2,0],[1,1],[1,0]]]]}`,
		},
		{
			wkt:  `MULTIPOLYGON M (((0 0 4,1 0 2,0 1 8,0 0 9)),EMPTY)`,
			want: `{"type":"MultiPolygon","coordinates":[[[[0,0],[1,0],[0,1],[0,0]]],[]]}`,
		},
		{
			wkt:  `MULTIPOLYGON ZM EMPTY`,
			want: `{"type":"MultiPolygon","coordinates":[]}`,
		},
		{
			wkt:  `MULTIPOLYGON ZM (((0 0 7 8,1 0 8 6,0 1 2 3,0 0 9 9)),((1 0 2 3,2 0 1 1,1 1 5 7,1 0 8 9)))`,
			want: `{"type":"MultiPolygon","coordinates":[[[[0,0,7],[1,0,8],[0,1,2],[0,0,9]]],[[[1,0,2],[2,0,1],[1,1,5],[1,0,8]]]]}`,
		},
		{
			wkt:  `MULTIPOLYGON ZM (((0 0 4 2,1 0 2 1,0 1 8 8,0 0 9 5)),EMPTY)`,
			want: `{"type":"MultiPolygon","coordinates":[[[[0,0,4],[1,0,2],[0,1,8],[0,0,9]]],[]]}`,
		},
		{
			wkt:  `GEOMETRYCOLLECTION EMPTY`,
			want: `{"type":"GeometryCollection","geometries":[]}`,
		},
		{
			wkt:  `GEOMETRYCOLLECTION(POINT(1 2),POINT(3 4))`,
			want: `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2]},{"type":"Point","coordinates":[3,4]}]}`,
		},
		{
			wkt:  `GEOMETRYCOLLECTION Z EMPTY`,
			want: `{"type":"GeometryCollection","geometries":[]}`,
		},
		{
			wkt:  `GEOMETRYCOLLECTION Z (POINT Z (1 2 3),POINT Z (3 4 5))`,
			want: `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2,3]},{"type":"Point","coordinates":[3,4,5]}]}`,
		},
		{
			wkt:  `GEOMETRYCOLLECTION M EMPTY`,
			want: `{"type":"GeometryCollection","geometries":[]}`,
		},
		{
			wkt:  `GEOMETRYCOLLECTION M (POINT M (1 2 3),POINT M (3 4 5))`,
			want: `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2]},{"type":"Point","coordinates":[3,4]}]}`,
		},
		{
			wkt:  `GEOMETRYCOLLECTION ZM EMPTY`,
			want: `{"type":"GeometryCollection","geometries":[]}`,
		},
		{
			wkt:  `GEOMETRYCOLLECTION ZM (POINT ZM (1 2 3 4),POINT ZM (3 4 5 6))`,
			want: `{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[1,2,3]},{"type":"Point","coordinates":[3,4,5]}]}`,
		},
	} {
		desc := strings.ReplaceAll(tt.wkt, "(", "_")
		desc = strings.ReplaceAll(desc, ")", "_")
		t.Run(desc, func(t *testing.T) {
			geom := geomFromWKT(t, tt.wkt)
			gotJSON, err := json.Marshal(geom)
			expectNoErr(t, err)
			if string(gotJSON) != tt.want {
				t.Error("json doesn't match")
				t.Logf("got:\n%v", hex.Dump(gotJSON))
				t.Logf("want:\n%v", hex.Dump([]byte(tt.want)))
			}
		})
	}
}
