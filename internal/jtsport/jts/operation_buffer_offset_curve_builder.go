package jts

import "math"

// OperationBuffer_OffsetCurveBuilder computes the raw offset curve for a
// single Geometry component (ring, line or point).
// A raw offset curve line is not noded -
// it may contain self-intersections (and usually will).
// The final buffer polygon is computed by forming a topological graph
// of all the noded raw curves and tracing outside contours.
// The points in the raw curve are rounded
// to a given PrecisionModel.
//
// Note: this may not produce correct results if the input
// contains repeated or invalid points.
// Repeated points should be removed before calling.
// See CoordinateArrays.removeRepeatedOrInvalidPoints.
type OperationBuffer_OffsetCurveBuilder struct {
	distance       float64
	precisionModel *Geom_PrecisionModel
	bufParams      *OperationBuffer_BufferParameters
}

// OperationBuffer_NewOffsetCurveBuilder creates a new OffsetCurveBuilder.
func OperationBuffer_NewOffsetCurveBuilder(precisionModel *Geom_PrecisionModel, bufParams *OperationBuffer_BufferParameters) *OperationBuffer_OffsetCurveBuilder {
	return &OperationBuffer_OffsetCurveBuilder{
		precisionModel: precisionModel,
		bufParams:      bufParams,
	}
}

// GetBufferParameters gets the buffer parameters being used to generate the curve.
func (ocb *OperationBuffer_OffsetCurveBuilder) GetBufferParameters() *OperationBuffer_BufferParameters {
	return ocb.bufParams
}

// GetLineCurve computes the offset curve for a line.
// This method handles single points as well as LineStrings.
// LineStrings are assumed NOT to be closed (the function will not
// fail for closed lines, but will generate superfluous line caps).
//
// Returns a Coordinate array representing the curve or nil if the curve is empty.
func (ocb *OperationBuffer_OffsetCurveBuilder) GetLineCurve(inputPts []*Geom_Coordinate, distance float64) []*Geom_Coordinate {
	ocb.distance = distance

	if ocb.IsLineOffsetEmpty(distance) {
		return nil
	}

	posDistance := math.Abs(distance)
	segGen := ocb.getSegGen(posDistance)
	if len(inputPts) <= 1 {
		ocb.computePointCurve(inputPts[0], segGen)
	} else {
		if ocb.bufParams.IsSingleSided() {
			isRightSide := distance < 0.0
			ocb.computeSingleSidedBufferCurve(inputPts, isRightSide, segGen)
		} else {
			ocb.computeLineBufferCurve(inputPts, segGen)
		}
	}

	lineCoord := segGen.GetCoordinates()
	return lineCoord
}

// IsLineOffsetEmpty tests whether the offset curve for line or point geometries
// at the given offset distance is empty (does not exist).
// This is the case if:
//   - the distance is zero,
//   - the distance is negative, except for the case of singled-sided buffers
func (ocb *OperationBuffer_OffsetCurveBuilder) IsLineOffsetEmpty(distance float64) bool {
	// a zero width buffer of a line or point is empty
	if distance == 0.0 {
		return true
	}
	// a negative width buffer of a line or point is empty,
	// except for single-sided buffers, where the sign indicates the side
	if distance < 0.0 && !ocb.bufParams.IsSingleSided() {
		return true
	}
	return false
}

// GetRingCurve computes the offset curve for a ring.
// This method handles the degenerate cases of single points and lines,
// as well as valid rings.
//
// Returns a Coordinate array representing the curve, or nil if the curve is empty.
func (ocb *OperationBuffer_OffsetCurveBuilder) GetRingCurve(inputPts []*Geom_Coordinate, side int, distance float64) []*Geom_Coordinate {
	ocb.distance = distance
	if len(inputPts) <= 2 {
		return ocb.GetLineCurve(inputPts, distance)
	}

	// optimize creating ring for for zero distance
	if distance == 0.0 {
		return operationBuffer_offsetCurveBuilder_copyCoordinates(inputPts)
	}
	segGen := ocb.getSegGen(distance)
	ocb.computeRingBufferCurve(inputPts, side, segGen)
	return segGen.GetCoordinates()
}

