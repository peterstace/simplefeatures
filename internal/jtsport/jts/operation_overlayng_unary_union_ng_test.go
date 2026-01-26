package jts

import "testing"

func checkUnaryUnionNG(t *testing.T, wkt string, scaleFactor float64, expectedWKT string) {
	t.Helper()
	geom := readWKT(t, wkt)
	expected := readWKT(t, expectedWKT)
	pm := Geom_NewPrecisionModelWithScale(scaleFactor)
	result := OperationOverlayng_UnaryUnionNG_UnionGeom(geom, pm)
	checkEqualGeomsNormalized(t, expected, result)
}

func checkUnaryUnionNGCollection(t *testing.T, wkts []string, scaleFactor float64, expectedWKT string) {
	t.Helper()
	geoms := make([]*Geom_Geometry, len(wkts))
	for i, wkt := range wkts {
		geoms[i] = readWKT(t, wkt)
	}
	expected := readWKT(t, expectedWKT)
	pm := Geom_NewPrecisionModelWithScale(scaleFactor)
	var result *Geom_Geometry
	if len(geoms) == 0 {
		result = OperationOverlayng_UnaryUnionNG_UnionCollectionWithFactory(geoms, Geom_NewGeometryFactoryDefault(), pm)
	} else {
		result = OperationOverlayng_UnaryUnionNG_UnionCollection(geoms, pm)
	}
	checkEqualGeomsNormalized(t, expected, result)
}

func TestUnaryUnionNGMultiPolygonNarrowGap(t *testing.T) {
	checkUnaryUnionNG(t,
		"MULTIPOLYGON (((1 9, 5.7 9, 5.7 1, 1 1, 1 9)), ((9 9, 9 1, 6 1, 6 9, 9 9)))",
		1,
		"POLYGON ((1 9, 6 9, 9 9, 9 1, 6 1, 1 1, 1 9))")
}

func TestUnaryUnionNGPolygonsRounded(t *testing.T) {
	checkUnaryUnionNG(t,
		"GEOMETRYCOLLECTION (POLYGON ((1 9, 6 9, 6 1, 1 1, 1 9)), POLYGON ((9 1, 2 8, 9 9, 9 1)))",
		1,
		"POLYGON ((1 9, 6 9, 9 9, 9 1, 6 4, 6 1, 1 1, 1 9))")
}

func TestUnaryUnionNGPolygonsOverlapping(t *testing.T) {
	checkUnaryUnionNG(t,
		"GEOMETRYCOLLECTION (POLYGON ((100 200, 200 200, 200 100, 100 100, 100 200)), POLYGON ((250 250, 250 150, 150 150, 150 250, 250 250)))",
		1,
		"POLYGON ((100 200, 150 200, 150 250, 250 250, 250 150, 200 150, 200 100, 100 100, 100 200))")
}

func TestUnaryUnionNGCollection(t *testing.T) {
	checkUnaryUnionNGCollection(t,
		[]string{
			"POLYGON ((100 200, 200 200, 200 100, 100 100, 100 200))",
			"POLYGON ((300 100, 200 100, 200 200, 300 200, 300 100))",
			"POLYGON ((100 300, 200 300, 200 200, 100 200, 100 300))",
			"POLYGON ((300 300, 300 200, 200 200, 200 300, 300 300))",
		},
		1,
		"POLYGON ((100 100, 100 200, 100 300, 200 300, 300 300, 300 200, 300 100, 200 100, 100 100))")
}

func TestUnaryUnionNGCollectionEmpty(t *testing.T) {
	checkUnaryUnionNGCollection(t,
		[]string{},
		1,
		"GEOMETRYCOLLECTION EMPTY")
}
