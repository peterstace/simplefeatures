package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationOverlay_OverlayNodeFactory creates nodes for use in the PlanarGraphs
// constructed during overlay operations.
type OperationOverlay_OverlayNodeFactory struct {
	*Geomgraph_NodeFactory
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (onf *OperationOverlay_OverlayNodeFactory) GetChild() java.Polymorphic {
	return onf.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (onf *OperationOverlay_OverlayNodeFactory) GetParent() java.Polymorphic {
	return onf.Geomgraph_NodeFactory
}

// OperationOverlay_NewOverlayNodeFactory creates a new OverlayNodeFactory.
func OperationOverlay_NewOverlayNodeFactory() *OperationOverlay_OverlayNodeFactory {
	nf := Geomgraph_NewNodeFactory()
	onf := &OperationOverlay_OverlayNodeFactory{
		Geomgraph_NodeFactory: nf,
	}
	nf.child = onf
	return onf
}

// CreateNode_BODY creates a node at the given coordinate with a
// DirectedEdgeStar.
func (onf *OperationOverlay_OverlayNodeFactory) CreateNode_BODY(coord *Geom_Coordinate) *Geomgraph_Node {
	return Geomgraph_NewNode(coord, Geomgraph_NewDirectedEdgeStar().Geomgraph_EdgeEndStar)
}
