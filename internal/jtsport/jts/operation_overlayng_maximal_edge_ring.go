package jts

const (
	operationOverlayng_MaximalEdgeRing_STATE_FIND_INCOMING = 1
	operationOverlayng_MaximalEdgeRing_STATE_LINK_OUTGOING = 2
)

// OperationOverlayng_MaximalEdgeRing represents a maximal edge ring formed by
// linking result edges around nodes.
type OperationOverlayng_MaximalEdgeRing struct {
	startEdge *OperationOverlayng_OverlayEdge
}

// OperationOverlayng_MaximalEdgeRing_LinkResultAreaMaxRingAtNode traverses the
// star of edges originating at a node and links consecutive result edges
// together into maximal edge rings. To link two edges the resultNextMax pointer
// for an incoming result edge is set to the next outgoing result edge.
//
// Edges are linked when:
//   - they belong to an area (i.e. they have sides)
//   - they are marked as being in the result
//
// Edges are linked in CCW order (which is the order they are linked in the
// underlying graph). This means that rings have their face on the Right (in
// other words, the topological location of the face is given by the RHS label
// of the DirectedEdge). This produces rings with CW orientation.
//
// PRECONDITIONS:
//   - This edge is in the result
//   - This edge is not yet linked
//   - The edge and its sym are NOT both marked as being in the result
func OperationOverlayng_MaximalEdgeRing_LinkResultAreaMaxRingAtNode(nodeEdge *OperationOverlayng_OverlayEdge) {
	Util_Assert_IsTrueWithMessage(nodeEdge.IsInResultArea(), "Attempt to link non-result edge")

	// Since the node edge is an out-edge, make it the last edge to be linked
	// by starting at the next edge. The node edge cannot be an in-edge as well,
	// but the next one may be the first in-edge.
	endOut := nodeEdge.ONextOE()
	currOut := endOut
	state := operationOverlayng_MaximalEdgeRing_STATE_FIND_INCOMING
	var currResultIn *OperationOverlayng_OverlayEdge
	for {
		// If an edge is linked this node has already been processed so can
		// skip further processing
		if currResultIn != nil && currResultIn.IsResultMaxLinked() {
			return
		}

		switch state {
		case operationOverlayng_MaximalEdgeRing_STATE_FIND_INCOMING:
			currIn := currOut.SymOE()
			if !currIn.IsInResultArea() {
				break
			}
			currResultIn = currIn
			state = operationOverlayng_MaximalEdgeRing_STATE_LINK_OUTGOING
		case operationOverlayng_MaximalEdgeRing_STATE_LINK_OUTGOING:
			if !currOut.IsInResultArea() {
				break
			}
			// link the in edge to the out edge
			currResultIn.SetNextResultMax(currOut)
			state = operationOverlayng_MaximalEdgeRing_STATE_FIND_INCOMING
		}
		currOut = currOut.ONextOE()
		if currOut == endOut {
			break
		}
	}
	if state == operationOverlayng_MaximalEdgeRing_STATE_LINK_OUTGOING {
		panic(Geom_NewTopologyExceptionWithCoordinate("no outgoing edge found", nodeEdge.GetCoordinate()))
	}
}

// OperationOverlayng_NewMaximalEdgeRing creates a new MaximalEdgeRing starting
// at the given edge.
func OperationOverlayng_NewMaximalEdgeRing(e *OperationOverlayng_OverlayEdge) *OperationOverlayng_MaximalEdgeRing {
	mer := &OperationOverlayng_MaximalEdgeRing{
		startEdge: e,
	}
	mer.attachEdges(e)
	return mer
}

func (mer *OperationOverlayng_MaximalEdgeRing) attachEdges(startEdge *OperationOverlayng_OverlayEdge) {
	edge := startEdge
	for {
		if edge == nil {
			panic(Geom_NewTopologyException("Ring edge is null"))
		}
		if edge.GetEdgeRingMax() == mer {
			panic(Geom_NewTopologyExceptionWithCoordinate("Ring edge visited twice at "+edge.GetCoordinate().String(), edge.GetCoordinate()))
		}
		if edge.NextResultMax() == nil {
			panic(Geom_NewTopologyExceptionWithCoordinate("Ring edge missing at", edge.Dest()))
		}
		edge.SetEdgeRingMax(mer)
		edge = edge.NextResultMax()
		if edge == startEdge {
			break
		}
	}
}

