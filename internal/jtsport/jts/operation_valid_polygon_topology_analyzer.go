package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationValid_PolygonTopologyAnalyzer_IsRingNested tests whether a ring is nested inside another ring.
//
// Preconditions:
//   - The rings do not cross (i.e. the test is wholly inside or outside the target)
//   - The rings may touch at discrete points only
//   - The target ring does not self-cross, but it may self-touch
//
// If the test ring start point is properly inside or outside, that provides the result.
// Otherwise the start point is on the target ring,
// and the incident start segment (accounting for repeated points) is
// tested for its topology relative to the target ring.
func OperationValid_PolygonTopologyAnalyzer_IsRingNested(test, target *Geom_LinearRing) bool {
	p0 := test.GetCoordinateN(0)
	targetPts := target.GetCoordinates()
	loc := Algorithm_PointLocation_LocateInRing(p0, targetPts)
	if loc == Geom_Location_Exterior {
		return false
	}
	if loc == Geom_Location_Interior {
		return true
	}

	// The start point is on the boundary of the ring.
	// Use the topology at the node to check if the segment
	// is inside or outside the ring.
	p1 := operationValid_PolygonTopologyAnalyzer_findNonEqualVertex(test, p0)
	return operationValid_PolygonTopologyAnalyzer_isIncidentSegmentInRing(p0, p1, targetPts)
}

func operationValid_PolygonTopologyAnalyzer_findNonEqualVertex(ring *Geom_LinearRing, p *Geom_Coordinate) *Geom_Coordinate {
	i := 1
	next := ring.GetCoordinateN(i)
	for next.Equals2D(p) && i < ring.GetNumPoints()-1 {
		i += 1
		next = ring.GetCoordinateN(i)
	}
	return next
}

// operationValid_PolygonTopologyAnalyzer_isIncidentSegmentInRing tests whether a touching segment is interior to a ring.
//
// Preconditions:
//   - The segment does not intersect the ring other than at the endpoints
//   - The segment vertex p0 lies on the ring
//   - The ring does not self-cross, but it may self-touch
//
// This works for both shells and holes, but the caller must know
// the ring role.
func operationValid_PolygonTopologyAnalyzer_isIncidentSegmentInRing(p0, p1 *Geom_Coordinate, ringPts []*Geom_Coordinate) bool {
	index := operationValid_PolygonTopologyAnalyzer_intersectingSegIndex(ringPts, p0)
	if index < 0 {
		panic("Segment vertex does not intersect ring")
	}
	rPrev := operationValid_PolygonTopologyAnalyzer_findRingVertexPrev(ringPts, index, p0)
	rNext := operationValid_PolygonTopologyAnalyzer_findRingVertexNext(ringPts, index, p0)
	// If ring orientation is not normalized, flip the corner orientation
	isInteriorOnRight := !Algorithm_Orientation_IsCCW(ringPts)
	if !isInteriorOnRight {
		temp := rPrev
		rPrev = rNext
		rNext = temp
	}
	return Algorithm_PolygonNodeTopology_IsInteriorSegment(p0, rPrev, rNext, p1)
}

// operationValid_PolygonTopologyAnalyzer_findRingVertexPrev finds the ring vertex previous to a node point on a ring
// (which is contained in the index'th segment,
// as either the start vertex or an interior point).
// Repeated points are skipped over.
func operationValid_PolygonTopologyAnalyzer_findRingVertexPrev(ringPts []*Geom_Coordinate, index int, node *Geom_Coordinate) *Geom_Coordinate {
	iPrev := index
	prev := ringPts[iPrev]
	for node.Equals2D(prev) {
		iPrev = operationValid_PolygonTopologyAnalyzer_ringIndexPrev(ringPts, iPrev)
		prev = ringPts[iPrev]
	}
	return prev
}

// operationValid_PolygonTopologyAnalyzer_findRingVertexNext finds the ring vertex next from a node point on a ring
// (which is contained in the index'th segment,
// as either the start vertex or an interior point).
// Repeated points are skipped over.
func operationValid_PolygonTopologyAnalyzer_findRingVertexNext(ringPts []*Geom_Coordinate, index int, node *Geom_Coordinate) *Geom_Coordinate {
	//-- safe, since index is always the start of a ring segment
	iNext := index + 1
	next := ringPts[iNext]
	for node.Equals2D(next) {
		iNext = operationValid_PolygonTopologyAnalyzer_ringIndexNext(ringPts, iNext)
		next = ringPts[iNext]
	}
	return next
}

func operationValid_PolygonTopologyAnalyzer_ringIndexPrev(ringPts []*Geom_Coordinate, index int) int {
	if index == 0 {
		return len(ringPts) - 2
	}
	return index - 1
}

