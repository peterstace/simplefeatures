package jts

// OperationOverlayng_LineBuilder finds and builds overlay result lines from the
// overlay graph. Output linework has the following semantics:
//
//  1. Linework is fully noded
//  2. Nodes in the input are preserved in the output
//  3. Output may contain more nodes than in the input (in particular, sequences
//     of coincident line segments are noded at each vertex)
type OperationOverlayng_LineBuilder struct {
	geometryFactory      *Geom_GeometryFactory
	graph                *OperationOverlayng_OverlayGraph
	opCode               int
	inputAreaIndex       int
	hasResultArea        bool
	isAllowMixedResult   bool
	isAllowCollapseLines bool
	lines                []*Geom_LineString
}

// OperationOverlayng_NewLineBuilder creates a builder for linear elements which
// may be present in the overlay result.
func OperationOverlayng_NewLineBuilder(inputGeom *OperationOverlayng_InputGeometry, graph *OperationOverlayng_OverlayGraph, hasResultArea bool, opCode int, geomFact *Geom_GeometryFactory) *OperationOverlayng_LineBuilder {
	return &OperationOverlayng_LineBuilder{
		geometryFactory:      geomFact,
		graph:                graph,
		opCode:               opCode,
		hasResultArea:        hasResultArea,
		inputAreaIndex:       inputGeom.GetAreaIndex(),
		isAllowMixedResult:   !OperationOverlayng_OverlayNG_STRICT_MODE_DEFAULT,
		isAllowCollapseLines: !OperationOverlayng_OverlayNG_STRICT_MODE_DEFAULT,
		lines:                make([]*Geom_LineString, 0),
	}
}

// SetStrictMode sets strict mode for the line builder.
func (lb *OperationOverlayng_LineBuilder) SetStrictMode(isStrictResultMode bool) {
	lb.isAllowCollapseLines = !isStrictResultMode
	lb.isAllowMixedResult = !isStrictResultMode
}

// GetLines returns the result lines from the overlay graph.
func (lb *OperationOverlayng_LineBuilder) GetLines() []*Geom_LineString {
	lb.markResultLines()
	lb.addResultLines()
	return lb.lines
}

func (lb *OperationOverlayng_LineBuilder) markResultLines() {
	edges := lb.graph.GetEdges()
	for _, edge := range edges {
		// If the edge linework is already marked as in the result, it is not
		// included as a line. This occurs when an edge either is in a result
		// area or has already been included as a line.
		if edge.IsInResultEither() {
			continue
		}
		if lb.isResultLine(edge.GetLabel()) {
			edge.MarkInResultLine()
		}
	}
}

// isResultLine checks if the topology indicated by an edge label determines
// that this edge should be part of a result line.
func (lb *OperationOverlayng_LineBuilder) isResultLine(lbl *OperationOverlayng_OverlayLabel) bool {
	// Omit edge which is a boundary of a single geometry (i.e. not a collapse
	// or line edge as well). These are only included if part of a result area.
	// This is a short-circuit for the most common area edge case.
	if lbl.IsBoundarySingleton() {
		return false
	}

	// Omit edge which is a collapse along a boundary. I.e. a result line edge
	// must be from an input line OR two coincident area boundaries.
	// This logic is only used if not including collapse lines in result.
	if !lb.isAllowCollapseLines && lbl.IsBoundaryCollapse() {
		return false
	}

	// Omit edge which is a collapse interior to its parent area.
	// (E.g. a narrow gore, or spike off a hole)
	if lbl.IsInteriorCollapse() {
		return false
	}

	// For ops other than Intersection, omit a line edge if it is interior to
	// the other area. For Intersection, a line edge interior to an area is
	// included.
	if lb.opCode != OperationOverlayng_OverlayNG_INTERSECTION {
		// Omit collapsed edge in other area interior.
		if lbl.IsCollapseAndNotPartInterior() {
			return false
		}

		// If there is a result area, omit line edge inside it. It is sufficient
		// to check against the input area rather than the result area, because
		// if line edges are present then there is only one input area, and the
		// result area must be the same as the input area.
		if lb.hasResultArea && lbl.IsLineInArea(lb.inputAreaIndex) {
			return false
		}
	}

	// Include line edge formed by touching area boundaries, if enabled.
	if lb.isAllowMixedResult && lb.opCode == OperationOverlayng_OverlayNG_INTERSECTION && lbl.IsBoundaryTouch() {
		return true
	}

	// Finally, determine included line edge according to overlay op boolean logic.
	aLoc := operationOverlayng_LineBuilder_effectiveLocation(lbl, 0)
	bLoc := operationOverlayng_LineBuilder_effectiveLocation(lbl, 1)
	isInResult := OperationOverlayng_OverlayNG_IsResultOfOp(lb.opCode, aLoc, bLoc)
	return isInResult
}

