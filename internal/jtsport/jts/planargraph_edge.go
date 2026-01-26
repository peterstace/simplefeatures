package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Planargraph_Edge represents an undirected edge of a PlanarGraph. An undirected edge
// in fact simply acts as a central point of reference for two opposite
// DirectedEdges.
//
// Usually a client using a PlanarGraph will subclass Edge
// to add its own application-specific data and methods.
type Planargraph_Edge struct {
	*Planargraph_GraphComponent
	child   java.Polymorphic
	dirEdge []*Planargraph_DirectedEdge
}

func (e *Planargraph_Edge) GetChild() java.Polymorphic {
	return e.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (e *Planargraph_Edge) GetParent() java.Polymorphic {
	return e.Planargraph_GraphComponent
}

// Planargraph_NewEdge constructs an Edge whose DirectedEdges are not yet set. Be sure to call
// SetDirectedEdges.
func Planargraph_NewEdge() *Planargraph_Edge {
	gc := &Planargraph_GraphComponent{}
	edge := &Planargraph_Edge{
		Planargraph_GraphComponent: gc,
	}
	gc.child = edge
	return edge
}

// Planargraph_NewEdgeWithDirectedEdges constructs an Edge initialized with the given DirectedEdges, and for each
// DirectedEdge: sets the Edge, sets the symmetric DirectedEdge, and adds
// this Edge to its from-Node.
func Planargraph_NewEdgeWithDirectedEdges(de0, de1 *Planargraph_DirectedEdge) *Planargraph_Edge {
	gc := &Planargraph_GraphComponent{}
	edge := &Planargraph_Edge{
		Planargraph_GraphComponent: gc,
	}
	gc.child = edge
	edge.SetDirectedEdges(de0, de1)
	return edge
}

// SetDirectedEdges initializes this Edge's two DirectedEdges, and for each DirectedEdge: sets the
// Edge, sets the symmetric DirectedEdge, and adds this Edge to its from-Node.
func (e *Planargraph_Edge) SetDirectedEdges(de0, de1 *Planargraph_DirectedEdge) {
	e.dirEdge = []*Planargraph_DirectedEdge{de0, de1}
	de0.SetEdge(e)
	de1.SetEdge(e)
	de0.SetSym(de1)
	de1.SetSym(de0)
	de0.GetFromNode().AddOutEdge(de0)
	de1.GetFromNode().AddOutEdge(de1)
}

// GetDirEdge returns one of the DirectedEdges associated with this Edge.
// i is 0 or 1. 0 returns the forward directed edge, 1 returns the reverse.
func (e *Planargraph_Edge) GetDirEdge(i int) *Planargraph_DirectedEdge {
	return e.dirEdge[i]
}

// GetDirEdgeNode returns the DirectedEdge that starts from the given node, or nil if the
// node is not one of the two nodes associated with this Edge.
func (e *Planargraph_Edge) GetDirEdgeNode(fromNode *Planargraph_Node) *Planargraph_DirectedEdge {
	if e.dirEdge[0].GetFromNode() == fromNode {
		return e.dirEdge[0]
	}
	if e.dirEdge[1].GetFromNode() == fromNode {
		return e.dirEdge[1]
	}
	// Node not found.
	return nil
}

// GetOppositeNode returns the other node if node is one of the two nodes associated with this Edge;
// otherwise returns nil.
func (e *Planargraph_Edge) GetOppositeNode(node *Planargraph_Node) *Planargraph_Node {
	if e.dirEdge[0].GetFromNode() == node {
		return e.dirEdge[0].GetToNode()
	}
	if e.dirEdge[1].GetFromNode() == node {
		return e.dirEdge[1].GetToNode()
	}
	// Node not found.
	return nil
}

// remove removes this edge from its containing graph.
func (e *Planargraph_Edge) remove() {
	e.dirEdge = nil
}

// IsRemoved_BODY tests whether this edge has been removed from its containing graph.
func (e *Planargraph_Edge) IsRemoved_BODY() bool {
	return e.dirEdge == nil
}
