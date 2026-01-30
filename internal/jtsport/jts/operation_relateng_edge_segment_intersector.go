package jts

var _ Noding_SegmentIntersector = (*OperationRelateng_EdgeSegmentIntersector)(nil)

// OperationRelateng_EdgeSegmentIntersector tests segments of
// RelateSegmentStrings and if they intersect adds the intersection(s) to the
// TopologyComputer.
type OperationRelateng_EdgeSegmentIntersector struct {
	li           *Algorithm_RobustLineIntersector
	topoComputer *OperationRelateng_TopologyComputer
}

// IsNoding_SegmentIntersector is a marker method for interface identification.
func (esi *OperationRelateng_EdgeSegmentIntersector) IsNoding_SegmentIntersector() {}

// OperationRelateng_NewEdgeSegmentIntersector creates a new
// EdgeSegmentIntersector.
func OperationRelateng_NewEdgeSegmentIntersector(topoComputer *OperationRelateng_TopologyComputer) *OperationRelateng_EdgeSegmentIntersector {
	return &OperationRelateng_EdgeSegmentIntersector{
		li:           Algorithm_NewRobustLineIntersector(),
		topoComputer: topoComputer,
	}
}

// IsDone implements the IsDone method.
func (esi *OperationRelateng_EdgeSegmentIntersector) IsDone() bool {
	return esi.topoComputer.IsResultKnown()
}

// ProcessIntersections processes intersections between two segment strings.
func (esi *OperationRelateng_EdgeSegmentIntersector) ProcessIntersections(
	ss0 Noding_SegmentString, segIndex0 int,
	ss1 Noding_SegmentString, segIndex1 int,
) {
	// Don't intersect a segment with itself.
	if ss0 == ss1 && segIndex0 == segIndex1 {
		return
	}

	rss0 := ss0.(*OperationRelateng_RelateSegmentString)
	rss1 := ss1.(*OperationRelateng_RelateSegmentString)

	// Order so that A is first.
	if rss0.IsA() {
		esi.addIntersections(rss0, segIndex0, rss1, segIndex1)
	} else {
		esi.addIntersections(rss1, segIndex1, rss0, segIndex0)
	}
}

func (esi *OperationRelateng_EdgeSegmentIntersector) addIntersections(
	ssA *OperationRelateng_RelateSegmentString, segIndexA int,
	ssB *OperationRelateng_RelateSegmentString, segIndexB int,
) {
	a0 := ssA.GetCoordinate(segIndexA)
	a1 := ssA.GetCoordinate(segIndexA + 1)
	b0 := ssB.GetCoordinate(segIndexB)
	b1 := ssB.GetCoordinate(segIndexB + 1)

	esi.li.ComputeIntersection(a0, a1, b0, b1)

	if !esi.li.HasIntersection() {
		return
	}

	for i := 0; i < esi.li.GetIntersectionNum(); i++ {
		intPt := esi.li.GetIntersection(i)
		// Ensure endpoint intersections are added once only, for their canonical
		// segments. Proper intersections lie on a unique segment so do not need
		// to be checked. And it is important that the Containing Segment check
		// not be used, since due to intersection computation roundoff, it is not
		// reliable in that situation.
		if esi.li.IsProper() ||
			(ssA.IsContainingSegment(segIndexA, intPt) &&
				ssB.IsContainingSegment(segIndexB, intPt)) {
			nsa := ssA.CreateNodeSection(segIndexA, intPt)
			nsb := ssB.CreateNodeSection(segIndexB, intPt)
			esi.topoComputer.AddIntersection(nsa, nsb)
		}
	}
}
