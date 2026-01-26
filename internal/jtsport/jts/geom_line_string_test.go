package jts_test

// Tests ported from org.locationtech.jts.geom.LineStringImplTest.java

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestLineStringIsCoordinate(t *testing.T) {
	reader := newLineStringTestReader()
	l := mustReadLineString(t, reader, "LINESTRING (0 0, 10 10, 10 0)")
	if !l.IsCoordinate(jts.Geom_NewCoordinateWithXY(0, 0)) {
		t.Error("expected (0,0) to be on linestring")
	}
	if l.IsCoordinate(jts.Geom_NewCoordinateWithXY(5, 0)) {
		t.Error("expected (5,0) to not be on linestring")
	}
}

func TestLineStringUnclosedLinearRing(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(
		jts.Geom_NewPrecisionModelWithScale(1000), 0)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for unclosed linear ring")
		}
	}()

	factory.CreateLinearRingFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(1, 0),
		jts.Geom_NewCoordinateWithXY(1, 1),
		jts.Geom_NewCoordinateWithXY(2, 1),
	})
}

func TestLineStringGetCoordinates(t *testing.T) {
	reader := newLineStringTestReader()
	l := mustReadLineString(t, reader, "LINESTRING(1.111 2.222, 5.555 6.666, 3.333 4.444)")
	coordinates := l.GetCoordinates()
	if len(coordinates) != 3 {
		t.Fatalf("expected 3 coordinates, got %d", len(coordinates))
	}
	c := coordinates[1]
	if c.X != 5.555 || c.Y != 6.666 {
		t.Errorf("expected (5.555, 6.666), got (%v, %v)", c.X, c.Y)
	}
}

func TestLineStringIsClosed(t *testing.T) {
	reader := newLineStringTestReader()
	factory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(
		jts.Geom_NewPrecisionModelWithScale(1000), 0)

	l := mustReadLineString(t, reader, "LINESTRING EMPTY")
	if !l.IsEmpty() {
		t.Error("expected empty linestring to be empty")
	}
	if l.IsClosed() {
		t.Error("expected empty linestring to not be closed")
	}

	r := factory.CreateLinearRingFromCoordinateSequence(nil)
	if !r.IsEmpty() {
		t.Error("expected empty linear ring to be empty")
	}
	if !r.IsClosed() {
		t.Error("expected empty linear ring to be closed")
	}

	m := factory.CreateMultiLineStringFromLineStrings([]*jts.Geom_LineString{l, r.Geom_LineString})
	if m.IsClosed() {
		t.Error("expected multilinestring with non-closed element to not be closed")
	}

	m2 := factory.CreateMultiLineStringFromLineStrings([]*jts.Geom_LineString{r.Geom_LineString})
	if m2.IsClosed() {
		t.Error("expected multilinestring with single empty ring to not be closed")
	}
}

func TestLineStringGetGeometryType(t *testing.T) {
	reader := newLineStringTestReader()
	l := mustReadLineString(t, reader, "LINESTRING EMPTY")
	if got := l.GetGeometryType(); got != "LineString" {
		t.Errorf("expected 'LineString', got %q", got)
	}
}

func TestLineStringFiveZeros(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()
	ls := factory.CreateLineStringFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(0, 0),
	})
	if !ls.IsClosed() {
		t.Error("expected linestring with identical endpoints to be closed")
	}
}

func TestLineStringEquals1(t *testing.T) {
	reader := newLineStringTestReader()
	l1 := mustReadLineString(t, reader, "LINESTRING(1.111 2.222, 3.333 4.444)")
	l2 := mustReadLineString(t, reader, "LINESTRING(1.111 2.222, 3.333 4.444)")
	if !l1.EqualsGeometry(l2.Geom_Geometry) {
		t.Error("expected l1 to equal l2")
	}
}

