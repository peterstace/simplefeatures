package jts

import "testing"

func TestLineLimiterEmptyEnv(t *testing.T) {
	checkLineLimit(t,
		"LINESTRING (5 15, 5 25, 25 25, 25 5, 5 5)",
		Geom_NewEnvelope(),
		"MULTILINESTRING EMPTY",
	)
}

func TestLineLimiterPointEnv(t *testing.T) {
	checkLineLimit(t,
		"LINESTRING (5 15, 5 25, 25 25, 25 5, 5 5)",
		Geom_NewEnvelopeFromXY(10, 10, 10, 10),
		"MULTILINESTRING EMPTY",
	)
}

func TestLineLimiterNonIntersecting(t *testing.T) {
	checkLineLimit(t,
		"LINESTRING (5 15, 5 25, 25 25, 25 5, 5 5)",
		Geom_NewEnvelopeFromXY(10, 20, 10, 20),
		"MULTILINESTRING EMPTY",
	)
}

func TestLineLimiterPartiallyInside(t *testing.T) {
	checkLineLimit(t,
		"LINESTRING (4 17, 8 14, 12 18, 15 15)",
		Geom_NewEnvelopeFromXY(10, 20, 10, 20),
		"LINESTRING (8 14, 12 18, 15 15)",
	)
}

func TestLineLimiterCrossing(t *testing.T) {
	checkLineLimit(t,
		"LINESTRING (5 17, 8 14, 12 18, 15 15, 18 18, 22 14, 25 18)",
		Geom_NewEnvelopeFromXY(10, 20, 10, 20),
		"LINESTRING (8 14, 12 18, 15 15, 18 18, 22 14)",
	)
}

func TestLineLimiterCrossesTwice(t *testing.T) {
	checkLineLimit(t,
		"LINESTRING (7 17, 23 17, 23 13, 7 13)",
		Geom_NewEnvelopeFromXY(10, 20, 10, 20),
		"MULTILINESTRING ((7 17, 23 17), (23 13, 7 13))",
	)
}

func TestLineLimiterDiamond(t *testing.T) {
	checkLineLimit(t,
		"LINESTRING (8 15, 15 22, 22 15, 15 8, 8 15)",
		Geom_NewEnvelopeFromXY(10, 20, 10, 20),
		"LINESTRING (8 15, 15 8, 22 15, 15 22, 8 15)",
	)
}

func TestLineLimiterOctagon(t *testing.T) {
	checkLineLimit(t,
		"LINESTRING (9 12, 12 9, 18 9, 21 12, 21 18, 18 21, 12 21, 9 18, 9 13)",
		Geom_NewEnvelopeFromXY(10, 20, 10, 20),
		"MULTILINESTRING ((9 12, 12 9), (18 9, 21 12), (21 18, 18 21), (12 21, 9 18))",
	)
}

func checkLineLimit(t *testing.T, wkt string, clipEnv *Geom_Envelope, wktExpected string) {
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

	limiter := OperationOverlayng_NewLineLimiter(clipEnv)
	sections := limiter.Limit(line.GetCoordinates())

	result := lineLimiterTestToLines(sections, line.GetFactory())
	resultNorm := result.Norm()
	expectedNorm := expected.Norm()
	if !resultNorm.EqualsExact(expectedNorm) {
		t.Errorf("expected %v, got %v", expectedNorm, resultNorm)
	}
}

func lineLimiterTestToLines(sections [][]*Geom_Coordinate, factory *Geom_GeometryFactory) *Geom_Geometry {
	lines := make([]*Geom_LineString, len(sections))
	for i, pts := range sections {
		lines[i] = factory.CreateLineStringFromCoordinates(pts)
	}
	if len(lines) == 1 {
		return lines[0].Geom_Geometry
	}
	return factory.CreateMultiLineStringFromLineStrings(lines).Geom_Geometry
}
