package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// The spatial functions supported by OverlayOp. These operations implement
// various boolean combinations of the resultants of the overlay.
const (
	// OperationOverlay_OverlayOp_Intersection is the code for the Intersection
	// overlay operation.
	OperationOverlay_OverlayOp_Intersection = 1

	// OperationOverlay_OverlayOp_Union is the code for the Union overlay
	// operation.
	OperationOverlay_OverlayOp_Union = 2

	// OperationOverlay_OverlayOp_Difference is the code for the Difference
	// overlay operation.
	OperationOverlay_OverlayOp_Difference = 3

	// OperationOverlay_OverlayOp_SymDifference is the code for the Symmetric
	// Difference overlay operation.
	OperationOverlay_OverlayOp_SymDifference = 4
)

// OperationOverlay_OverlayOp computes the geometric overlay of two Geometries.
// The overlay can be used to determine any boolean combination of the
// geometries.
type OperationOverlay_OverlayOp struct {
	*Operation_GeometryGraphOperation
	child java.Polymorphic

	ptLocator  *Algorithm_PointLocator
	geomFact   *Geom_GeometryFactory
	resultGeom *Geom_Geometry

	graph    *Geomgraph_PlanarGraph
	edgeList *Geomgraph_EdgeList

	resultPolyList  []*Geom_Polygon
	resultLineList  []*Geom_LineString
	resultPointList []*Geom_Point
}

