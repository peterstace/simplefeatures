package jts

import "math"

// OperationBuffer_BufferParameters_CAP_ROUND specifies a round line buffer end cap style.
const OperationBuffer_BufferParameters_CAP_ROUND = 1

// OperationBuffer_BufferParameters_CAP_FLAT specifies a flat line buffer end cap style.
const OperationBuffer_BufferParameters_CAP_FLAT = 2

// OperationBuffer_BufferParameters_CAP_SQUARE specifies a square line buffer end cap style.
const OperationBuffer_BufferParameters_CAP_SQUARE = 3

// OperationBuffer_BufferParameters_JOIN_ROUND specifies a round join style.
const OperationBuffer_BufferParameters_JOIN_ROUND = 1

// OperationBuffer_BufferParameters_JOIN_MITRE specifies a mitre join style.
const OperationBuffer_BufferParameters_JOIN_MITRE = 2

// OperationBuffer_BufferParameters_JOIN_BEVEL specifies a bevel join style.
const OperationBuffer_BufferParameters_JOIN_BEVEL = 3

// OperationBuffer_BufferParameters_DEFAULT_QUADRANT_SEGMENTS is the default number of facets
// into which to divide a fillet of 90 degrees.
// A value of 8 gives less than 2% max error in the buffer distance.
// For a max error of < 1%, use QS = 12.
// For a max error of < 0.1%, use QS = 18.
const OperationBuffer_BufferParameters_DEFAULT_QUADRANT_SEGMENTS = 8

// OperationBuffer_BufferParameters_DEFAULT_MITRE_LIMIT is the default mitre limit.
// Allows fairly pointy mitres.
const OperationBuffer_BufferParameters_DEFAULT_MITRE_LIMIT = 5.0

// OperationBuffer_BufferParameters_DEFAULT_SIMPLIFY_FACTOR is the default simplify factor.
// Provides an accuracy of about 1%, which matches the accuracy of the default Quadrant Segments parameter.
const OperationBuffer_BufferParameters_DEFAULT_SIMPLIFY_FACTOR = 0.01

// OperationBuffer_BufferParameters is a value class containing the parameters which
// specify how a buffer should be constructed.
//
// The parameters allow control over:
//   - Quadrant segments (accuracy of approximation for circular arcs)
//   - End Cap style
//   - Join style
//   - Mitre limit
//   - whether the buffer is single-sided
type OperationBuffer_BufferParameters struct {
	quadrantSegments int
	endCapStyle      int
	joinStyle        int
	mitreLimit       float64
	isSingleSided    bool
	simplifyFactor   float64
}

// OperationBuffer_NewBufferParameters creates a default set of parameters.
func OperationBuffer_NewBufferParameters() *OperationBuffer_BufferParameters {
	return &OperationBuffer_BufferParameters{
		quadrantSegments: OperationBuffer_BufferParameters_DEFAULT_QUADRANT_SEGMENTS,
		endCapStyle:      OperationBuffer_BufferParameters_CAP_ROUND,
		joinStyle:        OperationBuffer_BufferParameters_JOIN_ROUND,
		mitreLimit:       OperationBuffer_BufferParameters_DEFAULT_MITRE_LIMIT,
		isSingleSided:    false,
		simplifyFactor:   OperationBuffer_BufferParameters_DEFAULT_SIMPLIFY_FACTOR,
	}
}

// OperationBuffer_NewBufferParametersWithQuadrantSegments creates a set of parameters with the
// given quadrantSegments value.
func OperationBuffer_NewBufferParametersWithQuadrantSegments(quadrantSegments int) *OperationBuffer_BufferParameters {
	bp := OperationBuffer_NewBufferParameters()
	bp.SetQuadrantSegments(quadrantSegments)
	return bp
}

// OperationBuffer_NewBufferParametersWithQuadrantSegmentsAndEndCapStyle creates a set of parameters with the
// given quadrantSegments and endCapStyle values.
func OperationBuffer_NewBufferParametersWithQuadrantSegmentsAndEndCapStyle(quadrantSegments int, endCapStyle int) *OperationBuffer_BufferParameters {
	bp := OperationBuffer_NewBufferParameters()
	bp.SetQuadrantSegments(quadrantSegments)
	bp.SetEndCapStyle(endCapStyle)
	return bp
}

// OperationBuffer_NewBufferParametersWithQuadrantSegmentsEndCapStyleJoinStyleAndMitreLimit creates a set of parameters with the
// given parameter values.
func OperationBuffer_NewBufferParametersWithQuadrantSegmentsEndCapStyleJoinStyleAndMitreLimit(quadrantSegments int, endCapStyle int, joinStyle int, mitreLimit float64) *OperationBuffer_BufferParameters {
	bp := OperationBuffer_NewBufferParameters()
	bp.SetQuadrantSegments(quadrantSegments)
	bp.SetEndCapStyle(endCapStyle)
	bp.SetJoinStyle(joinStyle)
	bp.SetMitreLimit(mitreLimit)
	return bp
}

