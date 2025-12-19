package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationLinemerge_LineMergeEdge is an edge of a LineMergeGraph. The marked field indicates
// whether this Edge has been logically deleted from the graph.
type OperationLinemerge_LineMergeEdge struct {
	*Planargraph_Edge
	child java.Polymorphic
	line  *Geom_LineString
}

func (e *OperationLinemerge_LineMergeEdge) GetChild() java.Polymorphic {
	return e.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (e *OperationLinemerge_LineMergeEdge) GetParent() java.Polymorphic {
	return e.Planargraph_Edge
}

// OperationLinemerge_NewLineMergeEdge constructs a LineMergeEdge with vertices given by the specified LineString.
func OperationLinemerge_NewLineMergeEdge(line *Geom_LineString) *OperationLinemerge_LineMergeEdge {
	gc := &Planargraph_GraphComponent{}
	edge := &Planargraph_Edge{Planargraph_GraphComponent: gc}
	lme := &OperationLinemerge_LineMergeEdge{
		Planargraph_Edge: edge,
		line:            line,
	}
	gc.child = edge
	edge.child = lme
	return lme
}

// GetLine returns the LineString specifying the vertices of this edge.
func (e *OperationLinemerge_LineMergeEdge) GetLine() *Geom_LineString {
	return e.line
}
