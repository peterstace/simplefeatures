package jts

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

// Tests that IsValidOp validates polygons with
// Self-Touching Rings (inverted shells or exverted holes).
// Mainly tests that configuring IsValidOp to allow validating
// the STR validates polygons with this condition, and does not validate
// polygons with other kinds of self-intersection (such as ones with Disconnected Interiors).
// Includes some basic tests to confirm that other invalid cases remain detected correctly,
// but most of this testing is left to the existing XML validation tests.

var operationValidValidSelfTouchingRingTest_rdr = Io_NewWKTReader()

// TestValidSelfTouchingRingShellAndHoleSelfTouch tests a geometry with both a shell self-touch and a hole self-touch.
// This is valid if STR is allowed, but invalid in OGC
func TestValidSelfTouchingRingShellAndHoleSelfTouch(t *testing.T) {
	wkt := "POLYGON ((0 0, 0 340, 320 340, 320 0, 120 0, 180 100, 60 100, 120 0, 0 0),   (80 300, 80 180, 200 180, 200 240, 280 200, 280 280, 200 240, 200 300, 80 300))"
	validSelfTouchingRingTest_checkIsValidSTR(t, wkt, true)
	validSelfTouchingRingTest_checkIsValidOGC(t, wkt, false)
}

func TestValidSelfTouchingRingShellTouchAtHole(t *testing.T) {
	wkt := "POLYGON ((10 90, 90 90, 90 10, 50 50, 80 50, 80 80, 10 10, 10 90), (40 80, 20 60, 50 50, 40 80))"
	validSelfTouchingRingTest_checkIsValidSTR(t, wkt, true)
	validSelfTouchingRingTest_checkIsValidOGC(t, wkt, false)
}

func TestValidSelfTouchingRingShellTouchInChain(t *testing.T) {
	wkt := "POLYGON ((10 90, 90 90, 90 10, 10 10, 10 90, 20 70, 30 70, 30 50, 40 50, 40 70, 30 70, 30 80, 10 90))"
	validSelfTouchingRingTest_checkIsValidSTR(t, wkt, true)
	validSelfTouchingRingTest_checkIsValidOGC(t, wkt, false)
}

func TestValidSelfTouchingRingHoleTouchInChain(t *testing.T) {
	wkt := "POLYGON ((10 90, 90 90, 90 10, 10 10, 10 90), (20 20, 80 20, 80 50, 70 20, 70 50, 60 20, 60 50, 50 20, 50 50, 40 20, 40 50, 30 20, 30 50, 20 20))"
	validSelfTouchingRingTest_checkIsValidSTR(t, wkt, true)
	validSelfTouchingRingTest_checkIsValidOGC(t, wkt, false)
}

// TestValidSelfTouchingRingShellHoleAndHoleHoleTouch tests a geometry representing the same area as in testShellAndHoleSelfTouch
// but using a shell-hole touch and a hole-hole touch.
// This is valid in OGC.
func TestValidSelfTouchingRingShellHoleAndHoleHoleTouch(t *testing.T) {
	wkt := "POLYGON ((0 0, 0 340, 320 340, 320 0, 120 0, 0 0),   (120 0, 180 100, 60 100, 120 0),   (80 300, 80 180, 200 180, 200 240, 200 300, 80 300),  (200 240, 280 200, 280 280, 200 240))"
	validSelfTouchingRingTest_checkIsValidSTR(t, wkt, true)
	validSelfTouchingRingTest_checkIsValidOGC(t, wkt, true)
}

// TestValidSelfTouchingRingShellSelfTouchHoleOverlappingHole tests an overlapping hole condition, where one of the holes is created by a shell self-touch.
// This is never valid.
func TestValidSelfTouchingRingShellSelfTouchHoleOverlappingHole(t *testing.T) {
	wkt := "POLYGON ((0 0, 220 0, 220 200, 120 200, 140 100, 80 100, 120 200, 0 200, 0 0),   (200 80, 20 80, 120 200, 200 80))"
	validSelfTouchingRingTest_checkIsValidSTR(t, wkt, false)
	validSelfTouchingRingTest_checkIsValidOGC(t, wkt, false)
}

// TestValidSelfTouchingRingDisconnectedInteriorShellSelfTouchAtNonVertex ensures that the Disconnected Interior condition is not validated
func TestValidSelfTouchingRingDisconnectedInteriorShellSelfTouchAtNonVertex(t *testing.T) {
	wkt := "POLYGON ((40 180, 40 60, 240 60, 240 180, 140 60, 40 180))"
	validSelfTouchingRingTest_checkIsValidSTR(t, wkt, false)
	validSelfTouchingRingTest_checkIsValidOGC(t, wkt, false)
}