// GetQuadrantSegments gets the number of quadrant segments which will be used
// to approximate angle fillets in round endcaps and joins.
func (bp *OperationBuffer_BufferParameters) GetQuadrantSegments() int {
	return bp.quadrantSegments
}

// SetQuadrantSegments sets the number of line segments in a quarter-circle
// used to approximate angle fillets in round endcaps and joins.
// The value should be at least 1.
//
// This determines the error in the approximation to the true buffer curve.
// The default value of 8 gives less than 2% error in the buffer distance.
// For a error of < 1%, use QS = 12.
// For a error of < 0.1%, use QS = 18.
// The error is always less than the buffer distance
// (in other words, the computed buffer curve is always inside the true curve).
func (bp *OperationBuffer_BufferParameters) SetQuadrantSegments(quadSegs int) {
	bp.quadrantSegments = quadSegs
}

// OperationBuffer_BufferParameters_BufferDistanceError computes the maximum distance error due to a given level
// of approximation to a true arc.
func OperationBuffer_BufferParameters_BufferDistanceError(quadSegs int) float64 {
	alpha := Algorithm_Angle_PiOver2 / float64(quadSegs)
	return 1 - math.Cos(alpha/2.0)
}

// GetEndCapStyle gets the end cap style.
func (bp *OperationBuffer_BufferParameters) GetEndCapStyle() int {
	return bp.endCapStyle
}

// SetEndCapStyle specifies the end cap style of the generated buffer.
// The styles supported are CAP_ROUND, CAP_FLAT, and CAP_SQUARE.
// The default is CAP_ROUND.
func (bp *OperationBuffer_BufferParameters) SetEndCapStyle(endCapStyle int) {
	bp.endCapStyle = endCapStyle
}

// GetJoinStyle gets the join style.
func (bp *OperationBuffer_BufferParameters) GetJoinStyle() int {
	return bp.joinStyle
}

// SetJoinStyle sets the join style for outside (reflex) corners between line segments.
// The styles supported are JOIN_ROUND, JOIN_MITRE and JOIN_BEVEL.
// The default is JOIN_ROUND.
func (bp *OperationBuffer_BufferParameters) SetJoinStyle(joinStyle int) {
	bp.joinStyle = joinStyle
}

// GetMitreLimit gets the mitre ratio limit.
func (bp *OperationBuffer_BufferParameters) GetMitreLimit() float64 {
	return bp.mitreLimit
}

// SetMitreLimit sets the limit on the mitre ratio used for very sharp corners.
// The mitre ratio is the ratio of the distance from the corner
// to the end of the mitred offset corner.
// When two line segments meet at a sharp angle,
// a miter join will extend far beyond the original geometry.
// (and in the extreme case will be infinitely far.)
// To prevent unreasonable geometry, the mitre limit
// allows controlling the maximum length of the join corner.
// Corners with a ratio which exceed the limit will be beveled.
func (bp *OperationBuffer_BufferParameters) SetMitreLimit(mitreLimit float64) {
	bp.mitreLimit = mitreLimit
}

// SetSingleSided sets whether the computed buffer should be single-sided.
// A single-sided buffer is constructed on only one side of each input line.
//
// The side used is determined by the sign of the buffer distance:
//   - a positive distance indicates the left-hand side
//   - a negative distance indicates the right-hand side
//
// The single-sided buffer of point geometries is
// the same as the regular buffer.
//
// The End Cap Style for single-sided buffers is
// always ignored, and forced to the equivalent of CAP_FLAT.
func (bp *OperationBuffer_BufferParameters) SetSingleSided(isSingleSided bool) {
	bp.isSingleSided = isSingleSided
}

// IsSingleSided tests whether the buffer is to be generated on a single side only.
func (bp *OperationBuffer_BufferParameters) IsSingleSided() bool {
	return bp.isSingleSided
}

// GetSimplifyFactor gets the simplify factor.
func (bp *OperationBuffer_BufferParameters) GetSimplifyFactor() float64 {
	return bp.simplifyFactor
}

// SetSimplifyFactor sets the factor used to determine the simplify distance tolerance
// for input simplification.
// Simplifying can increase the performance of computing buffers.
// Generally the simplify factor should be greater than 0.
// Values between 0.01 and .1 produce relatively good accuracy for the generate buffer.
// Larger values sacrifice accuracy in return for performance.
func (bp *OperationBuffer_BufferParameters) SetSimplifyFactor(simplifyFactor float64) {
	if simplifyFactor < 0 {
		bp.simplifyFactor = 0
	} else {
		bp.simplifyFactor = simplifyFactor
	}
}

// Copy creates a copy of this BufferParameters.
func (bp *OperationBuffer_BufferParameters) Copy() *OperationBuffer_BufferParameters {
	newBp := OperationBuffer_NewBufferParameters()
	newBp.quadrantSegments = bp.quadrantSegments
	newBp.endCapStyle = bp.endCapStyle
	newBp.joinStyle = bp.joinStyle
	newBp.mitreLimit = bp.mitreLimit
	return newBp
}
