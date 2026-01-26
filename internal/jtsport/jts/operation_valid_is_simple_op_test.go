package jts_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

const isSimpleTolerance = 0.00005

// Test2TouchAtEndpoint tests 2 LineStrings touching at an endpoint.
func TestIsSimpleOp2TouchAtEndpoint(t *testing.T) {
	wkt := "MULTILINESTRING((0 1, 1 1, 2 1), (0 0, 1 0, 2 1))"
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE, true, coord(2, 1))
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_ENDPOINT_BOUNDARY_RULE, true, coord(2, 1))
}

// Test3TouchAtEndpoint tests 3 LineStrings touching at an endpoint.
func TestIsSimpleOp3TouchAtEndpoint(t *testing.T) {
	wkt := "MULTILINESTRING ((0 1, 1 1, 2 1),   (0 0, 1 0, 2 1),  (0 2, 1 2, 2 1))"
	// Rings are simple under all rules.
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE, true, coord(2, 1))
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_ENDPOINT_BOUNDARY_RULE, true, coord(2, 1))
}

func TestIsSimpleOpCross(t *testing.T) {
	wkt := "MULTILINESTRING ((20 120, 120 20), (20 20, 120 120))"
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE, false, coord(70, 70))
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_ENDPOINT_BOUNDARY_RULE, false, coord(70, 70))
}

func TestIsSimpleOpMultiLineStringWithRingTouchAtEndpoint(t *testing.T) {
	wkt := "MULTILINESTRING ((100 100, 20 20, 200 20, 100 100), (100 200, 100 100))"
	// Under Mod-2, the ring has no boundary, so the line intersects the interior ==> not simple.
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE, false, coord(100, 100))
	// Under Endpoint, the ring has a boundary point, so the line does NOT intersect the interior ==> simple.
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_ENDPOINT_BOUNDARY_RULE, true, nil)
}

func TestIsSimpleOpRing(t *testing.T) {
	wkt := "LINESTRING (100 100, 20 20, 200 20, 100 100)"
	// Rings are simple under all rules.
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE, true, nil)
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_ENDPOINT_BOUNDARY_RULE, true, nil)
}

func TestIsSimpleOpLineRepeatedStart(t *testing.T) {
	wkt := "LINESTRING (100 100, 100 100, 20 20, 200 20, 100 100)"
	// Rings are simple under all rules.
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE, true, nil)
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_ENDPOINT_BOUNDARY_RULE, true, nil)
}

func TestIsSimpleOpLineRepeatedEnd(t *testing.T) {
	wkt := "LINESTRING (100 100, 20 20, 200 20, 100 100, 100 100)"
	// Rings are simple under all rules.
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE, true, nil)
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_ENDPOINT_BOUNDARY_RULE, true, nil)
}

func TestIsSimpleOpLineRepeatedBothEnds(t *testing.T) {
	wkt := "LINESTRING (100 100, 100 100, 100 100, 20 20, 200 20, 100 100, 100 100)"
	// Rings are simple under all rules.
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE, true, nil)
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_ENDPOINT_BOUNDARY_RULE, true, nil)
}

func TestIsSimpleOpLineRepeatedAll(t *testing.T) {
	wkt := "LINESTRING (100 100, 100 100, 100 100)"
	// Rings are simple under all rules.
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE, true, nil)
	checkIsSimple(t, wkt, jts.Algorithm_BoundaryNodeRule_ENDPOINT_BOUNDARY_RULE, true, nil)
}

func TestIsSimpleOpLinesAll(t *testing.T) {
	checkIsSimpleAll(t,
		"MULTILINESTRING ((10 20, 90 20), (10 30, 90 30), (50 40, 50 10))",
		jts.Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE,
		"MULTIPOINT((50 20), (50 30))")
}

func TestIsSimpleOpPolygonAll(t *testing.T) {
	checkIsSimpleAll(t,
		"POLYGON ((0 0, 7 0, 6 -1, 6 -0.1, 6 0.1, 3 5.9, 3 6.1, 3.1 6, 2.9 6, 0 0))",
		jts.Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE,
		"MULTIPOINT((6 0), (3 6))")
}

func TestIsSimpleOpMultiPointAll(t *testing.T) {
	checkIsSimpleAll(t,
		"MULTIPOINT((1 1), (1 2), (1 2), (1 3), (1 4), (1 4), (1 5), (1 5))",
		jts.Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE,
		"MULTIPOINT((1 2), (1 4), (1 5))")
}

