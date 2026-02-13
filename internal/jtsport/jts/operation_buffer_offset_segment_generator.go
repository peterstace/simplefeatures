package jts

import "math"

// operationBuffer_OffsetSegmentGenerator_OFFSET_SEGMENT_SEPARATION_FACTOR is the factor controlling how close offset segments can be to
// skip adding a fillet or mitre.
// This eliminates very short fillet segments,
// reduces the number of offset curve vertices.
// and improves the robustness of mitre construction.
const operationBuffer_OffsetSegmentGenerator_OFFSET_SEGMENT_SEPARATION_FACTOR = 0.05

// operationBuffer_OffsetSegmentGenerator_INSIDE_TURN_VERTEX_SNAP_DISTANCE_FACTOR is the factor controlling how close curve vertices on inside turns can be to be snapped.
const operationBuffer_OffsetSegmentGenerator_INSIDE_TURN_VERTEX_SNAP_DISTANCE_FACTOR = 1.0e-3

// operationBuffer_OffsetSegmentGenerator_CURVE_VERTEX_SNAP_DISTANCE_FACTOR is the factor which controls how close curve vertices can be to be snapped.
const operationBuffer_OffsetSegmentGenerator_CURVE_VERTEX_SNAP_DISTANCE_FACTOR = 1.0e-6

// operationBuffer_OffsetSegmentGenerator_MAX_CLOSING_SEG_LEN_FACTOR is the factor which determines how short closing segs can be for round buffers.
const operationBuffer_OffsetSegmentGenerator_MAX_CLOSING_SEG_LEN_FACTOR = 80

// operationBuffer_OffsetSegmentGenerator generates segments which form an offset curve.
// Supports all end cap and join options provided for buffering.
// This algorithm implements various heuristics to produce smoother, simpler curves which are
// still within a reasonable tolerance of the true curve.
type operationBuffer_OffsetSegmentGenerator struct {
	// maxCurveSegmentError is the max error of approximation (distance) between a quad segment and the true fillet curve.
	maxCurveSegmentError float64

	// filletAngleQuantum is the angle quantum with which to approximate a fillet curve
	// (based on the input # of quadrant segments).
	filletAngleQuantum float64

	// closingSegLengthFactor controls how long "closing segments" are.
	// Closing segments are added at the middle of inside corners to ensure a smoother
	// boundary for the buffer offset curve.
	// In some cases (particularly for round joins with default-or-better
	// quantization) the closing segments can be made quite short.
	// This substantially improves performance (due to fewer intersections being created).
	//
	// A closingSegFactor of 0 results in lines to the corner vertex
	// A closingSegFactor of 1 results in lines halfway to the corner vertex
	// A closingSegFactor of 80 results in lines 1/81 of the way to the corner vertex
	// (this option is reasonable for the very common default situation of round joins
	// and quadrantSegs >= 8)
	closingSegLengthFactor int

	segList        *operationBuffer_OffsetSegmentString
	distance       float64
	precisionModel *Geom_PrecisionModel
	bufParams      *OperationBuffer_BufferParameters
	li             *Algorithm_LineIntersector

	s0      *Geom_Coordinate
	s1      *Geom_Coordinate
	s2      *Geom_Coordinate
	seg0    *Geom_LineSegment
	seg1    *Geom_LineSegment
	offset0 *Geom_LineSegment
	offset1 *Geom_LineSegment
	side    int

	hasNarrowConcaveAngle bool
}

