package jts_test

// Tests ported from org.locationtech.jts.geom.PointImplTest.java

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestPointEquals1(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1000)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	reader := jts.Io_NewWKTReaderWithFactory(geometryFactory)

	p1, err := reader.Read("POINT(1.234 5.678)")
	if err != nil {
		t.Fatalf("failed to read p1: %v", err)
	}
	p2, err := reader.Read("POINT(1.234 5.678)")
	if err != nil {
		t.Fatalf("failed to read p2: %v", err)
	}
	if !p1.EqualsExact(p2) {
		t.Errorf("expected p1 to equal p2")
	}
}

func TestPointEquals2(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1000)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	reader := jts.Io_NewWKTReaderWithFactory(geometryFactory)

	p1, err := reader.Read("POINT(1.23 5.67)")
	if err != nil {
		t.Fatalf("failed to read p1: %v", err)
	}
	p2, err := reader.Read("POINT(1.23 5.67)")
	if err != nil {
		t.Fatalf("failed to read p2: %v", err)
	}
	if !p1.EqualsExact(p2) {
		t.Errorf("expected p1 to equal p2")
	}
}

func TestPointEquals3(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1000)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	reader := jts.Io_NewWKTReaderWithFactory(geometryFactory)

	p1, err := reader.Read("POINT(1.235 5.678)")
	if err != nil {
		t.Fatalf("failed to read p1: %v", err)
	}
	p2, err := reader.Read("POINT(1.234 5.678)")
	if err != nil {
		t.Fatalf("failed to read p2: %v", err)
	}
	if p1.EqualsExact(p2) {
		t.Errorf("expected p1 to NOT equal p2")
	}
}

func TestPointEquals4(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1000)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	reader := jts.Io_NewWKTReaderWithFactory(geometryFactory)

	// Both 1.2334 and 1.2333 round to 1.233 with scale 1000.
	p1, err := reader.Read("POINT(1.2334 5.678)")
	if err != nil {
		t.Fatalf("failed to read p1: %v", err)
	}
	p2, err := reader.Read("POINT(1.2333 5.678)")
	if err != nil {
		t.Fatalf("failed to read p2: %v", err)
	}
	if !p1.EqualsExact(p2) {
		t.Errorf("expected p1 to equal p2 (both should round to 1.233)")
	}
}

func TestPointEquals5(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1000)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	reader := jts.Io_NewWKTReaderWithFactory(geometryFactory)

	// 1.2334 rounds to 1.233, 1.2335 rounds to 1.234 (different).
	p1, err := reader.Read("POINT(1.2334 5.678)")
	if err != nil {
		t.Fatalf("failed to read p1: %v", err)
	}
	p2, err := reader.Read("POINT(1.2335 5.678)")
	if err != nil {
		t.Fatalf("failed to read p2: %v", err)
	}
	if p1.EqualsExact(p2) {
		t.Errorf("expected p1 to NOT equal p2 (1.233 != 1.234)")
	}
}

func TestPointEquals6(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1000)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	reader := jts.Io_NewWKTReaderWithFactory(geometryFactory)

	// 1.2324 rounds to 1.232, 1.2325 rounds to 1.233 (different).
	p1, err := reader.Read("POINT(1.2324 5.678)")
	if err != nil {
		t.Fatalf("failed to read p1: %v", err)
	}
	p2, err := reader.Read("POINT(1.2325 5.678)")
	if err != nil {
		t.Fatalf("failed to read p2: %v", err)
	}
	if p1.EqualsExact(p2) {
		t.Errorf("expected p1 to NOT equal p2 (1.232 != 1.233)")
	}
}

func TestPointNegRounding1(t *testing.T) {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1000)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	reader := jts.Io_NewWKTReaderWithFactory(geometryFactory)

	pLo, err := reader.Read("POINT(-1.233 5.678)")
	if err != nil {
		t.Fatalf("failed to read pLo: %v", err)
	}
	pHi, err := reader.Read("POINT(-1.232 5.678)")
	if err != nil {
		t.Fatalf("failed to read pHi: %v", err)
	}

	// -1.2326 rounds to -1.233.
	p1, err := reader.Read("POINT(-1.2326 5.678)")
	if err != nil {
		t.Fatalf("failed to read p1: %v", err)
	}
	// -1.2325 rounds to -1.232 (round half away from zero for negative).
	p2, err := reader.Read("POINT(-1.2325 5.678)")
	if err != nil {
		t.Fatalf("failed to read p2: %v", err)
	}
	// -1.2324 rounds to -1.232.
	p3, err := reader.Read("POINT(-1.2324 5.678)")
	if err != nil {
		t.Fatalf("failed to read p3: %v", err)
	}

	// p1 (-1.233) != p2 (-1.232).
	if p1.EqualsExact(p2) {
		t.Errorf("expected p1 to NOT equal p2")
	}
	// p3 (-1.232) == p2 (-1.232).
	if !p3.EqualsExact(p2) {
		t.Errorf("expected p3 to equal p2")
	}

	// p1 (-1.233) == pLo (-1.233).
	if !p1.EqualsExact(pLo) {
		t.Errorf("expected p1 to equal pLo")
	}
	// p2 (-1.232) == pHi (-1.232).
	if !p2.EqualsExact(pHi) {
		t.Errorf("expected p2 to equal pHi")
	}
	// p3 (-1.232) == pHi (-1.232).
	if !p3.EqualsExact(pHi) {
		t.Errorf("expected p3 to equal pHi")
	}
}

func TestPointIsSimple(t *testing.T) {
	reader := jts.Io_NewWKTReader()

	p1, err := reader.Read("POINT(1.2324 5.678)")
	if err != nil {
		t.Fatalf("failed to read p1: %v", err)
	}
	if !p1.IsSimple() {
		t.Error("expected point to be simple")
	}

	p2, err := reader.Read("POINT EMPTY")
	if err != nil {
		t.Fatalf("failed to read p2: %v", err)
	}
	if !p2.IsSimple() {
		t.Error("expected empty point to be simple")
	}
}