// GetOffsetCurve computes the offset curve for a coordinate sequence.
func (ocb *OperationBuffer_OffsetCurveBuilder) GetOffsetCurve(inputPts []*Geom_Coordinate, distance float64) []*Geom_Coordinate {
	ocb.distance = distance

	// a zero width offset curve is empty
	if distance == 0.0 {
		return nil
	}

	isRightSide := distance < 0.0
	posDistance := math.Abs(distance)
	segGen := ocb.getSegGen(posDistance)
	if len(inputPts) <= 1 {
		ocb.computePointCurve(inputPts[0], segGen)
	} else {
		ocb.computeOffsetCurve(inputPts, isRightSide, segGen)
	}
	curvePts := segGen.GetCoordinates()
	// for right side line is traversed in reverse direction, so have to reverse generated line
	if isRightSide {
		Geom_CoordinateArrays_Reverse(curvePts)
	}
	return curvePts
}

func operationBuffer_offsetCurveBuilder_copyCoordinates(pts []*Geom_Coordinate) []*Geom_Coordinate {
	copyArr := make([]*Geom_Coordinate, len(pts))
	for i := 0; i < len(copyArr); i++ {
		copyArr[i] = pts[i].Copy()
	}
	return copyArr
}

func (ocb *OperationBuffer_OffsetCurveBuilder) getSegGen(distance float64) *operationBuffer_OffsetSegmentGenerator {
	return operationBuffer_newOffsetSegmentGenerator(ocb.precisionModel, ocb.bufParams, distance)
}

// simplifyTolerance computes the distance tolerance to use during input
// line simplification.
func (ocb *OperationBuffer_OffsetCurveBuilder) simplifyTolerance(bufDistance float64) float64 {
	return bufDistance * ocb.bufParams.GetSimplifyFactor()
}

func (ocb *OperationBuffer_OffsetCurveBuilder) computePointCurve(pt *Geom_Coordinate, segGen *operationBuffer_OffsetSegmentGenerator) {
	switch ocb.bufParams.GetEndCapStyle() {
	case OperationBuffer_BufferParameters_CAP_ROUND:
		segGen.CreateCircle(pt)
	case OperationBuffer_BufferParameters_CAP_SQUARE:
		segGen.CreateSquare(pt)
		// otherwise curve is empty (e.g. for a butt cap);
	}
}

func (ocb *OperationBuffer_OffsetCurveBuilder) computeLineBufferCurve(inputPts []*Geom_Coordinate, segGen *operationBuffer_OffsetSegmentGenerator) {
	distTol := ocb.simplifyTolerance(ocb.distance)

	//--------- compute points for left side of line
	// Simplify the appropriate side of the line before generating
	simp1 := OperationBuffer_BufferInputLineSimplifier_Simplify(inputPts, distTol)

	n1 := len(simp1) - 1
	segGen.InitSideSegments(simp1[0], simp1[1], Geom_Position_Left)
	for i := 2; i <= n1; i++ {
		segGen.AddNextSegment(simp1[i], true)
	}
	segGen.AddLastSegment()
	// add line cap for end of line
	segGen.AddLineEndCap(simp1[n1-1], simp1[n1])

	//---------- compute points for right side of line
	// Simplify the appropriate side of the line before generating
	simp2 := OperationBuffer_BufferInputLineSimplifier_Simplify(inputPts, -distTol)
	n2 := len(simp2) - 1

	// since we are traversing line in opposite order, offset position is still LEFT
	segGen.InitSideSegments(simp2[n2], simp2[n2-1], Geom_Position_Left)
	for i := n2 - 2; i >= 0; i-- {
		segGen.AddNextSegment(simp2[i], true)
	}
	segGen.AddLastSegment()
	// add line cap for start of line
	segGen.AddLineEndCap(simp2[1], simp2[0])

	segGen.CloseRing()
}