func operationValid_PolygonTopologyAnalyzer_ringIndexNext(ringPts []*Geom_Coordinate, index int) int {
	if index >= len(ringPts)-2 {
		return 0
	}
	return index + 1
}

// operationValid_PolygonTopologyAnalyzer_intersectingSegIndex computes the index of the segment which intersects a given point.
func operationValid_PolygonTopologyAnalyzer_intersectingSegIndex(ringPts []*Geom_Coordinate, pt *Geom_Coordinate) int {
	for i := 0; i < len(ringPts)-1; i++ {
		if Algorithm_PointLocation_IsOnSegment(pt, ringPts[i], ringPts[i+1]) {
			//-- check if pt is the start point of the next segment
			if pt.Equals2D(ringPts[i+1]) {
				return i + 1
			}
			return i
		}
	}
	return -1
}

// OperationValid_PolygonTopologyAnalyzer_FindSelfIntersection finds a self-intersection (if any) in a LinearRing.
func OperationValid_PolygonTopologyAnalyzer_FindSelfIntersection(ring *Geom_LinearRing) *Geom_Coordinate {
	ata := OperationValid_NewPolygonTopologyAnalyzerFromLinearRing(ring, false)
	if ata.HasInvalidIntersection() {
		return ata.GetInvalidLocation()
	}
	return nil
}

// OperationValid_PolygonTopologyAnalyzer analyzes the topology of polygonal geometry
// to determine whether it is valid.
//
// Analyzing polygons with inverted rings (shells or exverted holes)
// is performed if specified.
// Inverted rings may cause a disconnected interior due to a self-touch;
// this is reported by IsInteriorDisconnectedBySelfTouch().
type OperationValid_PolygonTopologyAnalyzer struct {
	isInvertedRingValid bool

	intFinder       *OperationValid_PolygonIntersectionAnalyzer
	polyRings       []*OperationValid_PolygonRing
	disconnectionPt *Geom_Coordinate
}

// OperationValid_NewPolygonTopologyAnalyzer creates a new analyzer for a Polygon or MultiPolygon.
func OperationValid_NewPolygonTopologyAnalyzer(geom *Geom_Geometry, isInvertedRingValid bool) *OperationValid_PolygonTopologyAnalyzer {
	pta := &OperationValid_PolygonTopologyAnalyzer{
		isInvertedRingValid: isInvertedRingValid,
	}
	pta.analyze(geom)
	return pta
}

// TRANSLITERATION NOTE: This constructor is added for Go convenience. Java's
// constructor accepts Geometry which includes LinearRing, but Go's type system
// requires a separate constructor to accept *Geom_LinearRing directly without
// requiring the caller to access the embedded Geom_Geometry.
func OperationValid_NewPolygonTopologyAnalyzerFromLinearRing(ring *Geom_LinearRing, isInvertedRingValid bool) *OperationValid_PolygonTopologyAnalyzer {
	pta := &OperationValid_PolygonTopologyAnalyzer{
		isInvertedRingValid: isInvertedRingValid,
	}
	pta.analyze(ring.Geom_Geometry)
	return pta
}

func (pta *OperationValid_PolygonTopologyAnalyzer) HasInvalidIntersection() bool {
	return pta.intFinder.IsInvalid()
}

func (pta *OperationValid_PolygonTopologyAnalyzer) GetInvalidCode() int {
	return pta.intFinder.GetInvalidCode()
}

func (pta *OperationValid_PolygonTopologyAnalyzer) GetInvalidLocation() *Geom_Coordinate {
	return pta.intFinder.GetInvalidLocation()
}

// IsInteriorDisconnected tests whether the interior of the polygonal geometry is
// disconnected.
// If true, the disconnection location is available from
// GetDisconnectionLocation().
func (pta *OperationValid_PolygonTopologyAnalyzer) IsInteriorDisconnected() bool {
	// May already be set by a double-touching hole
	if pta.disconnectionPt != nil {
		return true
	}
	if pta.isInvertedRingValid {
		pta.CheckInteriorDisconnectedBySelfTouch()
		if pta.disconnectionPt != nil {
			return true
		}
	}
	pta.CheckInteriorDisconnectedByHoleCycle()
	if pta.disconnectionPt != nil {
		return true
	}
	return false
}

// GetDisconnectionLocation gets a location where the polygonal interior is disconnected.
// IsInteriorDisconnected() must be called first.
func (pta *OperationValid_PolygonTopologyAnalyzer) GetDisconnectionLocation() *Geom_Coordinate {
	return pta.disconnectionPt
}

