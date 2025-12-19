package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestLineMerger1(t *testing.T) {
	// Three lines that should merge into one continuous linestring.
	doLineMergerTest(t,
		[]string{
			"LINESTRING (120 120, 180 140)",
			"LINESTRING (200 180, 180 140)",
			"LINESTRING (200 180, 240 180)",
		},
		[]string{
			"LINESTRING (120 120, 180 140, 200 180, 240 180)",
		},
	)
}

func TestLineMerger2(t *testing.T) {
	// Multiple groups including closed loops.
	doLineMergerTest(t,
		[]string{
			"LINESTRING (120 300, 80 340)",
			"LINESTRING (120 300, 140 320, 160 320)",
			"LINESTRING (40 320, 20 340, 0 320)",
			"LINESTRING (0 320, 20 300, 40 320)",
			"LINESTRING (40 320, 60 320, 80 340)",
			"LINESTRING (160 320, 180 340, 200 320)",
			"LINESTRING (200 320, 180 300, 160 320)",
		},
		[]string{
			"LINESTRING (160 320, 180 340, 200 320, 180 300, 160 320)",
			"LINESTRING (40 320, 20 340, 0 320, 20 300, 40 320)",
			"LINESTRING (40 320, 60 320, 80 340, 120 300, 140 320, 160 320)",
		},
	)
}

func TestLineMerger3(t *testing.T) {
	// Two lines that don't connect remain separate.
	doLineMergerTest(t,
		[]string{
			"LINESTRING (0 0, 100 100)",
			"LINESTRING (0 100, 100 0)",
		},
		[]string{
			"LINESTRING (0 0, 100 100)",
			"LINESTRING (0 100, 100 0)",
		},
	)
}

func TestLineMerger4(t *testing.T) {
	// Empty linestrings result in empty output.
	doLineMergerTest(t,
		[]string{
			"LINESTRING EMPTY",
			"LINESTRING EMPTY",
		},
		[]string{},
	)
}

func TestLineMerger5(t *testing.T) {
	// Empty input results in empty output.
	doLineMergerTest(t, []string{}, []string{})
}

func TestLineMergerSingleUniquePoint(t *testing.T) {
	// Single unique point and empty linestring result in empty output.
	doLineMergerTest(t,
		[]string{
			"LINESTRING (10642 31441, 10642 31441)",
			"LINESTRING EMPTY",
		},
		[]string{},
	)
}

func doLineMergerTest(t *testing.T, inputWKT, expectedOutputWKT []string) {
	t.Helper()
	reader := jts.Io_NewWKTReader()

	lineMerger := jts.OperationLinemerge_NewLineMerger()
	for _, wkt := range inputWKT {
		geom, err := reader.Read(wkt)
		if err != nil {
			t.Fatalf("failed to read input WKT %q: %v", wkt, err)
		}
		lineMerger.AddGeometry(geom)
	}

	expectedGeoms := make([]*jts.Geom_Geometry, len(expectedOutputWKT))
	for i, wkt := range expectedOutputWKT {
		geom, err := reader.Read(wkt)
		if err != nil {
			t.Fatalf("failed to read expected WKT %q: %v", wkt, err)
		}
		expectedGeoms[i] = geom
	}

	actualLineStrings := lineMerger.GetMergedLineStrings()

	if len(actualLineStrings) != len(expectedGeoms) {
		t.Fatalf("expected %d geometries, got %d", len(expectedGeoms), len(actualLineStrings))
	}

	// Check that each expected geometry is found in the actual results (using equalsExact).
	for _, expected := range expectedGeoms {
		found := false
		for _, actual := range actualLineStrings {
			if actual.Geom_Geometry.EqualsExact(expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected geometry not found: %v", expected)
		}
	}
}
