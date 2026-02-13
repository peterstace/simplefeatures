package jts

import (
	"testing"
)

// Tests for the effect of buffer parameter values.

func TestBufferParameterQuadSegsNeg(t *testing.T) {
	checkBufferParam(t, "LINESTRING (20 20, 80 20, 80 80)",
		10.0, -99,
		"POLYGON ((70 30, 70 80, 80 90, 90 80, 90 20, 80 10, 20 10, 10 20, 20 30, 70 30))")
}

func TestBufferParameterQuadSegs0(t *testing.T) {
	checkBufferParam(t, "LINESTRING (20 20, 80 20, 80 80)",
		10.0, 0,
		"POLYGON ((70 30, 70 80, 80 90, 90 80, 90 20, 80 10, 20 10, 10 20, 20 30, 70 30))")
}

func TestBufferParameterQuadSegs1(t *testing.T) {
	checkBufferParam(t, "LINESTRING (20 20, 80 20, 80 80)",
		10.0, 1,
		"POLYGON ((70 30, 70 80, 80 90, 90 80, 90 20, 80 10, 20 10, 10 20, 20 30, 70 30))")
}

func TestBufferParameterQuadSegs2(t *testing.T) {
	checkBufferParam(t, "LINESTRING (20 20, 80 20, 80 80)",
		10.0, 2,
		"POLYGON ((70 30, 70 80, 72.92893218813452 87.07106781186548, 80 90, 87.07106781186548 87.07106781186548, 90 80, 90 20, 87.07106781186548 12.928932188134524, 80 10, 20 10, 12.928932188134523 12.928932188134524, 10 20, 12.928932188134524 27.071067811865476, 20 30, 70 30))")
}

func TestBufferParameterQuadSegs2Bevel(t *testing.T) {
	checkBufferParamWithJoinStyle(t, "LINESTRING (20 20, 80 20, 80 80)",
		10.0, 2, OperationBuffer_BufferParameters_JOIN_BEVEL,
		"POLYGON ((70 30, 70 80, 72.92893218813452 87.07106781186548, 80 90, 87.07106781186548 87.07106781186548, 90 80, 90 20, 80 10, 20 10, 12.928932188134523 12.928932188134524, 10 20, 12.928932188134524 27.071067811865476, 20 30, 70 30))")
}

//----------------------------------------------------

func TestBufferParameterMitreRight0(t *testing.T) {
	checkBufferParamFlatMitre(t, "LINESTRING (20 20, 20 80, 80 80)",
		10.0, 0,
		"POLYGON ((10 80, 20 90, 80 90, 80 70, 30 70, 30 20, 10 20, 10 80))")
}

func TestBufferParameterMitreRight1(t *testing.T) {
	checkBufferParamFlatMitre(t, "LINESTRING (20 20, 20 80, 80 80)",
		10.0, 1,
		"POLYGON ((10 20, 10 84.14213562373095, 15.857864376269049 90, 80 90, 80 70, 30 70, 30 20, 10 20))")
}

func TestBufferParameterMitreRight2(t *testing.T) {
	checkBufferParamFlatMitre(t, "LINESTRING (20 20, 20 80, 80 80)",
		10.0, 2,
		"POLYGON ((10 20, 10 90, 80 90, 80 70, 30 70, 30 20, 10 20))")
}

func TestBufferParameterMitreNarrow0(t *testing.T) {
	checkBufferParamFlatMitre(t, "LINESTRING (10 20, 20 80, 30 20)",
		10.0, 0,
		"POLYGON ((10.136060761678563 81.64398987305357, 29.863939238321436 81.64398987305357, 39.863939238321436 21.643989873053574, 20.136060761678564 18.356010126946426, 20 19.172374697017812, 19.863939238321436 18.356010126946426, 0.1360607616785625 21.643989873053574, 10.136060761678563 81.64398987305357))")
}

func TestBufferParameterMitreNarrow1(t *testing.T) {
	checkBufferParamFlatMitre(t, "LINESTRING (10 20, 20 80, 30 20)",
		10.0, 1,
		"POLYGON ((11.528729116169634 90, 28.47127088383036 90, 39.863939238321436 21.643989873053574, 20.136060761678564 18.356010126946426, 20 19.172374697017812, 19.863939238321436 18.356010126946426, 0.1360607616785625 21.643989873053574, 11.528729116169634 90))")
}

func TestBufferParameterMitreNarrow5(t *testing.T) {
	checkBufferParamFlatMitre(t, "LINESTRING (10 20, 20 80, 30 20)",
		10.0, 5,
		"POLYGON ((18.1953957828363 130, 21.804604217163696 130, 39.863939238321436 21.643989873053574, 20.136060761678564 18.356010126946426, 20 19.172374697017812, 19.863939238321436 18.356010126946426, 0.1360607616785625 21.643989873053574, 18.1953957828363 130))")
}

func TestBufferParameterMitreNarrow10(t *testing.T) {
	checkBufferParamFlatMitre(t, "LINESTRING (10 20, 20 80, 30 20)",
		10.0, 10,
		"POLYGON ((20 140.82762530298217, 39.863939238321436 21.643989873053574, 20.136060761678564 18.356010126946426, 20 19.172374697017812, 19.863939238321436 18.356010126946426, 0.1360607616785625 21.643989873053574, 20 140.82762530298217))")
}

