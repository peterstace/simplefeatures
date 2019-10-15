package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/peterstace/simplefeatures/geom"
)

func TestFuzz(t *testing.T) {
	pg := setupDB(t)
	candidates := extractStringsFromSource(t)

	CheckWKTParse(t, pg, candidates)
	CheckWKBParse(t, pg, candidates)
	CheckGeoJSONParse(t, pg, candidates)

	geoms := convertToGeometries(t, candidates)

	for i, g := range geoms {
		// Use fmt log instead of t log in case of panic.
		fmt.Printf("index=%d WKT=%v\n", i, g.AsText())
	}
	for i, g := range geoms {
		t.Run(fmt.Sprintf("geom_%d_", i), func(t *testing.T) {
			CheckWKT(t, pg, g)
			CheckWKB(t, pg, g)
			CheckGeoJSON(t, pg, g)
			CheckIsEmpty(t, pg, g)
			CheckDimension(t, pg, g)
			CheckEnvelope(t, pg, g)
			CheckIsSimple(t, pg, g)
			CheckBoundary(t, pg, g)
			CheckConvexHull(t, pg, g)
			CheckIsValid(t, pg, g)
			CheckIsRing(t, pg, g)
			CheckLength(t, pg, g)
			CheckArea(t, pg, g)
		})
	}
	for i, g1 := range geoms {
		for j, g2 := range geoms {
			t.Run(fmt.Sprintf("geom_%d_%d_", i, j), func(t *testing.T) {
				CheckEqualsExact(t, pg, g1, g2)
				CheckEquals(t, pg, g1, g2)
				CheckIntersects(t, pg, g1, g2)
				// TODO: Intersection
			})
		}
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

func extractStringsFromSource(t *testing.T) []string {
	var strs []string
	if err := filepath.Walk("..", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() || strings.Contains(path, ".git") {
			return nil
		}
		pkgs, err := parser.ParseDir(new(token.FileSet), path, nil, 0)
		if err != nil {
			return err
		}
		for _, pkg := range pkgs {
			ast.Inspect(pkg, func(n ast.Node) bool {
				lit, ok := n.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				unquoted, err := strconv.Unquote(lit.Value)
				if !ok {
					// Shouldn't ever happen because we've validated that it's a string literal.
					panic(fmt.Sprintf("could not unquote string '%s'from ast: %v", lit.Value, err))
				}
				strs = append(strs, unquoted)
				return true
			})
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	strSet := map[string]struct{}{}
	for _, s := range strs {
		strSet[strings.TrimSpace(s)] = struct{}{}
	}
	strs = strs[:0]
	for s := range strSet {
		strs = append(strs, s)
	}
	sort.Strings(strs)
	return strs
}

func convertToGeometries(t *testing.T, candidates []string) []geom.Geometry {
	var geoms []geom.Geometry
	for _, c := range candidates {
		g, err := geom.UnmarshalWKT(strings.NewReader(c))
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
		g, err := geom.UnmarshalWKB(bytes.NewReader(buf))
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
