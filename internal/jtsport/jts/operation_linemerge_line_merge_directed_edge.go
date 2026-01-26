package jts

import (
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// OperationLinemerge_LineMergeDirectedEdge is a DirectedEdge of a LineMergeGraph.
type OperationLinemerge_LineMergeDirectedEdge struct {
	*Planargraph_DirectedEdge
	child java.Polymorphic
}

func (de *OperationLinemerge_LineMergeDirectedEdge) GetChild() java.Polymorphic {
	return de.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (de *OperationLinemerge_LineMergeDirectedEdge) GetParent() java.Polymorphic {
	return de.Planargraph_DirectedEdge
}

// OperationLinemerge_NewLineMergeDirectedEdge constructs a LineMergeDirectedEdge connecting
// the from node to the to node.
//
// directionPt specifies this DirectedEdge's direction (given by an imaginary
// line from the from node to directionPt).
//
// edgeDirection indicates whether this DirectedEdge's direction is the same as or
// opposite to that of the parent Edge (if any).
func OperationLinemerge_NewLineMergeDirectedEdge(from, to *Planargraph_Node, directionPt *Geom_Coordinate, edgeDirection bool) *OperationLinemerge_LineMergeDirectedEdge {
	gc := &Planargraph_GraphComponent{}
	de := &Planargraph_DirectedEdge{
		Planargraph_GraphComponent: gc,
		from:                      from,
		to:                        to,
		edgeDirection:             edgeDirection,
		p0:                        from.GetCoordinate(),
		p1:                        directionPt,
	}
	lmde := &OperationLinemerge_LineMergeDirectedEdge{
		Planargraph_DirectedEdge: de,
	}
	gc.child = de
	de.child = lmde

	dx := de.p1.GetX() - de.p0.GetX()
	dy := de.p1.GetY() - de.p0.GetY()
	de.quadrant = Geom_Quadrant_QuadrantFromDeltas(dx, dy)
	de.angle = math.Atan2(dy, dx)

	return lmde
}

// GetNext returns the directed edge that starts at this directed edge's end point, or nil
// if there are zero or multiple directed edges starting there.
func (de *OperationLinemerge_LineMergeDirectedEdge) GetNext() *OperationLinemerge_LineMergeDirectedEdge {
	if de.GetToNode().GetDegree() != 2 {
		return nil
	}
	edges := de.GetToNode().GetOutEdges().GetEdges()
	if edges[0] == de.GetSym() {
		return edges[1].GetChild().(*OperationLinemerge_LineMergeDirectedEdge)
	}
	Util_Assert_IsTrue(edges[1] == de.GetSym())
	return edges[0].GetChild().(*OperationLinemerge_LineMergeDirectedEdge)
}
