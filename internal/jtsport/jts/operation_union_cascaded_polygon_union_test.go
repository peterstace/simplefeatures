package jts_test

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/jts"
)

func TestCascadedPolygonUnionBoxes(t *testing.T) {
	inputWKTs := []string{
		"POLYGON ((80 260, 200 260, 200 30, 80 30, 80 260))",
		"POLYGON ((30 180, 300 180, 300 110, 30 110, 30 180))",
		"POLYGON ((30 280, 30 150, 140 150, 140 280, 30 280))",
	}
	checkCascadedUnion(t, inputWKTs)
}

func TestCascadedPolygonUnionSimple(t *testing.T) {
	inputWKTs := []string{
		"POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
		"POLYGON ((5 0, 15 0, 15 10, 5 10, 5 0))",
	}
	expectedWKT := "POLYGON ((0 0, 0 10, 5 10, 10 10, 15 10, 15 0, 10 0, 5 0, 0 0))"
	checkCascadedUnionExpected(t, inputWKTs, expectedWKT)
}

func TestCascadedPolygonUnionNonOverlapping(t *testing.T) {
	inputWKTs := []string{
		"POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))",
		"POLYGON ((20 0, 30 0, 30 10, 20 10, 20 0))",
	}
	expectedWKT := "MULTIPOLYGON (((0 0, 0 10, 10 10, 10 0, 0 0)), ((20 0, 20 10, 30 10, 30 0, 20 0)))"
	checkCascadedUnionExpected(t, inputWKTs, expectedWKT)
}

func checkCascadedUnion(t *testing.T, inputWKTs []string) {
	t.Helper()
	reader := jts.Io_NewWKTReader()

	geoms := make([]*jts.Geom_Geometry, 0, len(inputWKTs))
	for _, wkt := range inputWKTs {
		g, err := reader.Read(wkt)
		if err != nil {
			t.Fatalf("failed to read WKT %q: %v", wkt, err)
		}
		geoms = append(geoms, g)
	}

	// Compute cascaded union.
	cascadedResult := jts.OperationUnion_CascadedPolygonUnion_Union(geoms)

	// Compute iterated union for comparison.
	iteratedResult := unionIterated(t, geoms)

	// Compare the two results - they should be equivalent.
	if cascadedResult == nil && iteratedResult == nil {
		return
	}
	if cascadedResult == nil || iteratedResult == nil {
		t.Errorf("one result is nil: cascaded=%v, iterated=%v", cascadedResult, iteratedResult)
		return
	}

	cascadedNorm := cascadedResult.Norm()
	iteratedNorm := iteratedResult.Norm()

	if !cascadedNorm.EqualsExact(iteratedNorm) {
		t.Errorf("cascaded and iterated union results differ\ncascaded: %v\niterated: %v", cascadedNorm, iteratedNorm)
	}
}

func checkCascadedUnionExpected(t *testing.T, inputWKTs []string, expectedWKT string) {
	t.Helper()
	reader := jts.Io_NewWKTReader()

	geoms := make([]*jts.Geom_Geometry, 0, len(inputWKTs))
	for _, wkt := range inputWKTs {
		g, err := reader.Read(wkt)
		if err != nil {
			t.Fatalf("failed to read WKT %q: %v", wkt, err)
		}
		geoms = append(geoms, g)
	}

	result := jts.OperationUnion_CascadedPolygonUnion_Union(geoms)

	expected, err := reader.Read(expectedWKT)
	if err != nil {
		t.Fatalf("failed to read expected WKT %q: %v", expectedWKT, err)
	}

	checkGeomEqual(t, expected, result)
}

func unionIterated(t *testing.T, geoms []*jts.Geom_Geometry) *jts.Geom_Geometry {
	t.Helper()
	var unionAll *jts.Geom_Geometry
	for _, geom := range geoms {
		if unionAll == nil {
			unionAll = geom.Copy()
		} else {
			unionAll = unionAll.Union(geom)
		}
	}
	return unionAll
}
