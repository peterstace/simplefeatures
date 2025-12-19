package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Planargraph_Node_GetEdgesBetween returns all Edges that connect the two nodes (which are assumed to be different).
func Planargraph_Node_GetEdgesBetween(node0, node1 *Planargraph_Node) []*Planargraph_Edge {
	edges0 := Planargraph_DirectedEdge_ToEdges(node0.GetOutEdges().GetEdges())
	edges1 := Planargraph_DirectedEdge_ToEdges(node1.GetOutEdges().GetEdges())

	// Create a set from edges0.
	edgeSet := make(map[*Planargraph_Edge]bool)
	for _, e := range edges0 {
		if e != nil {
			edgeSet[e] = true
		}
	}

	// Retain only edges that are also in edges1.
	var commonEdges []*Planargraph_Edge
	for _, e := range edges1 {
		if e != nil && edgeSet[e] {
			commonEdges = append(commonEdges, e)
		}
	}
	return commonEdges
}

// Planargraph_Node is a node in a PlanarGraph. It is a location where 0 or more Edges
// meet. A node is connected to each of its incident Edges via an outgoing
// DirectedEdge. Some clients using a PlanarGraph may want to
// subclass Node to add their own application-specific data and methods.
type Planargraph_Node struct {
	*Planargraph_GraphComponent
	child  java.Polymorphic
	pt     *Geom_Coordinate
	deStar *Planargraph_DirectedEdgeStar
}

func (n *Planargraph_Node) GetChild() java.Polymorphic {
	return n.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (n *Planargraph_Node) GetParent() java.Polymorphic {
	return n.Planargraph_GraphComponent
}

// Planargraph_NewNode constructs a Node with the given location.
func Planargraph_NewNode(pt *Geom_Coordinate) *Planargraph_Node {
	return Planargraph_NewNodeWithStar(pt, Planargraph_NewDirectedEdgeStar())
}

// Planargraph_NewNodeWithStar constructs a Node with the given location and collection of outgoing DirectedEdges.
func Planargraph_NewNodeWithStar(pt *Geom_Coordinate, deStar *Planargraph_DirectedEdgeStar) *Planargraph_Node {
	gc := &Planargraph_GraphComponent{}
	n := &Planargraph_Node{
		Planargraph_GraphComponent: gc,
		pt:                         pt,
		deStar:                     deStar,
	}
	gc.child = n
	return n
}

// GetCoordinate returns the location of this Node.
func (n *Planargraph_Node) GetCoordinate() *Geom_Coordinate {
	return n.pt
}

// AddOutEdge adds an outgoing DirectedEdge to this Node.
func (n *Planargraph_Node) AddOutEdge(de *Planargraph_DirectedEdge) {
	n.deStar.Add(de)
}

// GetOutEdges returns the collection of DirectedEdges that leave this Node.
func (n *Planargraph_Node) GetOutEdges() *Planargraph_DirectedEdgeStar {
	return n.deStar
}

// GetDegree returns the number of edges around this Node.
func (n *Planargraph_Node) GetDegree() int {
	return n.deStar.GetDegree()
}

// GetIndex returns the zero-based index of the given Edge, after sorting in ascending order
// by angle with the positive x-axis.
func (n *Planargraph_Node) GetIndex(edge *Planargraph_Edge) int {
	return n.deStar.GetIndexByEdge(edge)
}

// Remove removes a DirectedEdge incident on this node.
// Does not change the state of the directed edge.
func (n *Planargraph_Node) Remove(de *Planargraph_DirectedEdge) {
	n.deStar.Remove(de)
}

// remove removes this node from its containing graph.
func (n *Planargraph_Node) remove() {
	n.pt = nil
}

// IsRemoved_BODY tests whether this node has been removed from its containing graph.
func (n *Planargraph_Node) IsRemoved_BODY() bool {
	return n.pt == nil
}