func TestValidSelfTouchingRingDisconnectedInteriorShellSelfTouchAtVertex(t *testing.T) {
	wkt := "POLYGON ((20 20, 20 100, 140 100, 140 180, 260 180, 260 100, 140 100, 140 20, 20 20))"
	validSelfTouchingRingTest_checkIsValidSTR(t, wkt, false)
	validSelfTouchingRingTest_checkIsValidOGC(t, wkt, false)
}

func TestValidSelfTouchingRingDisconnectedInteriorShellTouchAtVertices(t *testing.T) {
	wkt := "POLYGON ((10 10, 90 10, 50 50, 80 70, 90 10, 90 90, 10 90, 10 10, 50 50, 20 70, 10 10))"
	validSelfTouchingRingTest_checkIsValidSTR(t, wkt, false)
	validSelfTouchingRingTest_checkIsValidOGC(t, wkt, false)
}

func TestValidSelfTouchingRingDisconnectedInteriorHoleTouch(t *testing.T) {
	wkt := "POLYGON ((10 90, 90 90, 90 10, 10 10, 10 90), (20 20, 20 80, 80 80, 80 30, 30 30, 70 40, 70 70, 20 20))"
	validSelfTouchingRingTest_checkIsValidSTR(t, wkt, false)
	validSelfTouchingRingTest_checkIsValidOGC(t, wkt, false)
}

func TestValidSelfTouchingRingShellCross(t *testing.T) {
	wkt := "POLYGON ((20 20, 120 20, 120 220, 240 220, 240 120, 20 120, 20 20))"
	validSelfTouchingRingTest_checkIsValidSTR(t, wkt, false)
	validSelfTouchingRingTest_checkIsValidOGC(t, wkt, false)
}

func TestValidSelfTouchingRingShellCrossAndSTR(t *testing.T) {
	wkt := "POLYGON ((20 20, 120 20, 120 220, 180 220, 140 160, 200 160, 180 220, 240 220, 240 120, 20 120,  20 20))"
	validSelfTouchingRingTest_checkIsValidSTR(t, wkt, false)
	validSelfTouchingRingTest_checkIsValidOGC(t, wkt, false)
}

func TestValidSelfTouchingRingExvertedHoleStarTouchHoleCycle(t *testing.T) {
	wkt := "POLYGON ((10 90, 90 90, 90 10, 10 10, 10 90), (20 80, 50 30, 80 80, 80 30, 20 30, 20 80), (40 70, 50 70, 50 30, 40 70), (40 20, 60 20, 50 30, 40 20), (40 80, 20 80, 40 70, 40 80))"
	validSelfTouchingRingTest_checkInvalidSTR(t, wkt, OperationValid_TopologyValidationError_DISCONNECTED_INTERIOR)
	//checkIsValidOGC(wkt, false);
}

func TestValidSelfTouchingRingExvertedHoleStarTouch(t *testing.T) {
	wkt := "POLYGON ((10 90, 90 90, 90 10, 10 10, 10 90), (20 80, 50 30, 80 80, 80 30, 20 30, 20 80), (40 70, 50 70, 50 30, 40 70), (40 20, 60 20, 50 30, 40 20))"
	validSelfTouchingRingTest_checkIsValidSTR(t, wkt, true)
	validSelfTouchingRingTest_checkIsValidOGC(t, wkt, false)
}

func validSelfTouchingRingTest_checkInvalidSTR(t *testing.T, wkt string, expectedErrType int) {
	t.Helper()
	geom := validSelfTouchingRingTest_read(wkt)
	validOp := OperationValid_NewIsValidOp(geom)
	validOp.SetSelfTouchingRingFormingHoleValid(true)
	err := validOp.GetValidationError()
	junit.AssertEquals(t, expectedErrType, err.GetErrorType())
}

func validSelfTouchingRingTest_checkIsValidOGC(t *testing.T, wkt string, expected bool) {
	t.Helper()
	geom := validSelfTouchingRingTest_read(wkt)
	validator := OperationValid_NewIsValidOp(geom)
	isValid := validator.IsValid()
	junit.AssertTrue(t, isValid == expected)
}

func validSelfTouchingRingTest_checkIsValidSTR(t *testing.T, wkt string, expected bool) {
	t.Helper()
	geom := validSelfTouchingRingTest_read(wkt)
	validator := OperationValid_NewIsValidOp(geom)
	validator.SetSelfTouchingRingFormingHoleValid(true)
	isValid := validator.IsValid()
	junit.AssertTrue(t, isValid == expected)
}

func validSelfTouchingRingTest_read(wkt string) *Geom_Geometry {
	g, err := operationValidValidSelfTouchingRingTest_rdr.Read(wkt)
	if err != nil {
		panic(err)
	}
	return g
}
