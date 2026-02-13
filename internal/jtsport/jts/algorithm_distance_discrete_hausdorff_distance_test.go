package jts

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestDiscreteHausdorffDistance_LineSegments(t *testing.T) {
	algorithmDistance_discreteHausdorffDistance_runTest(t, "LINESTRING (0 0, 2 1)", "LINESTRING (0 0, 2 0)", 1.0)
}

func TestDiscreteHausdorffDistance_LineSegments2(t *testing.T) {
	algorithmDistance_discreteHausdorffDistance_runTest(t, "LINESTRING (0 0, 2 0)", "LINESTRING (0 1, 1 2, 2 1)", 2.0)
}

func TestDiscreteHausdorffDistance_LinePoints(t *testing.T) {
	algorithmDistance_discreteHausdorffDistance_runTest(t, "LINESTRING (0 0, 2 0)", "MULTIPOINT (0 1, 1 0, 2 1)", 1.0)
}

// TestDiscreteHausdorffDistance_LinesShowingDiscretenessEffect shows effects of limiting HD to vertices.
// Answer is not true Hausdorff distance.
func TestDiscreteHausdorffDistance_LinesShowingDiscretenessEffect(t *testing.T) {
	algorithmDistance_discreteHausdorffDistance_runTest(t, "LINESTRING (130 0, 0 0, 0 150)", "LINESTRING (10 10, 10 150, 130 10)", 14.142135623730951)
	// densifying provides accurate HD
	algorithmDistance_discreteHausdorffDistance_runTestWithDensifyFrac(t, "LINESTRING (130 0, 0 0, 0 150)", "LINESTRING (10 10, 10 150, 130 10)", 0.5, 70.0)
}

const algorithmDistance_discreteHausdorffDistance_tolerance = 0.00001

func algorithmDistance_discreteHausdorffDistance_runTest(t *testing.T, wkt1, wkt2 string, expectedDistance float64) {
	t.Helper()
	g1 := algorithmDistance_discreteHausdorffDistance_readWKT(t, wkt1)
	g2 := algorithmDistance_discreteHausdorffDistance_readWKT(t, wkt2)

	distance := AlgorithmDistance_DiscreteHausdorffDistance_Distance(g1, g2)
	junit.AssertEqualsFloat64(t, expectedDistance, distance, algorithmDistance_discreteHausdorffDistance_tolerance)
}

func algorithmDistance_discreteHausdorffDistance_runTestWithDensifyFrac(t *testing.T, wkt1, wkt2 string, densifyFrac, expectedDistance float64) {
	t.Helper()
	g1 := algorithmDistance_discreteHausdorffDistance_readWKT(t, wkt1)
	g2 := algorithmDistance_discreteHausdorffDistance_readWKT(t, wkt2)

	distance := AlgorithmDistance_DiscreteHausdorffDistance_DistanceWithDensifyFrac(g1, g2, densifyFrac)
	junit.AssertEqualsFloat64(t, expectedDistance, distance, algorithmDistance_discreteHausdorffDistance_tolerance)
}

func algorithmDistance_discreteHausdorffDistance_readWKT(t *testing.T, wkt string) *Geom_Geometry {
	t.Helper()
	reader := Io_NewWKTReader()
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("Failed to parse WKT: %v", err)
	}
	return geom
}
