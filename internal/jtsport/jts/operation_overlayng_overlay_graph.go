package jts

// OperationOverlayng_OverlayGraph is a planar graph of edges, representing the
// topology resulting from an overlay operation. Each source edge is
// represented by a pair of OverlayEdges, with opposite (symmetric) orientation.
// The pair of OverlayEdges share the edge coordinates and a single OverlayLabel.
type OperationOverlayng_OverlayGraph struct {
	edges   []*OperationOverlayng_OverlayEdge
	nodeMap map[planargraph_CoordKey]*OperationOverlayng_OverlayEdge
}

// OperationOverlayng_NewOverlayGraph creates an empty graph.
func OperationOverlayng_NewOverlayGraph() *OperationOverlayng_OverlayGraph {
	return &OperationOverlayng_OverlayGraph{
		edges:   make([]*OperationOverlayng_OverlayEdge, 0),
		nodeMap: make(map[planargraph_CoordKey]*OperationOverlayng_OverlayEdge),
	}
}

// GetEdges gets the set of edges in this graph. Only one of each symmetric pair
// of OverlayEdges is included. The opposing edge can be found by using Sym().
func (og *OperationOverlayng_OverlayGraph) GetEdges() []*OperationOverlayng_OverlayEdge {
	return og.edges
}

// GetNodeEdges gets the collection of edges representing the nodes in this
// graph. For each star of edges originating at a node a single representative
// edge is included. The other edges around the node can be found by following
// the next and prev links.
func (og *OperationOverlayng_OverlayGraph) GetNodeEdges() []*OperationOverlayng_OverlayEdge {
	result := make([]*OperationOverlayng_OverlayEdge, 0, len(og.nodeMap))
	for _, edge := range og.nodeMap {
		result = append(result, edge)
	}
	return result
}

// GetNodeEdge gets an edge originating at the given node point.
func (og *OperationOverlayng_OverlayGraph) GetNodeEdge(nodePt *Geom_Coordinate) *OperationOverlayng_OverlayEdge {
	key := planargraph_coordToKey(nodePt)
	return og.nodeMap[key]
}

// GetResultAreaEdges gets the representative edges marked as being in the
// result area.
func (og *OperationOverlayng_OverlayGraph) GetResultAreaEdges() []*OperationOverlayng_OverlayEdge {
	resultEdges := make([]*OperationOverlayng_OverlayEdge, 0)
	for _, edge := range og.GetEdges() {
		if edge.IsInResultArea() {
			resultEdges = append(resultEdges, edge)
		}
	}
	return resultEdges
}

// AddEdge adds a new edge to this graph, for the given linework and topology
// information. A pair of OverlayEdges with opposite (symmetric) orientation is
// added, sharing the same OverlayLabel.
func (og *OperationOverlayng_OverlayGraph) AddEdge(pts []*Geom_Coordinate, label *OperationOverlayng_OverlayLabel) *OperationOverlayng_OverlayEdge {
	e := OperationOverlayng_OverlayEdge_CreateEdgePair(pts, label)
	og.insert(e)
	og.insert(e.SymOE())
	return e
}

// insert inserts a single half-edge into the graph. The sym edge must also be
// inserted.
func (og *OperationOverlayng_OverlayGraph) insert(e *OperationOverlayng_OverlayEdge) {
	og.edges = append(og.edges, e)

	// If the edge origin node is already in the graph, insert the edge into
	// the star of edges around the node. Otherwise, add a new node for the origin.
	key := planargraph_coordToKey(e.Orig())
	nodeEdge, exists := og.nodeMap[key]
	if exists {
		nodeEdge.Insert(e.Edgegraph_HalfEdge)
	} else {
		og.nodeMap[key] = e
	}
}
