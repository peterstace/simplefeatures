package jts

import (
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

const operationValid_IsValidOp_MIN_SIZE_LINESTRING = 2
const operationValid_IsValidOp_MIN_SIZE_RING = 4

// OperationValid_IsValidOp_IsValid tests whether a Geometry is valid.
func OperationValid_IsValidOp_IsValid(geom *Geom_Geometry) bool {
	isValidOp := OperationValid_NewIsValidOp(geom)
	return isValidOp.IsValid()
}

// OperationValid_IsValidOp_IsValidCoordinate checks whether a coordinate is valid for processing.
// Coordinates are valid if their x and y ordinates are in the
// range of the floating point representation.
func OperationValid_IsValidOp_IsValidCoordinate(coord *Geom_Coordinate) bool {
	if math.IsNaN(coord.X) {
		return false
	}
	if math.IsInf(coord.X, 0) {
		return false
	}
	if math.IsNaN(coord.Y) {
		return false
	}
	if math.IsInf(coord.Y, 0) {
		return false
	}
	return true
}

// OperationValid_IsValidOp implements the algorithms required to compute the isValid() method
// for Geometrys.
// See the documentation for the various geometry types for a specification of validity.
type OperationValid_IsValidOp struct {
	// The geometry being validated
	inputGeometry *Geom_Geometry
	// If the following condition is TRUE JTS will validate inverted shells and exverted holes
	// (the ESRI SDE model)
	isInvertedRingValid bool

	validErr *OperationValid_TopologyValidationError
}

// OperationValid_NewIsValidOp creates a new validator for a geometry.
func OperationValid_NewIsValidOp(inputGeometry *Geom_Geometry) *OperationValid_IsValidOp {
	return &OperationValid_IsValidOp{
		inputGeometry:       inputGeometry,
		isInvertedRingValid: false,
	}
}

// SetSelfTouchingRingFormingHoleValid sets whether polygons using Self-Touching Rings to form
// holes are reported as valid.
// If this flag is set, the following Self-Touching conditions
// are treated as being valid:
//   - inverted shell - the shell ring self-touches to create a hole touching the shell
//   - exverted hole - a hole ring self-touches to create two holes touching at a point
//
// The default (following the OGC SFS standard)
// is that this condition is not valid (false).
//
// Self-Touching Rings which disconnect the
// the polygon interior are still considered to be invalid
// (these are invalid under the SFS, and many other
// spatial models as well).
// This includes:
//   - exverted ("bow-tie") shells which self-touch at a single point
//   - inverted shells with the inversion touching the shell at another point
//   - exverted holes with exversion touching the hole at another point
//   - inverted ("C-shaped") holes which self-touch at a single point causing an island to be formed
//   - inverted shells or exverted holes which form part of a chain of touching rings
//     (which disconnect the interior)
func (op *OperationValid_IsValidOp) SetSelfTouchingRingFormingHoleValid(isValid bool) {
	op.isInvertedRingValid = isValid
}

// IsValid tests the validity of the input geometry.
func (op *OperationValid_IsValidOp) IsValid() bool {
	return op.isValidGeometry(op.inputGeometry)
}

// GetValidationError computes the validity of the geometry,
// and if not valid returns the validation error for the geometry,
// or nil if the geometry is valid.
func (op *OperationValid_IsValidOp) GetValidationError() *OperationValid_TopologyValidationError {
	op.isValidGeometry(op.inputGeometry)
	return op.validErr
}

func (op *OperationValid_IsValidOp) logInvalid(code int, pt *Geom_Coordinate) {
	op.validErr = OperationValid_NewTopologyValidationError(code, pt)
}

func (op *OperationValid_IsValidOp) hasInvalidError() bool {
	return op.validErr != nil

}

func (op *OperationValid_IsValidOp) isValidGeometry(g *Geom_Geometry) bool {
	op.validErr = nil

	// empty geometries are always valid
	if g.IsEmpty() {
		return true
	}

	if java.InstanceOf[*Geom_Point](g) {
		return op.isValidPoint(java.Cast[*Geom_Point](g))
	}
	if java.InstanceOf[*Geom_MultiPoint](g) {
		return op.isValidMultiPoint(java.Cast[*Geom_MultiPoint](g))
	}
	if java.InstanceOf[*Geom_LinearRing](g) {
		return op.isValidLinearRing(java.Cast[*Geom_LinearRing](g))
	}
	if java.InstanceOf[*Geom_LineString](g) {
		return op.isValidLineString(java.Cast[*Geom_LineString](g))
	}
	if java.InstanceOf[*Geom_Polygon](g) {
		return op.isValidPolygon(java.Cast[*Geom_Polygon](g))
	}
	if java.InstanceOf[*Geom_MultiPolygon](g) {
		return op.isValidMultiPolygon(java.Cast[*Geom_MultiPolygon](g))
	}
	if java.InstanceOf[*Geom_GeometryCollection](g) {
		return op.isValidGeometryCollection(java.Cast[*Geom_GeometryCollection](g))
	}

	// geometry type not known
	panic("unsupported geometry type")
}

// isValidPoint tests validity of a Point.
func (op *OperationValid_IsValidOp) isValidPoint(g *Geom_Point) bool {
	op.checkCoordinatesValid(g.GetCoordinates())
	if op.hasInvalidError() {
		return false
	}
	return true
}

// isValidMultiPoint tests validity of a MultiPoint.
func (op *OperationValid_IsValidOp) isValidMultiPoint(g *Geom_MultiPoint) bool {
	op.checkCoordinatesValid(g.GetCoordinates())
	if op.hasInvalidError() {
		return false
	}
	return true
}

// isValidLineString tests validity of a LineString.
// Almost anything goes for linestrings!
func (op *OperationValid_IsValidOp) isValidLineString(g *Geom_LineString) bool {
	op.checkCoordinatesValid(g.GetCoordinates())
	if op.hasInvalidError() {
		return false
	}
	op.checkPointSize(g, operationValid_IsValidOp_MIN_SIZE_LINESTRING)
	if op.hasInvalidError() {
		return false
	}
	return true
}

// isValidLinearRing tests validity of a LinearRing.
func (op *OperationValid_IsValidOp) isValidLinearRing(g *Geom_LinearRing) bool {
	op.checkCoordinatesValid(g.GetCoordinates())
	if op.hasInvalidError() {
		return false
	}

	op.checkRingClosed(g)
	if op.hasInvalidError() {
		return false
	}

	op.checkRingPointSize(g)
	if op.hasInvalidError() {
		return false
	}

	op.checkRingSimple(g)
	return op.validErr == nil
}

// isValidPolygon tests the validity of a polygon.
// Sets the validErr flag.
func (op *OperationValid_IsValidOp) isValidPolygon(g *Geom_Polygon) bool {
	op.checkCoordinatesValidForPolygon(g)
	if op.hasInvalidError() {
		return false
	}

	op.checkRingsClosed(g)
	if op.hasInvalidError() {
		return false
	}

	op.checkRingsPointSize(g)
	if op.hasInvalidError() {
		return false
	}

	areaAnalyzer := OperationValid_NewPolygonTopologyAnalyzer(g.Geom_Geometry, op.isInvertedRingValid)

	op.checkAreaIntersections(areaAnalyzer)
	if op.hasInvalidError() {
		return false
	}

	op.checkHolesInShell(g)
	if op.hasInvalidError() {
		return false
	}

	op.checkHolesNotNested(g)
	if op.hasInvalidError() {
		return false
	}

	op.checkInteriorConnected(areaAnalyzer)
	if op.hasInvalidError() {
		return false
	}

	return true
}

// isValidMultiPolygon tests validity of a MultiPolygon.
func (op *OperationValid_IsValidOp) isValidMultiPolygon(g *Geom_MultiPolygon) bool {
	for i := 0; i < g.GetNumGeometries(); i++ {
		p := java.Cast[*Geom_Polygon](g.GetGeometryN(i))
		op.checkCoordinatesValidForPolygon(p)
		if op.hasInvalidError() {
			return false
		}

		op.checkRingsClosed(p)
		if op.hasInvalidError() {
			return false
		}
		op.checkRingsPointSize(p)
		if op.hasInvalidError() {
			return false
		}
	}

	areaAnalyzer := OperationValid_NewPolygonTopologyAnalyzer(g.Geom_Geometry, op.isInvertedRingValid)

	op.checkAreaIntersections(areaAnalyzer)
	if op.hasInvalidError() {
		return false
	}

	for i := 0; i < g.GetNumGeometries(); i++ {
		p := java.Cast[*Geom_Polygon](g.GetGeometryN(i))
		op.checkHolesInShell(p)
		if op.hasInvalidError() {
			return false
		}
	}
	for i := 0; i < g.GetNumGeometries(); i++ {
		p := java.Cast[*Geom_Polygon](g.GetGeometryN(i))
		op.checkHolesNotNested(p)
		if op.hasInvalidError() {
			return false
		}
	}
	op.checkShellsNotNested(g)
	if op.hasInvalidError() {
		return false
	}

	op.checkInteriorConnected(areaAnalyzer)
	if op.hasInvalidError() {
		return false
	}

	return true
}

// isValidGeometryCollection tests validity of a GeometryCollection.
func (op *OperationValid_IsValidOp) isValidGeometryCollection(gc *Geom_GeometryCollection) bool {
	for i := 0; i < gc.GetNumGeometries(); i++ {
		if !op.isValidGeometry(gc.GetGeometryN(i)) {
			return false
		}
	}
	return true
}

func (op *OperationValid_IsValidOp) checkCoordinatesValid(coords []*Geom_Coordinate) {
	for i := 0; i < len(coords); i++ {
		if !OperationValid_IsValidOp_IsValidCoordinate(coords[i]) {
			op.logInvalid(OperationValid_TopologyValidationError_INVALID_COORDINATE, coords[i])
			return
		}
	}
}

func (op *OperationValid_IsValidOp) checkCoordinatesValidForPolygon(poly *Geom_Polygon) {
	op.checkCoordinatesValid(poly.GetExteriorRing().GetCoordinates())
	if op.hasInvalidError() {
		return
	}
	for i := 0; i < poly.GetNumInteriorRing(); i++ {
		op.checkCoordinatesValid(poly.GetInteriorRingN(i).GetCoordinates())
		if op.hasInvalidError() {
			return
		}
	}
}

func (op *OperationValid_IsValidOp) checkRingClosed(ring *Geom_LinearRing) {
	if ring.IsEmpty() {
		return
	}
	if !ring.IsClosed() {
		var pt *Geom_Coordinate
		if ring.GetNumPoints() >= 1 {
			pt = ring.GetCoordinateN(0)
		}
		op.logInvalid(OperationValid_TopologyValidationError_RING_NOT_CLOSED, pt)
		return
	}
}

func (op *OperationValid_IsValidOp) checkRingsClosed(poly *Geom_Polygon) {
	op.checkRingClosed(poly.GetExteriorRing())
	if op.hasInvalidError() {
		return
	}
	for i := 0; i < poly.GetNumInteriorRing(); i++ {
		op.checkRingClosed(poly.GetInteriorRingN(i))
		if op.hasInvalidError() {
			return
		}
	}
}

func (op *OperationValid_IsValidOp) checkRingsPointSize(poly *Geom_Polygon) {
	op.checkRingPointSize(poly.GetExteriorRing())
	if op.hasInvalidError() {
		return
	}
	for i := 0; i < poly.GetNumInteriorRing(); i++ {
		op.checkRingPointSize(poly.GetInteriorRingN(i))
		if op.hasInvalidError() {
			return
		}
	}
}

func (op *OperationValid_IsValidOp) checkRingPointSize(ring *Geom_LinearRing) {
	if ring.IsEmpty() {
		return
	}
	op.checkPointSize(ring.Geom_LineString, operationValid_IsValidOp_MIN_SIZE_RING)
}

// checkPointSize checks the number of non-repeated points is at least a given size.
func (op *OperationValid_IsValidOp) checkPointSize(line *Geom_LineString, minSize int) {
	if !op.isNonRepeatedSizeAtLeast(line, minSize) {
		var pt *Geom_Coordinate
		if line.GetNumPoints() >= 1 {
			pt = line.GetCoordinateN(0)
		}
		op.logInvalid(OperationValid_TopologyValidationError_TOO_FEW_POINTS, pt)
	}
}

// isNonRepeatedSizeAtLeast tests if the number of non-repeated points in a line
// is at least a given minimum size.
func (op *OperationValid_IsValidOp) isNonRepeatedSizeAtLeast(line *Geom_LineString, minSize int) bool {
	numPts := 0
	var prevPt *Geom_Coordinate
	for i := 0; i < line.GetNumPoints(); i++ {
		if numPts >= minSize {
			return true
		}
		pt := line.GetCoordinateN(i)
		if prevPt == nil || !pt.Equals2D(prevPt) {
			numPts++
		}
		prevPt = pt
	}
	return numPts >= minSize
}

func (op *OperationValid_IsValidOp) checkAreaIntersections(areaAnalyzer *OperationValid_PolygonTopologyAnalyzer) {
	if areaAnalyzer.HasInvalidIntersection() {
		op.logInvalid(areaAnalyzer.GetInvalidCode(),
			areaAnalyzer.GetInvalidLocation())
		return
	}
}

// checkRingSimple checks whether a ring self-intersects (except at its endpoints).
func (op *OperationValid_IsValidOp) checkRingSimple(ring *Geom_LinearRing) {
	intPt := OperationValid_PolygonTopologyAnalyzer_FindSelfIntersection(ring)
	if intPt != nil {
		op.logInvalid(OperationValid_TopologyValidationError_RING_SELF_INTERSECTION,
			intPt)
	}
}

// checkHolesInShell tests that each hole is inside the polygon shell.
// This routine assumes that the holes have previously been tested
// to ensure that all vertices lie on the shell or on the same side of it
// (i.e. that the hole rings do not cross the shell ring).
// Given this, a simple point-in-polygon test of a single point in the hole can be used,
// provided the point is chosen such that it does not lie on the shell.
func (op *OperationValid_IsValidOp) checkHolesInShell(poly *Geom_Polygon) {
	// skip test if no holes are present
	if poly.GetNumInteriorRing() <= 0 {
		return
	}

	shell := poly.GetExteriorRing()
	isShellEmpty := shell.IsEmpty()

	for i := 0; i < poly.GetNumInteriorRing(); i++ {
		hole := poly.GetInteriorRingN(i)
		if hole.IsEmpty() {
			continue
		}

		var invalidPt *Geom_Coordinate
		if isShellEmpty {
			invalidPt = hole.GetCoordinate()
		} else {
			invalidPt = op.findHoleOutsideShellPoint(hole, shell)
		}
		if invalidPt != nil {
			op.logInvalid(OperationValid_TopologyValidationError_HOLE_OUTSIDE_SHELL,
				invalidPt)
			return
		}
	}
}

// findHoleOutsideShellPoint checks if a polygon hole lies inside its shell
// and if not returns a point indicating this.
// The hole is known to be wholly inside or outside the shell,
// so it suffices to find a single point which is interior or exterior,
// or check the edge topology at a point on the boundary of the shell.
func (op *OperationValid_IsValidOp) findHoleOutsideShellPoint(hole, shell *Geom_LinearRing) *Geom_Coordinate {
	holePt0 := hole.GetCoordinateN(0)
	// If hole envelope is not covered by shell, it must be outside
	if !shell.GetEnvelopeInternal().CoversEnvelope(hole.GetEnvelopeInternal()) {
		//TODO: find hole pt outside shell env
		return holePt0
	}

	if OperationValid_PolygonTopologyAnalyzer_IsRingNested(hole, shell) {
		return nil
	}
	//TODO: find hole point outside shell
	return holePt0
}

// checkHolesNotNested checks if any polygon hole is nested inside another.
// Assumes that holes do not cross (overlap),
// This is checked earlier.
func (op *OperationValid_IsValidOp) checkHolesNotNested(poly *Geom_Polygon) {
	// skip test if no holes are present
	if poly.GetNumInteriorRing() <= 0 {
		return
	}

	nestedTester := OperationValid_NewIndexedNestedHoleTester(poly)
	if nestedTester.IsNested() {
		op.logInvalid(OperationValid_TopologyValidationError_NESTED_HOLES,
			nestedTester.GetNestedPoint())
	}
}

// checkShellsNotNested checks that no element polygon is in the interior of another element polygon.
//
// Preconditions:
//   - shells do not partially overlap
//   - shells do not touch along an edge
//   - no duplicate rings exist
//
// These have been confirmed by the PolygonTopologyAnalyzer.
func (op *OperationValid_IsValidOp) checkShellsNotNested(mp *Geom_MultiPolygon) {
	// skip test if only one shell present
	if mp.GetNumGeometries() <= 1 {
		return
	}

	nestedTester := OperationValid_NewIndexedNestedPolygonTester(mp)
	if nestedTester.IsNested() {
		op.logInvalid(OperationValid_TopologyValidationError_NESTED_SHELLS,
			nestedTester.GetNestedPoint())
	}
}

func (op *OperationValid_IsValidOp) checkInteriorConnected(analyzer *OperationValid_PolygonTopologyAnalyzer) {
	if analyzer.IsInteriorDisconnected() {
		op.logInvalid(OperationValid_TopologyValidationError_DISCONNECTED_INTERIOR,
			analyzer.GetDisconnectionLocation())
	}
}
