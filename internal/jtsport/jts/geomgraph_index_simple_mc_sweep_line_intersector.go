package jts

import "sort"

// Compile-time interface check.
var _ GeomgraphIndex_EdgeSetIntersector = (*GeomgraphIndex_SimpleMCSweepLineIntersector)(nil)

// GeomgraphIndex_SimpleMCSweepLineIntersector finds all intersections in one or
// two sets of edges, using an x-axis sweepline algorithm in conjunction with
// Monotone Chains. While still O(n^2) in the worst case, this algorithm
// drastically improves the average-case time. The use of MonotoneChains as the
// items in the index seems to offer an improvement in performance over a
// sweep-line alone.
type GeomgraphIndex_SimpleMCSweepLineIntersector struct {
	events []*GeomgraphIndex_SweepLineEvent
	// Statistics information.
	nOverlaps int
}

// IsGeomgraphIndex_EdgeSetIntersector is a marker method for the interface.
func (s *GeomgraphIndex_SimpleMCSweepLineIntersector) IsGeomgraphIndex_EdgeSetIntersector() {}

// GeomgraphIndex_NewSimpleMCSweepLineIntersector creates a new SimpleMCSweepLineIntersector.
// A SimpleMCSweepLineIntersector creates monotone chains from the edges and
// compares them using a simple sweep-line along the x-axis.
func GeomgraphIndex_NewSimpleMCSweepLineIntersector() *GeomgraphIndex_SimpleMCSweepLineIntersector {
	return &GeomgraphIndex_SimpleMCSweepLineIntersector{}
}

// ComputeIntersectionsSingleList computes all self-intersections between edges in a set of edges.
func (s *GeomgraphIndex_SimpleMCSweepLineIntersector) ComputeIntersectionsSingleList(edges []*Geomgraph_Edge, si *GeomgraphIndex_SegmentIntersector, testAllSegments bool) {
	if testAllSegments {
		s.addEdgesWithEdgeSet(edges, nil)
	} else {
		s.addEdges(edges)
	}
	s.computeIntersections(si)
}

// ComputeIntersectionsTwoLists computes all mutual intersections between two sets of edges.
func (s *GeomgraphIndex_SimpleMCSweepLineIntersector) ComputeIntersectionsTwoLists(edges0, edges1 []*Geomgraph_Edge, si *GeomgraphIndex_SegmentIntersector) {
	s.addEdgesWithEdgeSet(edges0, edges0)
	s.addEdgesWithEdgeSet(edges1, edges1)
	s.computeIntersections(si)
}

func (s *GeomgraphIndex_SimpleMCSweepLineIntersector) addEdges(edges []*Geomgraph_Edge) {
	for _, edge := range edges {
		// Edge is its own group.
		s.addEdge(edge, edge)
	}
}

func (s *GeomgraphIndex_SimpleMCSweepLineIntersector) addEdgesWithEdgeSet(edges []*Geomgraph_Edge, edgeSet any) {
	for _, edge := range edges {
		s.addEdge(edge, edgeSet)
	}
}

func (s *GeomgraphIndex_SimpleMCSweepLineIntersector) addEdge(edge *Geomgraph_Edge, edgeSet any) {
	mce := edge.GetMonotoneChainEdge()
	startIndex := mce.GetStartIndexes()
	for i := 0; i < len(startIndex)-1; i++ {
		mc := GeomgraphIndex_NewMonotoneChain(mce, i)
		insertEvent := GeomgraphIndex_NewSweepLineEventInsert(edgeSet, mce.GetMinX(i), mc)
		s.events = append(s.events, insertEvent)
		s.events = append(s.events, GeomgraphIndex_NewSweepLineEventDelete(mce.GetMaxX(i), insertEvent))
	}
}

// prepareEvents sorts events and sets DELETE event indexes.
// Because Delete Events have a link to their corresponding Insert event, it is
// possible to compute exactly the range of events which must be compared to a
// given Insert event object.
func (s *GeomgraphIndex_SimpleMCSweepLineIntersector) prepareEvents() {
	sort.Slice(s.events, func(i, j int) bool {
		return s.events[i].CompareTo(s.events[j]) < 0
	})
	// Set DELETE event indexes.
	for i, ev := range s.events {
		if ev.IsDelete() {
			ev.GetInsertEvent().SetDeleteEventIndex(i)
		}
	}
}

func (s *GeomgraphIndex_SimpleMCSweepLineIntersector) computeIntersections(si *GeomgraphIndex_SegmentIntersector) {
	s.nOverlaps = 0
	s.prepareEvents()

	for i, ev := range s.events {
		if ev.IsInsert() {
			s.processOverlaps(i, ev.GetDeleteEventIndex(), ev, si)
		}
		if si.IsDone() {
			break
		}
	}
}

func (s *GeomgraphIndex_SimpleMCSweepLineIntersector) processOverlaps(start, end int, ev0 *GeomgraphIndex_SweepLineEvent, si *GeomgraphIndex_SegmentIntersector) {
	mc0 := ev0.GetObject().(*GeomgraphIndex_MonotoneChain)
	// Since we might need to test for self-intersections, include current
	// INSERT event object in list of event objects to test.
	// Last index can be skipped, because it must be a Delete event.
	for i := start; i < end; i++ {
		ev1 := s.events[i]
		if ev1.IsInsert() {
			mc1 := ev1.GetObject().(*GeomgraphIndex_MonotoneChain)
			// Don't compare edges in same group, if labels are present.
			if !ev0.IsSameLabel(ev1) {
				mc0.ComputeIntersections(mc1, si)
				s.nOverlaps++
			}
		}
	}
}
