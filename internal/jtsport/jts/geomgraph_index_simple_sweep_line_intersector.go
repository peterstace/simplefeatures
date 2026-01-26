package jts

import "sort"

// Compile-time interface check.
var _ GeomgraphIndex_EdgeSetIntersector = (*GeomgraphIndex_SimpleSweepLineIntersector)(nil)

// GeomgraphIndex_SimpleSweepLineIntersector finds all intersections in one or
// two sets of edges, using a simple x-axis sweepline algorithm. While still
// O(n^2) in the worst case, this algorithm drastically improves the
// average-case time.
type GeomgraphIndex_SimpleSweepLineIntersector struct {
	events []*GeomgraphIndex_SweepLineEvent
	// Statistics information.
	nOverlaps int
}

// IsGeomgraphIndex_EdgeSetIntersector is a marker method for the interface.
func (s *GeomgraphIndex_SimpleSweepLineIntersector) IsGeomgraphIndex_EdgeSetIntersector() {}

// GeomgraphIndex_NewSimpleSweepLineIntersector creates a new SimpleSweepLineIntersector.
func GeomgraphIndex_NewSimpleSweepLineIntersector() *GeomgraphIndex_SimpleSweepLineIntersector {
	return &GeomgraphIndex_SimpleSweepLineIntersector{}
}

// ComputeIntersectionsSingleList computes all self-intersections between edges in a set of edges.
func (s *GeomgraphIndex_SimpleSweepLineIntersector) ComputeIntersectionsSingleList(edges []*Geomgraph_Edge, si *GeomgraphIndex_SegmentIntersector, testAllSegments bool) {
	if testAllSegments {
		s.addEdgesWithEdgeSet(edges, nil)
	} else {
		s.addEdges(edges)
	}
	s.computeIntersections(si)
}

// ComputeIntersectionsTwoLists computes all mutual intersections between two sets of edges.
func (s *GeomgraphIndex_SimpleSweepLineIntersector) ComputeIntersectionsTwoLists(edges0, edges1 []*Geomgraph_Edge, si *GeomgraphIndex_SegmentIntersector) {
	s.addEdgesWithEdgeSet(edges0, edges0)
	s.addEdgesWithEdgeSet(edges1, edges1)
	s.computeIntersections(si)
}

func (s *GeomgraphIndex_SimpleSweepLineIntersector) addEdges(edges []*Geomgraph_Edge) {
	for _, edge := range edges {
		// Edge is its own group.
		s.addEdge(edge, edge)
	}
}

func (s *GeomgraphIndex_SimpleSweepLineIntersector) addEdgesWithEdgeSet(edges []*Geomgraph_Edge, edgeSet any) {
	for _, edge := range edges {
		s.addEdge(edge, edgeSet)
	}
}

func (s *GeomgraphIndex_SimpleSweepLineIntersector) addEdge(edge *Geomgraph_Edge, edgeSet any) {
	pts := edge.GetCoordinates()
	for i := 0; i < len(pts)-1; i++ {
		ss := GeomgraphIndex_NewSweepLineSegment(edge, i)
		insertEvent := GeomgraphIndex_NewSweepLineEventInsert(edgeSet, ss.GetMinX(), ss)
		s.events = append(s.events, insertEvent)
		s.events = append(s.events, GeomgraphIndex_NewSweepLineEventDelete(ss.GetMaxX(), insertEvent))
	}
}

// prepareEvents sorts events and sets DELETE event indexes.
// Because DELETE events have a link to their corresponding INSERT event,
// it is possible to compute exactly the range of events which must be
// compared to a given INSERT event object.
func (s *GeomgraphIndex_SimpleSweepLineIntersector) prepareEvents() {
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

func (s *GeomgraphIndex_SimpleSweepLineIntersector) computeIntersections(si *GeomgraphIndex_SegmentIntersector) {
	s.nOverlaps = 0
	s.prepareEvents()

	for i, ev := range s.events {
		if ev.IsInsert() {
			s.processOverlaps(i, ev.GetDeleteEventIndex(), ev, si)
		}
	}
}

func (s *GeomgraphIndex_SimpleSweepLineIntersector) processOverlaps(start, end int, ev0 *GeomgraphIndex_SweepLineEvent, si *GeomgraphIndex_SegmentIntersector) {
	ss0 := ev0.GetObject().(*GeomgraphIndex_SweepLineSegment)
	// Since we might need to test for self-intersections, include current
	// INSERT event object in list of event objects to test.
	// Last index can be skipped, because it must be a Delete event.
	for i := start; i < end; i++ {
		ev1 := s.events[i]
		if ev1.IsInsert() {
			ss1 := ev1.GetObject().(*GeomgraphIndex_SweepLineSegment)
			// Don't compare edges in same group, if labels are present.
			if !ev0.IsSameLabel(ev1) {
				ss0.ComputeIntersections(ss1, si)
				s.nOverlaps++
			}
		}
	}
}
