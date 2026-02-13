package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationBuffer_BufferSubgraph is a connected subset of the graph of
// DirectedEdges and Nodes.
// Its edges will generate either
//   - a single polygon in the complete buffer, with zero or more holes, or
//   - one or more connected holes
type OperationBuffer_BufferSubgraph struct {
	finder         *operationBuffer_RightmostEdgeFinder
	dirEdgeList    []*Geomgraph_DirectedEdge
	nodes          []*Geomgraph_Node
	rightMostCoord *Geom_Coordinate
	env            *Geom_Envelope
}

// OperationBuffer_NewBufferSubgraph creates a new BufferSubgraph.
func OperationBuffer_NewBufferSubgraph() *OperationBuffer_BufferSubgraph {
	return &OperationBuffer_BufferSubgraph{
		finder:      operationBuffer_newRightmostEdgeFinder(),
		dirEdgeList: make([]*Geomgraph_DirectedEdge, 0),
		nodes:       make([]*Geomgraph_Node, 0),
	}
}

// GetDirectedEdges returns the list of DirectedEdges.
func (bs *OperationBuffer_BufferSubgraph) GetDirectedEdges() []*Geomgraph_DirectedEdge {
	return bs.dirEdgeList
}

// GetNodes returns the list of nodes.
func (bs *OperationBuffer_BufferSubgraph) GetNodes() []*Geomgraph_Node {
	return bs.nodes
}

// GetEnvelope computes the envelope of the edges in the subgraph.
// The envelope is cached after being computed.
func (bs *OperationBuffer_BufferSubgraph) GetEnvelope() *Geom_Envelope {
	if bs.env == nil {
		edgeEnv := Geom_NewEnvelope()
		for _, dirEdge := range bs.dirEdgeList {
			pts := dirEdge.GetEdge().GetCoordinates()
			for i := 0; i < len(pts)-1; i++ {
				edgeEnv.ExpandToIncludeCoordinate(pts[i])
			}
		}
		bs.env = edgeEnv
	}
	return bs.env
}

// GetRightmostCoordinate gets the rightmost coordinate in the edges of the subgraph.
func (bs *OperationBuffer_BufferSubgraph) GetRightmostCoordinate() *Geom_Coordinate {
	return bs.rightMostCoord
}

// Create creates the subgraph consisting of all edges reachable from this node.
// Finds the edges in the graph and the rightmost coordinate.
func (bs *OperationBuffer_BufferSubgraph) Create(node *Geomgraph_Node) {
	bs.addReachable(node)
	bs.finder.FindEdge(bs.dirEdgeList)
	bs.rightMostCoord = bs.finder.GetCoordinate()
}

// addReachable adds all nodes and edges reachable from this node to the subgraph.
// Uses an explicit stack to avoid a large depth of recursion.
func (bs *OperationBuffer_BufferSubgraph) addReachable(startNode *Geomgraph_Node) {
	nodeStack := make([]*Geomgraph_Node, 0)
	nodeStack = append(nodeStack, startNode)
	for len(nodeStack) > 0 {
		// pop from stack
		node := nodeStack[len(nodeStack)-1]
		nodeStack = nodeStack[:len(nodeStack)-1]
		bs.add(node, &nodeStack)
	}
}

// add adds the argument node and all its out edges to the subgraph.
func (bs *OperationBuffer_BufferSubgraph) add(node *Geomgraph_Node, nodeStack *[]*Geomgraph_Node) {
	node.SetVisited(true)
	bs.nodes = append(bs.nodes, node)
	star := java.Cast[*Geomgraph_DirectedEdgeStar](node.GetEdges())
	for _, ee := range star.GetEdges() {
		de := java.Cast[*Geomgraph_DirectedEdge](ee)
		bs.dirEdgeList = append(bs.dirEdgeList, de)
		sym := de.GetSym()
		symNode := sym.GetNode()
		// NOTE: this is a depth-first traversal of the graph.
		// This will cause a large depth of recursion.
		// It might be better to do a breadth-first traversal.
		if !symNode.IsVisited() {
			*nodeStack = append(*nodeStack, symNode)
		}
	}
}

func (bs *OperationBuffer_BufferSubgraph) clearVisitedEdges() {
	for _, de := range bs.dirEdgeList {
		de.SetVisited(false)
	}
}