func (ocb *OperationBuffer_OffsetCurveBuilder) computeSingleSidedBufferCurve(inputPts []*Geom_Coordinate, isRightSide bool, segGen *operationBuffer_OffsetSegmentGenerator) {
	distTol := ocb.simplifyTolerance(ocb.distance)

	if isRightSide {
		// add original line
		segGen.AddSegments(inputPts, true)

		//---------- compute points for right side of line
		// Simplify the appropriate side of the line before generating
		simp2 := OperationBuffer_BufferInputLineSimplifier_Simplify(inputPts, -distTol)
		n2 := len(simp2) - 1

		// since we are traversing line in opposite order, offset position is still LEFT
		segGen.InitSideSegments(simp2[n2], simp2[n2-1], Geom_Position_Left)
		segGen.AddFirstSegment()
		for i := n2 - 2; i >= 0; i-- {
			segGen.AddNextSegment(simp2[i], true)
		}
	} else {
		// add original line
		segGen.AddSegments(inputPts, false)

		//--------- compute points for left side of line
		// Simplify the appropriate side of the line before generating
		simp1 := OperationBuffer_BufferInputLineSimplifier_Simplify(inputPts, distTol)

		n1 := len(simp1) - 1
		segGen.InitSideSegments(simp1[0], simp1[1], Geom_Position_Left)
		segGen.AddFirstSegment()
		for i := 2; i <= n1; i++ {
			segGen.AddNextSegment(simp1[i], true)
		}
	}
	segGen.AddLastSegment()
	segGen.CloseRing()
}

func (ocb *OperationBuffer_OffsetCurveBuilder) computeOffsetCurve(inputPts []*Geom_Coordinate, isRightSide bool, segGen *operationBuffer_OffsetSegmentGenerator) {
	distTol := ocb.simplifyTolerance(math.Abs(ocb.distance))

	if isRightSide {
		//---------- compute points for right side of line
		// Simplify the appropriate side of the line before generating
		simp2 := OperationBuffer_BufferInputLineSimplifier_Simplify(inputPts, -distTol)
		n2 := len(simp2) - 1

		// since we are traversing line in opposite order, offset position is still LEFT
		segGen.InitSideSegments(simp2[n2], simp2[n2-1], Geom_Position_Left)
		segGen.AddFirstSegment()
		for i := n2 - 2; i >= 0; i-- {
			segGen.AddNextSegment(simp2[i], true)
		}
	} else {
		//--------- compute points for left side of line
		// Simplify the appropriate side of the line before generating
		simp1 := OperationBuffer_BufferInputLineSimplifier_Simplify(inputPts, distTol)

		n1 := len(simp1) - 1
		segGen.InitSideSegments(simp1[0], simp1[1], Geom_Position_Left)
		segGen.AddFirstSegment()
		for i := 2; i <= n1; i++ {
			segGen.AddNextSegment(simp1[i], true)
		}
	}
	segGen.AddLastSegment()
}

func (ocb *OperationBuffer_OffsetCurveBuilder) computeRingBufferCurve(inputPts []*Geom_Coordinate, side int, segGen *operationBuffer_OffsetSegmentGenerator) {
	// simplify input line to improve performance
	distTol := ocb.simplifyTolerance(ocb.distance)
	// ensure that correct side is simplified
	if side == Geom_Position_Right {
		distTol = -distTol
	}
	simp := OperationBuffer_BufferInputLineSimplifier_Simplify(inputPts, distTol)

	n := len(simp) - 1
	segGen.InitSideSegments(simp[n-1], simp[0], side)
	for i := 1; i <= n; i++ {
		addStartPoint := i != 1
		segGen.AddNextSegment(simp[i], addStartPoint)
	}
	segGen.CloseRing()
}
