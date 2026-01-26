package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestLineSequencerSimple(t *testing.T) {
	doLineSequencerTest(t,
		[]string{
			"LINESTRING ( 0 0, 0 10 )",
			"LINESTRING ( 0 20, 0 30 )",
			"LINESTRING ( 0 10, 0 20 )",
		},
		"MULTILINESTRING ((0 0, 0 10), (0 10, 0 20), (0 20, 0 30))",
	)
}

func TestLineSequencerSimpleLoop(t *testing.T) {
	doLineSequencerTest(t,
		[]string{
			"LINESTRING ( 0 0, 0 10 )",
			"LINESTRING ( 0 10, 0 0 )",
		},
		"MULTILINESTRING ((0 0, 0 10), (0 10, 0 0))",
	)
}

func TestLineSequencerSimpleBigLoop(t *testing.T) {
	doLineSequencerTest(t,
		[]string{
			"LINESTRING ( 0 0, 0 10 )",
			"LINESTRING ( 0 20, 0 30 )",
			"LINESTRING ( 0 30, 0 00 )",
			"LINESTRING ( 0 10, 0 20 )",
		},
		"MULTILINESTRING ((0 0, 0 10), (0 10, 0 20), (0 20, 0 30), (0 30, 0 0))",
	)
}

func TestLineSequencer2SimpleLoops(t *testing.T) {
	doLineSequencerTest(t,
		[]string{
			"LINESTRING ( 0 0, 0 10 )",
			"LINESTRING ( 0 10, 0 0 )",
			"LINESTRING ( 0 0, 0 20 )",
			"LINESTRING ( 0 20, 0 0 )",
		},
		"MULTILINESTRING ((0 10, 0 0), (0 0, 0 20), (0 20, 0 0), (0 0, 0 10))",
	)
}

func TestLineSequencerWide8WithTail(t *testing.T) {
	// This is not sequenceable (expected result is nil).
	doLineSequencerTestNotSequenceable(t,
		[]string{
			"LINESTRING ( 0 0, 0 10 )",
			"LINESTRING ( 10 0, 10 10 )",
			"LINESTRING ( 0 0, 10 0 )",
			"LINESTRING ( 0 10, 10 10 )",
			"LINESTRING ( 0 10, 0 20 )",
			"LINESTRING ( 10 10, 10 20 )",
			"LINESTRING ( 0 20, 10 20 )",
			"LINESTRING ( 10 20, 30 30 )",
		},
	)
}

func TestLineSequencerSimpleLoopWithTail(t *testing.T) {
	doLineSequencerTest(t,
		[]string{
			"LINESTRING ( 0 0, 0 10 )",
			"LINESTRING ( 0 10, 10 10 )",
			"LINESTRING ( 10 10, 10 20, 0 10 )",
		},
		"MULTILINESTRING ((0 0, 0 10), (0 10, 10 10), (10 10, 10 20, 0 10))",
	)
}

func TestLineSequencerLineWithRing(t *testing.T) {
	doLineSequencerTest(t,
		[]string{
			"LINESTRING ( 0 0, 0 10 )",
			"LINESTRING ( 0 10, 10 10, 10 20, 0 10 )",
			"LINESTRING ( 0 30, 0 20 )",
			"LINESTRING ( 0 20, 0 10 )",
		},
		"MULTILINESTRING ((0 0, 0 10), (0 10, 10 10, 10 20, 0 10), (0 10, 0 20), (0 20, 0 30))",
	)
}

func TestLineSequencerMultipleGraphsWithRing(t *testing.T) {
	doLineSequencerTest(t,
		[]string{
			"LINESTRING ( 0 0, 0 10 )",
			"LINESTRING ( 0 10, 10 10, 10 20, 0 10 )",
			"LINESTRING ( 0 30, 0 20 )",
			"LINESTRING ( 0 20, 0 10 )",
			"LINESTRING ( 0 60, 0 50 )",
			"LINESTRING ( 0 40, 0 50 )",
		},
		"MULTILINESTRING ((0 0, 0 10), (0 10, 10 10, 10 20, 0 10), (0 10, 0 20), (0 20, 0 30), (0 40, 0 50), (0 50, 0 60))",
	)
}