// operationBuffer_newOffsetSegmentGenerator creates a new OffsetSegmentGenerator.
func operationBuffer_newOffsetSegmentGenerator(precisionModel *Geom_PrecisionModel, bufParams *OperationBuffer_BufferParameters, distance float64) *operationBuffer_OffsetSegmentGenerator {
	osg := &operationBuffer_OffsetSegmentGenerator{
		precisionModel:         precisionModel,
		bufParams:              bufParams,
		closingSegLengthFactor: 1,
		seg0:                   Geom_NewLineSegment(),
		seg1:                   Geom_NewLineSegment(),
		offset0:                Geom_NewLineSegment(),
		offset1:                Geom_NewLineSegment(),
		side:                   0,
	}

	// compute intersections in full precision, to provide accuracy
	// the points are rounded as they are inserted into the curve line
	osg.li = Algorithm_NewRobustLineIntersector().Algorithm_LineIntersector

	quadSegs := bufParams.GetQuadrantSegments()
	if quadSegs < 1 {
		quadSegs = 1
	}
	osg.filletAngleQuantum = Algorithm_Angle_PiOver2 / float64(quadSegs)

	// Non-round joins cause issues with short closing segments, so don't use
	// them. In any case, non-round joins only really make sense for relatively
	// small buffer distances.
	if bufParams.GetQuadrantSegments() >= 8 && bufParams.GetJoinStyle() == OperationBuffer_BufferParameters_JOIN_ROUND {
		osg.closingSegLengthFactor = operationBuffer_OffsetSegmentGenerator_MAX_CLOSING_SEG_LEN_FACTOR
	}
	osg.init(distance)
	return osg
}

// HasNarrowConcaveAngle tests whether the input has a narrow concave angle
// (relative to the offset distance).
// In this case the generated offset curve will contain self-intersections
// and heuristic closing segments.
// This is expected behaviour in the case of Buffer curves.
// For pure Offset Curves, the output needs to be further treated before it can be used.
func (osg *operationBuffer_OffsetSegmentGenerator) HasNarrowConcaveAngle() bool {
	return osg.hasNarrowConcaveAngle
}

func (osg *operationBuffer_OffsetSegmentGenerator) init(distance float64) {
	osg.distance = math.Abs(distance)
	osg.maxCurveSegmentError = osg.distance * (1 - math.Cos(osg.filletAngleQuantum/2.0))
	osg.segList = operationBuffer_newOffsetSegmentString()
	osg.segList.SetPrecisionModel(osg.precisionModel)
	// Choose the min vertex separation as a small fraction of the offset distance.
	osg.segList.SetMinimumVertexDistance(osg.distance * operationBuffer_OffsetSegmentGenerator_CURVE_VERTEX_SNAP_DISTANCE_FACTOR)
}

// InitSideSegments initializes the side segments for offset curve generation.
func (osg *operationBuffer_OffsetSegmentGenerator) InitSideSegments(s1 *Geom_Coordinate, s2 *Geom_Coordinate, side int) {
	osg.s1 = s1
	osg.s2 = s2
	osg.side = side
	osg.seg1.SetCoordinates(s1, s2)
	operationBuffer_OffsetSegmentGenerator_computeOffsetSegment(osg.seg1, side, osg.distance, osg.offset1)
}

// GetCoordinates returns the coordinates of the offset curve.
func (osg *operationBuffer_OffsetSegmentGenerator) GetCoordinates() []*Geom_Coordinate {
	pts := osg.segList.GetCoordinates()
	return pts
}

// CloseRing closes the ring.
func (osg *operationBuffer_OffsetSegmentGenerator) CloseRing() {
	osg.segList.CloseRing()
}

// AddSegments adds an array of segments to the offset curve.
func (osg *operationBuffer_OffsetSegmentGenerator) AddSegments(pt []*Geom_Coordinate, isForward bool) {
	osg.segList.AddPts(pt, isForward)
}

// AddFirstSegment adds the first segment point.
func (osg *operationBuffer_OffsetSegmentGenerator) AddFirstSegment() {
	osg.segList.AddPt(osg.offset1.P0)
}

// AddLastSegment adds the last offset point.
func (osg *operationBuffer_OffsetSegmentGenerator) AddLastSegment() {
	osg.segList.AddPt(osg.offset1.P1)
}

