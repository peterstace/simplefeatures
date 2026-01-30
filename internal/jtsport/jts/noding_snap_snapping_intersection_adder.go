package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

var _ Noding_SegmentIntersector = (*NodingSnap_SnappingIntersectionAdder)(nil)

// NodingSnap_SnappingIntersectionAdder finds intersections between line
// segments which are being snapped, and adds them as nodes.
type NodingSnap_SnappingIntersectionAdder struct {
	li             *Algorithm_LineIntersector
	snapTolerance  float64
	snapPointIndex *NodingSnap_SnappingPointIndex
}

// IsNoding_SegmentIntersector is a marker method for interface identification.
func (sia *NodingSnap_SnappingIntersectionAdder) IsNoding_SegmentIntersector() {}

// NodingSnap_NewSnappingIntersectionAdder creates an intersector which finds
// intersections, snaps them, and adds them as nodes.
func NodingSnap_NewSnappingIntersectionAdder(snapTolerance float64, snapPointIndex *NodingSnap_SnappingPointIndex) *NodingSnap_SnappingIntersectionAdder {
	rli := Algorithm_NewRobustLineIntersector()
	return &NodingSnap_SnappingIntersectionAdder{
		li:             rli.Algorithm_LineIntersector,
		snapTolerance:  snapTolerance,
		snapPointIndex: snapPointIndex,
	}
}

// ProcessIntersections is called by clients of the SegmentIntersector class to
// process intersections for two segments of the SegmentStrings being
// intersected.
func (sia *NodingSnap_SnappingIntersectionAdder) ProcessIntersections(
	seg0 Noding_SegmentString, segIndex0 int,
	seg1 Noding_SegmentString, segIndex1 int,
) {
	// Don't bother intersecting a segment with itself.
	if seg0 == seg1 && segIndex0 == segIndex1 {
		return
	}

	p00 := seg0.GetCoordinate(segIndex0)
	p01 := seg0.GetCoordinate(segIndex0 + 1)
	p10 := seg1.GetCoordinate(segIndex1)
	p11 := seg1.GetCoordinate(segIndex1 + 1)

	// Don't node intersections which are just due to the shared vertex of
	// adjacent segments.
	if !sia.isAdjacent(seg0, segIndex0, seg1, segIndex1) {
		sia.li.ComputeIntersection(p00, p01, p10, p11)

		// Process single point intersections only. Two-point (collinear) ones
		// are handled by the near-vertex code.
		if sia.li.HasIntersection() && sia.li.GetIntersectionNum() == 1 {
			intPt := sia.li.GetIntersection(0)
			snapPt := sia.snapPointIndex.Snap(intPt)

			nss0 := seg0.(*Noding_NodedSegmentString)
			nss1 := seg1.(*Noding_NodedSegmentString)
			nss0.AddIntersection(snapPt, segIndex0)
			nss1.AddIntersection(snapPt, segIndex1)
		}
	}

	// The segments must also be snapped to the other segment endpoints.
	sia.processNearVertex(seg0, segIndex0, p00, seg1, segIndex1, p10, p11)
	sia.processNearVertex(seg0, segIndex0, p01, seg1, segIndex1, p10, p11)
	sia.processNearVertex(seg1, segIndex1, p10, seg0, segIndex0, p00, p01)
	sia.processNearVertex(seg1, segIndex1, p11, seg0, segIndex0, p00, p01)
}

// processNearVertex adds an intersection if an endpoint of one segment is
// near the interior of the other segment. EXCEPT if the endpoint is also close
// to a segment endpoint (since this can introduce "zigs" in the linework).
func (sia *NodingSnap_SnappingIntersectionAdder) processNearVertex(
	srcSS Noding_SegmentString, srcIndex int, p *Geom_Coordinate,
	ss Noding_SegmentString, segIndex int, p0, p1 *Geom_Coordinate,
) {
	// Don't add intersection if candidate vertex is near endpoints of segment.
	// This avoids creating "zig-zag" linework (since the vertex could actually
	// be outside the segment envelope). Also, this should have already been
	// snapped.
	if p.Distance(p0) < sia.snapTolerance {
		return
	}
	if p.Distance(p1) < sia.snapTolerance {
		return
	}

	distSeg := Algorithm_Distance_PointToSegment(p, p0, p1)
	if distSeg < sia.snapTolerance {
		// Add node to target segment.
		nss := ss.(*Noding_NodedSegmentString)
		nss.AddIntersection(p, segIndex)
		// Add node at vertex to source SS.
		srcNss := srcSS.(*Noding_NodedSegmentString)
		srcNss.AddIntersection(p, srcIndex)
	}
}

// isAdjacent tests if segments are adjacent on the same SegmentString. Closed
// segStrings require a check for the point shared by the beginning and end
// segments.
func (sia *NodingSnap_SnappingIntersectionAdder) isAdjacent(
	ss0 Noding_SegmentString, segIndex0 int,
	ss1 Noding_SegmentString, segIndex1 int,
) bool {
	if ss0 != ss1 {
		return false
	}

	isAdjacent := java.AbsInt(segIndex0-segIndex1) == 1
	if isAdjacent {
		return true
	}
	if ss0.IsClosed() {
		maxSegIndex := ss0.Size() - 1
		if (segIndex0 == 0 && segIndex1 == maxSegIndex) ||
			(segIndex1 == 0 && segIndex0 == maxSegIndex) {
			return true
		}
	}
	return false
}

// IsDone always returns false since all intersections should be processed.
func (sia *NodingSnap_SnappingIntersectionAdder) IsDone() bool {
	return false
}
