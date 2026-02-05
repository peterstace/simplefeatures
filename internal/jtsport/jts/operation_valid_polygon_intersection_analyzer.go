package jts

// operationValid_PolygonIntersectionAnalyzer_noInvalidIntersection is the
// sentinel value indicating no invalid intersection was found.
const operationValid_PolygonIntersectionAnalyzer_noInvalidIntersection = -1

// OperationValid_PolygonIntersectionAnalyzer finds and analyzes intersections
// in and between polygons, to determine if they are valid.
//
// The Noding_SegmentStrings which are analyzed can have OperationValid_PolygonRings
// attached. If so they will be updated with intersection information
// to support further validity analysis which must be done after
// basic intersection validity has been confirmed.
type OperationValid_PolygonIntersectionAnalyzer struct {
	isInvertedRingValid bool

	li              *Algorithm_LineIntersector
	invalidCode     int
	invalidLocation *Geom_Coordinate

	hasDoubleTouch      bool
	doubleTouchLocation *Geom_Coordinate
}

var _ Noding_SegmentIntersector = (*OperationValid_PolygonIntersectionAnalyzer)(nil)

// OperationValid_NewPolygonIntersectionAnalyzer creates a new analyzer,
// allowing for the mode where inverted rings are valid.
func OperationValid_NewPolygonIntersectionAnalyzer(isInvertedRingValid bool) *OperationValid_PolygonIntersectionAnalyzer {
	return &OperationValid_PolygonIntersectionAnalyzer{
		isInvertedRingValid: isInvertedRingValid,
		li:                  Algorithm_NewRobustLineIntersector().Algorithm_LineIntersector,
		invalidCode:         operationValid_PolygonIntersectionAnalyzer_noInvalidIntersection,
	}
}

// TRANSLITERATION NOTE: Marker method required for Go interface implementation.
// Java implements SegmentIntersector interface implicitly.
func (pia *OperationValid_PolygonIntersectionAnalyzer) IsNoding_SegmentIntersector() {}

// IsDone reports whether the client needs to continue testing all
// intersections in an arrangement.
func (pia *OperationValid_PolygonIntersectionAnalyzer) IsDone() bool {
	return pia.IsInvalid() || pia.hasDoubleTouch
}

// IsInvalid reports whether an invalid intersection was found.
func (pia *OperationValid_PolygonIntersectionAnalyzer) IsInvalid() bool {
	return pia.invalidCode >= 0
}

// GetInvalidCode returns the code indicating the type of invalid intersection,
// or -1 if none was found.
func (pia *OperationValid_PolygonIntersectionAnalyzer) GetInvalidCode() int {
	return pia.invalidCode
}

// GetInvalidLocation returns the location of the invalid intersection.
func (pia *OperationValid_PolygonIntersectionAnalyzer) GetInvalidLocation() *Geom_Coordinate {
	return pia.invalidLocation
}

// HasDoubleTouch reports whether a double touch was found between rings.
func (pia *OperationValid_PolygonIntersectionAnalyzer) HasDoubleTouch() bool {
	return pia.hasDoubleTouch
}

// GetDoubleTouchLocation returns the location of the double touch.
func (pia *OperationValid_PolygonIntersectionAnalyzer) GetDoubleTouchLocation() *Geom_Coordinate {
	return pia.doubleTouchLocation
}

// ProcessIntersections is called by clients to process intersections for
// two segments of the SegmentStrings being intersected.
func (pia *OperationValid_PolygonIntersectionAnalyzer) ProcessIntersections(ss0 Noding_SegmentString, segIndex0 int, ss1 Noding_SegmentString, segIndex1 int) {
	// don't test a segment with itself
	isSameSegString := ss0 == ss1
	isSameSegment := isSameSegString && segIndex0 == segIndex1
	if isSameSegment {
		return
	}

	code := pia.findInvalidIntersection(ss0, segIndex0, ss1, segIndex1)
	// Ensure that invalidCode is only set once,
	// since the short-circuiting in SegmentIntersector is not guaranteed
	// to happen immediately.
	if code != operationValid_PolygonIntersectionAnalyzer_noInvalidIntersection {
		pia.invalidCode = code
		pia.invalidLocation = pia.li.GetIntersection(0)
	}
}

