package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

// Java method is named testArea but should be testLength.
func TestLengthArea(t *testing.T) {
	checkLengthOfLine(t, "LINESTRING (100 200, 200 200, 200 100, 100 100, 100 200)", 400.0)
}

func checkLengthOfLine(t *testing.T, wkt string, expectedLen float64) {
	t.Helper()
	reader := jts.Io_NewWKTReader()
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("failed to read WKT: %v", err)
	}
	ring := java.Cast[*jts.Geom_LineString](geom)
	pts := ring.GetCoordinateSequence()
	actual := jts.Algorithm_Length_OfLine(pts)
	junit.AssertEquals(t, expectedLen, actual)
}