// AddNextSegment adds the next segment to the offset curve.
func (osg *operationBuffer_OffsetSegmentGenerator) AddNextSegment(p *Geom_Coordinate, addStartPoint bool) {
	// s0-s1-s2 are the coordinates of the previous segment and the current one
	osg.s0 = osg.s1
	osg.s1 = osg.s2
	osg.s2 = p
	osg.seg0.SetCoordinates(osg.s0, osg.s1)
	operationBuffer_OffsetSegmentGenerator_computeOffsetSegment(osg.seg0, osg.side, osg.distance, osg.offset0)
	osg.seg1.SetCoordinates(osg.s1, osg.s2)
	operationBuffer_OffsetSegmentGenerator_computeOffsetSegment(osg.seg1, osg.side, osg.distance, osg.offset1)

	// do nothing if points are equal
	if osg.s1.Equals(osg.s2) {
		return
	}

	orientation := Algorithm_Orientation_Index(osg.s0, osg.s1, osg.s2)
	outsideTurn := (orientation == Algorithm_Orientation_Clockwise && osg.side == Geom_Position_Left) ||
		(orientation == Algorithm_Orientation_Counterclockwise && osg.side == Geom_Position_Right)

	if orientation == 0 { // lines are collinear
		osg.addCollinear(addStartPoint)
	} else if outsideTurn {
		osg.addOutsideTurn(orientation, addStartPoint)
	} else { // inside turn
		osg.addInsideTurn(orientation, addStartPoint)
	}
}

func (osg *operationBuffer_OffsetSegmentGenerator) addCollinear(addStartPoint bool) {
	// This test could probably be done more efficiently,
	// but the situation of exact collinearity should be fairly rare.
	osg.li.ComputeIntersection(osg.s0, osg.s1, osg.s1, osg.s2)
	numInt := osg.li.GetIntersectionNum()
	// if numInt is < 2, the lines are parallel and in the same direction. In
	// this case the point can be ignored, since the offset lines will also be
	// parallel.
	if numInt >= 2 {
		// segments are collinear but reversing.
		// Add an "end-cap" fillet all the way around to other direction.
		// This case should ONLY happen for LineStrings, so the orientation is always CW.
		// (Polygons can never have two consecutive segments which are parallel but reversed,
		// because that would be a self intersection.)
		if osg.bufParams.GetJoinStyle() == OperationBuffer_BufferParameters_JOIN_BEVEL ||
			osg.bufParams.GetJoinStyle() == OperationBuffer_BufferParameters_JOIN_MITRE {
			if addStartPoint {
				osg.segList.AddPt(osg.offset0.P1)
			}
			osg.segList.AddPt(osg.offset1.P0)
		} else {
			osg.addCornerFillet(osg.s1, osg.offset0.P1, osg.offset1.P0, Algorithm_Orientation_Clockwise, osg.distance)
		}
	}
}

// addOutsideTurn adds the offset points for an outside (convex) turn.
func (osg *operationBuffer_OffsetSegmentGenerator) addOutsideTurn(orientation int, addStartPoint bool) {
	// Heuristic: If offset endpoints are very close together,
	// (which happens for nearly-parallel segments),
	// use an endpoint as the single offset corner vertex.
	// This eliminates very short single-segment joins
	// and reduces the number of offset curve vertices.
	// This also avoids robustness problems with computing mitre corners
	// for nearly-parallel segments.
	if osg.offset0.P1.Distance(osg.offset1.P0) < osg.distance*operationBuffer_OffsetSegmentGenerator_OFFSET_SEGMENT_SEPARATION_FACTOR {
		// use endpoint of longest segment, to reduce change in area
		segLen0 := osg.s0.Distance(osg.s1)
		segLen1 := osg.s1.Distance(osg.s2)
		var offsetPt *Geom_Coordinate
		if segLen0 > segLen1 {
			offsetPt = osg.offset0.P1
		} else {
			offsetPt = osg.offset1.P0
		}
		osg.segList.AddPt(offsetPt)
		return
	}

	if osg.bufParams.GetJoinStyle() == OperationBuffer_BufferParameters_JOIN_MITRE {
		osg.addMitreJoin(osg.s1, osg.offset0, osg.offset1, osg.distance)
	} else if osg.bufParams.GetJoinStyle() == OperationBuffer_BufferParameters_JOIN_BEVEL {
		osg.addBevelJoin(osg.offset0, osg.offset1)
	} else {
		// add a circular fillet connecting the endpoints of the offset segments
		if addStartPoint {
			osg.segList.AddPt(osg.offset0.P1)
		}
		osg.addCornerFillet(osg.s1, osg.offset0.P1, osg.offset1.P0, orientation, osg.distance)
		osg.segList.AddPt(osg.offset1.P0)
	}
}