// BuildMinimalRings builds the minimal edge rings from this maximal edge ring.
func (mer *OperationOverlayng_MaximalEdgeRing) BuildMinimalRings(geometryFactory *Geom_GeometryFactory) []*OperationOverlayng_OverlayEdgeRing {
	mer.linkMinimalRings()

	minEdgeRings := make([]*OperationOverlayng_OverlayEdgeRing, 0)
	e := mer.startEdge
	for {
		if e.GetEdgeRing() == nil {
			minEr := OperationOverlayng_NewOverlayEdgeRing(e, geometryFactory)
			minEdgeRings = append(minEdgeRings, minEr)
		}
		e = e.NextResultMax()
		if e == mer.startEdge {
			break
		}
	}
	return minEdgeRings
}

func (mer *OperationOverlayng_MaximalEdgeRing) linkMinimalRings() {
	e := mer.startEdge
	for {
		operationOverlayng_MaximalEdgeRing_linkMinRingEdgesAtNode(e, mer)
		e = e.NextResultMax()
		if e == mer.startEdge {
			break
		}
	}
}

// linkMinRingEdgesAtNode links the edges of a MaximalEdgeRing around this node
// into minimal edge rings (OverlayEdgeRings). Minimal ring edges are linked in
// the opposite orientation (CW) to the maximal ring. This changes self-touching
// rings into a two or more separate rings, as per the OGC SFS polygon topology
// semantics. This relinking must be done to each max ring separately, rather
// than all the node result edges, since there may be more than one max ring
// incident at the node.
func operationOverlayng_MaximalEdgeRing_linkMinRingEdgesAtNode(nodeEdge *OperationOverlayng_OverlayEdge, maxRing *OperationOverlayng_MaximalEdgeRing) {
	// The node edge is an out-edge, so it is the first edge linked with the
	// next CCW in-edge
	endOut := nodeEdge
	currMaxRingOut := endOut
	currOut := endOut.ONextOE()
	for {
		if operationOverlayng_MaximalEdgeRing_isAlreadyLinked(currOut.SymOE(), maxRing) {
			return
		}

		if currMaxRingOut == nil {
			currMaxRingOut = operationOverlayng_MaximalEdgeRing_selectMaxOutEdge(currOut, maxRing)
		} else {
			currMaxRingOut = operationOverlayng_MaximalEdgeRing_linkMaxInEdge(currOut, currMaxRingOut, maxRing)
		}
		currOut = currOut.ONextOE()
		if currOut == endOut {
			break
		}
	}
	if currMaxRingOut != nil {
		panic(Geom_NewTopologyExceptionWithCoordinate("Unmatched edge found during min-ring linking", nodeEdge.GetCoordinate()))
	}
}

// isAlreadyLinked tests if an edge of the maximal edge ring is already linked
// into a minimal OverlayEdgeRing. If so, this node has already been processed
// earlier in the maximal edgering linking scan.
func operationOverlayng_MaximalEdgeRing_isAlreadyLinked(edge *OperationOverlayng_OverlayEdge, maxRing *OperationOverlayng_MaximalEdgeRing) bool {
	return edge.GetEdgeRingMax() == maxRing && edge.IsResultLinked()
}

func operationOverlayng_MaximalEdgeRing_selectMaxOutEdge(currOut *OperationOverlayng_OverlayEdge, maxEdgeRing *OperationOverlayng_MaximalEdgeRing) *OperationOverlayng_OverlayEdge {
	// select if currOut edge is part of this max ring
	if currOut.GetEdgeRingMax() == maxEdgeRing {
		return currOut
	}
	// otherwise skip this edge
	return nil
}

func operationOverlayng_MaximalEdgeRing_linkMaxInEdge(currOut, currMaxRingOut *OperationOverlayng_OverlayEdge, maxEdgeRing *OperationOverlayng_MaximalEdgeRing) *OperationOverlayng_OverlayEdge {
	currIn := currOut.SymOE()
	// currIn is not in this max-edgering, so keep looking
	if currIn.GetEdgeRingMax() != maxEdgeRing {
		return currMaxRingOut
	}

	currIn.SetNextResult(currMaxRingOut)
	// return null to indicate to scan for the next max-ring out-edge
	return nil
}

// String returns a WKT representation of this maximal edge ring.
func (mer *OperationOverlayng_MaximalEdgeRing) String() string {
	pts := mer.getCoordinates()
	return IO_WKTWriter_ToLineStringFromCoords(pts)
}

func (mer *OperationOverlayng_MaximalEdgeRing) getCoordinates() []*Geom_Coordinate {
	coords := Geom_NewCoordinateList()
	edge := mer.startEdge
	for {
		coords.AddCoordinate(edge.Orig(), true)
		if edge.NextResultMax() == nil {
			break
		}
		edge = edge.NextResultMax()
		if edge == mer.startEdge {
			break
		}
	}
	// add last coordinate
	coords.AddCoordinate(edge.Dest(), true)
	return coords.ToCoordinateArray()
}
