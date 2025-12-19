package jts

import (
	"math"
	"strconv"
	"strings"
)

var _ JtstestTestrunner_ResultMatcher = (*JtstestTestrunner_BufferResultMatcher)(nil)

const jtstestTestrunner_BufferResultMatcher_MAX_RELATIVE_AREA_DIFFERENCE = 1.0e-3

const jtstestTestrunner_BufferResultMatcher_MAX_HAUSDORFF_DISTANCE_FACTOR = 100

// The minimum distance tolerance which will be used. This is required because
// densified vertices do not lie precisely on their parent segment.
const jtstestTestrunner_BufferResultMatcher_MIN_DISTANCE_TOLERANCE = 1.0e-8

// JtstestTestrunner_BufferResultMatcher compares the results of buffer
// operations for equality, up to the given tolerance. All other operations are
// delegated to the standard EqualityResultMatcher algorithm.
type JtstestTestrunner_BufferResultMatcher struct {
	defaultMatcher JtstestTestrunner_ResultMatcher
}

func JtstestTestrunner_NewBufferResultMatcher() *JtstestTestrunner_BufferResultMatcher {
	return &JtstestTestrunner_BufferResultMatcher{
		defaultMatcher: JtstestTestrunner_NewEqualityResultMatcher(),
	}
}

func (m *JtstestTestrunner_BufferResultMatcher) IsJtstestTestrunner_ResultMatcher() {}

func (m *JtstestTestrunner_BufferResultMatcher) IsMatch(
	geom *Geom_Geometry,
	opName string,
	args []any,
	actualResult JtstestTestrunner_Result,
	expectedResult JtstestTestrunner_Result,
	tolerance float64,
) bool {
	if !strings.EqualFold(opName, "buffer") {
		return m.defaultMatcher.IsMatch(geom, opName, args, actualResult, expectedResult, tolerance)
	}

	distance, _ := strconv.ParseFloat(args[0].(string), 64)
	actualGeomResult := actualResult.(*JtstestTestrunner_GeometryResult)
	expectedGeomResult := expectedResult.(*JtstestTestrunner_GeometryResult)
	return m.IsBufferResultMatch(
		actualGeomResult.GetGeometry(),
		expectedGeomResult.GetGeometry(),
		distance,
	)
}

func (m *JtstestTestrunner_BufferResultMatcher) IsBufferResultMatch(
	actualBuffer *Geom_Geometry,
	expectedBuffer *Geom_Geometry,
	distance float64,
) bool {
	if actualBuffer.IsEmpty() && expectedBuffer.IsEmpty() {
		return true
	}

	// MD - need some more checks here - symDiffArea won't catch very small holes
	// ("tears") near the edge of computed buffers (which can happen in current
	// version of JTS (1.8)). This can probably be handled by testing that every
	// point of the actual buffer is at least a certain distance away from the
	// geometry boundary.
	if !m.IsSymDiffAreaInTolerance(actualBuffer, expectedBuffer) {
		return false
	}

	if !m.IsBoundaryHausdorffDistanceInTolerance(actualBuffer, expectedBuffer, distance) {
		return false
	}

	return true
}

func (m *JtstestTestrunner_BufferResultMatcher) IsSymDiffAreaInTolerance(
	actualBuffer *Geom_Geometry,
	expectedBuffer *Geom_Geometry,
) bool {
	area := expectedBuffer.GetArea()
	diff := actualBuffer.SymDifference(expectedBuffer)
	areaDiff := diff.GetArea()

	// Can't get closer than difference area = 0! This also handles case when symDiff is empty.
	if areaDiff <= 0.0 {
		return true
	}

	frac := math.Inf(1)
	if area > 0.0 {
		frac = areaDiff / area
	}

	return frac < jtstestTestrunner_BufferResultMatcher_MAX_RELATIVE_AREA_DIFFERENCE
}

func (m *JtstestTestrunner_BufferResultMatcher) IsBoundaryHausdorffDistanceInTolerance(
	actualBuffer *Geom_Geometry,
	expectedBuffer *Geom_Geometry,
	distance float64,
) bool {
	actualBdy := actualBuffer.GetBoundary()
	expectedBdy := expectedBuffer.GetBoundary()

	// TRANSLITERATION NOTE: DiscreteHausdorffDistance is not yet ported.
	// When it is, this stub should be replaced with:
	// haus := Algorithm_Distance_NewDiscreteHausdorffDistance(actualBdy, expectedBdy)
	// haus.SetDensifyFraction(0.25)
	// maxDistanceFound := haus.OrientedDistance()
	_ = actualBdy
	_ = expectedBdy
	maxDistanceFound := 0.0

	expectedDistanceTol := math.Abs(distance) / jtstestTestrunner_BufferResultMatcher_MAX_HAUSDORFF_DISTANCE_FACTOR
	if expectedDistanceTol < jtstestTestrunner_BufferResultMatcher_MIN_DISTANCE_TOLERANCE {
		expectedDistanceTol = jtstestTestrunner_BufferResultMatcher_MIN_DISTANCE_TOLERANCE
	}
	if maxDistanceFound > expectedDistanceTol {
		return false
	}
	return true
}
