package main

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/peterstace/simplefeatures/geom"
	"github.com/peterstace/simplefeatures/internal/extract"
)

func TestFuzz(t *testing.T) {
	pg := setupDB(t)
	candidates, err := extract.StringsFromSource("../../..")
	if err != nil {
		t.Fatalf("could not extract strings from source: %v", err)
	}

	checkWKTParse(t, pg, candidates)
	checkWKBParse(t, pg, candidates)
	checkGeoJSONParse(t, pg, candidates)

	geoms := convertToGeometries(t, candidates)

	for i, g := range geoms {
		// Use fmt log instead of t log in case of panic.
		fmt.Printf("index=%d WKT=%v\n", i, g.AsText())
	}
	for i, g := range geoms {
		t.Run(fmt.Sprintf("geom_%d_", i), func(t *testing.T) {
			// This geometry shows a problem with the simplefeature's convex hull.
			// See https://github.com/peterstace/simplefeatures/issues/246
			if g.AsText() == "LINESTRING(1 0,0.5000000000000001 0.5,0 1)" {
				t.Skip("Causes unmarshalling to fail for derivative geometry")
			}

			if isMultiPointWithEmptyPoint(g) {
				t.Skip("PostGIS cannot handle MultiPoints that contain empty Points")
			}
			want, err := BatchPostGIS(pg).Unary(g)
			if err != nil {
				t.Fatalf("could not get result from postgis: %v", err)
			}
			checkWKB(t, want, g)
			checkGeoJSON(t, want, g)
			checkIsEmpty(t, want, g)
			checkDimension(t, want, g)
			checkEnvelope(t, want, g)
			checkConvexHull(t, want, g)
			checkIsRing(t, want, g)
			checkLength(t, want, g)
			checkArea(t, want, g)
			checkCentroid(t, want, g)
			checkReverse(t, want, g)
			checkType(t, want, g)
			checkForceOrientation(t, want, g)
			checkDump(t, want, g)
			checkForceCoordinatesDimension(t, want, g)
		})
	}
}

func setupDB(t *testing.T) PostGIS {
	db, err := sql.Open("postgres", "postgres://postgres:password@postgis:5432/postgres?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
	return PostGIS{db}
}

func convertToGeometries(t *testing.T, candidates []string) []geom.Geometry {
	var geoms []geom.Geometry
	for _, c := range candidates {
		g, err := geom.UnmarshalWKT(c)
		if err == nil {
			geoms = append(geoms, g)
		}
	}
	if len(geoms) == 0 {
		t.Fatal("could not extract any WKT geoms")
	}

	oldCount := len(geoms)
	for _, c := range candidates {
		buf, err := hexStringToBytes(c)
		if err != nil {
			continue
		}
		g, err := geom.UnmarshalWKB(buf)
		if err == nil {
			geoms = append(geoms, g)
		}
	}
	if oldCount == len(geoms) {
		t.Fatal("could not extract any WKB geoms")
	}

	oldCount = len(geoms)
	for _, c := range candidates {
		g, err := geom.UnmarshalGeoJSON([]byte(c))
		if err == nil {
			geoms = append(geoms, g)
		}
	}
	if oldCount == len(geoms) {
		t.Fatal("could not extract any geojson")
	}

	return geoms
}

func isMultiPointWithEmptyPoint(g geom.Geometry) bool {
	mp, ok := g.AsMultiPoint()
	if !ok {
		return false
	}
	for i := 0; i < mp.NumPoints(); i++ {
		if mp.PointN(i).IsEmpty() {
			return true
		}
	}
	return false
}
