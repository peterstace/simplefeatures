package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// GeomgraphIndex_MonotoneChainEdge provides a way of partitioning the segments
// of an edge to allow for fast searching of intersections.
// Monotone chains have the following properties:
//  1. the segments within a monotone chain will never intersect each other
//  2. the envelope of any contiguous subset of the segments in a monotone chain
//     is simply the envelope of the endpoints of the subset.
//
// Property 1 means that there is no need to test pairs of segments from within
// the same monotone chain for intersection.
// Property 2 allows binary search to be used to find the intersection points of
// two monotone chains.
// For many types of real-world data, these properties eliminate a large number
// of segment comparisons, producing substantial speed gains.
type GeomgraphIndex_MonotoneChainEdge struct {
	child java.Polymorphic

	e *Geomgraph_Edge
	// Cache a reference to the coord array, for efficiency.
	pts []*Geom_Coordinate
	// The lists of start/end indexes of the monotone chains.
	// Includes the end point of the edge as a sentinel.
	startIndex []int
}

// GetChild returns the immediate child in the type hierarchy chain.
func (mce *GeomgraphIndex_MonotoneChainEdge) GetChild() java.Polymorphic {
	return mce.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (mce *GeomgraphIndex_MonotoneChainEdge) GetParent() java.Polymorphic {
	return nil
}

// GeomgraphIndex_NewMonotoneChainEdge creates a new MonotoneChainEdge.
func GeomgraphIndex_NewMonotoneChainEdge(e *Geomgraph_Edge) *GeomgraphIndex_MonotoneChainEdge {
	pts := e.GetCoordinates()
	mcb := GeomgraphIndex_NewMonotoneChainIndexer()
	return &GeomgraphIndex_MonotoneChainEdge{
		e:          e,
		pts:        pts,
		startIndex: mcb.GetChainStartIndices(pts),
	}
}

// GetCoordinates returns the coordinates of this edge.
func (mce *GeomgraphIndex_MonotoneChainEdge) GetCoordinates() []*Geom_Coordinate {
	return mce.pts
}

// GetStartIndexes returns the start indices of the monotone chains.
func (mce *GeomgraphIndex_MonotoneChainEdge) GetStartIndexes() []int {
	return mce.startIndex
}

// GetMinX returns the minimum x coordinate of the given chain.
func (mce *GeomgraphIndex_MonotoneChainEdge) GetMinX(chainIndex int) float64 {
	x1 := mce.pts[mce.startIndex[chainIndex]].X
	x2 := mce.pts[mce.startIndex[chainIndex+1]].X
	if x1 < x2 {
		return x1
	}
	return x2
}

// GetMaxX returns the maximum x coordinate of the given chain.
func (mce *GeomgraphIndex_MonotoneChainEdge) GetMaxX(chainIndex int) float64 {
	x1 := mce.pts[mce.startIndex[chainIndex]].X
	x2 := mce.pts[mce.startIndex[chainIndex+1]].X
	if x1 > x2 {
		return x1
	}
	return x2
}

// ComputeIntersects computes all intersections between this edge and another.
func (mce *GeomgraphIndex_MonotoneChainEdge) ComputeIntersects(other *GeomgraphIndex_MonotoneChainEdge, si *GeomgraphIndex_SegmentIntersector) {
	for i := 0; i < len(mce.startIndex)-1; i++ {
		for j := 0; j < len(other.startIndex)-1; j++ {
			mce.ComputeIntersectsForChain(i, other, j, si)
		}
	}
}

// ComputeIntersectsForChain computes intersections between two chains.
func (mce *GeomgraphIndex_MonotoneChainEdge) ComputeIntersectsForChain(chainIndex0 int, other *GeomgraphIndex_MonotoneChainEdge, chainIndex1 int, si *GeomgraphIndex_SegmentIntersector) {
	mce.computeIntersectsForChain(
		mce.startIndex[chainIndex0], mce.startIndex[chainIndex0+1],
		other,
		other.startIndex[chainIndex1], other.startIndex[chainIndex1+1],
		si,
	)
}

func (mce *GeomgraphIndex_MonotoneChainEdge) computeIntersectsForChain(start0, end0 int, other *GeomgraphIndex_MonotoneChainEdge, start1, end1 int, ei *GeomgraphIndex_SegmentIntersector) {
	// Terminating condition for the recursion.
	if end0-start0 == 1 && end1-start1 == 1 {
		ei.AddIntersections(mce.e, start0, other.e, start1)
		return
	}
	// Nothing to do if the envelopes of these chains don't overlap.
	if !mce.overlaps(start0, end0, other, start1, end1) {
		return
	}

	// The chains overlap, so split each in half and iterate (binary search).
	mid0 := (start0 + end0) / 2
	mid1 := (start1 + end1) / 2

	// Assert: mid != start or end (since we checked above for end - start <= 1).
	// Check terminating conditions before recursing.
	if start0 < mid0 {
		if start1 < mid1 {
			mce.computeIntersectsForChain(start0, mid0, other, start1, mid1, ei)
		}
		if mid1 < end1 {
			mce.computeIntersectsForChain(start0, mid0, other, mid1, end1, ei)
		}
	}
	if mid0 < end0 {
		if start1 < mid1 {
			mce.computeIntersectsForChain(mid0, end0, other, start1, mid1, ei)
		}
		if mid1 < end1 {
			mce.computeIntersectsForChain(mid0, end0, other, mid1, end1, ei)
		}
	}
}

// overlaps tests whether the envelopes of two chain sections overlap (intersect).
func (mce *GeomgraphIndex_MonotoneChainEdge) overlaps(start0, end0 int, other *GeomgraphIndex_MonotoneChainEdge, start1, end1 int) bool {
	return Geom_Envelope_IntersectsEnvelopeEnvelope(mce.pts[start0], mce.pts[end0], other.pts[start1], other.pts[end1])
}