// addInsideTurn adds the offset points for an inside (concave) turn.
func (osg *operationBuffer_OffsetSegmentGenerator) addInsideTurn(orientation int, addStartPoint bool) {
	// add intersection point of offset segments (if any)
	osg.li.ComputeIntersection(osg.offset0.P0, osg.offset0.P1, osg.offset1.P0, osg.offset1.P1)
	if osg.li.HasIntersection() {
		osg.segList.AddPt(osg.li.GetIntersection(0))
	} else {
		// If no intersection is detected,
		// it means the angle is so small and/or the offset so
		// large that the offsets segments don't intersect.
		// In this case we must add a "closing segment" to make sure the buffer curve is continuous,
		// fairly smooth (e.g. no sharp reversals in direction)
		// and tracks the buffer correctly around the corner. The curve connects
		// the endpoints of the segment offsets to points
		// which lie toward the centre point of the corner.
		// The joining curve will not appear in the final buffer outline, since it
		// is completely internal to the buffer polygon.
		//
		// In complex buffer cases the closing segment may cut across many other
		// segments in the generated offset curve. In order to improve the
		// performance of the noding, the closing segment should be kept as short as possible.
		// (But not too short, since that would defeat its purpose).
		// This is the purpose of the closingSegFactor heuristic value.

		// The intersection test above is vulnerable to robustness errors; i.e. it
		// may be that the offsets should intersect very close to their endpoints,
		// but aren't reported as such due to rounding. To handle this situation
		// appropriately, we use the following test: If the offset points are very
		// close, don't add closing segments but simply use one of the offset
		// points
		osg.hasNarrowConcaveAngle = true
		if osg.offset0.P1.Distance(osg.offset1.P0) < osg.distance*operationBuffer_OffsetSegmentGenerator_INSIDE_TURN_VERTEX_SNAP_DISTANCE_FACTOR {
			osg.segList.AddPt(osg.offset0.P1)
		} else {
			// add endpoint of this segment offset
			osg.segList.AddPt(osg.offset0.P1)

			// Add "closing segment" of required length.
			if osg.closingSegLengthFactor > 0 {
				mid0 := Geom_NewCoordinateWithXY(
					(float64(osg.closingSegLengthFactor)*osg.offset0.P1.X+osg.s1.X)/float64(osg.closingSegLengthFactor+1),
					(float64(osg.closingSegLengthFactor)*osg.offset0.P1.Y+osg.s1.Y)/float64(osg.closingSegLengthFactor+1))
				osg.segList.AddPt(mid0)
				mid1 := Geom_NewCoordinateWithXY(
					(float64(osg.closingSegLengthFactor)*osg.offset1.P0.X+osg.s1.X)/float64(osg.closingSegLengthFactor+1),
					(float64(osg.closingSegLengthFactor)*osg.offset1.P0.Y+osg.s1.Y)/float64(osg.closingSegLengthFactor+1))
				osg.segList.AddPt(mid1)
			} else {
				// This branch is not expected to be used except for testing purposes.
				// It is equivalent to the JTS 1.9 logic for closing segments
				// (which results in very poor performance for large buffer distances)
				osg.segList.AddPt(osg.s1)
			}

			// add start point of next segment offset
			osg.segList.AddPt(osg.offset1.P0)
		}
	}
}

