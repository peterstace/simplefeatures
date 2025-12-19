package jts

import "testing"

// Helper functions for OverlayNG tests.

func overlayNGTestIntersection(t *testing.T, a, b *Geom_Geometry, scaleFactor float64) *Geom_Geometry {
	t.Helper()
	pm := Geom_NewPrecisionModelWithScale(scaleFactor)
	return OperationOverlayng_OverlayNG_Overlay(a, b, OperationOverlayng_OverlayNG_INTERSECTION, pm)
}

func overlayNGTestUnion(t *testing.T, a, b *Geom_Geometry, scaleFactor float64) *Geom_Geometry {
	t.Helper()
	pm := Geom_NewPrecisionModelWithScale(scaleFactor)
	return OperationOverlayng_OverlayNG_Overlay(a, b, OperationOverlayng_OverlayNG_UNION, pm)
}

func overlayNGTestDifference(t *testing.T, a, b *Geom_Geometry, scaleFactor float64) *Geom_Geometry {
	t.Helper()
	pm := Geom_NewPrecisionModelWithScale(scaleFactor)
	return OperationOverlayng_OverlayNG_Overlay(a, b, OperationOverlayng_OverlayNG_DIFFERENCE, pm)
}

func overlayNGTestIntersectionFloat(t *testing.T, a, b *Geom_Geometry) *Geom_Geometry {
	t.Helper()
	pm := Geom_NewPrecisionModel()
	return OperationOverlayng_OverlayNG_Overlay(a, b, OperationOverlayng_OverlayNG_INTERSECTION, pm)
}

func overlayNGTestUnionFloat(t *testing.T, a, b *Geom_Geometry) *Geom_Geometry {
	t.Helper()
	pm := Geom_NewPrecisionModel()
	return OperationOverlayng_OverlayNG_Overlay(a, b, OperationOverlayng_OverlayNG_UNION, pm)
}

func checkEqualGeomsNormalized(t *testing.T, expected, actual *Geom_Geometry) {
	t.Helper()
	expected.Normalize()
	actual.Normalize()
	if !expected.EqualsExactWithTolerance(actual, 0.0) {
		t.Errorf("geometries not equal:\nexpected: %v\nactual:   %v",
			expected.ToText(), actual.ToText())
	}
}

func checkEqualGeomsExact(t *testing.T, expected, actual *Geom_Geometry) {
	t.Helper()
	if !expected.EqualsExactWithTolerance(actual, 0.0) {
		t.Errorf("geometries not equal:\nexpected: %v\nactual:   %v",
			expected.ToText(), actual.ToText())
	}
}

