package jts

// Compile-time interface check.
var _ GeomgraphIndex_EdgeSetIntersector = (*GeomgraphIndex_SimpleEdgeSetIntersector)(nil)

// GeomgraphIndex_SimpleEdgeSetIntersector finds all intersections in one or two
// sets of edges, using the straightforward method of comparing all segments.
// This algorithm is too slow for production use, but is useful for testing
// purposes.
type GeomgraphIndex_SimpleEdgeSetIntersector struct {
	// Statistics information.
	nOverlaps int
}

// IsGeomgraphIndex_EdgeSetIntersector is a marker method for the interface.
func (s *GeomgraphIndex_SimpleEdgeSetIntersector) IsGeomgraphIndex_EdgeSetIntersector() {}

// GeomgraphIndex_NewSimpleEdgeSetIntersector creates a new SimpleEdgeSetIntersector.
func GeomgraphIndex_NewSimpleEdgeSetIntersector() *GeomgraphIndex_SimpleEdgeSetIntersector {
	return &GeomgraphIndex_SimpleEdgeSetIntersector{}
}

// ComputeIntersectionsSingleList computes all self-intersections between edges in a set of edges.
func (s *GeomgraphIndex_SimpleEdgeSetIntersector) ComputeIntersectionsSingleList(edges []*Geomgraph_Edge, si *GeomgraphIndex_SegmentIntersector, testAllSegments bool) {
	s.nOverlaps = 0

	for _, edge0 := range edges {
		for _, edge1 := range edges {
			if testAllSegments || edge0 != edge1 {
				s.computeIntersects(edge0, edge1, si)
			}
		}
	}
}

// ComputeIntersectionsTwoLists computes all mutual intersections between two sets of edges.
func (s *GeomgraphIndex_SimpleEdgeSetIntersector) ComputeIntersectionsTwoLists(edges0, edges1 []*Geomgraph_Edge, si *GeomgraphIndex_SegmentIntersector) {
	s.nOverlaps = 0

	for _, edge0 := range edges0 {
		for _, edge1 := range edges1 {
			s.computeIntersects(edge0, edge1, si)
		}
	}
}

// computeIntersects performs a brute-force comparison of every segment in each
// Edge. This has n^2 performance, and is about 100 times slower than using
// monotone chains.
func (s *GeomgraphIndex_SimpleEdgeSetIntersector) computeIntersects(e0, e1 *Geomgraph_Edge, si *GeomgraphIndex_SegmentIntersector) {
	pts0 := e0.GetCoordinates()
	pts1 := e1.GetCoordinates()
	for i0 := 0; i0 < len(pts0)-1; i0++ {
		for i1 := 0; i1 < len(pts1)-1; i1++ {
			si.AddIntersections(e0, i0, e1, i1)
		}
	}
}