func TestIsSimpleOpGeometryCollectionAll(t *testing.T) {
	checkIsSimpleAll(t,
		"GEOMETRYCOLLECTION(MULTILINESTRING ((10 20, 90 20), (10 30, 90 30), (50 40, 50 10)), "+
			"MULTIPOINT((1 1), (1 2), (1 2), (1 3), (1 4), (1 4), (1 5), (1 5)))",
		jts.Algorithm_BoundaryNodeRule_MOD2_BOUNDARY_RULE,
		"MULTIPOINT((50 20), (50 30), (1 2), (1 4), (1 5))")
}

func coord(x, y float64) *jts.Geom_Coordinate {
	return jts.Geom_NewCoordinateWithXY(x, y)
}

func checkIsSimple(t *testing.T, wkt string, bnRule jts.Algorithm_BoundaryNodeRule, expectedResult bool, expectedLocation *jts.Geom_Coordinate) {
	t.Helper()
	reader := jts.Io_NewWKTReader()
	g, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("failed to read WKT: %v", err)
	}

	op := jts.OperationValid_NewIsSimpleOpWithBoundaryNodeRule(g, bnRule)
	isSimple := op.IsSimple()
	nonSimpleLoc := op.GetNonSimpleLocation()

	// If geom is not simple, should have a valid location.
	if !isSimple && nonSimpleLoc == nil {
		t.Errorf("geometry is not simple but no non-simple location returned")
	}

	if isSimple != expectedResult {
		t.Errorf("expected isSimple=%v, got %v", expectedResult, isSimple)
	}

	if !isSimple && expectedLocation != nil {
		dist := math.Sqrt(math.Pow(expectedLocation.X-nonSimpleLoc.X, 2) + math.Pow(expectedLocation.Y-nonSimpleLoc.Y, 2))
		if dist >= isSimpleTolerance {
			t.Errorf("expected non-simple location near %v, got %v (distance=%v)", expectedLocation, nonSimpleLoc, dist)
		}
	}
}

func checkIsSimpleAll(t *testing.T, wkt string, bnRule jts.Algorithm_BoundaryNodeRule, wktExpectedPts string) {
	t.Helper()
	reader := jts.Io_NewWKTReader()
	g, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("failed to read WKT: %v", err)
	}

	op := jts.OperationValid_NewIsSimpleOpWithBoundaryNodeRule(g, bnRule)
	op.SetFindAllLocations(true)
	op.IsSimple()
	nonSimpleCoords := op.GetNonSimpleLocations()

	nsPts := g.GetFactory().CreateMultiPointFromCoords(nonSimpleCoords)

	expectedPts, err := reader.Read(wktExpectedPts)
	if err != nil {
		t.Fatalf("failed to read expected WKT: %v", err)
	}

	if !checkEqualPoints(nsPts.Geom_Geometry, expectedPts) {
		t.Errorf("expected non-simple points %v, got %v", wktExpectedPts, coordsToString(nonSimpleCoords))
	}
}

func checkEqualPoints(actual, expected *jts.Geom_Geometry) bool {
	if actual.GetNumGeometries() != expected.GetNumGeometries() {
		return false
	}

	// Build a set of expected points.
	expectedSet := make(map[string]int)
	for i := 0; i < expected.GetNumGeometries(); i++ {
		pt := java.Cast[*jts.Geom_Point](expected.GetGeometryN(i))
		c := pt.GetCoordinate()
		if c != nil {
			key := coordKey(c)
			expectedSet[key]++
		}
	}

	// Check that all actual points are in the expected set.
	for i := 0; i < actual.GetNumGeometries(); i++ {
		pt := java.Cast[*jts.Geom_Point](actual.GetGeometryN(i))
		c := pt.GetCoordinate()
		if c != nil {
			key := coordKey(c)
			if expectedSet[key] > 0 {
				expectedSet[key]--
			} else {
				return false
			}
		}
	}

	// Check that all expected points were matched.
	for _, count := range expectedSet {
		if count > 0 {
			return false
		}
	}
	return true
}

func coordKey(c *jts.Geom_Coordinate) string {
	// Round to avoid floating point comparison issues.
	return fmt.Sprintf("%.6f,%.6f", c.X, c.Y)
}

func coordsToString(coords []*jts.Geom_Coordinate) string {
	var result string
	for i, c := range coords {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("(%v, %v)", c.X, c.Y)
	}
	return result
}
