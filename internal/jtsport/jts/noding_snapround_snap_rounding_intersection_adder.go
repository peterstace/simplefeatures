package jts

var _ Noding_SegmentIntersector = (*NodingSnapround_SnapRoundingIntersectionAdder)(nil)

// NodingSnapround_SnapRoundingIntersectionAdder finds intersections between
// line segments which will be snap-rounded, and adds them as nodes to the
// segments.
//
// Intersections are detected and computed using full precision. Snapping takes
// place in a subsequent phase.
//
// The intersection points are recorded, so that HotPixels can be created for
// them.
//
// To avoid robustness issues with vertices which lie very close to line
// segments a heuristic is used: nodes are created if a vertex lies within a
// tolerance distance of the interior of a segment. The tolerance distance is
// chosen to be significantly below the snap-rounding grid size. This has
// empirically proven to eliminate noding failures.
type NodingSnapround_SnapRoundingIntersectionAdder struct {
	li            *Algorithm_LineIntersector
	intersections []*Geom_Coordinate
	nearnessTol   float64
}

// IsNoding_SegmentIntersector is a marker method for interface identification.
func (sria *NodingSnapround_SnapRoundingIntersectionAdder) IsNoding_SegmentIntersector() {}

// NodingSnapround_NewSnapRoundingIntersectionAdder creates an intersector
// which finds all snapped interior intersections, and adds them as nodes.
func NodingSnapround_NewSnapRoundingIntersectionAdder(nearnessTol float64) *NodingSnapround_SnapRoundingIntersectionAdder {
	// Intersections are detected and computed using full precision. They are
	// snapped in a subsequent phase.
	rli := Algorithm_NewRobustLineIntersector()
	return &NodingSnapround_SnapRoundingIntersectionAdder{
		li:            rli.Algorithm_LineIntersector,
		intersections: make([]*Geom_Coordinate, 0),
		nearnessTol:   nearnessTol,
	}
}

// GetIntersections gets the created intersection nodes, so they can be
// processed as hot pixels.
func (sria *NodingSnapround_SnapRoundingIntersectionAdder) GetIntersections() []*Geom_Coordinate {
	return sria.intersections
}

// ProcessIntersections is called by clients of the SegmentIntersector class to
// process intersections for two segments of the SegmentStrings being
// intersected.
func (sria *NodingSnapround_SnapRoundingIntersectionAdder) ProcessIntersections(
	e0 Noding_SegmentString, segIndex0 int,
	e1 Noding_SegmentString, segIndex1 int,
) {
	// Don't bother intersecting a segment with itself.
	if e0 == e1 && segIndex0 == segIndex1 {
		return
	}

	p00 := e0.GetCoordinate(segIndex0)
	p01 := e0.GetCoordinate(segIndex0 + 1)
	p10 := e1.GetCoordinate(segIndex1)
	p11 := e1.GetCoordinate(segIndex1 + 1)

	sria.li.ComputeIntersection(p00, p01, p10, p11)

	if sria.li.HasIntersection() {
		if sria.li.IsInteriorIntersection() {
			for intIndex := 0; intIndex < sria.li.GetIntersectionNum(); intIndex++ {
				sria.intersections = append(sria.intersections, sria.li.GetIntersection(intIndex))
			}
			nss0 := e0.(*Noding_NodedSegmentString)
			nss1 := e1.(*Noding_NodedSegmentString)
			nss0.AddIntersections(sria.li, segIndex0, 0)
			nss1.AddIntersections(sria.li, segIndex1, 1)
			return
		}
	}

	// Segments did not actually intersect, within the limits of orientation
	// index robustness.
	//
	// To avoid certain robustness issues in snap-rounding, also treat very
	// near vertex-segment situations as intersections.
	sria.processNearVertex(p00, e1, segIndex1, p10, p11)
	sria.processNearVertex(p01, e1, segIndex1, p10, p11)
	sria.processNearVertex(p10, e0, segIndex0, p00, p01)
	sria.processNearVertex(p11, e0, segIndex0, p00, p01)
}

// processNearVertex adds an intersection if an endpoint of one segment is
// near the interior of the other segment. EXCEPT if the endpoint is also close
// to a segment endpoint (since this can introduce "zigs" in the linework).
func (sria *NodingSnapround_SnapRoundingIntersectionAdder) processNearVertex(
	p *Geom_Coordinate,
	edge Noding_SegmentString, segIndex int,
	p0, p1 *Geom_Coordinate,
) {
	// Don't add intersection if candidate vertex is near endpoints of segment.
	// This avoids creating "zig-zag" linework (since the vertex could actually
	// be outside the segment envelope).
	if p.Distance(p0) < sria.nearnessTol {
		return
	}
	if p.Distance(p1) < sria.nearnessTol {
		return
	}

	distSeg := Algorithm_Distance_PointToSegment(p, p0, p1)
	if distSeg < sria.nearnessTol {
		sria.intersections = append(sria.intersections, p)
		nss := edge.(*Noding_NodedSegmentString)
		nss.AddIntersection(p, segIndex)
	}
}

// IsDone always returns false since all intersections should be processed.
func (sria *NodingSnapround_SnapRoundingIntersectionAdder) IsDone() bool {
	return false
}
