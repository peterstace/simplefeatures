package jts

import (
	"testing"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
	"github.com/peterstace/simplefeatures/internal/jtsport/junit"
)

var operationValidValidClosedRingTest_rdr = Io_NewWKTReader()

func TestValidClosedRingBadLinearRing(t *testing.T) {
	ring := java.Cast[*Geom_LinearRing](operationValidValidClosedRingTest_fromWKT("LINEARRING (0 0, 0 10, 10 10, 10 0, 0 0)"))
	operationValidValidClosedRingTest_updateNonClosedRing(ring)
	operationValidValidClosedRingTest_checkIsValid(t, ring.Geom_Geometry, false)
}

func TestValidClosedRingGoodLinearRing(t *testing.T) {
	ring := java.Cast[*Geom_LinearRing](operationValidValidClosedRingTest_fromWKT("LINEARRING (0 0, 0 10, 10 10, 10 0, 0 0)"))
	operationValidValidClosedRingTest_checkIsValid(t, ring.Geom_Geometry, true)
}

func TestValidClosedRingBadPolygonShell(t *testing.T) {
	poly := java.Cast[*Geom_Polygon](operationValidValidClosedRingTest_fromWKT("POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))"))
	operationValidValidClosedRingTest_updateNonClosedRing(poly.GetExteriorRing())
	operationValidValidClosedRingTest_checkIsValid(t, poly.Geom_Geometry, false)
}

func TestValidClosedRingBadPolygonHole(t *testing.T) {
	poly := java.Cast[*Geom_Polygon](operationValidValidClosedRingTest_fromWKT("POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0), (1 1, 2 1, 2 2, 1 2, 1 1) ))"))
	operationValidValidClosedRingTest_updateNonClosedRing(poly.GetInteriorRingN(0))
	operationValidValidClosedRingTest_checkIsValid(t, poly.Geom_Geometry, false)
}

func TestValidClosedRingGoodPolygon(t *testing.T) {
	poly := java.Cast[*Geom_Polygon](operationValidValidClosedRingTest_fromWKT("POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))"))
	operationValidValidClosedRingTest_checkIsValid(t, poly.Geom_Geometry, true)
}

func TestValidClosedRingBadGeometryCollection(t *testing.T) {
	gc := java.Cast[*Geom_GeometryCollection](operationValidValidClosedRingTest_fromWKT("GEOMETRYCOLLECTION ( POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0), (1 1, 2 1, 2 2, 1 2, 1 1) )), POINT(0 0) )"))
	poly := java.Cast[*Geom_Polygon](gc.GetGeometryN(0))
	operationValidValidClosedRingTest_updateNonClosedRing(poly.GetInteriorRingN(0))
	operationValidValidClosedRingTest_checkIsValid(t, poly.Geom_Geometry, false)
}

func operationValidValidClosedRingTest_checkIsValid(t *testing.T, geom *Geom_Geometry, expected bool) {
	t.Helper()
	validator := OperationValid_NewIsValidOp(geom)
	isValid := validator.IsValid()
	junit.AssertTrue(t, isValid == expected)
}

func operationValidValidClosedRingTest_fromWKT(wkt string) *Geom_Geometry {
	geom, err := operationValidValidClosedRingTest_rdr.Read(wkt)
	if err != nil {
		panic(err)
	}
	return geom
}

func operationValidValidClosedRingTest_updateNonClosedRing(ring *Geom_LinearRing) {
	pts := ring.GetCoordinates()
	pts[0].X += 0.0001
}
