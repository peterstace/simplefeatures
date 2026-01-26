package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

const (
	operationOverlay_ConsistentPolygonRingChecker_SCANNING_FOR_INCOMING = 1
	operationOverlay_ConsistentPolygonRingChecker_LINKING_TO_OUTGOING   = 2
)

// OperationOverlay_ConsistentPolygonRingChecker tests whether the polygon rings
// in a GeometryGraph are consistent. Used for checking if Topology errors are
// present after noding.
type OperationOverlay_ConsistentPolygonRingChecker struct {
	child java.Polymorphic
	graph *Geomgraph_PlanarGraph
}

// GetChild returns the immediate child in the type hierarchy chain.
func (c *OperationOverlay_ConsistentPolygonRingChecker) GetChild() java.Polymorphic {
	return c.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (c *OperationOverlay_ConsistentPolygonRingChecker) GetParent() java.Polymorphic {
	return nil
}

// OperationOverlay_NewConsistentPolygonRingChecker creates a new checker.
func OperationOverlay_NewConsistentPolygonRingChecker(graph *Geomgraph_PlanarGraph) *OperationOverlay_ConsistentPolygonRingChecker {
	return &OperationOverlay_ConsistentPolygonRingChecker{
		graph: graph,
	}
}

// CheckAll checks all overlay operations for consistency.
func (c *OperationOverlay_ConsistentPolygonRingChecker) CheckAll() {
	c.Check(OperationOverlay_OverlayOp_Intersection)
	c.Check(OperationOverlay_OverlayOp_Difference)
	c.Check(OperationOverlay_OverlayOp_Union)
	c.Check(OperationOverlay_OverlayOp_SymDifference)
}

// Check tests whether the result geometry is consistent.
func (c *OperationOverlay_ConsistentPolygonRingChecker) Check(opCode int) {
	for _, node := range c.graph.GetNodeIterator() {
		des := java.GetLeaf(node.GetEdges()).(*Geomgraph_DirectedEdgeStar)
		c.testLinkResultDirectedEdges(des, opCode)
	}
}

func (c *OperationOverlay_ConsistentPolygonRingChecker) getPotentialResultAreaEdges(deStar *Geomgraph_DirectedEdgeStar, opCode int) []*Geomgraph_DirectedEdge {
	var resultAreaEdgeList []*Geomgraph_DirectedEdge
	for _, ee := range deStar.GetEdges() {
		de := java.Cast[*Geomgraph_DirectedEdge](ee)
		if c.isPotentialResultAreaEdge(de, opCode) || c.isPotentialResultAreaEdge(de.GetSym(), opCode) {
			resultAreaEdgeList = append(resultAreaEdgeList, de)
		}
	}
	return resultAreaEdgeList
}

func (c *OperationOverlay_ConsistentPolygonRingChecker) isPotentialResultAreaEdge(de *Geomgraph_DirectedEdge, opCode int) bool {
	// Mark all dirEdges with the appropriate label.
	label := de.GetLabel()
	if label.IsArea() && !de.IsInteriorAreaEdge() &&
		OperationOverlay_OverlayOp_IsResultOfOp(
			label.GetLocation(0, Geom_Position_Right),
			label.GetLocation(1, Geom_Position_Right),
			opCode) {
		return true
	}
	return false
}

func (c *OperationOverlay_ConsistentPolygonRingChecker) testLinkResultDirectedEdges(deStar *Geomgraph_DirectedEdgeStar, opCode int) {
	// Make sure edges are copied to resultAreaEdges list.
	ringEdges := c.getPotentialResultAreaEdges(deStar, opCode)
	// Find first area edge (if any) to start linking at.
	var firstOut *Geomgraph_DirectedEdge
	state := operationOverlay_ConsistentPolygonRingChecker_SCANNING_FOR_INCOMING
	// Link edges in CCW order.
	for i := 0; i < len(ringEdges); i++ {
		nextOut := ringEdges[i]
		nextIn := nextOut.GetSym()

		// Skip de's that we're not interested in.
		if !nextOut.GetLabel().IsArea() {
			continue
		}

		// Record first outgoing edge, in order to link the last incoming edge.
		if firstOut == nil && c.isPotentialResultAreaEdge(nextOut, opCode) {
			firstOut = nextOut
		}

		switch state {
		case operationOverlay_ConsistentPolygonRingChecker_SCANNING_FOR_INCOMING:
			if !c.isPotentialResultAreaEdge(nextIn, opCode) {
				continue
			}
			state = operationOverlay_ConsistentPolygonRingChecker_LINKING_TO_OUTGOING
		case operationOverlay_ConsistentPolygonRingChecker_LINKING_TO_OUTGOING:
			if !c.isPotentialResultAreaEdge(nextOut, opCode) {
				continue
			}
			state = operationOverlay_ConsistentPolygonRingChecker_SCANNING_FOR_INCOMING
		}
	}
	if state == operationOverlay_ConsistentPolygonRingChecker_LINKING_TO_OUTGOING {
		if firstOut == nil {
			panic(Geom_NewTopologyExceptionWithCoordinate("no outgoing dirEdge found", deStar.GetCoordinate()))
		}
	}
}
