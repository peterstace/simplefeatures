package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

func TestAdjacentEdgeLocatorAdjacent2(t *testing.T) {
	checkAdjacentEdgeLocation(t,
		"GEOMETRYCOLLECTION (POLYGON ((1 9, 5 9, 5 1, 1 1, 1 9)), POLYGON ((9 9, 9 1, 5 1, 5 9, 9 9)))",
		5, 5, jts.Geom_Location_Interior)
}

func TestAdjacentEdgeLocatorNonAdjacent(t *testing.T) {
	checkAdjacentEdgeLocation(t,
		"GEOMETRYCOLLECTION (POLYGON ((1 9, 4 9, 5 1, 1 1, 1 9)), POLYGON ((9 9, 9 1, 5 1, 5 9, 9 9)))",
		5, 5, jts.Geom_Location_Boundary)
}

func TestAdjacentEdgeLocatorAdjacent6WithFilledHoles(t *testing.T) {
	checkAdjacentEdgeLocation(t,
		"GEOMETRYCOLLECTION (POLYGON ((1 9, 5 9, 6 6, 1 5, 1 9), (2 6, 4 8, 6 6, 2 6)), POLYGON ((2 6, 4 8, 6 6, 2 6)), POLYGON ((9 9, 9 5, 6 6, 5 9, 9 9)), POLYGON ((9 1, 5 1, 6 6, 9 5, 9 1), (7 2, 6 6, 8 3, 7 2)), POLYGON ((7 2, 6 6, 8 3, 7 2)), POLYGON ((1 1, 1 5, 6 6, 5 1, 1 1)))",
		6, 6, jts.Geom_Location_Interior)
}

func TestAdjacentEdgeLocatorAdjacent5WithEmptyHole(t *testing.T) {
	checkAdjacentEdgeLocation(t,
		"GEOMETRYCOLLECTION (POLYGON ((1 9, 5 9, 6 6, 1 5, 1 9), (2 6, 4 8, 6 6, 2 6)), POLYGON ((2 6, 4 8, 6 6, 2 6)), POLYGON ((9 9, 9 5, 6 6, 5 9, 9 9)), POLYGON ((9 1, 5 1, 6 6, 9 5, 9 1), (7 2, 6 6, 8 3, 7 2)), POLYGON ((1 1, 1 5, 6 6, 5 1, 1 1)))",
		6, 6, jts.Geom_Location_Boundary)
}

func TestAdjacentEdgeLocatorContainedAndAdjacent(t *testing.T) {
	wkt := "GEOMETRYCOLLECTION (POLYGON ((1 9, 9 9, 9 1, 1 1, 1 9)), POLYGON ((9 2, 2 2, 2 8, 9 8, 9 2)))"
	checkAdjacentEdgeLocation(t, wkt, 9, 5, jts.Geom_Location_Boundary)
	checkAdjacentEdgeLocation(t, wkt, 9, 8, jts.Geom_Location_Boundary)
}

func TestAdjacentEdgeLocatorDisjointCollinear(t *testing.T) {
	checkAdjacentEdgeLocation(t,
		"GEOMETRYCOLLECTION (MULTIPOLYGON (((1 4, 4 4, 4 1, 1 1, 1 4)), ((5 4, 8 4, 8 1, 5 1, 5 4))))",
		2, 4, jts.Geom_Location_Boundary)
}

func checkAdjacentEdgeLocation(t *testing.T, wkt string, x, y float64, expectedLoc int) {
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("Failed to read geometry: %v", err)
	}
	ael := jts.OperationRelateng_NewAdjacentEdgeLocator(geom)
	coord := &jts.Geom_Coordinate{X: x, Y: y}
	loc := ael.Locate(coord)
	junit.AssertEquals(t, expectedLoc, loc)
}
