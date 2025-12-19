package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationRelate_RelateNodeFactory is used by the NodeMap in a RelateNodeGraph
// to create RelateNodes.
type OperationRelate_RelateNodeFactory struct {
	*Geomgraph_NodeFactory
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (rnf *OperationRelate_RelateNodeFactory) GetChild() java.Polymorphic {
	return rnf.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (rnf *OperationRelate_RelateNodeFactory) GetParent() java.Polymorphic {
	return rnf.Geomgraph_NodeFactory
}

// OperationRelate_NewRelateNodeFactory creates a new RelateNodeFactory.
func OperationRelate_NewRelateNodeFactory() *OperationRelate_RelateNodeFactory {
	nf := Geomgraph_NewNodeFactory()
	rnf := &OperationRelate_RelateNodeFactory{
		Geomgraph_NodeFactory: nf,
	}
	nf.child = rnf
	return rnf
}

// CreateNode_BODY creates a RelateNode with the given coordinate.
func (rnf *OperationRelate_RelateNodeFactory) CreateNode_BODY(coord *Geom_Coordinate) *Geomgraph_Node {
	eebs := OperationRelate_NewEdgeEndBundleStar()
	relateNode := OperationRelate_NewRelateNode(coord, eebs.Geomgraph_EdgeEndStar)
	return relateNode.Geomgraph_Node
}