func TestLineStringEquals2(t *testing.T) {
	reader := newLineStringTestReader()
	// Reversed coordinates - should still be topologically equal.
	l1 := mustReadLineString(t, reader, "LINESTRING(1.111 2.222, 3.333 4.444)")
	l2 := mustReadLineString(t, reader, "LINESTRING(3.333 4.444, 1.111 2.222)")
	if !l1.EqualsGeometry(l2.Geom_Geometry) {
		t.Error("expected l1 to equal l2 (reversed)")
	}
}

func TestLineStringEquals3(t *testing.T) {
	reader := newLineStringTestReader()
	// Different Y coordinate (4.444 vs 4.443).
	l1 := mustReadLineString(t, reader, "LINESTRING(1.111 2.222, 3.333 4.444)")
	l2 := mustReadLineString(t, reader, "LINESTRING(3.333 4.443, 1.111 2.222)")
	if l1.EqualsGeometry(l2.Geom_Geometry) {
		t.Error("expected l1 to NOT equal l2")
	}
}

func TestLineStringEquals4(t *testing.T) {
	reader := newLineStringTestReader()
	// 4.4445 rounds to 4.445, different from 4.444.
	l1 := mustReadLineString(t, reader, "LINESTRING(1.111 2.222, 3.333 4.444)")
	l2 := mustReadLineString(t, reader, "LINESTRING(3.333 4.4445, 1.111 2.222)")
	if l1.EqualsGeometry(l2.Geom_Geometry) {
		t.Error("expected l1 to NOT equal l2")
	}
}

func TestLineStringEquals5(t *testing.T) {
	reader := newLineStringTestReader()
	// 4.4446 rounds to 4.445, different from 4.444.
	l1 := mustReadLineString(t, reader, "LINESTRING(1.111 2.222, 3.333 4.444)")
	l2 := mustReadLineString(t, reader, "LINESTRING(3.333 4.4446, 1.111 2.222)")
	if l1.EqualsGeometry(l2.Geom_Geometry) {
		t.Error("expected l1 to NOT equal l2")
	}
}

func TestLineStringEquals6(t *testing.T) {
	reader := newLineStringTestReader()
	// Three-point linestring, same coordinates.
	l1 := mustReadLineString(t, reader, "LINESTRING(1.111 2.222, 3.333 4.444, 5.555 6.666)")
	l2 := mustReadLineString(t, reader, "LINESTRING(1.111 2.222, 3.333 4.444, 5.555 6.666)")
	if !l1.EqualsGeometry(l2.Geom_Geometry) {
		t.Error("expected l1 to equal l2")
	}
}

func TestLineStringEquals7(t *testing.T) {
	reader := newLineStringTestReader()
	// Different point order (not just reversed).
	l1 := mustReadLineString(t, reader, "LINESTRING(1.111 2.222, 5.555 6.666, 3.333 4.444)")
	l2 := mustReadLineString(t, reader, "LINESTRING(1.111 2.222, 3.333 4.444, 5.555 6.666)")
	if l1.EqualsGeometry(l2.Geom_Geometry) {
		t.Error("expected l1 to NOT equal l2")
	}
}

func TestLineStringEquals8(t *testing.T) {
	// MultiLineString with closed ring, different starting points.
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1000)
	factory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	reader := jts.Io_NewWKTReaderWithFactory(factory)

	l1, err := reader.Read("MULTILINESTRING((1732328800 519578384, 1732026179 519976285, 1731627364 519674014, 1731929984 519276112, 1732328800 519578384))")
	if err != nil {
		t.Fatalf("failed to read l1: %v", err)
	}
	l2, err := reader.Read("MULTILINESTRING((1731627364 519674014, 1731929984 519276112, 1732328800 519578384, 1732026179 519976285, 1731627364 519674014))")
	if err != nil {
		t.Fatalf("failed to read l2: %v", err)
	}
	if !l1.EqualsGeometry(l2) {
		t.Error("expected l1 to equal l2")
	}
}