// effectiveLocation determines the effective location for a line, for the
// purpose of overlay operation evaluation. Line edges and Collapses are
// reported as INTERIOR so they may be included in the result if warranted by
// the effect of the operation on the two edges.
func operationOverlayng_LineBuilder_effectiveLocation(lbl *OperationOverlayng_OverlayLabel, geomIndex int) int {
	if lbl.IsCollapse(geomIndex) {
		return Geom_Location_Interior
	}
	if lbl.IsLineIndex(geomIndex) {
		return Geom_Location_Interior
	}
	return lbl.GetLineLocation(geomIndex)
}

func (lb *OperationOverlayng_LineBuilder) addResultLines() {
	edges := lb.graph.GetEdges()
	for _, edge := range edges {
		if !edge.IsInResultLine() {
			continue
		}
		if edge.IsVisited() {
			continue
		}
		lb.lines = append(lb.lines, lb.toLine(edge))
		edge.MarkVisitedBoth()
	}
}

func (lb *OperationOverlayng_LineBuilder) toLine(edge *OperationOverlayng_OverlayEdge) *Geom_LineString {
	isForward := edge.IsForward()
	pts := Geom_NewCoordinateList()
	pts.AddCoordinate(edge.Orig(), false)
	edge.AddCoordinates(pts)

	ptsOut := pts.ToCoordinateArrayWithDirection(isForward)
	line := lb.geometryFactory.CreateLineStringFromCoordinates(ptsOut)
	return line
}

// NOTE: The following methods are for maximal line extraction logic, which is
// NOT USED currently. Instead the raw noded edges are output. This matches the
// original overlay semantics and is also faster.

func (lb *OperationOverlayng_LineBuilder) addResultLinesMerged() {
	lb.addResultLinesForNodes()
	lb.addResultLinesRings()
}

func (lb *OperationOverlayng_LineBuilder) addResultLinesForNodes() {
	edges := lb.graph.GetEdges()
	for _, edge := range edges {
		if !edge.IsInResultLine() {
			continue
		}
		if edge.IsVisited() {
			continue
		}
		// Choose line start point as a node. Nodes in the line graph are
		// degree-1 or degree >= 3 edges. This will find all lines originating
		// at nodes.
		if operationOverlayng_LineBuilder_degreeOfLines(edge) != 2 {
			lb.lines = append(lb.lines, lb.buildLine(edge))
		}
	}
}

// addResultLinesRings adds lines which form rings (i.e. have only degree-2
// vertices).
func (lb *OperationOverlayng_LineBuilder) addResultLinesRings() {
	edges := lb.graph.GetEdges()
	for _, edge := range edges {
		if !edge.IsInResultLine() {
			continue
		}
		if edge.IsVisited() {
			continue
		}
		lb.lines = append(lb.lines, lb.buildLine(edge))
	}
}

// buildLine traverses edges from edgeStart which lie in a single line (have
// degree = 2).
func (lb *OperationOverlayng_LineBuilder) buildLine(node *OperationOverlayng_OverlayEdge) *Geom_LineString {
	pts := Geom_NewCoordinateList()
	pts.AddCoordinate(node.Orig(), false)

	isForward := node.IsForward()

	e := node
	for {
		e.MarkVisitedBoth()
		e.AddCoordinates(pts)

		// End line if next vertex is a node.
		if operationOverlayng_LineBuilder_degreeOfLines(e.SymOE()) != 2 {
			break
		}
		e = operationOverlayng_LineBuilder_nextLineEdgeUnvisited(e.SymOE())
		// e will be nil if next edge has been visited, which indicates a ring.
		if e == nil {
			break
		}
	}

	ptsOut := pts.ToCoordinateArrayWithDirection(isForward)
	line := lb.geometryFactory.CreateLineStringFromCoordinates(ptsOut)
	return line
}

// nextLineEdgeUnvisited finds the next edge around a node which forms part of
// a result line.
func operationOverlayng_LineBuilder_nextLineEdgeUnvisited(node *OperationOverlayng_OverlayEdge) *OperationOverlayng_OverlayEdge {
	e := node
	for {
		e = e.ONextOE()
		if e.IsVisited() {
			continue
		}
		if e.IsInResultLine() {
			return e
		}
		if e == node {
			break
		}
	}
	return nil
}

// degreeOfLines computes the degree of the line edges incident on a node.
func operationOverlayng_LineBuilder_degreeOfLines(node *OperationOverlayng_OverlayEdge) int {
	degree := 0
	e := node
	for {
		if e.IsInResultLine() {
			degree++
		}
		e = e.ONextOE()
		if e == node {
			break
		}
	}
	return degree
}
