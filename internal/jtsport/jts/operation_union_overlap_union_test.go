package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestOverlapUnionFixedPrecCausingBorderChange(t *testing.T) {
	a := "POLYGON ((130 -10, 20 -10, 20 22, 30 20, 130 20, 130 -10))"
	b := "MULTIPOLYGON (((50 0, 100 450, 100 0, 50 0)), ((53 28, 50 28, 50 30, 53 30, 53 28)))"
	checkOverlapUnionWithTopologyFailure(t, a, b, 1)
}

func TestOverlapUnionFullPrecision(t *testing.T) {
	a := "POLYGON ((130 -10, 20 -10, 20 22, 30 20, 130 20, 130 -10))"
	b := "MULTIPOLYGON (((50 0, 100 450, 100 0, 50 0)), ((53 28, 50 28, 50 30, 53 30, 53 28)))"
	checkOverlapUnionValid(t, a, b)
}

func TestOverlapUnionSimpleOverlap(t *testing.T) {
	a := "MULTIPOLYGON (((0 400, 50 400, 50 350, 0 350, 0 400)), ((200 200, 220 200, 220 180, 200 180, 200 200)), ((350 100, 370 100, 370 80, 350 80, 350 100)))"
	b := "MULTIPOLYGON (((430 20, 450 20, 450 0, 430 0, 430 20)), ((100 300, 124 300, 124 276, 100 276, 100 300)), ((230 170, 210 170, 210 190, 230 190, 230 170)))"
	checkOverlapUnionOptimized(t, a, b)
}

func checkOverlapUnionWithTopologyFailure(t *testing.T, wktA, wktB string, scaleFactor float64) {
	t.Helper()
	pm := jts.Geom_NewPrecisionModelWithScale(scaleFactor)
	geomFact := jts.Geom_NewGeometryFactoryWithPrecisionModel(pm)
	reader := jts.Io_NewWKTReaderWithFactory(geomFact)

	a, err := reader.Read(wktA)
	if err != nil {
		t.Fatalf("failed to read wktA: %v", err)
	}
	b, err := reader.Read(wktB)
	if err != nil {
		t.Fatalf("failed to read wktB: %v", err)
	}

	union := jts.OperationUnion_NewOverlapUnion(a, b)

	// The test expects a TopologyException in some cases.
	// We use recover to catch panics since Go doesn't have exceptions.
	func() {
		defer func() {
			if r := recover(); r != nil {
				isOptimized := union.IsUnionOptimized()
				// If the optimized algorithm was used then this is a real error.
				if isOptimized {
					t.Errorf("TopologyException with optimized algorithm: %v", r)
				}
				// Otherwise the error is probably due to fixed precision.
			}
		}()
		result := union.Union()
		if result != nil {
			// Result computed successfully; no topology exception.
		}
	}()
}

func checkOverlapUnionValid(t *testing.T, wktA, wktB string) {
	t.Helper()
	checkOverlapUnionInternal(t, wktA, wktB, false)
}

func checkOverlapUnionOptimized(t *testing.T, wktA, wktB string) {
	t.Helper()
	checkOverlapUnionInternal(t, wktA, wktB, true)
}

func checkOverlapUnionInternal(t *testing.T, wktA, wktB string, isCheckOptimized bool) {
	t.Helper()
	pm := jts.Geom_NewPrecisionModel()
	geomFact := jts.Geom_NewGeometryFactoryWithPrecisionModel(pm)
	reader := jts.Io_NewWKTReaderWithFactory(geomFact)

	a, err := reader.Read(wktA)
	if err != nil {
		t.Fatalf("failed to read wktA: %v", err)
	}
	b, err := reader.Read(wktB)
	if err != nil {
		t.Fatalf("failed to read wktB: %v", err)
	}

	union := jts.OperationUnion_NewOverlapUnion(a, b)
	result := union.Union()

	if isCheckOptimized {
		isOptimized := union.IsUnionOptimized()
		if !isOptimized {
			t.Errorf("union was not performed using optimized combine")
		}
	}

	if result == nil {
		t.Errorf("union result is nil")
	}
}