func (pia *OperationValid_PolygonIntersectionAnalyzer) findInvalidIntersection(ss0 Noding_SegmentString, segIndex0 int, ss1 Noding_SegmentString, segIndex1 int) int {
	p00 := ss0.GetCoordinate(segIndex0)
	p01 := ss0.GetCoordinate(segIndex0 + 1)
	p10 := ss1.GetCoordinate(segIndex1)
	p11 := ss1.GetCoordinate(segIndex1 + 1)

	pia.li.ComputeIntersection(p00, p01, p10, p11)

	if !pia.li.HasIntersection() {
		return operationValid_PolygonIntersectionAnalyzer_noInvalidIntersection
	}

	isSameSegString := ss0 == ss1

	// Check for an intersection in the interior of both segments.
	// Collinear intersections by definition contain an interior intersection.
	if pia.li.IsProper() || pia.li.GetIntersectionNum() >= 2 {
		return OperationValid_TopologyValidationError_SELF_INTERSECTION
	}

	// Now know there is exactly one intersection,
	// at a vertex of at least one segment.
	intPt := pia.li.GetIntersection(0)

	// If segments are adjacent the intersection must be their common endpoint.
	// (since they are not collinear).
	// This is valid.
	isAdjacentSegments := isSameSegString && operationValid_PolygonIntersectionAnalyzer_isAdjacentInRing(ss0, segIndex0, segIndex1)
	// Assert: intersection is an endpoint of both segs
	if isAdjacentSegments {
		return operationValid_PolygonIntersectionAnalyzer_noInvalidIntersection
	}

	// Under OGC semantics, rings cannot self-intersect.
	// So the intersection is invalid.
	//
	// The return of RING_SELF_INTERSECTION is to match the previous IsValid semantics.
	if isSameSegString && !pia.isInvertedRingValid {
		return OperationValid_TopologyValidationError_RING_SELF_INTERSECTION
	}

	// Optimization: don't analyze intPts at the endpoint of a segment.
	// This is because they are also start points, so don't need to be
	// evaluated twice.
	// This simplifies following logic, by removing the segment endpoint case.
	if intPt.Equals2D(p01) || intPt.Equals2D(p11) {
		return operationValid_PolygonIntersectionAnalyzer_noInvalidIntersection
	}

	// Check topology of a vertex intersection.
	// The ring(s) must not cross.
	e00 := p00
	e01 := p01
	if intPt.Equals2D(p00) {
		e00 = operationValid_PolygonIntersectionAnalyzer_prevCoordinateInRing(ss0, segIndex0)
		e01 = p01
	}
	e10 := p10
	e11 := p11
	if intPt.Equals2D(p10) {
		e10 = operationValid_PolygonIntersectionAnalyzer_prevCoordinateInRing(ss1, segIndex1)
		e11 = p11
	}
	hasCrossing := Algorithm_PolygonNodeTopology_IsCrossing(intPt, e00, e01, e10, e11)
	if hasCrossing {
		return OperationValid_TopologyValidationError_SELF_INTERSECTION
	}

	// If allowing inverted rings, record a self-touch to support later checking
	// that it does not disconnect the interior.
	if isSameSegString && pia.isInvertedRingValid {
		pia.addSelfTouch(ss0, intPt, e00, e01, e10, e11)
	}

	// If the rings are in the same polygon
	// then record the touch to support connected interior checking.
	//
	// Also check for an invalid double-touch situation,
	// if the rings are different.
	isDoubleTouch := pia.addDoubleTouch(ss0, ss1, intPt)
	if isDoubleTouch && !isSameSegString {
		pia.hasDoubleTouch = true
		pia.doubleTouchLocation = intPt
		// TODO: for poly-hole or hole-hole touch, check if it has bad topology. If so return invalid code
	}

	return operationValid_PolygonIntersectionAnalyzer_noInvalidIntersection
}

func (pia *OperationValid_PolygonIntersectionAnalyzer) addDoubleTouch(ss0, ss1 Noding_SegmentString, intPt *Geom_Coordinate) bool {
	return OperationValid_PolygonRing_AddTouch(ss0.GetData().(*OperationValid_PolygonRing), ss1.GetData().(*OperationValid_PolygonRing), intPt)
}

func (pia *OperationValid_PolygonIntersectionAnalyzer) addSelfTouch(ss Noding_SegmentString, intPt, e00, e01, e10, e11 *Geom_Coordinate) {
	polyRing := ss.GetData().(*OperationValid_PolygonRing)
	if polyRing == nil {
		panic("SegmentString missing PolygonRing data when checking self-touches")
	}
	polyRing.AddSelfTouch(intPt, e00, e01, e10, e11)
}

// operationValid_PolygonIntersectionAnalyzer_prevCoordinateInRing gets the
// coordinate previous to the given index for a segment string for a ring
// (wrapping if the index is 0).
func operationValid_PolygonIntersectionAnalyzer_prevCoordinateInRing(ringSS Noding_SegmentString, segIndex int) *Geom_Coordinate {
	prevIndex := segIndex - 1
	if prevIndex < 0 {
		prevIndex = ringSS.Size() - 2
	}
	return ringSS.GetCoordinate(prevIndex)
}

// operationValid_PolygonIntersectionAnalyzer_isAdjacentInRing tests if two
// segments in a closed SegmentString are adjacent. This handles determining
// adjacency across the start/end of the ring.
func operationValid_PolygonIntersectionAnalyzer_isAdjacentInRing(ringSS Noding_SegmentString, segIndex0, segIndex1 int) bool {
	delta := segIndex1 - segIndex0
	if delta < 0 {
		delta = -delta
	}
	if delta <= 1 {
		return true
	}
	// A string with N vertices has maximum segment index of N-2.
	// If the delta is at least N-2, the segments must be
	// at the start and end of the string and thus adjacent.
	if delta >= ringSS.Size()-2 {
		return true
	}
	return false
}
