package jts

import (
	"math"
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

var operationValidIsValidTest_precisionModel = Geom_NewPrecisionModel()
var operationValidIsValidTest_geometryFactory = Geom_NewGeometryFactoryWithPrecisionModelAndSRID(operationValidIsValidTest_precisionModel, 0)
var operationValidIsValidTest_reader = Io_NewWKTReaderWithFactory(operationValidIsValidTest_geometryFactory)

func TestIsValidInvalidCoordinate(t *testing.T) {
	badCoord := Geom_NewCoordinateWithXY(1.0, math.NaN())
	pts := []*Geom_Coordinate{Geom_NewCoordinateWithXY(0.0, 0.0), badCoord}
	line := operationValidIsValidTest_geometryFactory.CreateLineStringFromCoordinates(pts)
	isValidOp := OperationValid_NewIsValidOp(line.Geom_Geometry)
	valid := isValidOp.IsValid()
	err := isValidOp.GetValidationError()
	errCoord := err.GetCoordinate()

	junit.AssertEquals(t, OperationValid_TopologyValidationError_INVALID_COORDINATE, err.GetErrorType())
	junit.AssertTrue(t, math.IsNaN(errCoord.Y))
	junit.AssertEquals(t, false, valid)
}

func TestIsValidZeroAreaPolygon(t *testing.T) {
	isValid_checkInvalid(t, "POLYGON((0 0, 0 0, 0 0, 0 0, 0 0))")
}

func TestIsValidValidSimplePolygon(t *testing.T) {
	isValid_checkValid(t, "POLYGON ((10 89, 90 89, 90 10, 10 10, 10 89))")
}

func TestIsValidInvalidSimplePolygonRingSelfIntersection(t *testing.T) {
	isValid_checkInvalidWithCode(t, OperationValid_TopologyValidationError_SELF_INTERSECTION,
		"POLYGON ((10 90, 90 10, 90 90, 10 10, 10 90))")
}

func TestIsValidInvalidPolygonInverted(t *testing.T) {
	isValid_checkInvalidWithCode(t, OperationValid_TopologyValidationError_RING_SELF_INTERSECTION,
		"POLYGON ((70 250, 40 500, 100 400, 70 250, 80 350, 60 350, 70 250))")
}

func TestIsValidInvalidPolygonSelfCrossing(t *testing.T) {
	isValid_checkInvalidWithCode(t, OperationValid_TopologyValidationError_SELF_INTERSECTION,
		"POLYGON ((70 250, 70 500, 80 400, 40 400, 70 250))")
}

func TestIsValidSimplePolygonHole(t *testing.T) {
	isValid_checkValid(t,
		"POLYGON ((10 90, 90 90, 90 10, 10 10, 10 90), (60 20, 20 70, 90 90, 60 20))")
}

func TestIsValidPolygonTouchingHoleAtVertex(t *testing.T) {
	isValid_checkValid(t,
		"POLYGON ((240 260, 40 260, 40 80, 240 80, 240 260), (140 180, 40 260, 140 240, 140 180))")
}

func TestIsValidPolygonMultipleHolesTouchAtSamePoint(t *testing.T) {
	isValid_checkValid(t,
		"POLYGON ((10 90, 90 90, 90 10, 10 10, 10 90), (40 80, 60 80, 50 50, 40 80), (20 60, 20 40, 50 50, 20 60), (40 20, 60 20, 50 50, 40 20))")
}

func TestIsValidPolygonHoleOutsideShellAllTouch(t *testing.T) {
	isValid_checkInvalidWithCode(t, OperationValid_TopologyValidationError_HOLE_OUTSIDE_SHELL,
		"POLYGON ((10 10, 30 10, 30 50, 70 50, 70 10, 90 10, 90 90, 10 90, 10 10), (50 50, 30 10, 70 10, 50 50))")
}

func TestIsValidPolygonHoleOutsideShellDoubleTouch(t *testing.T) {
	isValid_checkInvalidWithCode(t, OperationValid_TopologyValidationError_HOLE_OUTSIDE_SHELL,
		"POLYGON ((10 90, 90 90, 90 10, 10 10, 10 90), (20 80, 80 80, 80 20, 20 20, 20 80), (90 70, 150 50, 90 20, 110 40, 90 70))")
}

func TestIsValidPolygonNestedHolesAllTouch(t *testing.T) {
	isValid_checkInvalidWithCode(t, OperationValid_TopologyValidationError_NESTED_HOLES,
		"POLYGON ((10 90, 90 90, 90 10, 10 10, 10 90), (20 80, 80 80, 80 20, 20 20, 20 80), (50 80, 80 50, 50 20, 20 50, 50 80))")
}

func TestIsValidInvalidPolygonHoleProperIntersection(t *testing.T) {
	isValid_checkInvalidWithCode(t, OperationValid_TopologyValidationError_SELF_INTERSECTION,
		"POLYGON ((10 90, 50 50, 10 10, 10 90), (20 50, 60 70, 60 30, 20 50))")
}

func TestIsValidInvalidPolygonDisconnectedInterior(t *testing.T) {
	isValid_checkInvalidWithCode(t, OperationValid_TopologyValidationError_DISCONNECTED_INTERIOR,
		"POLYGON ((10 90, 90 90, 90 10, 10 10, 10 90), (20 80, 30 80, 20 20, 20 80), (80 30, 20 20, 80 20, 80 30), (80 80, 30 80, 80 30, 80 80))")
}

func TestIsValidValidMultiPolygonTouchAtVertices(t *testing.T) {
	isValid_checkValid(t,
		"MULTIPOLYGON (((10 10, 10 90, 90 90, 90 10, 80 80, 50 20, 20 80, 10 10)), ((90 10, 10 10, 50 20, 90 10)))")
}

func TestIsValidInvalidMultiPolygonHoleOverlapCrossing(t *testing.T) {
	isValid_checkInvalidWithCode(t, OperationValid_TopologyValidationError_SELF_INTERSECTION,
		"MULTIPOLYGON (((20 380, 420 380, 420 20, 20 20, 20 380), (220 340, 180 240, 60 200, 140 100, 340 60, 300 240, 220 340)), ((60 200, 340 60, 220 340, 60 200)))")
}

func TestIsValidValidMultiPolygonTouchAtVerticesSegments(t *testing.T) {
	isValid_checkValid(t,
		"MULTIPOLYGON (((60 40, 90 10, 90 90, 10 90, 10 10, 40 40, 60 40)), ((50 40, 20 20, 80 20, 50 40)))")
}

func TestIsValidInvalidMultiPolygonNestedAllTouchAtVertices(t *testing.T) {
	isValid_checkInvalidWithCode(t, OperationValid_TopologyValidationError_NESTED_SHELLS,
		"MULTIPOLYGON (((10 10, 20 30, 10 90, 90 90, 80 30, 90 10, 50 20, 10 10)), ((80 30, 20 30, 50 20, 80 30)))")
}

func TestIsValidValidMultiPolygonHoleTouchVertices(t *testing.T) {
	isValid_checkValid(t,
		"MULTIPOLYGON (((20 380, 420 380, 420 20, 20 20, 20 380), (220 340, 80 320, 60 200, 140 100, 340 60, 300 240, 220 340)), ((60 200, 340 60, 220 340, 60 200)))")
}

func TestIsValidLineString(t *testing.T) {
	isValid_checkInvalid(t, "LINESTRING(0 0, 0 0)")
}

func TestIsValidLinearRingTriangle(t *testing.T) {
	isValid_checkValid(t, "LINEARRING (100 100, 150 200, 200 100, 100 100)")
}

func TestIsValidLinearRingSelfCrossing(t *testing.T) {
	isValid_checkInvalidWithCode(t, OperationValid_TopologyValidationError_RING_SELF_INTERSECTION,
		"LINEARRING (150 100, 300 300, 100 300, 350 100, 150 100)")
}

func TestIsValidLinearRingSelfCrossing2(t *testing.T) {
	isValid_checkInvalidWithCode(t, OperationValid_TopologyValidationError_RING_SELF_INTERSECTION,
		"LINEARRING (0 0, 100 100, 100 0, 0 100, 0 0)")
}

// TestIsValidPolygonHoleWithRepeatedShellPointTouch tests that repeated points at nodes are handled correctly.
//
// See https://github.com/locationtech/jts/issues/843
func TestIsValidPolygonHoleWithRepeatedShellPointTouch(t *testing.T) {
	isValid_checkValid(t, "POLYGON ((90 10, 10 10, 50 90, 50 90, 90 10), (50 90, 60 30, 40 30, 50 90))")
}

func TestIsValidPolygonHoleWithRepeatedShellPointTouchMultiple(t *testing.T) {
	isValid_checkValid(t, "POLYGON ((90 10, 10 10, 50 90, 50 90, 50 90, 50 90, 90 10), (50 90, 60 30, 40 30, 50 90))")
}

func TestIsValidPolygonHoleWithRepeatedTouchEndPoint(t *testing.T) {
	isValid_checkValid(t, "POLYGON ((90 10, 10 10, 50 90, 90 10, 90 10), (90 10, 40 30, 60 50, 90 10))")
}

func TestIsValidPolygonHoleWithRepeatedHolePointTouch(t *testing.T) {
	isValid_checkValid(t, "POLYGON ((50 90, 10 10, 90 10, 50 90), (50 90, 50 90, 60 40, 60 40, 40 40, 50 90))")
}

//=============================================

func isValid_checkValid(t *testing.T, wkt string) {
	t.Helper()
	isValid_checkValidExpected(t, true, wkt)
}

func isValid_checkValidExpected(t *testing.T, isExpectedValid bool, wkt string) {
	t.Helper()
	geom := isValid_read(wkt)
	isValid := geom.IsValid()
	junit.AssertEquals(t, isExpectedValid, isValid)
}

func isValid_checkInvalid(t *testing.T, wkt string) {
	t.Helper()
	isValid_checkValidExpected(t, false, wkt)
}

func isValid_checkInvalidWithCode(t *testing.T, expectedErrType int, wkt string) {
	t.Helper()
	geom := isValid_read(wkt)
	validOp := OperationValid_NewIsValidOp(geom)
	err := validOp.GetValidationError()
	junit.AssertEquals(t, expectedErrType, err.GetErrorType())
}

func isValid_read(wkt string) *Geom_Geometry {
	g, err := operationValidIsValidTest_reader.Read(wkt)
	if err != nil {
		panic(err)
	}
	return g
}