// ComputeDepth computes the depth for all edges in the subgraph.
func (bs *OperationBuffer_BufferSubgraph) ComputeDepth(outsideDepth int) {
	bs.clearVisitedEdges()
	// find an outside edge to assign depth to
	de := bs.finder.GetEdge()
	// right side of line returned by finder is on the outside
	de.SetEdgeDepths(Geom_Position_Right, outsideDepth)
	bs.copySymDepths(de)

	bs.computeDepths(de)
}

// computeDepths computes depths for all dirEdges via breadth-first traversal of nodes in graph.
func (bs *OperationBuffer_BufferSubgraph) computeDepths(startEdge *Geomgraph_DirectedEdge) {
	nodesVisited := make(map[*Geomgraph_Node]bool)
	nodeQueue := make([]*Geomgraph_Node, 0)

	startNode := startEdge.GetNode()
	nodeQueue = append(nodeQueue, startNode)
	nodesVisited[startNode] = true
	startEdge.SetVisited(true)

	for len(nodeQueue) > 0 {
		// remove from front of queue
		n := nodeQueue[0]
		nodeQueue = nodeQueue[1:]
		nodesVisited[n] = true
		// compute depths around node, starting at this edge since it has depths assigned
		bs.computeNodeDepth(n)

		// add all adjacent nodes to process queue,
		// unless the node has been visited already
		star := java.Cast[*Geomgraph_DirectedEdgeStar](n.GetEdges())
		for _, ee := range star.GetEdges() {
			de := java.Cast[*Geomgraph_DirectedEdge](ee)
			sym := de.GetSym()
			if sym.IsVisited() {
				continue
			}
			adjNode := sym.GetNode()
			if !nodesVisited[adjNode] {
				nodeQueue = append(nodeQueue, adjNode)
				nodesVisited[adjNode] = true
			}
		}
	}
}

func (bs *OperationBuffer_BufferSubgraph) computeNodeDepth(n *Geomgraph_Node) {
	// find a visited dirEdge to start at
	var startEdge *Geomgraph_DirectedEdge
	star := java.Cast[*Geomgraph_DirectedEdgeStar](n.GetEdges())
	for _, ee := range star.GetEdges() {
		de := java.Cast[*Geomgraph_DirectedEdge](ee)
		if de.IsVisited() || de.GetSym().IsVisited() {
			startEdge = de
			break
		}
	}

	// only compute string append if assertion would fail
	if startEdge == nil {
		panic("unable to find edge to compute depths at " + n.GetCoordinate().String())
	}

	star.ComputeDepths(startEdge)

	// copy depths to sym edges
	for _, ee := range star.GetEdges() {
		de := java.Cast[*Geomgraph_DirectedEdge](ee)
		de.SetVisited(true)
		bs.copySymDepths(de)
	}
}

func (bs *OperationBuffer_BufferSubgraph) copySymDepths(de *Geomgraph_DirectedEdge) {
	sym := de.GetSym()
	sym.SetDepth(Geom_Position_Left, de.GetDepth(Geom_Position_Right))
	sym.SetDepth(Geom_Position_Right, de.GetDepth(Geom_Position_Left))
}

// FindResultEdges finds all edges whose depths indicates that they are in the result area(s).
// Since we want polygon shells to be oriented CW, choose dirEdges with the interior of the result on the RHS.
// Mark them as being in the result.
// Interior Area edges are the result of dimensional collapses.
// They do not form part of the result area boundary.
func (bs *OperationBuffer_BufferSubgraph) FindResultEdges() {
	for _, de := range bs.dirEdgeList {
		// Select edges which have an interior depth on the RHS
		// and an exterior depth on the LHS.
		// Note that because of weird rounding effects there may be
		// edges which have negative depths! Negative depths
		// count as "outside".
		// <FIX> - handle negative depths
		if de.GetDepth(Geom_Position_Right) >= 1 &&
			de.GetDepth(Geom_Position_Left) <= 0 &&
			!de.IsInteriorAreaEdge() {
			de.SetInResult(true)
		}
	}
}

// CompareTo compares BufferSubgraphs on the x-value of their rightmost Coordinate.
// This defines a partial ordering on the graphs such that:
//
//	g1 >= g2 <==> Ring(g2) does not contain Ring(g1)
//
// where Polygon(g) is the buffer polygon that is built from g.
//
// This relationship is used to sort the BufferSubgraphs so that shells are guaranteed to
// be built before holes.
func (bs *OperationBuffer_BufferSubgraph) CompareTo(other *OperationBuffer_BufferSubgraph) int {
	if bs.rightMostCoord.X < other.rightMostCoord.X {
		return -1
	}
	if bs.rightMostCoord.X > other.rightMostCoord.X {
		return 1
	}
	return 0
}
