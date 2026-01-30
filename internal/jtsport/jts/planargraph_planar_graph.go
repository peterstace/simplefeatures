package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Planargraph_PlanarGraph represents a directed graph which is embeddable in a planar surface.
//
// This class and the other classes in this package serve as a framework for
// building planar graphs for specific algorithms. This class must be
// subclassed to expose appropriate methods to construct the graph. This allows
// controlling the types of graph components (DirectedEdges, Edges and Nodes) which can be added to the graph.
// An application which uses the graph framework will almost always provide
// subclasses for one or more graph components, which hold application-specific
// data and graph algorithms.
type Planargraph_PlanarGraph struct {
	child    java.Polymorphic
	edges    map[*Planargraph_Edge]bool
	dirEdges map[*Planargraph_DirectedEdge]bool
	nodeMap  *Planargraph_NodeMap
}

func (pg *Planargraph_PlanarGraph) GetChild() java.Polymorphic {
	return pg.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (pg *Planargraph_PlanarGraph) GetParent() java.Polymorphic {
	return nil
}

// Planargraph_NewPlanarGraph constructs an empty graph.
func Planargraph_NewPlanarGraph() *Planargraph_PlanarGraph {
	return &Planargraph_PlanarGraph{
		edges:    make(map[*Planargraph_Edge]bool),
		dirEdges: make(map[*Planargraph_DirectedEdge]bool),
		nodeMap:  Planargraph_NewNodeMap(),
	}
}

// FindNode returns the Node at the given location, or nil if no Node was there.
func (pg *Planargraph_PlanarGraph) FindNode(pt *Geom_Coordinate) *Planargraph_Node {
	return pg.nodeMap.Find(pt)
}

// addNode adds a node to the map, replacing any that is already at that location.
// Only subclasses can add Nodes, to ensure Nodes are of the right type.
func (pg *Planargraph_PlanarGraph) addNode(node *Planargraph_Node) {
	pg.nodeMap.Add(node)
}

// addEdge adds the Edge and its DirectedEdges with this PlanarGraph.
// Assumes that the Edge has already been created with its associated DirectedEdges.
// Only subclasses can add Edges, to ensure the edges added are of the right class.
func (pg *Planargraph_PlanarGraph) addEdge(edge *Planargraph_Edge) {
	pg.edges[edge] = true
	pg.addDirectedEdge(edge.GetDirEdge(0))
	pg.addDirectedEdge(edge.GetDirEdge(1))
}

// addDirectedEdge adds the DirectedEdge to this PlanarGraph; only subclasses can add DirectedEdges,
// to ensure the edges added are of the right class.
func (pg *Planargraph_PlanarGraph) addDirectedEdge(dirEdge *Planargraph_DirectedEdge) {
	pg.dirEdges[dirEdge] = true
}

// nodeIterator returns an Iterator over the Nodes in this PlanarGraph.
func (pg *Planargraph_PlanarGraph) nodeIterator() []*Planargraph_Node {
	return pg.nodeMap.iterator()
}

// TRANSLITERATION NOTE: Java has both iterator methods (lines 97, 133, 140) and
// collection getter methods (lines 101, 124, 146). Go preserves both patterns for
// 1-1 correspondence.

// ContainsEdge tests whether this graph contains the given Edge.
func (pg *Planargraph_PlanarGraph) ContainsEdge(e *Planargraph_Edge) bool {
	return pg.edges[e]
}

// ContainsDirectedEdge tests whether this graph contains the given DirectedEdge.
func (pg *Planargraph_PlanarGraph) ContainsDirectedEdge(de *Planargraph_DirectedEdge) bool {
	return pg.dirEdges[de]
}

// GetNodes returns the Nodes in this PlanarGraph.
func (pg *Planargraph_PlanarGraph) GetNodes() []*Planargraph_Node {
	return pg.nodeMap.Values()
}

// dirEdgeIterator returns an Iterator over the DirectedEdges in this PlanarGraph, in the order in which they
// were added.
func (pg *Planargraph_PlanarGraph) dirEdgeIterator() []*Planargraph_DirectedEdge {
	result := make([]*Planargraph_DirectedEdge, 0, len(pg.dirEdges))
	for de := range pg.dirEdges {
		result = append(result, de)
	}
	return result
}

// edgeIterator returns an Iterator over the Edges in this PlanarGraph, in the order in which they
// were added.
func (pg *Planargraph_PlanarGraph) edgeIterator() []*Planargraph_Edge {
	result := make([]*Planargraph_Edge, 0, len(pg.edges))
	for e := range pg.edges {
		result = append(result, e)
	}
	return result
}

// GetEdges returns the Edges that have been added to this PlanarGraph.
func (pg *Planargraph_PlanarGraph) GetEdges() []*Planargraph_Edge {
	return pg.edgeIterator()
}

// RemoveEdge removes an Edge and its associated DirectedEdges
// from their from-Nodes and from the graph.
// Note: This method does not remove the Nodes associated
// with the Edge, even if the removal of the Edge
// reduces the degree of a Node to zero.
func (pg *Planargraph_PlanarGraph) RemoveEdge(edge *Planargraph_Edge) {
	pg.RemoveDirectedEdge(edge.GetDirEdge(0))
	pg.RemoveDirectedEdge(edge.GetDirEdge(1))
	delete(pg.edges, edge)
	edge.remove()
}

// RemoveDirectedEdge removes a DirectedEdge from its from-Node and from this graph.
// This method does not remove the Nodes associated with the DirectedEdge,
// even if the removal of the DirectedEdge reduces the degree of a Node to zero.
func (pg *Planargraph_PlanarGraph) RemoveDirectedEdge(de *Planargraph_DirectedEdge) {
	sym := de.GetSym()
	if sym != nil {
		sym.SetSym(nil)
	}
	de.GetFromNode().Remove(de)
	de.remove()
	delete(pg.dirEdges, de)
}

// RemoveNode removes a node from the graph, along with any associated DirectedEdges and Edges.
func (pg *Planargraph_PlanarGraph) RemoveNode(node *Planargraph_Node) {
	// Unhook all directed edges.
	outEdges := node.GetOutEdges().GetEdges()
	for _, de := range outEdges {
		sym := de.GetSym()
		// Remove the directed edge that points to this node.
		if sym != nil {
			pg.RemoveDirectedEdge(sym)
		}
		// Remove this directed edge from the graph collection.
		delete(pg.dirEdges, de)

		edge := de.GetEdge()
		if edge != nil {
			delete(pg.edges, edge)
		}
	}
	// Remove the node from the graph.
	pg.nodeMap.Remove(node.GetCoordinate())
	node.remove()
}

// FindNodesOfDegree returns all Nodes with the given number of Edges around it.
func (pg *Planargraph_PlanarGraph) FindNodesOfDegree(degree int) []*Planargraph_Node {
	var nodesFound []*Planargraph_Node
	for _, node := range pg.nodeIterator() {
		if node.GetDegree() == degree {
			nodesFound = append(nodesFound, node)
		}
	}
	return nodesFound
}
