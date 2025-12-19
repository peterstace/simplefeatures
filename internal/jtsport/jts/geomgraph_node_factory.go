package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Geomgraph_NodeFactory is a factory for creating Node objects.
type Geomgraph_NodeFactory struct {
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (nf *Geomgraph_NodeFactory) GetChild() java.Polymorphic {
	return nf.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (nf *Geomgraph_NodeFactory) GetParent() java.Polymorphic {
	return nil
}

// Geomgraph_NewNodeFactory creates a new NodeFactory.
func Geomgraph_NewNodeFactory() *Geomgraph_NodeFactory {
	nf := &Geomgraph_NodeFactory{}
	return nf
}

// CreateNode creates a basic node. The basic node constructor does not allow
// for incident edges.
func (nf *Geomgraph_NodeFactory) CreateNode(coord *Geom_Coordinate) *Geomgraph_Node {
	if impl, ok := java.GetLeaf(nf).(interface {
		CreateNode_BODY(*Geom_Coordinate) *Geomgraph_Node
	}); ok {
		return impl.CreateNode_BODY(coord)
	}
	return nf.CreateNode_BODY(coord)
}

// CreateNode_BODY is the default implementation of CreateNode.
func (nf *Geomgraph_NodeFactory) CreateNode_BODY(coord *Geom_Coordinate) *Geomgraph_Node {
	return Geomgraph_NewNode(coord, nil)
}
