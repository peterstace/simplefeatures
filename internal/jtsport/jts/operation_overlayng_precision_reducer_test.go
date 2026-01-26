package jts

import "testing"

func checkPrecisionReduce(t *testing.T, wkt string, scaleFactor float64, expectedWKT string) {
	t.Helper()
	geom := readWKT(t, wkt)
	expected := readWKT(t, expectedWKT)
	pm := Geom_NewPrecisionModelWithScale(scaleFactor)
	result := OperationOverlayng_PrecisionReducer_ReducePrecision(geom, pm)
	checkEqualGeomsNormalized(t, expected, result)
}

func TestPrecisionReducerPolygonBoxEmpty(t *testing.T) {
	checkPrecisionReduce(t,
		"POLYGON ((1 1.4, 7.3 1.4, 7.3 1.2, 1 1.2, 1 1.4))",
		1,
		"POLYGON EMPTY")
}

func TestPrecisionReducerPolygonThinEmpty(t *testing.T) {
	checkPrecisionReduce(t,
		"POLYGON ((1 1.4, 3.05 1.4, 3 4.1, 6 5, 3.2 4, 3.2 1.4, 7.3 1.4, 7.3 1.2, 1 1.2, 1 1.4))",
		1,
		"POLYGON EMPTY")
}

func TestPrecisionReducerPolygonGore(t *testing.T) {
	checkPrecisionReduce(t,
		"POLYGON ((2 1, 9 1, 9 5, 3 5, 9 5.3, 9 9, 2 9, 2 1))",
		1,
		"POLYGON ((9 1, 2 1, 2 9, 9 9, 9 5, 9 1))")
}

func TestPrecisionReducerPolygonGore2(t *testing.T) {
	checkPrecisionReduce(t,
		"POLYGON ((9 1, 1 1, 1 9, 9 9, 9 5, 5 5.1, 5 4.9, 9 4.9, 9 1))",
		1,
		"POLYGON ((9 1, 1 1, 1 9, 9 9, 9 5, 9 1))")
}

func TestPrecisionReducerPolygonGoreToHole(t *testing.T) {
	checkPrecisionReduce(t,
		"POLYGON ((9 1, 1 1, 1 9, 9 9, 9 5, 5 5.9, 5 4.9, 9 4.9, 9 1))",
		1,
		"POLYGON ((9 1, 1 1, 1 9, 9 9, 9 5, 9 1), (9 5, 5 6, 5 5, 9 5))")
}

func TestPrecisionReducerPolygonSpike(t *testing.T) {
	checkPrecisionReduce(t,
		"POLYGON ((1 1, 9 1, 5 1.4, 5 5, 1 5, 1 1))",
		1,
		"POLYGON ((5 5, 5 1, 1 1, 1 5, 5 5))")
}

func TestPrecisionReducerPolygonNarrowHole(t *testing.T) {
	checkPrecisionReduce(t,
		"POLYGON ((1 9, 9 9, 9 1, 1 1, 1 9), (2 5, 8 5, 8 5.3, 2 5))",
		1,
		"POLYGON ((9 1, 1 1, 1 9, 9 9, 9 1))")
}

func TestPrecisionReducerPolygonWideHole(t *testing.T) {
	checkPrecisionReduce(t,
		"POLYGON ((1 9, 9 9, 9 1, 1 1, 1 9), (2 5, 8 5, 8 5.8, 2 5))",
		1,
		"POLYGON ((9 1, 1 1, 1 9, 9 9, 9 1), (8 5, 8 6, 2 5, 8 5))")
}

func TestPrecisionReducerMultiPolygonGap(t *testing.T) {
	checkPrecisionReduce(t,
		"MULTIPOLYGON (((1 9, 9.1 9.1, 9 9, 9 4, 1 4.3, 1 9)), ((1 1, 1 4, 9 3.6, 9 1, 1 1)))",
		1,
		"POLYGON ((9 1, 1 1, 1 4, 1 9, 9 9, 9 4, 9 1))")
}

func TestPrecisionReducerMultiPolygonGapToHole(t *testing.T) {
	checkPrecisionReduce(t,
		"MULTIPOLYGON (((1 9, 9 9, 9.05 4.35, 6 4.35, 4 6, 2.6 4.25, 1 4, 1 9)), ((1 1, 1 4, 9 4, 9 1, 1 1)))",
		1,
		"POLYGON ((9 1, 1 1, 1 4, 1 9, 9 9, 9 4, 9 1), (6 4, 4 6, 3 4, 6 4))")
}

func TestPrecisionReducerLine(t *testing.T) {
	checkPrecisionReduce(t,
		"LINESTRING(-3 6, 9 1)",
		0.5,
		"LINESTRING (-2 6, 10 2)")
}

func TestPrecisionReducerCollapsedLine(t *testing.T) {
	checkPrecisionReduce(t,
		"LINESTRING(1 1, 1 9, 1.1 1)",
		1,
		"LINESTRING (1 1, 1 9)")
}

func TestPrecisionReducerCollapsedNodedLine(t *testing.T) {
	checkPrecisionReduce(t,
		"LINESTRING(1 1, 3 3, 9 9, 5.1 5, 2.1 2)",
		1,
		"MULTILINESTRING ((1 1, 2 2), (2 2, 3 3), (3 3, 5 5), (5 5, 9 9))")
}
