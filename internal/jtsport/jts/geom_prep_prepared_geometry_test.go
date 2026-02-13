package jts

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

// TRANSLITERATION NOTE: Java main() method (JUnit TestRunner entry point) not
// ported - Go uses `go test`.

// TRANSLITERATION NOTE: Java constructor
// PreparedGeometryTest(String name) not ported - JUnit TestCase infrastructure
// not needed in Go.

func TestPreparedGeometryEmptyElement(t *testing.T) {
	reader := Io_NewWKTReader()
	geomA, err := reader.Read("MULTIPOLYGON (((9 9, 9 1, 1 1, 2 4, 7 7, 9 9)), EMPTY)")
	if err != nil {
		t.Fatalf("failed to read geomA: %v", err)
	}
	geomB, err := reader.Read("MULTIPOLYGON (((7 6, 7 3, 4 3, 7 6)), EMPTY)")
	if err != nil {
		t.Fatalf("failed to read geomB: %v", err)
	}
	prepA := GeomPrep_PreparedGeometryFactory_Prepare(geomA)
	junit.AssertTrue(t, prepA.Covers(geomB))
	junit.AssertTrue(t, prepA.Contains(geomB))
	junit.AssertTrue(t, prepA.Intersects(geomB))
}
