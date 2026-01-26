package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestUnaryUnionEmptyCollection(t *testing.T) {
	checkUnaryUnionFromSlice(t, []string{}, "GEOMETRYCOLLECTION EMPTY")
}

func TestUnaryUnionEmptyPolygon(t *testing.T) {
	checkUnaryUnionFromSingle(t, "POLYGON EMPTY", "POLYGON EMPTY")
}

func TestUnaryUnionEmptyPointWithLine(t *testing.T) {
	checkUnaryUnionFromSlice(t, []string{"POINT EMPTY", "LINESTRING (0 0, 1 1)"}, "LINESTRING (0 0, 1 1)")
}

func TestUnaryUnionPoints(t *testing.T) {
	checkUnaryUnionFromSlice(t, []string{"POINT (1 1)", "POINT (2 2)"}, "MULTIPOINT ((1 1), (2 2))")
}

func TestUnaryUnionLineNoding(t *testing.T) {
	checkUnaryUnionFromSlice(t,
		[]string{"LINESTRING (0 0, 10 0, 5 -5, 5 5)"},
		"MULTILINESTRING ((0 0, 5 0), (5 0, 10 0, 5 -5, 5 0), (5 0, 5 5))")
}

func TestUnaryUnionAll(t *testing.T) {
	checkUnaryUnionFromSlice(t,
		[]string{"GEOMETRYCOLLECTION (POLYGON ((0 0, 0 90, 90 90, 90 0, 0 0)),   POLYGON ((120 0, 120 90, 210 90, 210 0, 120 0)),  LINESTRING (40 50, 40 140),  LINESTRING (160 50, 160 140),  POINT (60 50),  POINT (60 140),  POINT (40 140))"},
		"GEOMETRYCOLLECTION (POINT (60 140),   LINESTRING (40 90, 40 140), LINESTRING (160 90, 160 140), POLYGON ((0 0, 0 90, 40 90, 90 90, 90 0, 0 0)), POLYGON ((120 0, 120 90, 160 90, 210 90, 210 0, 120 0)))")
}

func checkUnaryUnionFromSlice(t *testing.T, inputWKTs []string, expectedWKT string) {
	t.Helper()
	geomFact := jts.Geom_NewGeometryFactoryDefault()
	reader := jts.Io_NewWKTReader()

	geoms := make([]*jts.Geom_Geometry, 0, len(inputWKTs))
	for _, wkt := range inputWKTs {
		g, err := reader.Read(wkt)
		if err != nil {
			t.Fatalf("failed to read WKT %q: %v", wkt, err)
		}
		geoms = append(geoms, g)
	}

	var result *jts.Geom_Geometry
	if len(geoms) == 0 {
		result = jts.OperationUnion_UnaryUnionOp_UnionCollectionWithFactory(geoms, geomFact)
	} else {
		result = jts.OperationUnion_UnaryUnionOp_UnionCollection(geoms)
	}

	expected, err := reader.Read(expectedWKT)
	if err != nil {
		t.Fatalf("failed to read expected WKT %q: %v", expectedWKT, err)
	}

	checkGeomEqual(t, expected, result)
}

func checkUnaryUnionFromSingle(t *testing.T, inputWKT, expectedWKT string) {
	t.Helper()
	reader := jts.Io_NewWKTReader()

	geom, err := reader.Read(inputWKT)
	if err != nil {
		t.Fatalf("failed to read input WKT %q: %v", inputWKT, err)
	}

	result := jts.OperationUnion_UnaryUnionOp_Union(geom)

	expected, err := reader.Read(expectedWKT)
	if err != nil {
		t.Fatalf("failed to read expected WKT %q: %v", expectedWKT, err)
	}

	checkGeomEqual(t, expected, result)
}

func checkGeomEqual(t *testing.T, expected, actual *jts.Geom_Geometry) {
	t.Helper()
	var actualNorm, expectedNorm *jts.Geom_Geometry
	if actual != nil {
		actualNorm = actual.Norm()
	}
	if expected != nil {
		expectedNorm = expected.Norm()
	}

	var equal bool
	if actualNorm == nil || expectedNorm == nil {
		equal = actualNorm == nil && expectedNorm == nil
	} else {
		equal = actualNorm.EqualsExact(expectedNorm)
	}

	if !equal {
		t.Errorf("geometries not equal\nexpected: %v\nactual:   %v", expectedNorm, actualNorm)
	}
}
