package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationOverlay_LineBuilder forms LineStrings out of a the graph of
// DirectedEdges created by an OverlayOp.
type OperationOverlay_LineBuilder struct {
	child java.Polymorphic

	op              *OperationOverlay_OverlayOp
	geometryFactory *Geom_GeometryFactory
	ptLocator       *Algorithm_PointLocator

	lineEdgesList  []*Geomgraph_Edge
	resultLineList []*Geom_LineString
}

// GetChild returns the immediate child in the type hierarchy chain.
func (lb *OperationOverlay_LineBuilder) GetChild() java.Polymorphic {
	return lb.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (lb *OperationOverlay_LineBuilder) GetParent() java.Polymorphic {
	return nil
}

// OperationOverlay_NewLineBuilder creates a new LineBuilder.
func OperationOverlay_NewLineBuilder(op *OperationOverlay_OverlayOp, geometryFactory *Geom_GeometryFactory, ptLocator *Algorithm_PointLocator) *OperationOverlay_LineBuilder {
	return &OperationOverlay_LineBuilder{
		op:              op,
		geometryFactory: geometryFactory,
		ptLocator:       ptLocator,
	}
}

// Build returns a list of the LineStrings in the result of the specified
// overlay operation.
func (lb *OperationOverlay_LineBuilder) Build(opCode int) []*Geom_LineString {
	lb.findCoveredLineEdges()
	lb.collectLines(opCode)
	lb.buildLines(opCode)
	return lb.resultLineList
}

// findCoveredLineEdges finds and marks L edges which are "covered" by the
// result area (if any). L edges at nodes which also have A edges can be checked
// by checking their depth at that node. L edges at nodes which do not have A
// edges can be checked by doing a point-in-polygon test with the previously
// computed result areas.
func (lb *OperationOverlay_LineBuilder) findCoveredLineEdges() {
	// First set covered for all L edges at nodes which have A edges too.
	for _, node := range lb.op.GetGraph().GetNodes() {
		des := java.GetLeaf(node.GetEdges()).(*Geomgraph_DirectedEdgeStar)
		des.FindCoveredLineEdges()
	}

	// For all L edges which weren't handled by the above, use a point-in-poly
	// test to determine whether they are covered.
	for _, ee := range lb.op.GetGraph().GetEdgeEnds() {
		de := java.Cast[*Geomgraph_DirectedEdge](ee)
		e := de.GetEdge()
		if de.IsLineEdge() && !e.IsCoveredSet() {
			isCovered := lb.op.IsCoveredByA(de.GetCoordinate())
			e.SetCovered(isCovered)
		}
	}
}

func (lb *OperationOverlay_LineBuilder) collectLines(opCode int) {
	for _, ee := range lb.op.GetGraph().GetEdgeEnds() {
		de := java.Cast[*Geomgraph_DirectedEdge](ee)
		lb.collectLineEdge(de, opCode)
		lb.collectBoundaryTouchEdge(de, opCode)
	}
}

// collectLineEdge collects line edges which are in the result. Line edges are
// in the result if they are not part of an area boundary, if they are in the
// result of the overlay operation, and if they are not covered by a result
// area.
func (lb *OperationOverlay_LineBuilder) collectLineEdge(de *Geomgraph_DirectedEdge, opCode int) {
	label := de.GetLabel()
	e := de.GetEdge()
	// Include L edges which are in the result.
	if de.IsLineEdge() {
		if !de.IsVisited() && OperationOverlay_OverlayOp_IsResultOfOpLabel(label, opCode) && !e.IsCovered() {
			lb.lineEdgesList = append(lb.lineEdgesList, e)
			de.SetVisitedEdge(true)
		}
	}
}

// collectBoundaryTouchEdge collects edges from Area inputs which should be in
// the result but which have not been included in a result area. This happens
// ONLY:
//   - during an intersection when the boundaries of two areas touch in a line
//     segment
//   - OR as a result of a dimensional collapse.
func (lb *OperationOverlay_LineBuilder) collectBoundaryTouchEdge(de *Geomgraph_DirectedEdge, opCode int) {
	label := de.GetLabel()
	if de.IsLineEdge() {
		return // Only interested in area edges.
	}
	if de.IsVisited() {
		return // Already processed.
	}
	if de.IsInteriorAreaEdge() {
		return // Added to handle dimensional collapses.
	}
	if de.GetEdge().IsInResult() {
		return // If the edge linework is already included, don't include it again.
	}

	// Sanity check for labelling of result edgerings.
	Util_Assert_IsTrueWithMessage(!(de.IsInResult() || de.GetSym().IsInResult()) || !de.GetEdge().IsInResult(), "")

	// Include the linework if it's in the result of the operation.
	if OperationOverlay_OverlayOp_IsResultOfOpLabel(label, opCode) && opCode == OperationOverlay_OverlayOp_Intersection {
		lb.lineEdgesList = append(lb.lineEdgesList, de.GetEdge())
		de.SetVisitedEdge(true)
	}
}

func (lb *OperationOverlay_LineBuilder) buildLines(opCode int) {
	for _, e := range lb.lineEdgesList {
		line := lb.geometryFactory.CreateLineStringFromCoordinates(e.GetCoordinates())
		lb.resultLineList = append(lb.resultLineList, line)
		e.SetInResult(true)
	}
}

func (lb *OperationOverlay_LineBuilder) labelIsolatedLines(edgesList []*Geomgraph_Edge) {
	for _, e := range edgesList {
		label := e.GetLabel()
		if e.IsIsolated() {
			if label.IsNull(0) {
				lb.labelIsolatedLine(e, 0)
			} else {
				lb.labelIsolatedLine(e, 1)
			}
		}
	}
}

// labelIsolatedLine labels an isolated node with its relationship to the
// target geometry.
func (lb *OperationOverlay_LineBuilder) labelIsolatedLine(e *Geomgraph_Edge, targetIndex int) {
	loc := lb.ptLocator.Locate(e.GetCoordinate(), lb.op.GetArgGeometry(targetIndex))
	e.GetLabel().SetLocationOn(targetIndex, loc)
}
