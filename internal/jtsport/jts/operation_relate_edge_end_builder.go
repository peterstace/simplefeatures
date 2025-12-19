package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationRelate_EdgeEndBuilder creates EdgeEnds for all the "split edges"
// created by the intersections determined for an Edge.
type OperationRelate_EdgeEndBuilder struct {
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (eeb *OperationRelate_EdgeEndBuilder) GetChild() java.Polymorphic {
	return eeb.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (eeb *OperationRelate_EdgeEndBuilder) GetParent() java.Polymorphic {
	return nil
}

// OperationRelate_NewEdgeEndBuilder creates a new EdgeEndBuilder.
func OperationRelate_NewEdgeEndBuilder() *OperationRelate_EdgeEndBuilder {
	return &OperationRelate_EdgeEndBuilder{}
}

// ComputeEdgeEndsFromIterator computes EdgeEnds for all the edges in the given
// iterator and returns them in a list.
func (eeb *OperationRelate_EdgeEndBuilder) ComputeEdgeEndsFromIterator(edges []*Geomgraph_Edge) []*Geomgraph_EdgeEnd {
	l := make([]*Geomgraph_EdgeEnd, 0)
	for _, e := range edges {
		eeb.ComputeEdgeEnds(e, &l)
	}
	return l
}

// ComputeEdgeEnds creates stub edges for all the intersections in this Edge (if
// any) and inserts them into the list.
func (eeb *OperationRelate_EdgeEndBuilder) ComputeEdgeEnds(edge *Geomgraph_Edge, l *[]*Geomgraph_EdgeEnd) {
	eiList := edge.GetEdgeIntersectionList()
	// Ensure that the list has entries for the first and last point of the
	// edge.
	eiList.AddEndpoints()

	intersections := eiList.Iterator()
	// No intersections, so there is nothing to do.
	if len(intersections) == 0 {
		return
	}

	var eiPrev *Geomgraph_EdgeIntersection
	var eiCurr *Geomgraph_EdgeIntersection
	idx := 0
	eiNext := intersections[idx]
	idx++

	for {
		eiPrev = eiCurr
		eiCurr = eiNext
		eiNext = nil
		if idx < len(intersections) {
			eiNext = intersections[idx]
			idx++
		}

		if eiCurr != nil {
			eeb.createEdgeEndForPrev(edge, l, eiCurr, eiPrev)
			eeb.createEdgeEndForNext(edge, l, eiCurr, eiNext)
		}

		if eiCurr == nil {
			break
		}
	}
}

// createEdgeEndForPrev creates a EdgeStub for the edge before the intersection
// eiCurr. The previous intersection is provided in case it is the endpoint for
// the stub edge. Otherwise, the previous point from the parent edge will be the
// endpoint.
//
// eiCurr will always be an EdgeIntersection, but eiPrev may be nil.
func (eeb *OperationRelate_EdgeEndBuilder) createEdgeEndForPrev(
	edge *Geomgraph_Edge,
	l *[]*Geomgraph_EdgeEnd,
	eiCurr, eiPrev *Geomgraph_EdgeIntersection,
) {
	iPrev := eiCurr.SegmentIndex
	if eiCurr.Dist == 0.0 {
		// If at the start of the edge there is no previous edge.
		if iPrev == 0 {
			return
		}
		iPrev--
	}
	pPrev := edge.GetCoordinateAtIndex(iPrev)
	// If prev intersection is past the previous vertex, use it instead.
	if eiPrev != nil && eiPrev.SegmentIndex >= iPrev {
		pPrev = eiPrev.Coord
	}

	label := Geomgraph_NewLabelFromLabel(edge.GetLabel())
	// Since edgeStub is oriented opposite to its parent edge, have to flip
	// sides for edge label.
	label.Flip()
	e := Geomgraph_NewEdgeEndWithLabel(edge, eiCurr.Coord, pPrev, label)
	*l = append(*l, e)
}

// createEdgeEndForNext creates a StubEdge for the edge after the intersection
// eiCurr. The next intersection is provided in case it is the endpoint for the
// stub edge. Otherwise, the next point from the parent edge will be the
// endpoint.
//
// eiCurr will always be an EdgeIntersection, but eiNext may be nil.
func (eeb *OperationRelate_EdgeEndBuilder) createEdgeEndForNext(
	edge *Geomgraph_Edge,
	l *[]*Geomgraph_EdgeEnd,
	eiCurr, eiNext *Geomgraph_EdgeIntersection,
) {
	iNext := eiCurr.SegmentIndex + 1
	// If there is no next edge there is nothing to do.
	if iNext >= edge.GetNumPoints() && eiNext == nil {
		return
	}

	pNext := edge.GetCoordinateAtIndex(iNext)

	// If the next intersection is in the same segment as the current, use it as
	// the endpoint.
	if eiNext != nil && eiNext.SegmentIndex == eiCurr.SegmentIndex {
		pNext = eiNext.Coord
	}

	e := Geomgraph_NewEdgeEndWithLabel(edge, eiCurr.Coord, pNext, Geomgraph_NewLabelFromLabel(edge.GetLabel()))
	*l = append(*l, e)
}
