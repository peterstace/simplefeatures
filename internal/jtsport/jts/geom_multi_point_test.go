package jts_test

// NOTE: testIsSimple1/2 are commented out in the Java source.

import (
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestMultiPointGetGeometryN(t *testing.T) {
	reader := newMultiPointTestReader()
	m := mustReadMultiPoint(t, reader, "MULTIPOINT(1.111 2.222, 3.333 4.444, 3.333 4.444)")
	g := m.GetGeometryN(1)
	if !java.InstanceOf[*jts.Geom_Point](g) {
		t.Fatal("expected Point")
	}
	p := java.Cast[*jts.Geom_Point](g)
	coord := p.GetCoordinate()
	if math.Abs(coord.X-3.333) > 1e-10 {
		t.Errorf("expected x=3.333, got %v", coord.X)
	}
	if math.Abs(coord.Y-4.444) > 1e-10 {
		t.Errorf("expected y=4.444, got %v", coord.Y)
	}
}

func TestMultiPointGetEnvelope(t *testing.T) {
	reader := newMultiPointTestReader()
	m := mustReadMultiPoint(t, reader, "MULTIPOINT(1.111 2.222, 3.333 4.444, 3.333 4.444)")
	e := m.GetEnvelopeInternal()
	if math.Abs(e.GetMinX()-1.111) > 1e-10 {
		t.Errorf("expected minX=1.111, got %v", e.GetMinX())
	}
	if math.Abs(e.GetMaxX()-3.333) > 1e-10 {
		t.Errorf("expected maxX=3.333, got %v", e.GetMaxX())
	}
	if math.Abs(e.GetMinY()-2.222) > 1e-10 {
		t.Errorf("expected minY=2.222, got %v", e.GetMinY())
	}
	if math.Abs(e.GetMaxY()-4.444) > 1e-10 {
		t.Errorf("expected maxY=4.444, got %v", e.GetMaxY())
	}
}

func TestMultiPointEquals(t *testing.T) {
	reader := newMultiPointTestReader()
	m1 := mustReadMultiPoint(t, reader, "MULTIPOINT(5 6, 7 8)")
	m2 := mustReadMultiPoint(t, reader, "MULTIPOINT(5 6, 7 8)")
	if !m1.EqualsGeometry(m2.Geom_Geometry) {
		t.Error("expected m1 to equal m2")
	}
}

func newMultiPointTestReader() *jts.Io_WKTReader {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1000)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	return jts.Io_NewWKTReaderWithFactory(geometryFactory)
}

func mustReadMultiPoint(t *testing.T, reader *jts.Io_WKTReader, wkt string) *jts.Geom_MultiPoint {
	t.Helper()
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("failed to read WKT %q: %v", wkt, err)
	}
	return java.Cast[*jts.Geom_MultiPoint](geom)
}
