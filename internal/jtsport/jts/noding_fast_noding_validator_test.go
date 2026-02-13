package jts

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

var noding_FastNodingValidatorTest_VERTEX_INT = []string{
	"LINESTRING (100 100, 200 200, 300 300)",
	"LINESTRING (100 300, 200 200)",
}

var noding_FastNodingValidatorTest_INTERIOR_INT = []string{
	"LINESTRING (100 100, 300 300)",
	"LINESTRING (100 300, 300 100)",
}

var noding_FastNodingValidatorTest_NO_INT = []string{
	"LINESTRING (100 100, 200 200)",
	"LINESTRING (200 200, 300 300)",
	"LINESTRING (100 300, 200 200)",
}

var noding_FastNodingValidatorTest_SELF_INTERIOR_INT = []string{
	"LINESTRING (100 100, 300 300, 300 100, 100 300)",
}

var noding_FastNodingValidatorTest_SELF_VERTEX_INT = []string{
	"LINESTRING (100 100, 200 200, 300 300, 400 200, 200 200)",
}

func TestFastNodingValidator_InteriorIntersection(t *testing.T) {
	noding_FastNodingValidatorTest_checkValid(t, noding_FastNodingValidatorTest_INTERIOR_INT, false)
	noding_FastNodingValidatorTest_checkIntersection(t, noding_FastNodingValidatorTest_INTERIOR_INT, "POINT(200 200)")
}

func TestFastNodingValidator_VertexIntersection(t *testing.T) {
	noding_FastNodingValidatorTest_checkValid(t, noding_FastNodingValidatorTest_VERTEX_INT, false)
	// checkIntersection(VERTEX_INT, "POINT(200 200)");
}

func TestFastNodingValidator_NoIntersection(t *testing.T) {
	noding_FastNodingValidatorTest_checkValid(t, noding_FastNodingValidatorTest_NO_INT, true)
}

func TestFastNodingValidator_SelfInteriorIntersection(t *testing.T) {
	noding_FastNodingValidatorTest_checkValid(t, noding_FastNodingValidatorTest_SELF_INTERIOR_INT, false)
}

func TestFastNodingValidator_SelfVertexIntersection(t *testing.T) {
	noding_FastNodingValidatorTest_checkValid(t, noding_FastNodingValidatorTest_SELF_VERTEX_INT, false)
}

func noding_FastNodingValidatorTest_checkValid(t *testing.T, inputWKT []string, isValidExpected bool) {
	t.Helper()
	input := noding_FastNodingValidatorTest_readList(t, inputWKT)
	segStrings := noding_FastNodingValidatorTest_toSegmentStrings(input)
	fnv := Noding_NewFastNodingValidator(segStrings)
	isValid := fnv.IsValid()

	junit.AssertTrue(t, isValidExpected == isValid)
}

func noding_FastNodingValidatorTest_checkIntersection(t *testing.T, inputWKT []string, expectedWKT string) {
	t.Helper()
	input := noding_FastNodingValidatorTest_readList(t, inputWKT)
	expected := noding_FastNodingValidatorTest_read(t, expectedWKT)
	pts := expected.GetCoordinates()
	intPtsExpected := Geom_NewCoordinateListFromCoordinates(pts)

	segStrings := noding_FastNodingValidatorTest_toSegmentStrings(input)
	intPtsActual := Noding_FastNodingValidator_ComputeIntersections(segStrings)

	isSameNumberOfIntersections := intPtsExpected.Size() == len(intPtsActual)
	junit.AssertTrue(t, isSameNumberOfIntersections)

	noding_FastNodingValidatorTest_checkIntersections(t, intPtsActual, intPtsExpected)
}

func noding_FastNodingValidatorTest_checkIntersections(t *testing.T, intPtsActual []*Geom_Coordinate, intPtsExpected *Geom_CoordinateList) {
	t.Helper()
	// TODO: sort intersections so they can be compared
	for i := 0; i < len(intPtsActual); i++ {
		ptActual := intPtsActual[i]
		ptExpected := intPtsExpected.Get(i)

		isEqual := ptActual.Equals2D(ptExpected)
		junit.AssertTrue(t, isEqual)
	}
}

func noding_FastNodingValidatorTest_toSegmentStrings(geoms []*Geom_Geometry) []Noding_SegmentString {
	var segStrings []Noding_SegmentString
	for _, geom := range geoms {
		extracted := noding_FastNodingValidatorTest_extractSegmentStrings(geom)
		segStrings = append(segStrings, extracted...)
	}
	return segStrings
}

func noding_FastNodingValidatorTest_extractSegmentStrings(geom *Geom_Geometry) []Noding_SegmentString {
	var segStr []Noding_SegmentString
	lines := GeomUtil_LinearComponentExtracter_GetLines(geom)
	for _, line := range lines {
		pts := line.GetCoordinates()
		segStr = append(segStr, Noding_NewNodedSegmentString(pts, geom))
	}
	return segStr
}

func noding_FastNodingValidatorTest_readList(t *testing.T, wktList []string) []*Geom_Geometry {
	t.Helper()
	var result []*Geom_Geometry
	for _, wkt := range wktList {
		geom := noding_FastNodingValidatorTest_read(t, wkt)
		result = append(result, geom)
	}
	return result
}

func noding_FastNodingValidatorTest_read(t *testing.T, wkt string) *Geom_Geometry {
	t.Helper()
	reader := Io_NewWKTReader()
	geom, err := reader.Read(wkt)
	if err != nil {
		t.Fatalf("failed to read WKT: %v", err)
	}
	return geom
}
