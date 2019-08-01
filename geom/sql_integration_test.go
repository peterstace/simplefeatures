package simplefeatures_test

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	. "github.com/peterstace/simplefeatures/geom"
)

func newDB(t *testing.T) *sql.DB {
	t.Helper()
	const uri = "postgres://postgres:password@postgis:5432/postgres?sslmode=disable"
	db, err := sql.Open("postgres", uri)
	if err != nil {
		t.Fatalf("could not connect to db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("could not ping db: %v", err)
	}
	return db
}

func TestIntegrationValuerScanner(t *testing.T) {
	db := newDB(t)
	defer db.Close()

	input := AnyGeometry{geomFromWKT(t, "POINT(4 2)")}
	var output AnyGeometry
	if err := db.QueryRow("SELECT ST_AsText(ST_GeomFromText($1))", &input).Scan(&output); err != nil {
		t.Fatal(err)
	}
	expectDeepEqual(t, output, input)
}