// operationBuffer_OffsetSegmentGenerator_computeOffsetSegment computes an offset segment for an input segment on a given side and at a given distance.
// The offset points are computed in full double precision, for accuracy.
func operationBuffer_OffsetSegmentGenerator_computeOffsetSegment(seg *Geom_LineSegment, side int, distance float64, offset *Geom_LineSegment) {
	sideSign := 1
	if side != Geom_Position_Left {
		sideSign = -1
	}
	dx := seg.P1.X - seg.P0.X
	dy := seg.P1.Y - seg.P0.Y
	length := math.Hypot(dx, dy)
	// u is the vector that is the length of the offset, in the direction of the segment
	ux := float64(sideSign) * distance * dx / length
	uy := float64(sideSign) * distance * dy / length
	offset.P0.X = seg.P0.X - uy
	offset.P0.Y = seg.P0.Y + ux
	offset.P1.X = seg.P1.X - uy
	offset.P1.Y = seg.P1.Y + ux
}

// AddLineEndCap adds an end cap around point p1, terminating a line segment coming from p0.
func (osg *operationBuffer_OffsetSegmentGenerator) AddLineEndCap(p0 *Geom_Coordinate, p1 *Geom_Coordinate) {
	seg := Geom_NewLineSegmentFromCoordinates(p0, p1)

	offsetL := Geom_NewLineSegment()
	operationBuffer_OffsetSegmentGenerator_computeOffsetSegment(seg, Geom_Position_Left, osg.distance, offsetL)
	offsetR := Geom_NewLineSegment()
	operationBuffer_OffsetSegmentGenerator_computeOffsetSegment(seg, Geom_Position_Right, osg.distance, offsetR)

	dx := p1.X - p0.X
	dy := p1.Y - p0.Y
	angle := math.Atan2(dy, dx)

	switch osg.bufParams.GetEndCapStyle() {
	case OperationBuffer_BufferParameters_CAP_ROUND:
		// add offset seg points with a fillet between them
		osg.segList.AddPt(offsetL.P1)
		osg.addDirectedFillet(p1, angle+Algorithm_Angle_PiOver2, angle-Algorithm_Angle_PiOver2, Algorithm_Orientation_Clockwise, osg.distance)
		osg.segList.AddPt(offsetR.P1)
	case OperationBuffer_BufferParameters_CAP_FLAT:
		// only offset segment points are added
		osg.segList.AddPt(offsetL.P1)
		osg.segList.AddPt(offsetR.P1)
	case OperationBuffer_BufferParameters_CAP_SQUARE:
		// add a square defined by extensions of the offset segment endpoints
		squareCapSideOffset := Geom_NewCoordinate()
		squareCapSideOffset.X = math.Abs(osg.distance) * Algorithm_Angle_CosSnap(angle)
		squareCapSideOffset.Y = math.Abs(osg.distance) * Algorithm_Angle_SinSnap(angle)

		squareCapLOffset := Geom_NewCoordinateWithXY(
			offsetL.P1.X+squareCapSideOffset.X,
			offsetL.P1.Y+squareCapSideOffset.Y)
		squareCapROffset := Geom_NewCoordinateWithXY(
			offsetR.P1.X+squareCapSideOffset.X,
			offsetR.P1.Y+squareCapSideOffset.Y)
		osg.segList.AddPt(squareCapLOffset)
		osg.segList.AddPt(squareCapROffset)
	}
}

