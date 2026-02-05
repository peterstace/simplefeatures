package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/peterstace/simplefeatures/geom"
)

func TestFuzz(t *testing.T) {
	pg := setupDB(t)
	candidates := loadStringsFromFile(t, "../testdata/strings.txt")

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
			// Sending large coordinates to PostGIS works fine, but comparing the
			// results fails because geom.ExactEquals only supports absolute
			// tolerance. A relative tolerance option for ExactEquals doesn't
			// exist yet, but would allow these comparisons to succeed.
			if hasLargeCoordinates(g) {
				t.Skip("Geometry has large coordinates that cause floating point precision issues in absolute comparisons")
			}
			want, err := BatchPostGIS(pg).Unary(g)
			if err != nil {
				t.Fatalf("could not get result from postgis: %v", err)
			}
			checkWKB(t, want, g)
			checkGeoJSON(t, want, g)
			checkIsEmpty(t, want, g)
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

func loadStringsFromFile(t *testing.T, path string) []string {
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("could not open strings file: %v", err)
	}
	defer f.Close()

	strSet := map[string]struct{}{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		strSet[line] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("could not read strings file: %v", err)
	}

	var strs []string
	for s := range strSet {
		strs = append(strs, s)
	}
	sort.Strings(strs)
	return strs
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

// hasLargeCoordinates returns true if the geometry has any coordinates with
// magnitude large enough to cause floating point precision issues in
// comparisons.
func hasLargeCoordinates(g geom.Geometry) bool {
	env := g.Envelope()
	lo, hi, ok := env.MinMaxXYs()
	if !ok {
		return false
	}
	const threshold = 1e6
	return math.Abs(lo.X) > threshold ||
		math.Abs(lo.Y) > threshold ||
		math.Abs(hi.X) > threshold ||
		math.Abs(hi.Y) > threshold
}
