package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

// Tests for SimplePointInAreaLocator ported from
// org.locationtech.jts.algorithm.locate.SimplePointInAreaLocatorTest
// which extends AbstractPointInRingTest.

func runSimplePointInAreaLocatorPtInRing(t *testing.T, expectedLoc int, pt *jts.Geom_Coordinate, geom *jts.Geom_Geometry) {
	t.Helper()
	loc := jts.AlgorithmLocate_NewSimplePointInAreaLocator(geom)
	result := loc.Locate(pt)
	if result != expectedLoc {
		t.Errorf("expected location %d, got %d", expectedLoc, result)
	}
}

func TestSimplePointInAreaLocatorBox(t *testing.T) {
	geom := createPolygonFromCoords(t, [][2]float64{
		{0, 0}, {0, 20}, {20, 20}, {20, 0}, {0, 0},
	})
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Interior, jts.Geom_NewCoordinateWithXY(10, 10), geom)
}

func TestSimplePointInAreaLocatorComplexRing(t *testing.T) {
	geom := createPolygonFromCoords(t, [][2]float64{
		{-40, 80}, {-40, -80}, {20, 0}, {20, -100}, {40, 40}, {80, -80},
		{100, 80}, {140, -20}, {120, 140}, {40, 180}, {60, 40}, {0, 120},
		{-20, -20}, {-40, 80},
	})
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Interior, jts.Geom_NewCoordinateWithXY(0, 0), geom)
}

func TestSimplePointInAreaLocatorComb(t *testing.T) {
	geom := createPolygonFromCoords(t, [][2]float64{
		{0, 0}, {0, 10}, {4, 5}, {6, 10}, {7, 5}, {9, 10}, {10, 5}, {13, 5},
		{15, 10}, {16, 3}, {17, 10}, {18, 3}, {25, 10}, {30, 10}, {30, 0},
		{15, 0}, {14, 5}, {13, 0}, {9, 0}, {8, 5}, {6, 0}, {0, 0},
	})

	// Boundary tests.
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Boundary, jts.Geom_NewCoordinateWithXY(0, 0), geom)
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Boundary, jts.Geom_NewCoordinateWithXY(0, 1), geom)
	// At vertex.
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Boundary, jts.Geom_NewCoordinateWithXY(4, 5), geom)
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Boundary, jts.Geom_NewCoordinateWithXY(8, 5), geom)
	// On horizontal segment.
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Boundary, jts.Geom_NewCoordinateWithXY(11, 5), geom)
	// On vertical segment.
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Boundary, jts.Geom_NewCoordinateWithXY(30, 5), geom)
	// On angled segment.
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Boundary, jts.Geom_NewCoordinateWithXY(22, 7), geom)

	// Interior tests.
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Interior, jts.Geom_NewCoordinateWithXY(1, 5), geom)
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Interior, jts.Geom_NewCoordinateWithXY(5, 5), geom)
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Interior, jts.Geom_NewCoordinateWithXY(1, 7), geom)

	// Exterior tests.
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Exterior, jts.Geom_NewCoordinateWithXY(12, 10), geom)
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Exterior, jts.Geom_NewCoordinateWithXY(16, 5), geom)
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Exterior, jts.Geom_NewCoordinateWithXY(35, 5), geom)
}

func TestSimplePointInAreaLocatorRepeatedPts(t *testing.T) {
	geom := createPolygonFromCoords(t, [][2]float64{
		{0, 0}, {0, 10}, {2, 5}, {2, 5}, {2, 5}, {2, 5}, {2, 5}, {3, 10},
		{6, 10}, {8, 5}, {8, 5}, {8, 5}, {8, 5}, {10, 10}, {10, 5}, {10, 5},
		{10, 5}, {10, 5}, {10, 0}, {0, 0},
	})

	// Boundary tests.
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Boundary, jts.Geom_NewCoordinateWithXY(0, 0), geom)
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Boundary, jts.Geom_NewCoordinateWithXY(0, 1), geom)
	// At vertex.
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Boundary, jts.Geom_NewCoordinateWithXY(2, 5), geom)
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Boundary, jts.Geom_NewCoordinateWithXY(8, 5), geom)
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Boundary, jts.Geom_NewCoordinateWithXY(10, 5), geom)

	// Interior tests.
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Interior, jts.Geom_NewCoordinateWithXY(1, 5), geom)
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Interior, jts.Geom_NewCoordinateWithXY(3, 5), geom)
}

func TestSimplePointInAreaLocatorRobustStressTriangles(t *testing.T) {
	geom1 := createPolygonFromCoords(t, [][2]float64{
		{0.0, 0.0}, {0.0, 172.0}, {100.0, 0.0}, {0.0, 0.0},
	})
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Exterior, jts.Geom_NewCoordinateWithXY(25.374625374625374, 128.35564435564436), geom1)

	geom2 := createPolygonFromCoords(t, [][2]float64{
		{642.0, 815.0}, {69.0, 764.0}, {394.0, 966.0}, {642.0, 815.0},
	})
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Interior, jts.Geom_NewCoordinateWithXY(97.96039603960396, 782.0), geom2)
}

func TestSimplePointInAreaLocatorRobustTriangle(t *testing.T) {
	geom := createPolygonFromCoords(t, [][2]float64{
		{2.152214146946829, 50.470470727186765},
		{18.381941666723034, 19.567250592139274},
		{2.390837642830135, 49.228045261718165},
		{2.152214146946829, 50.470470727186765},
	})
	runSimplePointInAreaLocatorPtInRing(t, jts.Geom_Location_Exterior, jts.Geom_NewCoordinateWithXY(3.166572116932842, 48.5390194687463), geom)
}

// createPolygonFromCoords is a helper function to create a polygon from a list of coordinates.
func createPolygonFromCoords(t *testing.T, coords [][2]float64) *jts.Geom_Geometry {
	t.Helper()
	gf := jts.Geom_NewGeometryFactoryDefault()
	coordList := make([]*jts.Geom_Coordinate, len(coords))
	for i, c := range coords {
		coordList[i] = jts.Geom_NewCoordinateWithXY(c[0], c[1])
	}
	ring := gf.CreateLinearRingFromCoordinates(coordList)
	poly := gf.CreatePolygonFromLinearRing(ring)
	return poly.Geom_Geometry
}
