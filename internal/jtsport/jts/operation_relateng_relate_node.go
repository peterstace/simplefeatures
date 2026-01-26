package jts

import "strings"

// OperationRelateng_RelateNode represents a node in the RelateNG topology graph.
type OperationRelateng_RelateNode struct {
	nodePt *Geom_Coordinate
	// A list of the edges around the node in CCW order, ordered by their CCW
	// angle with the positive X-axis.
	edges []*OperationRelateng_RelateEdge
}

// OperationRelateng_NewRelateNode creates a new RelateNode at the given point.
func OperationRelateng_NewRelateNode(pt *Geom_Coordinate) *OperationRelateng_RelateNode {
	return &OperationRelateng_RelateNode{
		nodePt: pt,
		edges:  make([]*OperationRelateng_RelateEdge, 0),
	}
}

// GetCoordinate returns the coordinate of this node.
func (n *OperationRelateng_RelateNode) GetCoordinate() *Geom_Coordinate {
	return n.nodePt
}

// GetEdges returns the edges around this node.
func (n *OperationRelateng_RelateNode) GetEdges() []*OperationRelateng_RelateEdge {
	return n.edges
}

// AddEdges adds edges for a list of node sections.
func (n *OperationRelateng_RelateNode) AddEdges(nss []*OperationRelateng_NodeSection) {
	for _, ns := range nss {
		n.AddEdgesFromSection(ns)
	}
}

// AddEdgesFromSection adds edges for a single node section.
func (n *OperationRelateng_RelateNode) AddEdgesFromSection(ns *OperationRelateng_NodeSection) {
	switch ns.Dimension() {
	case Geom_Dimension_L:
		n.addLineEdge(ns.IsA(), ns.GetVertex(0))
		n.addLineEdge(ns.IsA(), ns.GetVertex(1))
	case Geom_Dimension_A:
		// Assumes node edges have CW orientation (as per JTS norm).
		// Entering edge - interior on L.
		e0 := n.addAreaEdge(ns.IsA(), ns.GetVertex(0), false)
		// Exiting edge - interior on R.
		e1 := n.addAreaEdge(ns.IsA(), ns.GetVertex(1), true)

		index0 := n.indexOf(e0)
		index1 := n.indexOf(e1)
		n.updateEdgesInArea(ns.IsA(), index0, index1)
		n.updateIfAreaPrev(ns.IsA(), index0)
		n.updateIfAreaNext(ns.IsA(), index1)
	}
}

func (n *OperationRelateng_RelateNode) indexOf(e *OperationRelateng_RelateEdge) int {
	for i, edge := range n.edges {
		if edge == e {
			return i
		}
	}
	return -1
}

func (n *OperationRelateng_RelateNode) updateEdgesInArea(isA bool, indexFrom, indexTo int) {
	index := operationRelateng_RelateNode_nextIndex(n.edges, indexFrom)
	for index != indexTo {
		edge := n.edges[index]
		edge.SetAreaInterior(isA)
		index = operationRelateng_RelateNode_nextIndex(n.edges, index)
	}
}

func (n *OperationRelateng_RelateNode) updateIfAreaPrev(isA bool, index int) {
	indexPrev := operationRelateng_RelateNode_prevIndex(n.edges, index)
	edgePrev := n.edges[indexPrev]
	if edgePrev.IsInterior(isA, Geom_Position_Left) {
		edge := n.edges[index]
		edge.SetAreaInterior(isA)
	}
}

func (n *OperationRelateng_RelateNode) updateIfAreaNext(isA bool, index int) {
	indexNext := operationRelateng_RelateNode_nextIndex(n.edges, index)
	edgeNext := n.edges[indexNext]
	if edgeNext.IsInterior(isA, Geom_Position_Right) {
		edge := n.edges[index]
		edge.SetAreaInterior(isA)
	}
}

func (n *OperationRelateng_RelateNode) addLineEdge(isA bool, dirPt *Geom_Coordinate) *OperationRelateng_RelateEdge {
	return n.addEdge(isA, dirPt, Geom_Dimension_L, false)
}

