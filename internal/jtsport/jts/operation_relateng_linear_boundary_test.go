package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestLinearBoundaryLineMod2(t *testing.T) {
	checkLinearBoundary(t, "LINESTRING (0 0, 9 9)",
		jts.Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE,
		"MULTIPOINT((0 0), (9 9))")
}

func TestLinearBoundaryLines2Mod2(t *testing.T) {
	checkLinearBoundary(t, "MULTILINESTRING ((0 0, 9 9), (9 9, 5 1))",
		jts.Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE,
		"MULTIPOINT((0 0), (5 1))")
}

func TestLinearBoundaryLines3Mod2(t *testing.T) {
	checkLinearBoundary(t, "MULTILINESTRING ((0 0, 9 9), (9 9, 5 1), (9 9, 1 5))",
		jts.Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE,
		"MULTIPOINT((0 0), (5 1), (1 5), (9 9))")
}

func TestLinearBoundaryLines3Monvalent(t *testing.T) {
	checkLinearBoundary(t, "MULTILINESTRING ((0 0, 9 9), (9 9, 5 1), (9 9, 1 5))",
		jts.Algorithm_BoundaryNodeRule_MONOVALENT_ENDPOINT_BOUNDARY_RULE,
		"MULTIPOINT((0 0), (5 1), (1 5))")
}

func checkLinearBoundary(t *testing.T, wkt string, bnr jts.Algorithm_BoundaryNodeRule, wktBdyExpected string) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("Failed to read geometry: %v", err)
	}

	lines := jts.GeomUtil_LineStringExtracter_GetLines(geom)
	lb := jts.OperationRelateng_NewLinearBoundary(lines, bnr)

	hasBoundaryExpected := wktBdyExpected != ""
	junit.AssertEquals(t, hasBoundaryExpected, lb.HasBoundary())

	checkBoundaryPoints(t, lb, geom, wktBdyExpected, reader)
}

// coord2DKey is a 2D coordinate key for map lookups (only X, Y).
// This is needed because Geom_Coordinate includes Z which may be NaN,
// and NaN != NaN in Go, breaking map lookups.
type coord2DKey struct {
	x, y float64
}

func checkBoundaryPoints(t *testing.T, lb *jts.OperationRelateng_LinearBoundary, geom *jts.Geom_Geometry, wktBdyExpected string, reader *jts.Io_WKTReader) {
	bdySet := extractPointsSet(t, wktBdyExpected, reader)

	for p := range bdySet {
		coord := &jts.Geom_Coordinate{X: p.x, Y: p.y}
		junit.AssertTrue(t, lb.IsBoundary(coord))
	}

	allPts := geom.GetCoordinates()
	for _, p := range allPts {
		key := coord2DKey{x: p.X, y: p.Y}
		if !bdySet[key] {
			junit.AssertFalse(t, lb.IsBoundary(p))
		}
	}
}

func extractPointsSet(t *testing.T, wkt string, reader *jts.Io_WKTReader) map[coord2DKey]bool {
	ptSet := make(map[coord2DKey]bool)
	if wkt == "" {
		return ptSet
	}
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("Failed to read WKT: %v", err)
	}
	pts := geom.GetCoordinates()
	for _, p := range pts {
		key := coord2DKey{x: p.X, y: p.Y}
		ptSet[key] = true
	}
	return ptSet
}