func TestLineStringEquals9(t *testing.T) {
	// Same as equals8 but with precision 1.
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1)
	factory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	reader := jts.Io_NewWKTReaderWithFactory(factory)

	l1, err := reader.Read("MULTILINESTRING((1732328800 519578384, 1732026179 519976285, 1731627364 519674014, 1731929984 519276112, 1732328800 519578384))")
	if err != nil {
		t.Fatalf("failed to read l1: %v", err)
	}
	l2, err := reader.Read("MULTILINESTRING((1731627364 519674014, 1731929984 519276112, 1732328800 519578384, 1732026179 519976285, 1731627364 519674014))")
	if err != nil {
		t.Fatalf("failed to read l2: %v", err)
	}
	if !l1.EqualsGeometry(l2) {
		t.Error("expected l1 to equal l2")
	}
}

func TestLineStringEquals10(t *testing.T) {
	// Polygon with different starting vertex, normalize then equalsExact.
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1)
	factory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	reader := jts.Io_NewWKTReaderWithFactory(factory)

	l1, err := reader.Read("POLYGON((1732328800 519578384, 1732026179 519976285, 1731627364 519674014, 1731929984 519276112, 1732328800 519578384))")
	if err != nil {
		t.Fatalf("failed to read l1: %v", err)
	}
	l2, err := reader.Read("POLYGON((1731627364 519674014, 1731929984 519276112, 1732328800 519578384, 1732026179 519976285, 1731627364 519674014))")
	if err != nil {
		t.Fatalf("failed to read l2: %v", err)
	}
	l1.Normalize()
	l2.Normalize()
	if !l1.EqualsExact(l2) {
		t.Error("expected normalized l1 to equalsExact normalized l2")
	}
}

func TestLinearRingConstructor(t *testing.T) {
	factory := jts.Geom_NewGeometryFactoryDefault()
	ring := factory.CreateLinearRingFromCoordinates([]*jts.Geom_Coordinate{
		jts.Geom_NewCoordinateWithXY(0, 0),
		jts.Geom_NewCoordinateWithXY(10, 10),
		jts.Geom_NewCoordinateWithXY(0, 0),
	})

	reader := jts.Io_NewWKTReaderWithFactory(factory)
	ringFromWKT, err := reader.Read("LINEARRING (0 0, 10 10, 0 0)")
	if err != nil {
		t.Fatalf("failed to read WKT: %v", err)
	}

	// checkEqual normalizes and compares with equalsExact.
	if !ring.Geom_Geometry.EqualsNorm(ringFromWKT) {
		t.Error("expected ring to equal ringFromWKT after normalization")
	}
}

func newLineStringTestReader() *jts.Io_WKTReader {
	precisionModel := jts.Geom_NewPrecisionModelWithScale(1000)
	geometryFactory := jts.Geom_NewGeometryFactoryWithPrecisionModelAndSRID(precisionModel, 0)
	return jts.Io_NewWKTReaderWithFactory(geometryFactory)
}

func mustReadLineString(t *testing.T, reader *jts.Io_WKTReader, wkt string) *jts.Geom_LineString {
	t.Helper()
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("failed to read WKT %q: %v", wkt, err)
	}
	return java.Cast[*jts.Geom_LineString](geom)
}

func TestLineStringIsSimple(t *testing.T) {
	reader := jts.Io_NewWKTReader()

	// Self-intersecting linestring (figure-8 shape).
	l1, err := reader.Read("LINESTRING (0 0, 10 10, 10 0, 0 10, 0 0)")
	if err != nil {
		t.Fatalf("failed to read l1: %v", err)
	}
	if l1.IsSimple() {
		t.Error("expected self-intersecting linestring to NOT be simple")
	}

	// Self-intersecting linestring (X shape, not closed).
	l2, err := reader.Read("LINESTRING (0 0, 10 10, 10 0, 0 10)")
	if err != nil {
		t.Fatalf("failed to read l2: %v", err)
	}
	if l2.IsSimple() {
		t.Error("expected self-intersecting linestring to NOT be simple")
	}
}
