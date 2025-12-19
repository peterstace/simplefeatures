package jts

// OperationOverlayng_OverlayLabeller implements the logic to compute the full
// labeling for the edges in an OverlayGraph.
type OperationOverlayng_OverlayLabeller struct {
	graph         *OperationOverlayng_OverlayGraph
	inputGeometry *OperationOverlayng_InputGeometry
	edges         []*OperationOverlayng_OverlayEdge
}

// OperationOverlayng_NewOverlayLabeller creates a new OverlayLabeller.
func OperationOverlayng_NewOverlayLabeller(graph *OperationOverlayng_OverlayGraph, inputGeometry *OperationOverlayng_InputGeometry) *OperationOverlayng_OverlayLabeller {
	return &OperationOverlayng_OverlayLabeller{
		graph:         graph,
		inputGeometry: inputGeometry,
		edges:         graph.GetEdges(),
	}
}

// ComputeLabelling computes the topological labelling for the edges in the
// graph.
func (ol *OperationOverlayng_OverlayLabeller) ComputeLabelling() {
	nodes := ol.graph.GetNodeEdges()

	ol.labelAreaNodeEdges(nodes)
	ol.labelConnectedLinearEdges()

	// At this point collapsed edges labeled with location UNKNOWN must be
	// disconnected from the area edges of the parent (because otherwise the
	// location would have been propagated from them). They can be labeled based
	// on their parent ring role (shell or hole).
	ol.labelCollapsedEdges()
	ol.labelConnectedLinearEdges()

	ol.labelDisconnectedEdges()
}

// labelAreaNodeEdges labels edges around nodes based on the arrangement of
// incident area boundary edges. Also propagates the labeling to connected
// linear edges.
func (ol *OperationOverlayng_OverlayLabeller) labelAreaNodeEdges(nodes []*OperationOverlayng_OverlayEdge) {
	for _, nodeEdge := range nodes {
		ol.PropagateAreaLocations(nodeEdge, 0)
		if ol.inputGeometry.HasEdges(1) {
			ol.PropagateAreaLocations(nodeEdge, 1)
		}
	}
}

// PropagateAreaLocations scans around a node CCW, propagating the side labels
// for a given area geometry to all edges (and their sym) with unknown locations
// for that geometry.
func (ol *OperationOverlayng_OverlayLabeller) PropagateAreaLocations(nodeEdge *OperationOverlayng_OverlayEdge, geomIndex int) {
	// Only propagate for area geometries
	if !ol.inputGeometry.IsArea(geomIndex) {
		return
	}
	// No need to propagate if node has only one edge. This handles dangling
	// edges created by overlap limiting.
	if nodeEdge.Degree() == 1 {
		return
	}

	eStart := operationOverlayng_OverlayLabeller_findPropagationStartEdge(nodeEdge, geomIndex)
	// no labelled edge found, so nothing to propagate
	if eStart == nil {
		return
	}

	// initialize currLoc to location of L side
	currLoc := eStart.GetLocation(geomIndex, Geom_Position_Left)
	e := eStart.ONextOE()

	for e != eStart {
		label := e.GetLabel()
		if !label.IsBoundary(geomIndex) {
			// If this is not a Boundary edge for this input area, its location
			// is now known relative to this input area
			label.SetLocationLine(geomIndex, currLoc)
		} else {
			// must be a boundary edge
			Util_Assert_IsTrueWithMessage(label.HasSides(geomIndex), "")
			// This is a boundary edge for the input area geom. Update the
			// current location from its labels. Also check for topological
			// consistency.
			locRight := e.GetLocation(geomIndex, Geom_Position_Right)
			if locRight != currLoc {
				panic(Geom_NewTopologyExceptionWithCoordinate("side location conflict: arg "+string(rune('0'+geomIndex)), e.GetCoordinate()))
			}
			locLeft := e.GetLocation(geomIndex, Geom_Position_Left)
			if locLeft == Geom_Location_None {
				Util_Assert_ShouldNeverReachHereWithMessage("found single null side at " + e.String())
			}
			currLoc = locLeft
		}
		e = e.ONextOE()
	}
}

// findPropagationStartEdge finds a boundary edge for this geom originating at
// the given node, if one exists.
func operationOverlayng_OverlayLabeller_findPropagationStartEdge(nodeEdge *OperationOverlayng_OverlayEdge, geomIndex int) *OperationOverlayng_OverlayEdge {
	eStart := nodeEdge
	for {
		label := eStart.GetLabel()
		if label.IsBoundary(geomIndex) {
			Util_Assert_IsTrueWithMessage(label.HasSides(geomIndex), "")
			return eStart
		}
		eStart = eStart.ONextOE()
		if eStart == nodeEdge {
			break
		}
	}
	return nil
}

