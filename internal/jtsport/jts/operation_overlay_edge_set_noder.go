package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationOverlay_EdgeSetNoder nodes a set of edges. Takes one or more sets of
// edges and constructs a new set of edges consisting of all the split edges
// created by noding the input edges together.
type OperationOverlay_EdgeSetNoder struct {
	child java.Polymorphic

	li         *Algorithm_LineIntersector
	inputEdges []*Geomgraph_Edge
}

// GetChild returns the immediate child in the type hierarchy chain.
func (esn *OperationOverlay_EdgeSetNoder) GetChild() java.Polymorphic {
	return esn.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (esn *OperationOverlay_EdgeSetNoder) GetParent() java.Polymorphic {
	return nil
}

// OperationOverlay_NewEdgeSetNoder creates a new EdgeSetNoder.
func OperationOverlay_NewEdgeSetNoder(li *Algorithm_LineIntersector) *OperationOverlay_EdgeSetNoder {
	return &OperationOverlay_EdgeSetNoder{
		li: li,
	}
}

// AddEdges adds edges to be noded.
func (esn *OperationOverlay_EdgeSetNoder) AddEdges(edges []*Geomgraph_Edge) {
	esn.inputEdges = append(esn.inputEdges, edges...)
}

// GetNodedEdges returns the noded edges.
func (esn *OperationOverlay_EdgeSetNoder) GetNodedEdges() []*Geomgraph_Edge {
	esi := GeomgraphIndex_NewSimpleMCSweepLineIntersector()
	si := GeomgraphIndex_NewSegmentIntersector(esn.li, true, false)
	esi.ComputeIntersectionsSingleList(esn.inputEdges, si, true)

	var splitEdges []*Geomgraph_Edge
	for _, e := range esn.inputEdges {
		e.GetEdgeIntersectionList().AddSplitEdges(&splitEdges)
	}
	return splitEdges
}
