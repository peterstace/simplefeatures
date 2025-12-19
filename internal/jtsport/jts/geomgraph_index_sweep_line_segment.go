package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// GeomgraphIndex_SweepLineSegment represents a segment used in sweep line algorithms.
type GeomgraphIndex_SweepLineSegment struct {
	child java.Polymorphic

	edge    *Geomgraph_Edge
	pts     []*Geom_Coordinate
	ptIndex int
}

// GetChild returns the immediate child in the type hierarchy chain.
func (s *GeomgraphIndex_SweepLineSegment) GetChild() java.Polymorphic {
	return s.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (s *GeomgraphIndex_SweepLineSegment) GetParent() java.Polymorphic {
	return nil
}

// GeomgraphIndex_NewSweepLineSegment creates a new SweepLineSegment.
func GeomgraphIndex_NewSweepLineSegment(edge *Geomgraph_Edge, ptIndex int) *GeomgraphIndex_SweepLineSegment {
	return &GeomgraphIndex_SweepLineSegment{
		edge:    edge,
		ptIndex: ptIndex,
		pts:     edge.GetCoordinates(),
	}
}

// GetMinX returns the minimum x coordinate of this segment.
func (s *GeomgraphIndex_SweepLineSegment) GetMinX() float64 {
	x1 := s.pts[s.ptIndex].X
	x2 := s.pts[s.ptIndex+1].X
	if x1 < x2 {
		return x1
	}
	return x2
}

// GetMaxX returns the maximum x coordinate of this segment.
func (s *GeomgraphIndex_SweepLineSegment) GetMaxX() float64 {
	x1 := s.pts[s.ptIndex].X
	x2 := s.pts[s.ptIndex+1].X
	if x1 > x2 {
		return x1
	}
	return x2
}

// ComputeIntersections computes intersections between this segment and another.
func (s *GeomgraphIndex_SweepLineSegment) ComputeIntersections(ss *GeomgraphIndex_SweepLineSegment, si *GeomgraphIndex_SegmentIntersector) {
	si.AddIntersections(s.edge, s.ptIndex, ss.edge, ss.ptIndex)
}
