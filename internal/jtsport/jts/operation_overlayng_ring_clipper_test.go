package jts

import "testing"

func TestRingClipperEmptyEnv(t *testing.T) {
	checkRingClip(t,
		"POLYGON ((2 9, 7 27, 26 34, 45 10, 26 9, 17 -7, 14 4, 2 9))",
		Geom_NewEnvelope(),
		"LINESTRING EMPTY",
	)
}

func TestRingClipperPointEnv(t *testing.T) {
	checkRingClip(t,
		"POLYGON ((2 9, 7 27, 26 34, 45 10, 26 9, 17 -7, 14 4, 2 9))",
		Geom_NewEnvelopeFromXY(10, 10, 10, 10),
		"LINESTRING EMPTY",
	)
}

func TestRingClipperClipCompletely(t *testing.T) {
	checkRingClip(t,
		"POLYGON ((2 9, 7 27, 26 34, 45 10, 26 9, 17 -7, 14 4, 2 9))",
		Geom_NewEnvelopeFromXY(10, 20, 10, 20),
		"LINESTRING (10 20, 20 20, 20 10, 10 10, 10 20)",
	)
}

func TestRingClipperInside(t *testing.T) {
	checkRingClip(t,
		"POLYGON ((12 13, 13 17, 18 17, 15 16, 17 12, 14 14, 12 13))",
		Geom_NewEnvelopeFromXY(10, 20, 10, 20),
		"LINESTRING (12 13, 13 17, 18 17, 15 16, 17 12, 14 14, 12 13)",
	)
}

func TestRingClipperStarClipped(t *testing.T) {
	checkRingClip(t,
		"POLYGON ((7 15, 12 18, 15 23, 18 18, 24 15, 18 12, 15 7, 12 12, 7 15))",
		Geom_NewEnvelopeFromXY(10, 20, 10, 20),
		"LINESTRING (10 16.8, 12 18, 13.2 20, 16.8 20, 18 18, 20 17, 20 13, 18 12, 16.8 10, 13.2 10, 12 12, 10 13.2, 10 16.8)",
	)
}

func TestRingClipperWrapPartial(t *testing.T) {
	checkRingClip(t,
		"POLYGON ((30 60, 60 60, 40 80, 40 110, 110 110, 110 80, 90 60, 120 60, 120 120, 30 120, 30 60))",
		Geom_NewEnvelopeFromXY(50, 100, 50, 100),
		"LINESTRING (50 60, 60 60, 50 70, 50 100, 100 100, 100 70, 90 60, 100 60, 100 100, 50 100, 50 60)",
	)
}

func TestRingClipperWrapAllSides(t *testing.T) {
	checkRingClip(t,
		"POLYGON ((30 80, 60 80, 60 90, 40 90, 40 110, 110 110, 110 40, 40 40, 40 59, 60 59, 60 70, 30 70, 30 30, 120 30, 120 120, 30 120, 30 80))",
		Geom_NewEnvelopeFromXY(50, 100, 50, 100),
		"LINESTRING (50 80, 60 80, 60 90, 50 90, 50 100, 100 100, 100 50, 50 50, 50 59, 60 59, 60 70, 50 70, 50 50, 100 50, 100 100, 50 100, 50 80)",
	)
}

func TestRingClipperWrapOverlap(t *testing.T) {
	checkRingClip(t,
		"POLYGON ((30 80, 60 80, 60 90, 40 90, 40 110, 110 110, 110 40, 40 40, 40 59, 30 70, 20 100, 10 100, 10 30, 120 30, 120 120, 30 120, 30 80))",
		Geom_NewEnvelopeFromXY(50, 100, 50, 100),
		"LINESTRING (50 80, 60 80, 60 90, 50 90, 50 100, 100 100, 100 50, 50 50, 100 50, 100 100, 50 100, 50 80)",
	)
}

func checkRingClip(t *testing.T, wkt string, clipEnv *Geom_Envelope, wktExpected string) {
	t.Helper()
	reader := Io_NewWKTReader()
	line, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("failed to read wkt: %v", err)
	}
	expected, err := reader.Read(wktExpected)
	if err != nil {
		t.Fatalf("failed to read wktExpected: %v", err)
	}

	clipper := OperationOverlayng_NewRingClipper(clipEnv)
	pts := clipper.Clip(line.GetCoordinates())

	result := line.GetFactory().CreateLineStringFromCoordinates(pts)
	if !result.Geom_Geometry.EqualsExact(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
