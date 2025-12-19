package jts

// GeomgraphIndex_EdgeSetIntersector computes all the intersections between the
// edges in the set. It adds the computed intersections to each edge they are
// found on. It may be used in two scenarios:
//   - determining the internal intersections between a single set of edges
//   - determining the mutual intersections between two different sets of edges
//
// It uses a SegmentIntersector to compute the intersections between segments
// and to record statistics about what kinds of intersections were found.
type GeomgraphIndex_EdgeSetIntersector interface {
	// ComputeIntersectionsSingleList computes all self-intersections between edges in a set of edges,
	// allowing client to choose whether self-intersections are computed.
	//
	// edges is a list of edges to test for intersections.
	// si is the SegmentIntersector to use.
	// testAllSegments is true if self-intersections are to be tested as well.
	ComputeIntersectionsSingleList(edges []*Geomgraph_Edge, si *GeomgraphIndex_SegmentIntersector, testAllSegments bool)

	// ComputeIntersectionsTwoLists computes all mutual intersections between two sets of edges.
	ComputeIntersectionsTwoLists(edges0, edges1 []*Geomgraph_Edge, si *GeomgraphIndex_SegmentIntersector)

	// IsGeomgraphIndex_EdgeSetIntersector is a marker method for the interface.
	IsGeomgraphIndex_EdgeSetIntersector()
}