// addMitreJoin adds a mitre join connecting two convex offset segments.
// The mitre is beveled if it exceeds the mitre limit factor.
// The mitre limit is intended to prevent extremely long corners occurring.
// If the mitre limit is very small it can cause unwanted artifacts around fairly flat corners.
// This is prevented by using a simple bevel join in this case.
// In other words, the limit prevents the corner from getting too long,
// but it won't force it to be very short/flat.
func (osg *operationBuffer_OffsetSegmentGenerator) addMitreJoin(cornerPt *Geom_Coordinate, offset0 *Geom_LineSegment, offset1 *Geom_LineSegment, distance float64) {
	mitreLimitDistance := osg.bufParams.GetMitreLimit() * distance
	// First try a non-beveled join.
	// Compute the intersection point of the lines determined by the offsets.
	// Parallel or collinear lines will return a null point ==> need to be beveled
	//
	// Note: This computation is unstable if the offset segments are nearly collinear.
	// However, this situation should have been eliminated earlier by the check
	// for whether the offset segment endpoints are almost coincident
	intPt := Algorithm_Intersection_Intersection(offset0.P0, offset0.P1, offset1.P0, offset1.P1)
	if intPt != nil && intPt.Distance(cornerPt) <= mitreLimitDistance {
		osg.segList.AddPt(intPt)
		return
	}
	// In case the mitre limit is very small, try a plain bevel.
	// Use it if it's further than the limit.
	bevelDist := Algorithm_Distance_PointToSegment(cornerPt, offset0.P1, offset1.P0)
	if bevelDist >= mitreLimitDistance {
		osg.addBevelJoin(offset0, offset1)
		return
	}
	// Have to construct a limited mitre bevel.
	osg.addLimitedMitreJoin(offset0, offset1, distance, mitreLimitDistance)
}

// addLimitedMitreJoin adds a limited mitre join connecting two convex offset segments.
// A limited mitre join is beveled at the distance determined by the mitre limit factor,
// or as a standard bevel join, whichever is further.
func (osg *operationBuffer_OffsetSegmentGenerator) addLimitedMitreJoin(offset0 *Geom_LineSegment, offset1 *Geom_LineSegment, distance float64, mitreLimitDistance float64) {
	cornerPt := osg.seg0.P1
	// oriented angle of the corner formed by segments
	angInterior := Algorithm_Angle_AngleBetweenOriented(osg.seg0.P0, cornerPt, osg.seg1.P1)
	// half of the interior angle
	angInterior2 := angInterior / 2

	// direction of bisector of the interior angle between the segments
	dir0 := Algorithm_Angle_AngleBetweenPoints(cornerPt, osg.seg0.P0)
	dirBisector := Algorithm_Angle_Normalize(dir0 + angInterior2)

	// midpoint of the bevel segment
	bevelMidPt := operationBuffer_OffsetSegmentGenerator_project(cornerPt, -mitreLimitDistance, dirBisector)

	// direction of bevel segment (at right angle to corner bisector)
	dirBevel := Algorithm_Angle_Normalize(dirBisector + Algorithm_Angle_PiOver2)

	// compute the candidate bevel segment by projecting both sides of the midpoint
	bevel0 := operationBuffer_OffsetSegmentGenerator_project(bevelMidPt, distance, dirBevel)
	bevel1 := operationBuffer_OffsetSegmentGenerator_project(bevelMidPt, distance, dirBevel+math.Pi)

	// compute actual bevel segment between the offset lines
	bevelInt0 := Algorithm_Intersection_LineSegment(offset0.P0, offset0.P1, bevel0, bevel1)
	bevelInt1 := Algorithm_Intersection_LineSegment(offset1.P0, offset1.P1, bevel0, bevel1)

	// add the limited bevel, if it intersects the offsets
	if bevelInt0 != nil && bevelInt1 != nil {
		osg.segList.AddPt(bevelInt0)
		osg.segList.AddPt(bevelInt1)
		return
	}
	// If the corner is very flat or the mitre limit is very small
	// the limited bevel segment may not intersect the offsets.
	// In this case just bevel the join.
	osg.addBevelJoin(offset0, offset1)
}

// operationBuffer_OffsetSegmentGenerator_project projects a point to a given distance in a given direction angle.
func operationBuffer_OffsetSegmentGenerator_project(pt *Geom_Coordinate, d float64, dir float64) *Geom_Coordinate {
	x := pt.X + d*Algorithm_Angle_CosSnap(dir)
	y := pt.Y + d*Algorithm_Angle_SinSnap(dir)
	return Geom_NewCoordinateWithXY(x, y)
}