func TestBufferParameterMitreObtuse0(t *testing.T) {
	checkBufferParamFlatMitre(t, "LINESTRING (10 10, 50 20, 90 10)",
		1.0, 0,
		"POLYGON ((49.75746437496367 20.970142500145332, 50.24253562503633 20.970142500145332, 90.24253562503634 10.970142500145332, 89.75746437496366 9.029857499854668, 50 18.969223593595583, 10.242535625036332 9.029857499854668, 9.757464374963668 10.970142500145332, 49.75746437496367 20.970142500145332))")
}

func TestBufferParameterMitreObtuse1(t *testing.T) {
	checkBufferParamFlatMitre(t, "LINESTRING (10 10, 50 20, 90 10)",
		1.0, 1,
		"POLYGON ((9.757464374963668 10.970142500145332, 49.876894374382324 21, 50.12310562561766 20.999999999999996, 90.24253562503634 10.970142500145332, 89.75746437496366 9.029857499854668, 50 18.969223593595583, 10.242535625036332 9.029857499854668, 9.757464374963668 10.970142500145332))")
}

func TestBufferParameterMitreObtuse2(t *testing.T) {
	checkBufferParamFlatMitre(t, "LINESTRING (10 10, 50 20, 90 10)",
		1.0, 2,
		"POLYGON ((50 21.030776406404417, 90.24253562503634 10.970142500145332, 89.75746437496366 9.029857499854668, 50 18.969223593595583, 10.242535625036332 9.029857499854668, 9.757464374963668 10.970142500145332, 50 21.030776406404417))")
}

//----------------------------------------------------

func TestBufferParameterMitreSquareCCW1(t *testing.T) {
	checkBufferParamFlatMitre(t, "POLYGON((0 0, 100 0, 100 100, 0 100, 0 0))",
		10.0, 1,
		"POLYGON ((-10 -4.142135623730949, -10 104.14213562373095, -4.142135623730949 110, 104.14213562373095 110, 110 104.14213562373095, 110 -4.142135623730949, 104.14213562373095 -10, -4.142135623730949 -10, -10 -4.142135623730949))")
}

func TestBufferParameterMitreSquare1(t *testing.T) {
	checkBufferParamFlatMitre(t, "POLYGON ((0 0, 0 100, 100 100, 100 0, 0 0))",
		10.0, 1,
		"POLYGON ((-4.14213562373095 -10, -10 -4.14213562373095, -10 104.14213562373095, -4.14213562373095 110, 104.14213562373095 110, 110 104.14213562373095, 110 -4.142135623730951, 104.14213562373095 -10, -4.14213562373095 -10))")
}

//----------------------------------------------------

func checkBufferParam(t *testing.T, wkt string, dist float64, quadSegs int, wktExpected string) {
	t.Helper()
	checkBufferParamWithJoinStyle(t, wkt, dist, quadSegs, OperationBuffer_BufferParameters_JOIN_ROUND, wktExpected)
}

func checkBufferParamWithJoinStyle(t *testing.T, wkt string, dist float64, quadSegs int, joinStyle int, wktExpected string) {
	t.Helper()
	param := OperationBuffer_NewBufferParameters()
	param.SetQuadrantSegments(quadSegs)
	param.SetJoinStyle(joinStyle)
	checkBufferWithParams(t, wkt, dist, param, wktExpected)
}

func checkBufferWithParams(t *testing.T, wkt string, dist float64, param *OperationBuffer_BufferParameters, wktExpected string) {
	t.Helper()
	reader := Io_NewWKTReader()
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("failed to read input WKT: %v", err)
	}
	result := OperationBuffer_BufferOp_BufferOpWithParams(geom, dist, param)
	expected, err := reader.Read(wktExpected)
	if err != nil {
		t.Fatalf("failed to read expected WKT: %v", err)
	}
	checkGeomEqualWithTolerance(t, expected, result, 0.00001)
}

func checkBufferParamFlatMitre(t *testing.T, wkt string, dist float64, mitreLimit float64, wktExpected string) {
	t.Helper()
	param := bufParamFlatMitre(mitreLimit)
	checkBufferWithParams(t, wkt, dist, param, wktExpected)
}

func bufParamFlatMitre(mitreLimit float64) *OperationBuffer_BufferParameters {
	param := OperationBuffer_NewBufferParameters()
	param.SetJoinStyle(OperationBuffer_BufferParameters_JOIN_MITRE)
	param.SetMitreLimit(mitreLimit)
	param.SetEndCapStyle(OperationBuffer_BufferParameters_CAP_FLAT)
	return param
}

// TRANSLITERATION NOTE: This helper replaces GeometryTestCase.checkEqual() which
// is inherited by Java test classes. Go doesn't have test class inheritance, so
// the helper is defined locally.
func checkGeomEqualWithTolerance(t *testing.T, expected, actual *Geom_Geometry, tolerance float64) {
	t.Helper()
	actualNorm := actual.Norm()
	expectedNorm := expected.Norm()
	equal := actualNorm.EqualsExactWithTolerance(expectedNorm, tolerance)
	if !equal {
		t.Errorf("geometries not equal\nexpected: %v\nactual:   %v", expectedNorm, actualNorm)
	}
}