// labelCollapsedEdges labels collapsed edges based on their ring role.
func (ol *OperationOverlayng_OverlayLabeller) labelCollapsedEdges() {
	for _, edge := range ol.edges {
		if edge.GetLabel().IsLineLocationUnknown(0) {
			ol.labelCollapsedEdge(edge, 0)
		}
		if edge.GetLabel().IsLineLocationUnknown(1) {
			ol.labelCollapsedEdge(edge, 1)
		}
	}
}

func (ol *OperationOverlayng_OverlayLabeller) labelCollapsedEdge(edge *OperationOverlayng_OverlayEdge, geomIndex int) {
	label := edge.GetLabel()
	if !label.IsCollapse(geomIndex) {
		return
	}
	// This must be a collapsed edge which is disconnected from any area edges
	// (e.g. a fully collapsed shell or hole). It can be labeled according to
	// its parent source ring role.
	label.SetLocationCollapse(geomIndex)
}

// labelConnectedLinearEdges propagates linear location to connected edges.
func (ol *OperationOverlayng_OverlayLabeller) labelConnectedLinearEdges() {
	ol.propagateLinearLocations(0)
	if ol.inputGeometry.HasEdges(1) {
		ol.propagateLinearLocations(1)
	}
}

// propagateLinearLocations performs a breadth-first graph traversal to find
// and label connected linear edges.
func (ol *OperationOverlayng_OverlayLabeller) propagateLinearLocations(geomIndex int) {
	// find located linear edges
	linearEdges := operationOverlayng_OverlayLabeller_findLinearEdgesWithLocation(ol.edges, geomIndex)
	if len(linearEdges) <= 0 {
		return
	}

	// Use a slice as a deque
	edgeStack := make([]*OperationOverlayng_OverlayEdge, len(linearEdges))
	copy(edgeStack, linearEdges)
	isInputLine := ol.inputGeometry.IsLine(geomIndex)

	// traverse connected linear edges, labeling unknown ones
	for len(edgeStack) > 0 {
		// removeFirst
		lineEdge := edgeStack[0]
		edgeStack = edgeStack[1:]

		// for any edges around origin with unknown location for this geomIndex,
		// add those edges to stack to continue traversal
		operationOverlayng_OverlayLabeller_propagateLinearLocationAtNode(lineEdge, geomIndex, isInputLine, &edgeStack)
	}
}

func operationOverlayng_OverlayLabeller_propagateLinearLocationAtNode(eNode *OperationOverlayng_OverlayEdge, geomIndex int, isInputLine bool, edgeStack *[]*OperationOverlayng_OverlayEdge) {
	lineLoc := eNode.GetLabel().GetLineLocation(geomIndex)
	// If the parent geom is a Line then only propagate EXTERIOR locations.
	if isInputLine && lineLoc != Geom_Location_Exterior {
		return
	}

	e := eNode.ONextOE()
	for e != eNode {
		label := e.GetLabel()
		if label.IsLineLocationUnknown(geomIndex) {
			// If edge is not a boundary edge, its location is now known for
			// this area
			label.SetLocationLine(geomIndex, lineLoc)

			// Add sym edge to stack for graph traversal (Don't add e itself,
			// since e origin node has now been scanned)
			*edgeStack = append([]*OperationOverlayng_OverlayEdge{e.SymOE()}, *edgeStack...)
		}
		e = e.ONextOE()
	}
}

// findLinearEdgesWithLocation finds all OverlayEdges which are linear (i.e.
// line or collapsed) and have a known location for the given input geometry.
func operationOverlayng_OverlayLabeller_findLinearEdgesWithLocation(edges []*OperationOverlayng_OverlayEdge, geomIndex int) []*OperationOverlayng_OverlayEdge {
	linearEdges := make([]*OperationOverlayng_OverlayEdge, 0)
	for _, edge := range edges {
		lbl := edge.GetLabel()
		// keep if linear with known location
		if lbl.IsLinear(geomIndex) && !lbl.IsLineLocationUnknown(geomIndex) {
			linearEdges = append(linearEdges, edge)
		}
	}
	return linearEdges
}

