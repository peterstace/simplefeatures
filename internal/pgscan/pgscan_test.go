package pgscan_test

import (
	"database/sql"
	"strconv"
	"testing"

	_ "github.com/lib/pq"
	"github.com/peterstace/simplefeatures/geom"
)

func TestPostgresScan(t *testing.T) {
	const dbURL = "postgres://postgres:password@postgis:5432/postgres?sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	for i, tc := range []struct {
		wkt      string
		concrete interface{ AsText() string }
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
			if err := db.QueryRow(
				`SELECT ST_AsBinary(ST_GeomFromText($1))`,
				tc.wkt,
			).Scan(tc.concrete); err != nil {
				t.Error(err)
			}
			if got := tc.concrete.AsText(); got != tc.wkt {
				t.Errorf("want=%v got=%v", tc.wkt, got)
			}
		})
	}
}