// addBevelJoin adds a bevel join connecting two offset segments around a convex corner.
func (osg *operationBuffer_OffsetSegmentGenerator) addBevelJoin(offset0 *Geom_LineSegment, offset1 *Geom_LineSegment) {
	osg.segList.AddPt(offset0.P1)
	osg.segList.AddPt(offset1.P0)
}

// addCornerFillet adds points for a circular fillet around a convex corner.
// Adds the start and end points.
func (osg *operationBuffer_OffsetSegmentGenerator) addCornerFillet(p *Geom_Coordinate, p0 *Geom_Coordinate, p1 *Geom_Coordinate, direction int, radius float64) {
	dx0 := p0.X - p.X
	dy0 := p0.Y - p.Y
	startAngle := math.Atan2(dy0, dx0)
	dx1 := p1.X - p.X
	dy1 := p1.Y - p.Y
	endAngle := math.Atan2(dy1, dx1)

	if direction == Algorithm_Orientation_Clockwise {
		if startAngle <= endAngle {
			startAngle += Algorithm_Angle_PiTimes2
		}
	} else { // direction == COUNTERCLOCKWISE
		if startAngle >= endAngle {
			startAngle -= Algorithm_Angle_PiTimes2
		}
	}
	osg.segList.AddPt(p0)
	osg.addDirectedFillet(p, startAngle, endAngle, direction, radius)
	osg.segList.AddPt(p1)
}

// addDirectedFillet adds points for a circular fillet arc between two specified angles.
// The start and end point for the fillet are not added - the caller must add them if required.
func (osg *operationBuffer_OffsetSegmentGenerator) addDirectedFillet(p *Geom_Coordinate, startAngle float64, endAngle float64, direction int, radius float64) {
	directionFactor := -1
	if direction != Algorithm_Orientation_Clockwise {
		directionFactor = 1
	}

	totalAngle := math.Abs(startAngle - endAngle)
	nSegs := int(totalAngle/osg.filletAngleQuantum + 0.5)

	if nSegs < 1 {
		return // no segments because angle is less than increment - nothing to do!
	}

	// choose angle increment so that each segment has equal length
	angleInc := totalAngle / float64(nSegs)

	pt := Geom_NewCoordinate()
	for i := 0; i < nSegs; i++ {
		angle := startAngle + float64(directionFactor)*float64(i)*angleInc
		pt.X = p.X + radius*Algorithm_Angle_CosSnap(angle)
		pt.Y = p.Y + radius*Algorithm_Angle_SinSnap(angle)
		osg.segList.AddPt(pt)
	}
}

// CreateCircle creates a CW circle around a point.
func (osg *operationBuffer_OffsetSegmentGenerator) CreateCircle(p *Geom_Coordinate) {
	// add start point
	pt := Geom_NewCoordinateWithXY(p.X+osg.distance, p.Y)
	osg.segList.AddPt(pt)
	osg.addDirectedFillet(p, 0.0, Algorithm_Angle_PiTimes2, -1, osg.distance)
	osg.segList.CloseRing()
}

// CreateSquare creates a CW square around a point.
func (osg *operationBuffer_OffsetSegmentGenerator) CreateSquare(p *Geom_Coordinate) {
	osg.segList.AddPt(Geom_NewCoordinateWithXY(p.X+osg.distance, p.Y+osg.distance))
	osg.segList.AddPt(Geom_NewCoordinateWithXY(p.X+osg.distance, p.Y-osg.distance))
	osg.segList.AddPt(Geom_NewCoordinateWithXY(p.X-osg.distance, p.Y-osg.distance))
	osg.segList.AddPt(Geom_NewCoordinateWithXY(p.X-osg.distance, p.Y+osg.distance))
	osg.segList.CloseRing()
}