// CheckInteriorDisconnectedByHoleCycle tests whether any polygon with holes has a disconnected interior
// by virtue of the holes (and possibly shell) forming a hole cycle.
//
// This is a global check, which relies on determining
// the touching graph of all holes in a polygon.
//
// If inverted rings disconnect the interior
// via a self-touch, this is checked by the PolygonIntersectionAnalyzer.
// If inverted rings are part of a hole cycle
// this is detected here as well.
func (pta *OperationValid_PolygonTopologyAnalyzer) CheckInteriorDisconnectedByHoleCycle() {
	// PolyRings will be null for empty, no hole or LinearRing inputs
	if pta.polyRings != nil {
		pta.disconnectionPt = OperationValid_PolygonRing_FindHoleCycleLocation(pta.polyRings)
	}
}

// CheckInteriorDisconnectedBySelfTouch tests if an area interior is disconnected by a self-touching ring.
// This must be evaluated after other self-intersections have been analyzed
// and determined to not exist, since the logic relies on
// the rings not self-crossing (winding).
//
// If self-touching rings are not allowed,
// then the self-touch will previously trigger a self-intersection error.
func (pta *OperationValid_PolygonTopologyAnalyzer) CheckInteriorDisconnectedBySelfTouch() {
	if pta.polyRings != nil {
		pta.disconnectionPt = OperationValid_PolygonRing_FindInteriorSelfNode(pta.polyRings)
	}
}

func (pta *OperationValid_PolygonTopologyAnalyzer) analyze(geom *Geom_Geometry) {
	if geom.IsEmpty() {
		return
	}
	segStrings := operationValid_PolygonTopologyAnalyzer_createSegmentStrings(geom, pta.isInvertedRingValid)
	pta.polyRings = operationValid_PolygonTopologyAnalyzer_getPolygonRings(segStrings)
	pta.intFinder = pta.analyzeIntersections(segStrings)

	if pta.intFinder.HasDoubleTouch() {
		pta.disconnectionPt = pta.intFinder.GetDoubleTouchLocation()
		return
	}
}

func (pta *OperationValid_PolygonTopologyAnalyzer) analyzeIntersections(segStrings []Noding_SegmentString) *OperationValid_PolygonIntersectionAnalyzer {
	segInt := OperationValid_NewPolygonIntersectionAnalyzer(pta.isInvertedRingValid)
	noder := Noding_NewMCIndexNoder()
	noder.SetSegmentIntersector(segInt)
	noder.ComputeNodes(segStrings)
	return segInt
}

func operationValid_PolygonTopologyAnalyzer_createSegmentStrings(geom *Geom_Geometry, isInvertedRingValid bool) []Noding_SegmentString {
	segStrings := make([]Noding_SegmentString, 0)
	if java.InstanceOf[*Geom_LinearRing](geom) {
		ring := java.Cast[*Geom_LinearRing](geom)
		segStrings = append(segStrings, operationValid_PolygonTopologyAnalyzer_createSegString(ring, nil))
		return segStrings
	}
	for i := 0; i < geom.GetNumGeometries(); i++ {
		poly := java.Cast[*Geom_Polygon](geom.GetGeometryN(i))
		if poly.IsEmpty() {
			continue
		}
		hasHoles := poly.GetNumInteriorRing() > 0

		//--- polygons with no holes do not need connected interior analysis
		var shellRing *OperationValid_PolygonRing
		if hasHoles || isInvertedRingValid {
			shellRing = OperationValid_NewPolygonRing(poly.GetExteriorRing())
		}
		segStrings = append(segStrings, operationValid_PolygonTopologyAnalyzer_createSegString(poly.GetExteriorRing(), shellRing))

		for j := 0; j < poly.GetNumInteriorRing(); j++ {
			hole := poly.GetInteriorRingN(j)
			if hole.IsEmpty() {
				continue
			}
			holeRing := OperationValid_NewPolygonRingWithIndexAndShell(hole, j, shellRing)
			segStrings = append(segStrings, operationValid_PolygonTopologyAnalyzer_createSegString(hole, holeRing))
		}
	}
	return segStrings
}

func operationValid_PolygonTopologyAnalyzer_getPolygonRings(segStrings []Noding_SegmentString) []*OperationValid_PolygonRing {
	var polyRings []*OperationValid_PolygonRing
	for _, ss := range segStrings {
		data := ss.GetData()
		polyRing, ok := data.(*OperationValid_PolygonRing)
		if ok && polyRing != nil {
			if polyRings == nil {
				polyRings = make([]*OperationValid_PolygonRing, 0)
			}
			polyRings = append(polyRings, polyRing)
		}
	}
	return polyRings
}

func operationValid_PolygonTopologyAnalyzer_createSegString(ring *Geom_LinearRing, polyRing *OperationValid_PolygonRing) Noding_SegmentString {
	pts := ring.GetCoordinates()

	//--- repeated points must be removed for accurate intersection detection
	if Geom_CoordinateArrays_HasRepeatedPoints(pts) {
		pts = Geom_CoordinateArrays_RemoveRepeatedPoints(pts)
	}

	ss := Noding_NewBasicSegmentString(pts, polyRing)
	return ss
}