func TestOverlayNGAreaLineIntersection(t *testing.T) {
	a := readWKT(t, "POLYGON ((360 200, 220 200, 220 180, 300 180, 300 160, 300 140, 360 200))")
	b := readWKT(t, "MULTIPOLYGON (((280 180, 280 160, 300 160, 300 180, 280 180)), ((220 230, 240 230, 240 180, 220 180, 220 230)))")
	expected := readWKT(t, "GEOMETRYCOLLECTION (LINESTRING (280 180, 300 180), LINESTRING (300 160, 300 180), POLYGON ((220 180, 220 200, 240 200, 240 180, 220 180)))")
	actual := overlayNGTestIntersection(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGAreaLinePointIntersection(t *testing.T) {
	a := readWKT(t, "POLYGON ((100 100, 200 100, 200 150, 250 100, 300 100, 300 150, 350 100, 350 200, 100 200, 100 100))")
	b := readWKT(t, "POLYGON ((100 140, 170 140, 200 100, 400 100, 400 30, 100 30, 100 140))")
	expected := readWKT(t, "GEOMETRYCOLLECTION (POINT (350 100), LINESTRING (250 100, 300 100), POLYGON ((100 100, 100 140, 170 140, 200 100, 100 100)))")
	actual := overlayNGTestIntersection(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGTriangleFillingHoleUnion(t *testing.T) {
	a := readWKT(t, "POLYGON ((0 0, 4 0, 4 4, 0 4, 0 0), (1 1, 1 2, 2 1, 1 1), (1 2, 1 3, 2 3, 1 2), (2 3, 3 3, 3 2, 2 3))")
	b := readWKT(t, "POLYGON ((2 1, 3 1, 3 2, 2 1))")
	expected := readWKT(t, "POLYGON ((0 0, 0 4, 4 4, 4 0, 0 0), (1 2, 1 1, 2 1, 1 2), (2 3, 1 3, 1 2, 2 3))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGTriangleFillingHoleUnionPrec10(t *testing.T) {
	a := readWKT(t, "POLYGON ((0 0, 4 0, 4 4, 0 4, 0 0), (1 1, 1 2, 2 1, 1 1), (1 2, 1 3, 2 3, 1 2), (2 3, 3 3, 3 2, 2 3))")
	b := readWKT(t, "POLYGON ((2 1, 3 1, 3 2, 2 1))")
	expected := readWKT(t, "POLYGON ((0 0, 0 4, 4 4, 4 0, 0 0), (1 2, 1 1, 2 1, 1 2), (2 3, 1 3, 1 2, 2 3), (3 2, 3 3, 2 3, 3 2))")
	actual := overlayNGTestUnion(t, a, b, 10)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGBoxTriIntersection(t *testing.T) {
	a := readWKT(t, "POLYGON ((0 6, 4 6, 4 2, 0 2, 0 6))")
	b := readWKT(t, "POLYGON ((1 0, 2 5, 3 0, 1 0))")
	expected := readWKT(t, "POLYGON ((3 2, 1 2, 2 5, 3 2))")
	actual := overlayNGTestIntersection(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGBoxTriUnion(t *testing.T) {
	a := readWKT(t, "POLYGON ((0 6, 4 6, 4 2, 0 2, 0 6))")
	b := readWKT(t, "POLYGON ((1 0, 2 5, 3 0, 1 0))")
	expected := readWKT(t, "POLYGON ((0 6, 4 6, 4 2, 3 2, 3 0, 1 0, 1 2, 0 2, 0 6))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNG2SpikesIntersection(t *testing.T) {
	a := readWKT(t, "POLYGON ((0 100, 40 100, 40 0, 0 0, 0 100))")
	b := readWKT(t, "POLYGON ((70 80, 10 80, 60 50, 11 20, 69 11, 70 80))")
	expected := readWKT(t, "MULTIPOLYGON (((40 80, 40 62, 10 80, 40 80)), ((40 38, 40 16, 11 20, 40 38)))")
	actual := overlayNGTestIntersection(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNG2SpikesUnion(t *testing.T) {
	a := readWKT(t, "POLYGON ((0 100, 40 100, 40 0, 0 0, 0 100))")
	b := readWKT(t, "POLYGON ((70 80, 10 80, 60 50, 11 20, 69 11, 70 80))")
	expected := readWKT(t, "POLYGON ((0 100, 40 100, 40 80, 70 80, 69 11, 40 16, 40 0, 0 0, 0 100), (40 62, 40 38, 60 50, 40 62))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGTriBoxIntersection(t *testing.T) {
	a := readWKT(t, "POLYGON ((68 35, 35 42, 40 9, 68 35))")
	b := readWKT(t, "POLYGON ((20 60, 50 60, 50 30, 20 30, 20 60))")
	expected := readWKT(t, "POLYGON ((37 30, 35 42, 50 39, 50 30, 37 30))")
	actual := overlayNGTestIntersection(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGNestedShellsIntersection(t *testing.T) {
	a := readWKT(t, "POLYGON ((100 200, 200 200, 200 100, 100 100, 100 200))")
	b := readWKT(t, "POLYGON ((120 180, 180 180, 180 120, 120 120, 120 180))")
	expected := readWKT(t, "POLYGON ((120 180, 180 180, 180 120, 120 120, 120 180))")
	actual := overlayNGTestIntersection(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGNestedShellsUnion(t *testing.T) {
	a := readWKT(t, "POLYGON ((100 200, 200 200, 200 100, 100 100, 100 200))")
	b := readWKT(t, "POLYGON ((120 180, 180 180, 180 120, 120 120, 120 180))")
	expected := readWKT(t, "POLYGON ((100 200, 200 200, 200 100, 100 100, 100 200))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGATouchingNestedPolyUnion(t *testing.T) {
	a := readWKT(t, "MULTIPOLYGON (((0 200, 200 200, 200 0, 0 0, 0 200), (50 50, 190 50, 50 200, 50 50)), ((60 100, 100 60, 50 50, 60 100)))")
	b := readWKT(t, "POLYGON ((135 176, 180 176, 180 130, 135 130, 135 176))")
	expected := readWKT(t, "MULTIPOLYGON (((0 0, 0 200, 50 200, 200 200, 200 0, 0 0), (50 50, 190 50, 50 200, 50 50)), ((50 50, 60 100, 100 60, 50 50)))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGTouchingPolyDifference(t *testing.T) {
	a := readWKT(t, "POLYGON ((200 200, 200 0, 0 0, 0 200, 200 200), (100 100, 50 100, 50 200, 100 100))")
	b := readWKT(t, "POLYGON ((150 100, 100 100, 150 200, 150 100))")
	expected := readWKT(t, "MULTIPOLYGON (((0 0, 0 200, 50 200, 50 100, 100 100, 150 100, 150 200, 200 200, 200 0, 0 0)), ((50 200, 150 200, 100 100, 50 200)))")
	actual := overlayNGTestDifference(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGTouchingHoleUnion(t *testing.T) {
	a := readWKT(t, "POLYGON ((100 300, 300 300, 300 100, 100 100, 100 300), (200 200, 150 200, 200 300, 200 200))")
	b := readWKT(t, "POLYGON ((130 160, 260 160, 260 120, 130 120, 130 160))")
	expected := readWKT(t, "POLYGON ((100 100, 100 300, 200 300, 300 300, 300 100, 100 100), (150 200, 200 200, 200 300, 150 200))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGTouchingMultiHoleUnion(t *testing.T) {
	a := readWKT(t, "POLYGON ((100 300, 300 300, 300 100, 100 100, 100 300), (200 200, 150 200, 200 300, 200 200), (250 230, 216 236, 250 300, 250 230), (235 198, 300 200, 237 175, 235 198))")
	b := readWKT(t, "POLYGON ((130 160, 260 160, 260 120, 130 120, 130 160))")
	expected := readWKT(t, "POLYGON ((100 300, 200 300, 250 300, 300 300, 300 200, 300 100, 100 100, 100 300), (200 300, 150 200, 200 200, 200 300), (250 300, 216 236, 250 230, 250 300), (300 200, 235 198, 237 175, 300 200))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGBoxLineIntersection(t *testing.T) {
	a := readWKT(t, "POLYGON ((100 200, 200 200, 200 100, 100 100, 100 200))")
	b := readWKT(t, "LINESTRING (50 150, 150 150)")
	expected := readWKT(t, "LINESTRING (100 150, 150 150)")
	actual := overlayNGTestIntersection(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGBoxLineUnion(t *testing.T) {
	a := readWKT(t, "POLYGON ((100 200, 200 200, 200 100, 100 100, 100 200))")
	b := readWKT(t, "LINESTRING (50 150, 150 150)")
	expected := readWKT(t, "GEOMETRYCOLLECTION (POLYGON ((200 200, 200 100, 100 100, 100 150, 100 200, 200 200)), LINESTRING (50 150, 100 150))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGAdjacentBoxesIntersection(t *testing.T) {
	a := readWKT(t, "POLYGON ((100 200, 200 200, 200 100, 100 100, 100 200))")
	b := readWKT(t, "POLYGON ((300 200, 300 100, 200 100, 200 200, 300 200))")
	expected := readWKT(t, "LINESTRING (200 100, 200 200)")
	actual := overlayNGTestIntersection(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGAdjacentBoxesUnion(t *testing.T) {
	a := readWKT(t, "POLYGON ((100 200, 200 200, 200 100, 100 100, 100 200))")
	b := readWKT(t, "POLYGON ((300 200, 300 100, 200 100, 200 200, 300 200))")
	expected := readWKT(t, "POLYGON ((100 100, 100 200, 200 200, 300 200, 300 100, 200 100, 100 100))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGCollapseBoxGoreIntersection(t *testing.T) {
	a := readWKT(t, "MULTIPOLYGON (((1 1, 5 1, 5 0, 1 0, 1 1)), ((1 1, 5 2, 5 4, 1 4, 1 1)))")
	b := readWKT(t, "POLYGON ((1 0, 1 2, 2 2, 2 0, 1 0))")
	expected := readWKT(t, "POLYGON ((2 0, 1 0, 1 1, 1 2, 2 2, 2 1, 2 0))")
	actual := overlayNGTestIntersection(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGCollapseBoxGoreUnion(t *testing.T) {
	a := readWKT(t, "MULTIPOLYGON (((1 1, 5 1, 5 0, 1 0, 1 1)), ((1 1, 5 2, 5 4, 1 4, 1 1)))")
	b := readWKT(t, "POLYGON ((1 0, 1 2, 2 2, 2 0, 1 0))")
	expected := readWKT(t, "POLYGON ((2 0, 1 0, 1 1, 1 2, 1 4, 5 4, 5 2, 2 1, 5 1, 5 0, 2 0))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGSnapBoxGoreIntersection(t *testing.T) {
	a := readWKT(t, "MULTIPOLYGON (((1 1, 5 1, 5 0, 1 0, 1 1)), ((1 1, 5 2, 5 4, 1 4, 1 1)))")
	b := readWKT(t, "POLYGON ((4 3, 5 3, 5 0, 4 0, 4 3))")
	expected := readWKT(t, "MULTIPOLYGON (((4 3, 5 3, 5 2, 4 2, 4 3)), ((4 0, 4 1, 5 1, 5 0, 4 0)))")
	actual := overlayNGTestIntersection(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGSnapBoxGoreUnion(t *testing.T) {
	a := readWKT(t, "MULTIPOLYGON (((1 1, 5 1, 5 0, 1 0, 1 1)), ((1 1, 5 2, 5 4, 1 4, 1 1)))")
	b := readWKT(t, "POLYGON ((4 3, 5 3, 5 0, 4 0, 4 3))")
	expected := readWKT(t, "POLYGON ((1 1, 1 4, 5 4, 5 3, 5 2, 5 1, 5 0, 4 0, 1 0, 1 1), (1 1, 4 1, 4 2, 1 1))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGCollapseTriBoxIntersection(t *testing.T) {
	a := readWKT(t, "POLYGON ((1 2, 1 1, 9 1, 1 2))")
	b := readWKT(t, "POLYGON ((9 2, 9 1, 8 1, 8 2, 9 2))")
	expected := readWKT(t, "LINESTRING (8 1, 9 1)")
	actual := overlayNGTestIntersection(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGCollapseTriBoxUnion(t *testing.T) {
	a := readWKT(t, "POLYGON ((1 2, 1 1, 9 1, 1 2))")
	b := readWKT(t, "POLYGON ((9 2, 9 1, 8 1, 8 2, 9 2))")
	expected := readWKT(t, "MULTIPOLYGON (((1 1, 1 2, 8 1, 1 1)), ((8 1, 8 2, 9 2, 9 1, 8 1)))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGCollapseAIncompleteRingUnion(t *testing.T) {
	a := readWKT(t, "POLYGON ((0.9 1.7, 1.3 1.4, 2.1 1.4, 2.1 0.9, 1.3 0.9, 0.9 0, 0.9 1.7))")
	b := readWKT(t, "POLYGON ((1 3, 3 3, 3 1, 1.3 0.9, 1 0.4, 1 3))")
	expected := readWKT(t, "GEOMETRYCOLLECTION (LINESTRING (1 0, 1 1), POLYGON ((1 1, 1 2, 1 3, 3 3, 3 1, 2 1, 1 1)))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGCollapseResultShouldHavePolygonUnion(t *testing.T) {
	a := readWKT(t, "POLYGON ((1 3.3, 1.3 1.4, 3.1 1.4, 3.1 0.9, 1.3 0.9, 1 -0.2, 0.8 1.3, 1 3.3))")
	b := readWKT(t, "POLYGON ((1 2.9, 2.9 2.9, 2.9 1.3, 1.7 1, 1.3 0.9, 1 0.4, 1 2.9))")
	expected := readWKT(t, "GEOMETRYCOLLECTION (LINESTRING (1 0, 1 1), POLYGON ((1 1, 1 3, 3 3, 3 1, 2 1, 1 1)))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGCollapseHoleAlongEdgeOfBIntersection(t *testing.T) {
	a := readWKT(t, "POLYGON ((0 3, 3 3, 3 0, 0 0, 0 3), (1 1.2, 1 1.1, 2.3 1.1, 1 1.2))")
	b := readWKT(t, "POLYGON ((1 1, 2 1, 2 0, 1 0, 1 1))")
	expected := readWKT(t, "POLYGON ((1 1, 2 1, 2 0, 1 0, 1 1))")
	actual := overlayNGTestIntersection(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGCollapseHolesAlongAllEdgesOfBIntersection(t *testing.T) {
	a := readWKT(t, "POLYGON ((0 3, 3 3, 3 0, 0 0, 0 3), (1 2.2, 1 2.1, 2 2.1, 1 2.2), (2.1 2, 2.2 2, 2.1 1, 2.1 2), (2 0.9, 2 0.8, 1 0.9, 2 0.9), (0.9 1, 0.8 1, 0.9 2, 0.9 1))")
	b := readWKT(t, "POLYGON ((1 2, 2 2, 2 1, 1 1, 1 2))")
	expected := readWKT(t, "POLYGON ((1 2, 2 2, 2 1, 1 1, 1 2))")
	actual := overlayNGTestIntersection(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGVerySmallBIntersection(t *testing.T) {
	a := readWKT(t, "POLYGON ((2.526855443750341 48.82324221874807, 2.5258255 48.8235855, 2.5251389 48.8242722, 2.5241089 48.8246155, 2.5254822 48.8246155, 2.5265121 48.8242722, 2.526855443750341 48.82324221874807))")
	b := readWKT(t, "POLYGON ((2.526512100000002 48.824272199999996, 2.5265120999999953 48.8242722, 2.5265121 48.8242722, 2.526512100000002 48.824272199999996))")
	expected := readWKT(t, "POLYGON EMPTY")
	actual := overlayNGTestIntersection(t, a, b, 100000000)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGEdgeDisappears(t *testing.T) {
	a := readWKT(t, "LINESTRING (2.1279144 48.8445282, 2.126884443750796 48.84555818124935, 2.1268845 48.8455582, 2.1268845 48.8462448)")
	b := readWKT(t, "LINESTRING EMPTY")
	expected := readWKT(t, "LINESTRING EMPTY")
	actual := overlayNGTestIntersection(t, a, b, 1000000)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGBcollapseLocateIssue(t *testing.T) {
	a := readWKT(t, "POLYGON ((2.3442078 48.9331054, 2.3435211 48.9337921, 2.3428345 48.9358521, 2.3428345 48.9372253, 2.3433495 48.9370537, 2.3440361 48.936367, 2.3442078 48.9358521, 2.3442078 48.9331054))")
	b := readWKT(t, "POLYGON ((2.3442078 48.9331054, 2.3435211 48.9337921, 2.3433494499999985 48.934307100000005, 2.3438644 48.9341354, 2.3442078 48.9331055, 2.3442078 48.9331054))")
	expected := readWKT(t, "MULTILINESTRING ((2.343 48.934, 2.344 48.934), (2.344 48.933, 2.344 48.934))")
	actual := overlayNGTestIntersection(t, a, b, 1000)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGBcollapseEdgeLabeledInterior(t *testing.T) {
	a := readWKT(t, "POLYGON ((2.384376506250038 48.91765596875102, 2.3840332 48.916626, 2.3840332 48.9138794, 2.3833466 48.9118195, 2.3812866 48.9111328, 2.37854 48.9111328, 2.3764801 48.9118195, 2.3723602 48.9159393, 2.3703003 48.916626, 2.3723602 48.9173126, 2.3737335 48.9186859, 2.3757935 48.9193726, 2.3812866 48.9193726, 2.3833466 48.9186859, 2.384376506250038 48.91765596875102))")
	b := readWKT(t, "MULTIPOLYGON (((2.3751067666731345 48.919143677778855, 2.3757935 48.9193726, 2.3812866 48.9193726, 2.3812866 48.9179993, 2.3809433 48.9169693, 2.3799133 48.916626, 2.3771667 48.916626, 2.3761368 48.9169693, 2.3754501 48.9190292, 2.3751067666731345 48.919143677778855)), ((2.3826108673454116 48.91893115612326, 2.3833466 48.9186859, 2.3840331750033394 48.91799930833141, 2.3830032 48.9183426, 2.3826108673454116 48.91893115612326)))")
	expected := readWKT(t, "POLYGON ((2.375 48.91833333333334, 2.375 48.92, 2.381666666666667 48.92, 2.381666666666667 48.91833333333334, 2.381666666666667 48.916666666666664, 2.38 48.916666666666664, 2.3766666666666665 48.916666666666664, 2.375 48.91833333333334))")
	actual := overlayNGTestIntersection(t, a, b, 600)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGBNearVertexSnappingCausesInversion(t *testing.T) {
	a := readWKT(t, "POLYGON ((2.2494507 48.8864136, 2.2484207 48.8867569, 2.2477341 48.8874435, 2.2470474 48.8874435, 2.2463608 48.8853836, 2.2453308 48.8850403, 2.2439575 48.8850403, 2.2429276 48.8853836, 2.2422409 48.8860703, 2.2360611 48.8970566, 2.2504807 48.8956833, 2.2494507 48.8864136))")
	b := readWKT(t, "POLYGON ((2.247734099999997 48.8874435, 2.2467041 48.8877869, 2.2453308 48.8877869, 2.2443008 48.8881302, 2.243957512499544 48.888473487500455, 2.2443008 48.8888168, 2.2453308 48.8891602, 2.2463608 48.8888168, 2.247734099999997 48.8874435))")
	expected := readWKT(t, "LINESTRING (2.245 48.89, 2.25 48.885)")
	actual := overlayNGTestIntersection(t, a, b, 200)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGBCollapsedHoleEdgeLabelledExterior(t *testing.T) {
	a := readWKT(t, "POLYGON ((309500 3477900, 309900 3477900, 309900 3477600, 309500 3477600, 309500 3477900), (309741.87561330193 3477680.6737848604, 309745.53718649445 3477677.607851833, 309779.0333599192 3477653.585555199, 309796.8051681937 3477642.143583868, 309741.87561330193 3477680.6737848604))")
	b := readWKT(t, "POLYGON ((309500 3477900, 309900 3477900, 309900 3477600, 309500 3477600, 309500 3477900), (309636.40806633036 3477777.2910157656, 309692.56085444096 3477721.966349552, 309745.53718649445 3477677.607851833, 309779.0333599192 3477653.585555199, 309792.0991800499 3477645.1734264474, 309779.03383125085 3477653.5853248164, 309745.53756275156 3477677.6076231804, 309692.5613257677 3477721.966119165, 309636.40806633036 3477777.2910157656))")
	expected := readWKT(t, "POLYGON ((309500 3477600, 309500 3477900, 309900 3477900, 309900 3477600, 309500 3477600), (309741.88 3477680.67, 309745.54 3477677.61, 309779.03 3477653.59, 309792.1 3477645.17, 309796.81 3477642.14, 309741.88 3477680.67))")
	actual := overlayNGTestIntersection(t, a, b, 100)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGLineUnion(t *testing.T) {
	a := readWKT(t, "LINESTRING (0 0, 1 1)")
	b := readWKT(t, "LINESTRING (1 1, 2 2)")
	expected := readWKT(t, "MULTILINESTRING ((0 0, 1 1), (1 1, 2 2))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGLine2Union(t *testing.T) {
	a := readWKT(t, "LINESTRING (0 0, 1 1, 0 1)")
	b := readWKT(t, "LINESTRING (1 1, 2 2, 3 3)")
	expected := readWKT(t, "MULTILINESTRING ((0 0, 1 1), (0 1, 1 1), (1 1, 2 2, 3 3))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGLine3Union(t *testing.T) {
	a := readWKT(t, "MULTILINESTRING ((0 1, 1 1), (2 2, 2 0))")
	b := readWKT(t, "LINESTRING (0 0, 1 1, 2 2, 3 3)")
	expected := readWKT(t, "MULTILINESTRING ((0 0, 1 1), (0 1, 1 1), (1 1, 2 2), (2 0, 2 2), (2 2, 3 3))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGLine4Union(t *testing.T) {
	a := readWKT(t, "LINESTRING (100 300, 200 300, 200 100, 100 100)")
	b := readWKT(t, "LINESTRING (300 300, 200 300, 200 300, 200 100, 300 100)")
	expected := readWKT(t, "MULTILINESTRING ((200 100, 100 100), (300 300, 200 300), (200 300, 200 100), (200 100, 300 100), (100 300, 200 300))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGLineFigure8Union(t *testing.T) {
	a := readWKT(t, "LINESTRING (5 1, 2 2, 5 3, 2 4, 5 5)")
	b := readWKT(t, "LINESTRING (5 1, 8 2, 5 3, 8 4, 5 5)")
	expected := readWKT(t, "MULTILINESTRING ((5 1, 2 2, 5 3), (5 1, 8 2, 5 3), (5 3, 2 4, 5 5), (5 3, 8 4, 5 5))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGLineRingUnion(t *testing.T) {
	a := readWKT(t, "LINESTRING (1 1, 5 5, 9 1)")
	b := readWKT(t, "LINESTRING (1 1, 9 1)")
	expected := readWKT(t, "MULTILINESTRING ((1 1, 5 5, 9 1), (1 1, 9 1))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGDisjointLinesRoundedIntersection(t *testing.T) {
	a := readWKT(t, "LINESTRING (3 2, 3 4)")
	b := readWKT(t, "LINESTRING (1.1 1.6, 3.8 1.9)")
	expected := readWKT(t, "POINT (3 2)")
	actual := overlayNGTestIntersection(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGPolygonMultiLineUnion(t *testing.T) {
	a := readWKT(t, "POLYGON ((100 200, 200 200, 200 100, 100 100, 100 200))")
	b := readWKT(t, "MULTILINESTRING ((150 250, 150 50), (250 250, 250 50))")
	expected := readWKT(t, "GEOMETRYCOLLECTION (LINESTRING (150 50, 150 100), LINESTRING (150 200, 150 250), LINESTRING (250 50, 250 250), POLYGON ((100 100, 100 200, 150 200, 200 200, 200 100, 150 100, 100 100)))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGLinePolygonUnion(t *testing.T) {
	a := readWKT(t, "LINESTRING (50 150, 150 150)")
	b := readWKT(t, "POLYGON ((100 200, 200 200, 200 100, 100 100, 100 200))")
	expected := readWKT(t, "GEOMETRYCOLLECTION (LINESTRING (50 150, 100 150), POLYGON ((100 200, 200 200, 200 100, 100 100, 100 150, 100 200)))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGLinePolygonUnionAlongPolyBoundary(t *testing.T) {
	a := readWKT(t, "LINESTRING (150 300, 250 300)")
	b := readWKT(t, "POLYGON ((100 400, 200 400, 200 300, 100 300, 100 400))")
	expected := readWKT(t, "GEOMETRYCOLLECTION (LINESTRING (200 300, 250 300), POLYGON ((200 300, 150 300, 100 300, 100 400, 200 400, 200 300)))")
	actual := overlayNGTestUnion(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGLinePolygonIntersectionAlongPolyBoundary(t *testing.T) {
	a := readWKT(t, "LINESTRING (150 300, 250 300)")
	b := readWKT(t, "POLYGON ((100 400, 200 400, 200 300, 100 300, 100 400))")
	expected := readWKT(t, "LINESTRING (200 300, 150 300)")
	actual := overlayNGTestIntersection(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGPolygonFlatCollapseIntersection(t *testing.T) {
	a := readWKT(t, "POLYGON ((200 100, 150 200, 250 200, 150 200, 100 100, 200 100))")
	b := readWKT(t, "POLYGON ((50 150, 250 150, 250 50, 50 50, 50 150))")
	expected := readWKT(t, "POLYGON ((175 150, 200 100, 100 100, 125 150, 175 150))")
	actual := overlayNGTestIntersection(t, a, b, 1)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGPolygonLineIntersectionOrder(t *testing.T) {
	a := readWKT(t, "POLYGON ((1 1, 1 9, 9 9, 9 7, 3 7, 3 3, 9 3, 9 1, 1 1))")
	b := readWKT(t, "MULTILINESTRING ((2 10, 2 0), (4 10, 4 0))")
	expected := readWKT(t, "MULTILINESTRING ((2 9, 2 1), (4 9, 4 7), (4 3, 4 1))")
	actual := overlayNGTestIntersection(t, a, b, 1)
	checkEqualGeomsExact(t, expected, actual)
}

func TestOverlayNGPolygonLineVerticalIntersection(t *testing.T) {
	a := readWKT(t, "POLYGON ((-200 -200, 200 -200, 200 200, -200 200, -200 -200))")
	b := readWKT(t, "LINESTRING (-100 100, -100 -100)")
	expected := readWKT(t, "LINESTRING (-100 100, -100 -100)")
	actual := overlayNGTestIntersectionFloat(t, a, b)
	checkEqualGeomsNormalized(t, expected, actual)
}

func TestOverlayNGPolygonLineHorizontalIntersection(t *testing.T) {
	a := readWKT(t, "POLYGON ((10 90, 90 90, 90 10, 10 10, 10 90))")
	b := readWKT(t, "LINESTRING (20 50, 80 50)")
	expected := readWKT(t, "LINESTRING (20 50, 80 50)")
	actual := overlayNGTestIntersectionFloat(t, a, b)
	checkEqualGeomsNormalized(t, expected, actual)
}
