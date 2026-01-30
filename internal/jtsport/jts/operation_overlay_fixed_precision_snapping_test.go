package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

// Tests ported from FixedPrecisionSnappingTest.java.

func TestFixedPrecisionSnappingTriangles(t *testing.T) {
	pm := jts.Geom_NewPrecisionModelWithScale(1.0)
	fact := jts.Geom_NewGeometryFactoryWithPrecisionModel(pm)
	reader := jts.Io_NewWKTReaderWithFactory(fact)

	a, err := reader.Read("POLYGON ((545 317, 617 379, 581 321, 545 317))")
	if err != nil {
		t.Fatalf("failed to read geometry a: %v", err)
	}
	b, err := reader.Read("POLYGON ((484 290, 558 359, 543 309, 484 290))")
	if err != nil {
		t.Fatalf("failed to read geometry b: %v", err)
	}

	// The test in Java just calls intersection and verifies it doesn't throw.
	// If the operation completes without panic, the test passes.
	result := a.Intersection(b)

	// Sanity check: result should not be nil.
	if result == nil {
		t.Fatal("intersection result should not be nil")
	}
}
