package jts

import (
	"fmt"
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

var operationBufferValidate_BufferDistanceValidator_VERBOSE = false

// operationBufferValidate_BufferDistanceValidator_MAX_DISTANCE_DIFF_FRAC is the
// maximum allowable fraction of buffer distance the actual distance can differ by.
// 1% sometimes causes an error - 1.2% should be safe.
const operationBufferValidate_BufferDistanceValidator_MAX_DISTANCE_DIFF_FRAC = .012

// OperationBufferValidate_BufferDistanceValidator validates that a given buffer curve lies
// an appropriate distance from the input generating it.
// Useful only for round buffers (cap and join).
// Can be used for either positive or negative distances.
//
// This is a heuristic test, and may return false positive results
// (I.e. it may fail to detect an invalid result.)
// It should never return a false negative result, however
// (I.e. it should never report a valid result as invalid.)
type OperationBufferValidate_BufferDistanceValidator struct {
	input       *Geom_Geometry
	bufDistance float64
	result      *Geom_Geometry

	minValidDistance float64
	maxValidDistance float64

	minDistanceFound float64
	maxDistanceFound float64

	isValid        bool
	errMsg         string
	errorLocation  *Geom_Coordinate
	errorIndicator *Geom_Geometry
}

// OperationBufferValidate_NewBufferDistanceValidator creates a new BufferDistanceValidator.
func OperationBufferValidate_NewBufferDistanceValidator(input *Geom_Geometry, bufDistance float64, result *Geom_Geometry) *OperationBufferValidate_BufferDistanceValidator {
	return &OperationBufferValidate_BufferDistanceValidator{
		input:       input,
		bufDistance: bufDistance,
		result:      result,
		isValid:     true,
	}
}

// IsValid validates the buffer distance.
func (v *OperationBufferValidate_BufferDistanceValidator) IsValid() bool {
	posDistance := math.Abs(v.bufDistance)
	distDelta := operationBufferValidate_BufferDistanceValidator_MAX_DISTANCE_DIFF_FRAC * posDistance
	v.minValidDistance = posDistance - distDelta
	v.maxValidDistance = posDistance + distDelta

	// can't use this test if either is empty
	if v.input.IsEmpty() || v.result.IsEmpty() {
		return true
	}

	if v.bufDistance > 0.0 {
		v.checkPositiveValid()
	} else {
		v.checkNegativeValid()
	}
	if operationBufferValidate_BufferDistanceValidator_VERBOSE {
		Util_Debug_Println("Min Dist= " + fmt.Sprintf("%v", v.minDistanceFound) + "  err= " +
			fmt.Sprintf("%v", 1.0-v.minDistanceFound/v.bufDistance) +
			"  Max Dist= " + fmt.Sprintf("%v", v.maxDistanceFound) + "  err= " +
			fmt.Sprintf("%v", v.maxDistanceFound/v.bufDistance-1.0))
	}
	return v.isValid
}

// GetErrorMessage returns an appropriate error message if the buffer is invalid.
func (v *OperationBufferValidate_BufferDistanceValidator) GetErrorMessage() string {
	return v.errMsg
}

// GetErrorLocation returns the location of the error.
func (v *OperationBufferValidate_BufferDistanceValidator) GetErrorLocation() *Geom_Coordinate {
	return v.errorLocation
}

// GetErrorIndicator gets a geometry which indicates the location and nature of a validation failure.
//
// The indicator is a line segment showing the location and size
// of the distance discrepancy.
//
// Returns a geometric error indicator or nil if no error was found.
func (v *OperationBufferValidate_BufferDistanceValidator) GetErrorIndicator() *Geom_Geometry {
	return v.errorIndicator
}

func (v *OperationBufferValidate_BufferDistanceValidator) checkPositiveValid() {
	bufCurve := v.result.GetBoundary()
	v.checkMinimumDistance(v.input, bufCurve, v.minValidDistance)
	if !v.isValid {
		return
	}

	v.checkMaximumDistance(v.input, bufCurve, v.maxValidDistance)
}

func (v *OperationBufferValidate_BufferDistanceValidator) checkNegativeValid() {
	// Assert: only polygonal inputs can be checked for negative buffers

	// MD - could generalize this to handle GCs too
	if !java.InstanceOf[*Geom_Polygon](v.input) &&
		!java.InstanceOf[*Geom_MultiPolygon](v.input) &&
		!java.InstanceOf[*Geom_GeometryCollection](v.input) {
		return
	}
	inputCurve := v.getPolygonLines(v.input)
	v.checkMinimumDistance(inputCurve, v.result, v.minValidDistance)
	if !v.isValid {
		return
	}

	v.checkMaximumDistance(inputCurve, v.result, v.maxValidDistance)
}

func (v *OperationBufferValidate_BufferDistanceValidator) getPolygonLines(g *Geom_Geometry) *Geom_Geometry {
	var lines []*Geom_Geometry
	lineExtracter := GeomUtil_NewLinearComponentExtracter(nil)
	// Store lines temporarily and convert to []*Geom_Geometry
	polys := GeomUtil_PolygonExtracter_GetPolygons(g)
	for _, poly := range polys {
		poly.Geom_Geometry.Apply(lineExtracter)
	}
	// Get the extracted lines
	extractedLines := lineExtracter.lines
	for _, line := range extractedLines {
		lines = append(lines, line.Geom_Geometry)
	}
	return g.GetFactory().BuildGeometry(lines)
}

// checkMinimumDistance checks that two geometries are at least a minimum distance apart.
func (v *OperationBufferValidate_BufferDistanceValidator) checkMinimumDistance(g1, g2 *Geom_Geometry, minDist float64) {
	distOp := OperationDistance_NewDistanceOpWithTerminate(g1, g2, minDist)
	v.minDistanceFound = distOp.Distance()

	if v.minDistanceFound < minDist {
		v.isValid = false
		pts := distOp.NearestPoints()
		v.errorLocation = distOp.NearestPoints()[1]
		v.errorIndicator = g1.GetFactory().CreateLineStringFromCoordinates(pts).Geom_Geometry
		v.errMsg = "Distance between buffer curve and input is too small " +
			"(" + fmt.Sprintf("%v", v.minDistanceFound) +
			" at " + Io_WKTWriter_ToLineStringFromTwoCoords(pts[0], pts[1]) + " )"
	}
}

// checkMaximumDistance checks that the furthest distance from the buffer curve to the input
// is less than the given maximum distance.
// This uses the Oriented Hausdorff distance metric.
// It corresponds to finding
// the point on the buffer curve which is furthest from some point on the input.
func (v *OperationBufferValidate_BufferDistanceValidator) checkMaximumDistance(input, bufCurve *Geom_Geometry, maxDist float64) {
	// BufferCurveMaximumDistanceFinder maxDistFinder = new BufferCurveMaximumDistanceFinder(input);
	// maxDistanceFound = maxDistFinder.findDistance(bufCurve);

	haus := AlgorithmDistance_NewDiscreteHausdorffDistance(bufCurve, input)
	haus.SetDensifyFraction(0.25)
	v.maxDistanceFound = haus.OrientedDistance()

	if v.maxDistanceFound > maxDist {
		v.isValid = false
		pts := haus.GetCoordinates()
		v.errorLocation = pts[1]
		v.errorIndicator = input.GetFactory().CreateLineStringFromCoordinates(pts).Geom_Geometry
		v.errMsg = "Distance between buffer curve and input is too large " +
			"(" + fmt.Sprintf("%v", v.maxDistanceFound) +
			" at " + Io_WKTWriter_ToLineStringFromTwoCoords(pts[0], pts[1]) + ")"
	}
}