func TestLineSequencerMultipleGraphsWithMultipleRings(t *testing.T) {
	doLineSequencerTest(t,
		[]string{
			"LINESTRING ( 0 0, 0 10 )",
			"LINESTRING ( 0 10, 10 10, 10 20, 0 10 )",
			"LINESTRING ( 0 10, 40 40, 40 20, 0 10 )",
			"LINESTRING ( 0 30, 0 20 )",
			"LINESTRING ( 0 20, 0 10 )",
			"LINESTRING ( 0 60, 0 50 )",
			"LINESTRING ( 0 40, 0 50 )",
		},
		"MULTILINESTRING ((0 0, 0 10), (0 10, 40 40, 40 20, 0 10), (0 10, 10 10, 10 20, 0 10), (0 10, 0 20), (0 20, 0 30), (0 40, 0 50), (0 50, 0 60))",
	)
}

// IsSequenced tests.

func TestLineSequencerLineSequence(t *testing.T) {
	doIsSequencedTest(t, "LINESTRING ( 0 0, 0 10 )", true)
}

func TestLineSequencerSplitLineSequence(t *testing.T) {
	doIsSequencedTest(t, "MULTILINESTRING ((0 0, 0 1), (0 2, 0 3), (0 3, 0 4) )", true)
}

func TestLineSequencerBadLineSequence(t *testing.T) {
	doIsSequencedTest(t, "MULTILINESTRING ((0 0, 0 1), (0 2, 0 3), (0 1, 0 4) )", false)
}

func doLineSequencerTest(t *testing.T, inputWKT []string, expectedWKT string) {
	t.Helper()
	reader := jts.Io_NewWKTReader()

	sequencer := jts.OperationLinemerge_NewLineSequencer()
	for _, wkt := range inputWKT {
		geom, err := reader.Read(wkt)
		if err != nil {
			t.Fatalf("failed to read input WKT %q: %v", wkt, err)
		}
		sequencer.AddGeometry(geom)
	}

	if !sequencer.IsSequenceable() {
		t.Fatal("expected geometry to be sequenceable but it wasn't")
	}

	expected, err := reader.Read(expectedWKT)
	if err != nil {
		t.Fatalf("failed to read expected WKT %q: %v", expectedWKT, err)
	}

	result := sequencer.GetSequencedLineStrings()
	if !expected.EqualsNorm(result) {
		t.Errorf("expected %v but got %v", expectedWKT, result)
	}

	// Verify that the result is itself sequenced.
	if !jts.OperationLinemerge_LineSequencer_IsSequenced(result) {
		t.Error("result is not sequenced")
	}
}

func doLineSequencerTestNotSequenceable(t *testing.T, inputWKT []string) {
	t.Helper()
	reader := jts.Io_NewWKTReader()

	sequencer := jts.OperationLinemerge_NewLineSequencer()
	for _, wkt := range inputWKT {
		geom, err := reader.Read(wkt)
		if err != nil {
			t.Fatalf("failed to read input WKT %q: %v", wkt, err)
		}
		sequencer.AddGeometry(geom)
	}

	if sequencer.IsSequenceable() {
		t.Error("expected geometry to NOT be sequenceable but it was")
	}
}

func doIsSequencedTest(t *testing.T, inputWKT string, expected bool) {
	t.Helper()
	reader := jts.Io_NewWKTReader()

	geom, err := reader.Read(inputWKT)
	if err != nil {
		t.Fatalf("failed to read WKT %q: %v", inputWKT, err)
	}

	actual := jts.OperationLinemerge_LineSequencer_IsSequenced(geom)
	if actual != expected {
		t.Errorf("isSequenced: expected %v but got %v", expected, actual)
	}
}