// labelDisconnectedEdges labels edges that are disconnected from any edges of
// the input geometry via point-in-polygon tests.
func (ol *OperationOverlayng_OverlayLabeller) labelDisconnectedEdges() {
	for _, edge := range ol.edges {
		if edge.GetLabel().IsLineLocationUnknown(0) {
			ol.labelDisconnectedEdge(edge, 0)
		}
		if edge.GetLabel().IsLineLocationUnknown(1) {
			ol.labelDisconnectedEdge(edge, 1)
		}
	}
}

// labelDisconnectedEdge determines the location of an edge relative to a
// target input geometry.
func (ol *OperationOverlayng_OverlayLabeller) labelDisconnectedEdge(edge *OperationOverlayng_OverlayEdge, geomIndex int) {
	label := edge.GetLabel()

	// if target geom is not an area then edge must be EXTERIOR, since to be
	// INTERIOR it would have been labelled when it was created.
	if !ol.inputGeometry.IsArea(geomIndex) {
		label.SetLocationAll(geomIndex, Geom_Location_Exterior)
		return
	}

	// Locate edge in input area using a Point-In-Poly check. This should be
	// safe even with precision reduction, because since the edge has remained
	// disconnected its interior-exterior relationship can be determined
	// relative to the original input geometry.
	edgeLoc := ol.locateEdgeBothEnds(geomIndex, edge)
	label.SetLocationAll(geomIndex, edgeLoc)
}

// locateEdge determines the Location for an edge within an Area geometry via
// point-in-polygon location.
func (ol *OperationOverlayng_OverlayLabeller) locateEdge(geomIndex int, edge *OperationOverlayng_OverlayEdge) int {
	loc := ol.inputGeometry.LocatePointInArea(geomIndex, edge.Orig())
	edgeLoc := Geom_Location_Interior
	if loc == Geom_Location_Exterior {
		edgeLoc = Geom_Location_Exterior
	}
	return edgeLoc
}

// locateEdgeBothEnds determines the Location for an edge within an Area
// geometry via point-in-polygon location, by checking that both endpoints are
// interior to the target geometry.
func (ol *OperationOverlayng_OverlayLabeller) locateEdgeBothEnds(geomIndex int, edge *OperationOverlayng_OverlayEdge) int {
	// To improve the robustness of the point location, check both ends of the
	// edge. Edge is only labelled INTERIOR if both ends are.
	locOrig := ol.inputGeometry.LocatePointInArea(geomIndex, edge.Orig())
	locDest := ol.inputGeometry.LocatePointInArea(geomIndex, edge.Dest())
	isInt := locOrig != Geom_Location_Exterior && locDest != Geom_Location_Exterior
	if isInt {
		return Geom_Location_Interior
	}
	return Geom_Location_Exterior
}

// MarkResultAreaEdges marks edges which form part of the boundary of the result
// area.
func (ol *OperationOverlayng_OverlayLabeller) MarkResultAreaEdges(overlayOpCode int) {
	for _, edge := range ol.edges {
		ol.MarkInResultArea(edge, overlayOpCode)
	}
}

// MarkInResultArea marks an edge which forms part of the boundary of the result
// area.
func (ol *OperationOverlayng_OverlayLabeller) MarkInResultArea(e *OperationOverlayng_OverlayEdge, overlayOpCode int) {
	label := e.GetLabel()
	if label.IsBoundaryEither() &&
		OperationOverlayng_OverlayNG_IsResultOfOp(
			overlayOpCode,
			label.GetLocationBoundaryOrLine(0, Geom_Position_Right, e.IsForward()),
			label.GetLocationBoundaryOrLine(1, Geom_Position_Right, e.IsForward())) {
		e.MarkInResultArea()
	}
}

// UnmarkDuplicateEdgesFromResultArea unmarks result area edges where the sym
// edge is also marked as in the result.
func (ol *OperationOverlayng_OverlayLabeller) UnmarkDuplicateEdgesFromResultArea() {
	for _, edge := range ol.edges {
		if edge.IsInResultAreaBoth() {
			edge.UnmarkFromResultAreaBoth()
		}
	}
}

// OperationOverlayng_OverlayLabeller_ToString returns a string representation
// of the edges around a node.
func OperationOverlayng_OverlayLabeller_ToString(nodeEdge *OperationOverlayng_OverlayEdge) string {
	orig := nodeEdge.Orig()
	sb := "Node( " + IO_WKTWriter_Format(orig) + " )\n"
	e := nodeEdge
	for {
		sb += "  -> " + e.String()
		if e.IsResultLinked() {
			sb += " Link: " + e.NextResult().String()
		}
		sb += "\n"
		e = e.ONextOE()
		if e == nodeEdge {
			break
		}
	}
	return sb
}
