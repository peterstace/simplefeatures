package jts

// Planargraph_Subgraph is a subgraph of a PlanarGraph.
// A subgraph may contain any subset of Edges from the parent graph.
// It will also automatically contain all DirectedEdges and Nodes associated with those edges.
// No new objects are created when edges are added -
// all associated components must already exist in the parent graph.
type Planargraph_Subgraph struct {
	parentGraph *Planargraph_PlanarGraph
	edges       map[*Planargraph_Edge]bool
	dirEdges    []*Planargraph_DirectedEdge
	nodeMap     *Planargraph_NodeMap
}

// Planargraph_NewSubgraph creates a new subgraph of the given PlanarGraph.
func Planargraph_NewSubgraph(parentGraph *Planargraph_PlanarGraph) *Planargraph_Subgraph {
	return &Planargraph_Subgraph{
		parentGraph: parentGraph,
		edges:       make(map[*Planargraph_Edge]bool),
		dirEdges:    make([]*Planargraph_DirectedEdge, 0),
		nodeMap:     Planargraph_NewNodeMap(),
	}
}

// GetParent gets the PlanarGraph which this subgraph is part of.
func (s *Planargraph_Subgraph) GetParent() *Planargraph_PlanarGraph {
	return s.parentGraph
}

// Add adds an Edge to the subgraph.
// The associated DirectedEdges and Nodes are also added.
func (s *Planargraph_Subgraph) Add(e *Planargraph_Edge) {
	if s.edges[e] {
		return
	}
	s.edges[e] = true
	s.dirEdges = append(s.dirEdges, e.GetDirEdge(0))
	s.dirEdges = append(s.dirEdges, e.GetDirEdge(1))
	s.nodeMap.Add(e.GetDirEdge(0).GetFromNode())
	s.nodeMap.Add(e.GetDirEdge(1).GetFromNode())
}

// dirEdgeIterator returns an Iterator over the DirectedEdges in this graph,
// in the order in which they were added.
func (s *Planargraph_Subgraph) dirEdgeIterator() []*Planargraph_DirectedEdge {
	return s.dirEdges
}

// edgeIterator returns an Iterator over the Edges in this graph,
// in the order in which they were added.
func (s *Planargraph_Subgraph) edgeIterator() []*Planargraph_Edge {
	result := make([]*Planargraph_Edge, 0, len(s.edges))
	for e := range s.edges {
		result = append(result, e)
	}
	return result
}

// nodeIterator returns an Iterator over the Nodes in this graph.
func (s *Planargraph_Subgraph) nodeIterator() []*Planargraph_Node {
	return s.nodeMap.iterator()
}

// TRANSLITERATION NOTE: GetNodes, GetEdges, and GetDirectedEdges are convenience methods
// added for Go idiomatic usage, wrapping the iterator methods. These are called from
// other parts of the codebase.

// GetNodes returns the Nodes in this graph.
func (s *Planargraph_Subgraph) GetNodes() []*Planargraph_Node {
	return s.nodeIterator()
}

// GetEdges returns the Edges in this graph.
func (s *Planargraph_Subgraph) GetEdges() []*Planargraph_Edge {
	return s.edgeIterator()
}

// GetDirectedEdges returns the DirectedEdges in this graph.
func (s *Planargraph_Subgraph) GetDirectedEdges() []*Planargraph_DirectedEdge {
	return s.dirEdgeIterator()
}

// Contains tests whether an Edge is contained in this subgraph.
func (s *Planargraph_Subgraph) Contains(e *Planargraph_Edge) bool {
	return s.edges[e]
}