// GetChild returns the immediate child in the type hierarchy chain.
func (oo *OperationOverlay_OverlayOp) GetChild() java.Polymorphic {
	return oo.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (oo *OperationOverlay_OverlayOp) GetParent() java.Polymorphic {
	return oo.Operation_GeometryGraphOperation
}

// OperationOverlay_OverlayOp_OverlayOp computes an overlay operation for the
// given geometry arguments.
func OperationOverlay_OverlayOp_OverlayOp(geom0, geom1 *Geom_Geometry, opCode int) *Geom_Geometry {
	gov := OperationOverlay_NewOverlayOp(geom0, geom1)
	geomOv := gov.GetResultGeometry(opCode)
	return geomOv
}

// OperationOverlay_OverlayOp_IsResultOfOpLabel tests whether a point with a
// given topological Label relative to two geometries is contained in the result
// of overlaying the geometries using a given overlay operation. The method
// handles arguments of Location.NONE correctly.
func OperationOverlay_OverlayOp_IsResultOfOpLabel(label *Geomgraph_Label, opCode int) bool {
	loc0 := label.GetLocationOn(0)
	loc1 := label.GetLocationOn(1)
	return OperationOverlay_OverlayOp_IsResultOfOp(loc0, loc1, opCode)
}

// OperationOverlay_OverlayOp_IsResultOfOp tests whether a point with given
// Locations relative to two geometries is contained in the result of overlaying
// the geometries using a given overlay operation. The method handles arguments
// of Location.NONE correctly.
func OperationOverlay_OverlayOp_IsResultOfOp(loc0, loc1, overlayOpCode int) bool {
	if loc0 == Geom_Location_Boundary {
		loc0 = Geom_Location_Interior
	}
	if loc1 == Geom_Location_Boundary {
		loc1 = Geom_Location_Interior
	}
	switch overlayOpCode {
	case OperationOverlay_OverlayOp_Intersection:
		return loc0 == Geom_Location_Interior && loc1 == Geom_Location_Interior
	case OperationOverlay_OverlayOp_Union:
		return loc0 == Geom_Location_Interior || loc1 == Geom_Location_Interior
	case OperationOverlay_OverlayOp_Difference:
		return loc0 == Geom_Location_Interior && loc1 != Geom_Location_Interior
	case OperationOverlay_OverlayOp_SymDifference:
		return (loc0 == Geom_Location_Interior && loc1 != Geom_Location_Interior) ||
			(loc0 != Geom_Location_Interior && loc1 == Geom_Location_Interior)
	}
	return false
}

// OperationOverlay_NewOverlayOp constructs an instance to compute a single
// overlay operation for the given geometries.
func OperationOverlay_NewOverlayOp(g0, g1 *Geom_Geometry) *OperationOverlay_OverlayOp {
	ggo := Operation_NewGeometryGraphOperation(g0, g1)
	oo := &OperationOverlay_OverlayOp{
		Operation_GeometryGraphOperation: ggo,
		ptLocator:                       Algorithm_NewPointLocator(),
		graph:                           Geomgraph_NewPlanarGraph(OperationOverlay_NewOverlayNodeFactory().Geomgraph_NodeFactory),
		edgeList:                        Geomgraph_NewEdgeList(),
		// Use factory of primary geometry. Note that this does NOT handle
		// mixed-precision arguments where the second arg has greater precision
		// than the first.
		geomFact: g0.GetFactory(),
	}
	ggo.child = oo
	return oo
}

// GetResultGeometry gets the result of the overlay for a given overlay
// operation. Note: this method can be called once only.
func (oo *OperationOverlay_OverlayOp) GetResultGeometry(overlayOpCode int) *Geom_Geometry {
	oo.computeOverlay(overlayOpCode)
	return oo.resultGeom
}

// GetGraph gets the graph constructed to compute the overlay.
func (oo *OperationOverlay_OverlayOp) GetGraph() *Geomgraph_PlanarGraph {
	return oo.graph
}

func (oo *OperationOverlay_OverlayOp) computeOverlay(opCode int) {
	// Copy points from input Geometries. This ensures that any Point geometries
	// in the input are considered for inclusion in the result set.
	oo.copyPoints(0)
	oo.copyPoints(1)

	// Node the input Geometries.
	oo.arg[0].ComputeSelfNodes(oo.li, false)
	oo.arg[1].ComputeSelfNodes(oo.li, false)

	// Compute intersections between edges of the two input geometries.
	oo.arg[0].ComputeEdgeIntersections(oo.arg[1], oo.li, true)

	var baseSplitEdges []*Geomgraph_Edge
	oo.arg[0].ComputeSplitEdges(&baseSplitEdges)
	oo.arg[1].ComputeSplitEdges(&baseSplitEdges)
	// Add the noded edges to this result graph.
	oo.insertUniqueEdges(baseSplitEdges)

	oo.computeLabelsFromDepths()
	oo.replaceCollapsedEdges()

	// Check that the noding completed correctly. This test is slow, but
	// necessary in order to catch robustness failure situations.
	Geomgraph_EdgeNodingValidator_CheckValid(oo.edgeList.GetEdges())

	oo.graph.AddEdges(oo.edgeList.GetEdges())
	oo.computeLabelling()
	oo.labelIncompleteNodes()

	// The ordering of building the result Geometries is important. Areas must
	// be built before lines, which must be built before points. This is so that
	// lines which are covered by areas are not included explicitly, and
	// similarly for points.
	oo.findResultAreaEdges(opCode)
	oo.cancelDuplicateResultEdges()

	polyBuilder := OperationOverlay_NewPolygonBuilder(oo.geomFact)
	polyBuilder.AddFromGraph(oo.graph)
	oo.resultPolyList = polyBuilder.GetPolygons()

	lineBuilder := OperationOverlay_NewLineBuilder(oo, oo.geomFact, oo.ptLocator)
	oo.resultLineList = lineBuilder.Build(opCode)

	pointBuilder := OperationOverlay_NewPointBuilder(oo, oo.geomFact, oo.ptLocator)
	oo.resultPointList = pointBuilder.Build(opCode)

	// Gather the results from all calculations into a single Geometry for the
	// result set.
	oo.resultGeom = oo.computeGeometry(oo.resultPointList, oo.resultLineList, oo.resultPolyList, opCode)
}

func (oo *OperationOverlay_OverlayOp) insertUniqueEdges(edges []*Geomgraph_Edge) {
	for _, e := range edges {
		oo.insertUniqueEdge(e)
	}
}

// insertUniqueEdge inserts an edge from one of the noded input graphs. Checks
// edges that are inserted to see if an identical edge already exists. If so,
// the edge is not inserted, but its label is merged with the existing edge.
func (oo *OperationOverlay_OverlayOp) insertUniqueEdge(e *Geomgraph_Edge) {
	existingEdge := oo.edgeList.FindEqualEdge(e)

	// If an identical edge already exists, simply update its label.
	if existingEdge != nil {
		existingLabel := existingEdge.GetLabel()

		labelToMerge := e.GetLabel()
		// Check if new edge is in reverse direction to existing edge. If so,
		// must flip the label before merging it.
		if !existingEdge.IsPointwiseEqual(e) {
			labelToMerge = Geomgraph_NewLabelFromLabel(e.GetLabel())
			labelToMerge.Flip()
		}
		depth := existingEdge.GetDepth()
		// If this is the first duplicate found for this edge, initialize the
		// depths.
		if depth.IsNull() {
			depth.AddLabel(existingLabel)
		}
		depth.AddLabel(labelToMerge)
		existingLabel.Merge(labelToMerge)
	} else {
		// No matching existing edge was found. Add this new edge to the list of
		// edges in this graph.
		oo.edgeList.Add(e)
	}
}

// computeLabelsFromDepths updates the labels for edges according to their
// depths. For each edge, the depths are first normalized. Then, if the depths
// for the edge are equal, this edge must have collapsed into a line edge. If
// the depths are not equal, update the label with the locations corresponding
// to the depths (i.e. a depth of 0 corresponds to a Location of EXTERIOR, a
// depth of 1 corresponds to INTERIOR).
func (oo *OperationOverlay_OverlayOp) computeLabelsFromDepths() {
	for _, e := range oo.edgeList.GetEdges() {
		lbl := e.GetLabel()
		depth := e.GetDepth()
		// Only check edges for which there were duplicates, since these are the
		// only ones which might be the result of dimensional collapses.
		if !depth.IsNull() {
			depth.Normalize()
			for i := 0; i < 2; i++ {
				if !lbl.IsNull(i) && lbl.IsArea() && !depth.IsNullAt(i) {
					// If the depths are equal, this edge is the result of the
					// dimensional collapse of two or more edges. It has the
					// same location on both sides of the edge, so it has
					// collapsed to a line.
					if depth.GetDelta(i) == 0 {
						lbl.ToLine(i)
					} else {
						// This edge may be the result of a dimensional
						// collapse, but it still has different locations on
						// both sides. The label of the edge must be updated to
						// reflect the resultant side locations indicated by the
						// depth values.
						Util_Assert_IsTrueWithMessage(!depth.IsNullAtPos(i, Geom_Position_Left), "depth of LEFT side has not been initialized")
						lbl.SetLocation(i, Geom_Position_Left, depth.GetLocation(i, Geom_Position_Left))
						Util_Assert_IsTrueWithMessage(!depth.IsNullAtPos(i, Geom_Position_Right), "depth of RIGHT side has not been initialized")
						lbl.SetLocation(i, Geom_Position_Right, depth.GetLocation(i, Geom_Position_Right))
					}
				}
			}
		}
	}
}

// replaceCollapsedEdges replaces edges which have undergone dimensional
// collapse with a new edge which is a L edge.
func (oo *OperationOverlay_OverlayOp) replaceCollapsedEdges() {
	var newEdges []*Geomgraph_Edge
	var keptEdges []*Geomgraph_Edge
	for _, e := range oo.edgeList.GetEdges() {
		if e.IsCollapsed() {
			newEdges = append(newEdges, e.GetCollapsedEdge())
		} else {
			keptEdges = append(keptEdges, e)
		}
	}
	oo.edgeList.Clear()
	oo.edgeList.AddAll(keptEdges)
	oo.edgeList.AddAll(newEdges)
}

// copyPoints copies all nodes from an arg geometry into this graph. The node
// label in the arg geometry overrides any previously computed label for that
// argIndex.
func (oo *OperationOverlay_OverlayOp) copyPoints(argIndex int) {
	for _, graphNode := range oo.arg[argIndex].GetNodeIterator() {
		newNode := oo.graph.AddNodeFromCoord(graphNode.GetCoordinate())
		newNode.SetLabelAt(argIndex, graphNode.GetLabel().GetLocationOn(argIndex))
	}
}

// computeLabelling computes initial labelling for all DirectedEdges at each
// node. In this step, DirectedEdges will acquire a complete labelling (i.e. one
// with labels for both Geometries) only if they are incident on a node which
// has edges for both Geometries.
func (oo *OperationOverlay_OverlayOp) computeLabelling() {
	for _, node := range oo.graph.GetNodes() {
		node.GetEdges().ComputeLabelling(oo.arg)
	}
	oo.mergeSymLabels()
	oo.updateNodeLabelling()
}

// mergeSymLabels merges labels for nodes which have edges from only one
// Geometry. For these nodes, the previous step will have left their dirEdges
// with no labelling for the other Geometry. However, the sym dirEdge may have a
// labelling for the other Geometry, so merge the two labels.
func (oo *OperationOverlay_OverlayOp) mergeSymLabels() {
	for _, node := range oo.graph.GetNodes() {
		des := java.GetLeaf(node.GetEdges()).(*Geomgraph_DirectedEdgeStar)
		des.MergeSymLabels()
	}
}

func (oo *OperationOverlay_OverlayOp) updateNodeLabelling() {
	// Update the labels for nodes. The label for a node is updated from the
	// edges incident on it.
	for _, node := range oo.graph.GetNodes() {
		des := java.GetLeaf(node.GetEdges()).(*Geomgraph_DirectedEdgeStar)
		lbl := des.GetLabel()
		node.GetLabel().Merge(lbl)
	}
}

// labelIncompleteNodes labels nodes whose labels are incomplete. Isolated
// nodes are found because nodes in one graph which don't intersect nodes in the
// other are not completely labelled by the initial process of adding nodes to
// the nodeList.
func (oo *OperationOverlay_OverlayOp) labelIncompleteNodes() {
	for _, n := range oo.graph.GetNodes() {
		label := n.GetLabel()
		if n.IsIsolated() {
			if label.IsNull(0) {
				oo.labelIncompleteNode(n, 0)
			} else {
				oo.labelIncompleteNode(n, 1)
			}
		}
		// Now update the labelling for the DirectedEdges incident on this node.
		des := java.GetLeaf(n.GetEdges()).(*Geomgraph_DirectedEdgeStar)
		des.UpdateLabelling(label)
	}
}

// labelIncompleteNode labels an isolated node with its relationship to the
// target geometry.
func (oo *OperationOverlay_OverlayOp) labelIncompleteNode(n *Geomgraph_Node, targetIndex int) {
	loc := oo.ptLocator.Locate(n.GetCoordinate(), oo.arg[targetIndex].GetGeometry())
	n.GetLabel().SetLocationOn(targetIndex, loc)
}

// findResultAreaEdges finds all edges whose label indicates that they are in
// the result area(s), according to the operation being performed. Since we want
// polygon shells to be oriented CW, choose dirEdges with the interior of the
// result on the RHS. Mark them as being in the result. Interior Area edges are
// the result of dimensional collapses. They do not form part of the result area
// boundary.
func (oo *OperationOverlay_OverlayOp) findResultAreaEdges(opCode int) {
	for _, ee := range oo.graph.GetEdgeEnds() {
		de := java.Cast[*Geomgraph_DirectedEdge](ee)
		// Mark all dirEdges with the appropriate label.
		label := de.GetLabel()
		if label.IsArea() && !de.IsInteriorAreaEdge() &&
			OperationOverlay_OverlayOp_IsResultOfOp(
				label.GetLocation(0, Geom_Position_Right),
				label.GetLocation(1, Geom_Position_Right),
				opCode) {
			de.SetInResult(true)
		}
	}
}

// cancelDuplicateResultEdges cancels out dirEdges whose sym is also marked as
// being in the result.
func (oo *OperationOverlay_OverlayOp) cancelDuplicateResultEdges() {
	for _, ee := range oo.graph.GetEdgeEnds() {
		de := java.Cast[*Geomgraph_DirectedEdge](ee)
		sym := de.GetSym()
		if de.IsInResult() && sym.IsInResult() {
			de.SetInResult(false)
			sym.SetInResult(false)
		}
	}
}

// IsCoveredByLA tests if a point node should be included in the result or not.
func (oo *OperationOverlay_OverlayOp) IsCoveredByLA(coord *Geom_Coordinate) bool {
	if oo.isCoveredByLines(coord) {
		return true
	}
	if oo.isCoveredByPolys(coord) {
		return true
	}
	return false
}

// IsCoveredByA tests if an L edge should be included in the result or not.
func (oo *OperationOverlay_OverlayOp) IsCoveredByA(coord *Geom_Coordinate) bool {
	return oo.isCoveredByPolys(coord)
}

func (oo *OperationOverlay_OverlayOp) isCoveredByLines(coord *Geom_Coordinate) bool {
	for _, line := range oo.resultLineList {
		loc := oo.ptLocator.Locate(coord, line.Geom_Geometry)
		if loc != Geom_Location_Exterior {
			return true
		}
	}
	return false
}

func (oo *OperationOverlay_OverlayOp) isCoveredByPolys(coord *Geom_Coordinate) bool {
	for _, poly := range oo.resultPolyList {
		loc := oo.ptLocator.Locate(coord, poly.Geom_Geometry)
		if loc != Geom_Location_Exterior {
			return true
		}
	}
	return false
}

func (oo *OperationOverlay_OverlayOp) computeGeometry(resultPointList []*Geom_Point, resultLineList []*Geom_LineString, resultPolyList []*Geom_Polygon, opcode int) *Geom_Geometry {
	var geomList []*Geom_Geometry
	// Element geometries of the result are always in the order P,L,A.
	for _, p := range resultPointList {
		geomList = append(geomList, p.Geom_Geometry)
	}
	for _, l := range resultLineList {
		geomList = append(geomList, l.Geom_Geometry)
	}
	for _, a := range resultPolyList {
		geomList = append(geomList, a.Geom_Geometry)
	}

	if len(geomList) == 0 {
		return OperationOverlay_OverlayOp_CreateEmptyResult(opcode, oo.arg[0].GetGeometry(), oo.arg[1].GetGeometry(), oo.geomFact)
	}

	// Build the most specific geometry possible.
	return oo.geomFact.BuildGeometry(geomList)
}

// OperationOverlay_OverlayOp_CreateEmptyResult creates an empty result geometry
// of the appropriate dimension, based on the given overlay operation and the
// dimensions of the inputs.
func OperationOverlay_OverlayOp_CreateEmptyResult(overlayOpCode int, a, b *Geom_Geometry, geomFact *Geom_GeometryFactory) *Geom_Geometry {
	resultDim := OperationOverlay_OverlayOp_ResultDimension(overlayOpCode, a, b)
	return geomFact.CreateEmpty(resultDim)
}

// OperationOverlay_OverlayOp_ResultDimension returns the dimension of the
// result for the given overlay operation.
func OperationOverlay_OverlayOp_ResultDimension(opCode int, g0, g1 *Geom_Geometry) int {
	dim0 := g0.GetDimension()
	dim1 := g1.GetDimension()

	resultDimension := -1
	switch opCode {
	case OperationOverlay_OverlayOp_Intersection:
		if dim0 < dim1 {
			resultDimension = dim0
		} else {
			resultDimension = dim1
		}
	case OperationOverlay_OverlayOp_Union:
		if dim0 > dim1 {
			resultDimension = dim0
		} else {
			resultDimension = dim1
		}
	case OperationOverlay_OverlayOp_Difference:
		resultDimension = dim0
	case OperationOverlay_OverlayOp_SymDifference:
		// This result is chosen because SymDiff = Union(Diff(A, B), Diff(B, A))
		// and Union has the dimension of the highest-dimension argument.
		if dim0 > dim1 {
			resultDimension = dim0
		} else {
			resultDimension = dim1
		}
	}
	return resultDimension
}
