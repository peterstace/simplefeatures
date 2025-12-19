package jts

// PlanargraphAlgorithm_ConnectedSubgraphFinder finds all connected Subgraphs of a PlanarGraph.
//
// Note: uses the isVisited flag on the nodes.
type PlanargraphAlgorithm_ConnectedSubgraphFinder struct {
	graph *Planargraph_PlanarGraph
}

// PlanargraphAlgorithm_NewConnectedSubgraphFinder creates a new ConnectedSubgraphFinder for the given graph.
func PlanargraphAlgorithm_NewConnectedSubgraphFinder(graph *Planargraph_PlanarGraph) *PlanargraphAlgorithm_ConnectedSubgraphFinder {
	return &PlanargraphAlgorithm_ConnectedSubgraphFinder{
		graph: graph,
	}
}

// GetConnectedSubgraphs returns all connected subgraphs of the graph.
func (csf *PlanargraphAlgorithm_ConnectedSubgraphFinder) GetConnectedSubgraphs() []*Planargraph_Subgraph {
	var subgraphs []*Planargraph_Subgraph

	// Set visited to false on all nodes.
	for _, node := range csf.graph.GetNodes() {
		node.SetVisited(false)
	}

	for _, e := range csf.graph.GetEdges() {
		node := e.GetDirEdge(0).GetFromNode()
		if !node.IsVisited() {
			subgraphs = append(subgraphs, csf.findSubgraph(node))
		}
	}
	return subgraphs
}

func (csf *PlanargraphAlgorithm_ConnectedSubgraphFinder) findSubgraph(node *Planargraph_Node) *Planargraph_Subgraph {
	subgraph := Planargraph_NewSubgraph(csf.graph)
	csf.addReachable(node, subgraph)
	return subgraph
}

// addReachable adds all nodes and edges reachable from this node to the subgraph.
// Uses an explicit stack to avoid a large depth of recursion.
func (csf *PlanargraphAlgorithm_ConnectedSubgraphFinder) addReachable(startNode *Planargraph_Node, subgraph *Planargraph_Subgraph) {
	nodeStack := []*Planargraph_Node{startNode}
	for len(nodeStack) > 0 {
		// Pop from stack.
		node := nodeStack[len(nodeStack)-1]
		nodeStack = nodeStack[:len(nodeStack)-1]
		csf.addEdges(node, &nodeStack, subgraph)
	}
}

// addEdges adds the argument node and all its out edges to the subgraph.
func (csf *PlanargraphAlgorithm_ConnectedSubgraphFinder) addEdges(node *Planargraph_Node, nodeStack *[]*Planargraph_Node, subgraph *Planargraph_Subgraph) {
	node.SetVisited(true)
	for _, de := range node.GetOutEdges().GetEdges() {
		subgraph.Add(de.GetEdge())
		toNode := de.GetToNode()
		if !toNode.IsVisited() {
			*nodeStack = append(*nodeStack, toNode)
		}
	}
}