func (n *OperationRelateng_RelateNode) addAreaEdge(isA bool, dirPt *Geom_Coordinate, isForward bool) *OperationRelateng_RelateEdge {
	return n.addEdge(isA, dirPt, Geom_Dimension_A, isForward)
}

// addEdge adds or merges an edge to the node.
func (n *OperationRelateng_RelateNode) addEdge(isA bool, dirPt *Geom_Coordinate, dim int, isForward bool) *OperationRelateng_RelateEdge {
	// Check for well-formed edge - skip null or zero-len input.
	if dirPt == nil {
		return nil
	}
	if n.nodePt.Equals2D(dirPt) {
		return nil
	}

	insertIndex := -1
	for i, e := range n.edges {
		comp := e.CompareToEdge(dirPt)
		if comp == 0 {
			e.Merge(isA, dirPt, dim, isForward)
			return e
		}
		if comp == 1 {
			// Found further edge, so insert a new edge at this position.
			insertIndex = i
			break
		}
	}
	// Add a new edge.
	e := OperationRelateng_RelateEdge_Create(n, dirPt, isA, dim, isForward)
	if insertIndex < 0 {
		// Add edge at end of list.
		n.edges = append(n.edges, e)
	} else {
		// Add edge before higher edge found.
		n.edges = append(n.edges[:insertIndex], append([]*OperationRelateng_RelateEdge{e}, n.edges[insertIndex:]...)...)
	}
	return e
}

// Finish computes the final topology for the edges around this node. Although
// nodes lie on the boundary of areas or the interior of lines, in a mixed GC
// they may also lie in the interior of an area. This changes the locations of
// the sides and line to Interior.
func (n *OperationRelateng_RelateNode) Finish(isAreaInteriorA, isAreaInteriorB bool) {
	n.finishNode(OperationRelateng_RelateGeometry_GEOM_A, isAreaInteriorA)
	n.finishNode(OperationRelateng_RelateGeometry_GEOM_B, isAreaInteriorB)
}

func (n *OperationRelateng_RelateNode) finishNode(isA, isAreaInterior bool) {
	if isAreaInterior {
		OperationRelateng_RelateEdge_SetAreaInteriorAll(n.edges, isA)
	} else {
		startIndex := OperationRelateng_RelateEdge_FindKnownEdgeIndex(n.edges, isA)
		// Only interacting nodes are finished, so this should never happen.
		n.propagateSideLocations(isA, startIndex)
	}
}

func (n *OperationRelateng_RelateNode) propagateSideLocations(isA bool, startIndex int) {
	currLoc := n.edges[startIndex].Location(isA, Geom_Position_Left)
	// Edges are stored in CCW order.
	index := operationRelateng_RelateNode_nextIndex(n.edges, startIndex)
	for index != startIndex {
		e := n.edges[index]
		e.SetUnknownLocations(isA, currLoc)
		currLoc = e.Location(isA, Geom_Position_Left)
		index = operationRelateng_RelateNode_nextIndex(n.edges, index)
	}
}

func operationRelateng_RelateNode_prevIndex(list []*OperationRelateng_RelateEdge, index int) int {
	if index > 0 {
		return index - 1
	}
	// index == 0.
	return len(list) - 1
}

func operationRelateng_RelateNode_nextIndex(list []*OperationRelateng_RelateEdge, i int) int {
	if i >= len(list)-1 {
		return 0
	}
	return i + 1
}

// String returns a string representation of this node.
func (n *OperationRelateng_RelateNode) String() string {
	var buf strings.Builder
	buf.WriteString("Node[")
	buf.WriteString(Io_WKTWriter_ToPoint(n.nodePt))
	buf.WriteString("]:\n")
	for _, e := range n.edges {
		buf.WriteString(e.String())
		buf.WriteString("\n")
	}
	return buf.String()
}

// HasExteriorEdge tests if this node has any exterior edges for the given
// geometry.
func (n *OperationRelateng_RelateNode) HasExteriorEdge(isA bool) bool {
	for _, e := range n.edges {
		if Geom_Location_Exterior == e.Location(isA, Geom_Position_Left) ||
			Geom_Location_Exterior == e.Location(isA, Geom_Position_Right) {
			return true
		}
	}
	return false
}
