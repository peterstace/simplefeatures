package jts

// OperationRelateng_TopologyComputer manages the computation of topology
// relationships between two geometries in RelateNG.
type OperationRelateng_TopologyComputer struct {
	predicate OperationRelateng_TopologyPredicate
	geomA     *OperationRelateng_RelateGeometry
	geomB     *OperationRelateng_RelateGeometry
	nodeMap   map[coord2DKey]*OperationRelateng_NodeSections
}

// OperationRelateng_NewTopologyComputer creates a new TopologyComputer.
func OperationRelateng_NewTopologyComputer(predicate OperationRelateng_TopologyPredicate, geomA, geomB *OperationRelateng_RelateGeometry) *OperationRelateng_TopologyComputer {
	tc := &OperationRelateng_TopologyComputer{
		predicate: predicate,
		geomA:     geomA,
		geomB:     geomB,
		nodeMap:   make(map[coord2DKey]*OperationRelateng_NodeSections),
	}
	tc.initExteriorDims()
	return tc
}

// initExteriorDims determines a priori partial EXTERIOR topology based on
// dimensions.
func (tc *OperationRelateng_TopologyComputer) initExteriorDims() {
	dimRealA := tc.geomA.GetDimensionReal()
	dimRealB := tc.geomB.GetDimensionReal()

	// For P/L case, P exterior intersects L interior.
	if dimRealA == Geom_Dimension_P && dimRealB == Geom_Dimension_L {
		tc.updateDim(Geom_Location_Exterior, Geom_Location_Interior, Geom_Dimension_L)
	} else if dimRealA == Geom_Dimension_L && dimRealB == Geom_Dimension_P {
		tc.updateDim(Geom_Location_Interior, Geom_Location_Exterior, Geom_Dimension_L)
	} else if dimRealA == Geom_Dimension_P && dimRealB == Geom_Dimension_A {
		// For P/A case, the Area Int and Bdy intersect the Point exterior.
		tc.updateDim(Geom_Location_Exterior, Geom_Location_Interior, Geom_Dimension_A)
		tc.updateDim(Geom_Location_Exterior, Geom_Location_Boundary, Geom_Dimension_L)
	} else if dimRealA == Geom_Dimension_A && dimRealB == Geom_Dimension_P {
		tc.updateDim(Geom_Location_Interior, Geom_Location_Exterior, Geom_Dimension_A)
		tc.updateDim(Geom_Location_Boundary, Geom_Location_Exterior, Geom_Dimension_L)
	} else if dimRealA == Geom_Dimension_L && dimRealB == Geom_Dimension_A {
		tc.updateDim(Geom_Location_Exterior, Geom_Location_Interior, Geom_Dimension_A)
	} else if dimRealA == Geom_Dimension_A && dimRealB == Geom_Dimension_L {
		tc.updateDim(Geom_Location_Interior, Geom_Location_Exterior, Geom_Dimension_A)
	} else if dimRealA == Geom_Dimension_False || dimRealB == Geom_Dimension_False {
		// Cases where one geom is EMPTY.
		if dimRealA != Geom_Dimension_False {
			tc.initExteriorEmpty(OperationRelateng_RelateGeometry_GEOM_A)
		}
		if dimRealB != Geom_Dimension_False {
			tc.initExteriorEmpty(OperationRelateng_RelateGeometry_GEOM_B)
		}
	}
}

func (tc *OperationRelateng_TopologyComputer) initExteriorEmpty(geomNonEmpty bool) {
	dimNonEmpty := tc.GetDimension(geomNonEmpty)
	switch dimNonEmpty {
	case Geom_Dimension_P:
		tc.updateDimAB(geomNonEmpty, Geom_Location_Interior, Geom_Location_Exterior, Geom_Dimension_P)
	case Geom_Dimension_L:
		if tc.getGeometry(geomNonEmpty).HasBoundary() {
			tc.updateDimAB(geomNonEmpty, Geom_Location_Boundary, Geom_Location_Exterior, Geom_Dimension_P)
		}
		tc.updateDimAB(geomNonEmpty, Geom_Location_Interior, Geom_Location_Exterior, Geom_Dimension_L)
	case Geom_Dimension_A:
		tc.updateDimAB(geomNonEmpty, Geom_Location_Boundary, Geom_Location_Exterior, Geom_Dimension_L)
		tc.updateDimAB(geomNonEmpty, Geom_Location_Interior, Geom_Location_Exterior, Geom_Dimension_A)
	}
}

func (tc *OperationRelateng_TopologyComputer) getGeometry(isA bool) *OperationRelateng_RelateGeometry {
	if isA {
		return tc.geomA
	}
	return tc.geomB
}

// GetDimension returns the dimension of the specified geometry.
func (tc *OperationRelateng_TopologyComputer) GetDimension(isA bool) int {
	return tc.getGeometry(isA).GetDimension()
}

// IsAreaArea tests if both geometries are areas.
func (tc *OperationRelateng_TopologyComputer) IsAreaArea() bool {
	return tc.GetDimension(OperationRelateng_RelateGeometry_GEOM_A) == Geom_Dimension_A &&
		tc.GetDimension(OperationRelateng_RelateGeometry_GEOM_B) == Geom_Dimension_A
}

// IsSelfNodingRequired indicates whether the input geometries require
// self-noding for correct evaluation of specific spatial predicates.
// Self-noding is required for geometries which may have self-crossing linework.
func (tc *OperationRelateng_TopologyComputer) IsSelfNodingRequired() bool {
	if tc.predicate.RequireSelfNoding() {
		if tc.geomA.IsSelfNodingRequired() || tc.geomB.IsSelfNodingRequired() {
			return true
		}
	}
	return false
}

// IsExteriorCheckRequired tests if exterior check is required for the given
// geometry.
func (tc *OperationRelateng_TopologyComputer) IsExteriorCheckRequired(isA bool) bool {
	return tc.predicate.RequireExteriorCheck(isA)
}

func (tc *OperationRelateng_TopologyComputer) updateDim(locA, locB, dimension int) {
	tc.predicate.UpdateDimension(locA, locB, dimension)
}

func (tc *OperationRelateng_TopologyComputer) updateDimAB(isAB bool, loc1, loc2, dimension int) {
	if isAB {
		tc.updateDim(loc1, loc2, dimension)
	} else {
		// Is ordered BA.
		tc.updateDim(loc2, loc1, dimension)
	}
}

// IsResultKnown tests if the result of the predicate is already known.
func (tc *OperationRelateng_TopologyComputer) IsResultKnown() bool {
	return tc.predicate.IsKnown()
}

// GetResult returns the result of the predicate.
func (tc *OperationRelateng_TopologyComputer) GetResult() bool {
	return tc.predicate.Value()
}

// Finish finalizes the evaluation.
func (tc *OperationRelateng_TopologyComputer) Finish() {
	tc.predicate.Finish()
}

func (tc *OperationRelateng_TopologyComputer) getNodeSections(nodePt *Geom_Coordinate) *OperationRelateng_NodeSections {
	key := coord2DKey{x: nodePt.X, y: nodePt.Y}
	node, ok := tc.nodeMap[key]
	if !ok {
		node = OperationRelateng_NewNodeSections(nodePt)
		tc.nodeMap[key] = node
	}
	return node
}

// AddIntersection adds an intersection between two node sections.
func (tc *OperationRelateng_TopologyComputer) AddIntersection(a, b *OperationRelateng_NodeSection) {
	if !a.IsSameGeometry(b) {
		tc.updateIntersectionAB(a, b)
	}
	// Add edges to node to allow full topology evaluation later.
	tc.addNodeSections(a, b)
}

// updateIntersectionAB updates topology for an intersection between A and B.
func (tc *OperationRelateng_TopologyComputer) updateIntersectionAB(a, b *OperationRelateng_NodeSection) {
	if OperationRelateng_NodeSection_IsAreaArea(a, b) {
		tc.updateAreaAreaCross(a, b)
	}
	tc.updateNodeLocation(a, b)
}

// updateAreaAreaCross updates topology for an AB Area-Area crossing node.
// Sections cross at a node if (a) the intersection is proper (i.e. in the
// interior of two segments) or (b) if non-proper then whether the linework
// crosses is determined by the geometry of the segments on either side of the
// node. In these situations the area geometry interiors intersect (in
// dimension 2).
func (tc *OperationRelateng_TopologyComputer) updateAreaAreaCross(a, b *OperationRelateng_NodeSection) {
	isProper := OperationRelateng_NodeSection_IsProperSections(a, b)
	if isProper || Algorithm_PolygonNodeTopology_IsCrossing(a.NodePt(),
		a.GetVertex(0), a.GetVertex(1),
		b.GetVertex(0), b.GetVertex(1)) {
		tc.updateDim(Geom_Location_Interior, Geom_Location_Interior, Geom_Dimension_A)
	}
}

// updateNodeLocation updates topology for a node at an AB edge intersection.
func (tc *OperationRelateng_TopologyComputer) updateNodeLocation(a, b *OperationRelateng_NodeSection) {
	pt := a.NodePt()
	locA := tc.geomA.LocateNode(pt, a.GetPolygonal())
	locB := tc.geomB.LocateNode(pt, b.GetPolygonal())
	tc.updateDim(locA, locB, Geom_Dimension_P)
}

func (tc *OperationRelateng_TopologyComputer) addNodeSections(ns0, ns1 *OperationRelateng_NodeSection) {
	sections := tc.getNodeSections(ns0.NodePt())
	sections.AddNodeSection(ns0)
	sections.AddNodeSection(ns1)
}

// AddPointOnPointInterior adds topology for two points intersecting.
func (tc *OperationRelateng_TopologyComputer) AddPointOnPointInterior(pt *Geom_Coordinate) {
	tc.updateDim(Geom_Location_Interior, Geom_Location_Interior, Geom_Dimension_P)
}

// AddPointOnPointExterior adds topology for a point in the exterior of another
// point geometry.
func (tc *OperationRelateng_TopologyComputer) AddPointOnPointExterior(isGeomA bool, pt *Geom_Coordinate) {
	tc.updateDimAB(isGeomA, Geom_Location_Interior, Geom_Location_Exterior, Geom_Dimension_P)
}

// AddPointOnGeometry adds topology for a point intersecting a target geometry.
func (tc *OperationRelateng_TopologyComputer) AddPointOnGeometry(isA bool, locTarget, dimTarget int, pt *Geom_Coordinate) {
	tc.updateDimAB(isA, Geom_Location_Interior, locTarget, Geom_Dimension_P)
	switch dimTarget {
	case Geom_Dimension_P:
		return
	case Geom_Dimension_L:
		// Because zero-length lines are handled, a point lying in the exterior
		// of the line target may imply either P or L for the Exterior
		// interaction.
		return
	case Geom_Dimension_A:
		// If a point intersects an area target, then the area interior and
		// boundary must extend beyond the point and thus interact with its
		// exterior.
		tc.updateDimAB(isA, Geom_Location_Exterior, Geom_Location_Interior, Geom_Dimension_A)
		tc.updateDimAB(isA, Geom_Location_Exterior, Geom_Location_Boundary, Geom_Dimension_L)
		return
	}
	panic("Unknown target dimension")
}

// AddLineEndOnGeometry adds topology for a line end. The line end point must be
// "significant"; i.e. not contained in an area if the source is a mixed-
// dimension GC.
func (tc *OperationRelateng_TopologyComputer) AddLineEndOnGeometry(isLineA bool, locLineEnd, locTarget, dimTarget int, pt *Geom_Coordinate) {
	// Record topology at line end point.
	tc.updateDimAB(isLineA, locLineEnd, locTarget, Geom_Dimension_P)

	// Line and Area targets may have additional topology.
	switch dimTarget {
	case Geom_Dimension_P:
		return
	case Geom_Dimension_L:
		tc.addLineEndOnLine(isLineA, locLineEnd, locTarget, pt)
		return
	case Geom_Dimension_A:
		tc.addLineEndOnArea(isLineA, locLineEnd, locTarget, pt)
		return
	}
	panic("Unknown target dimension")
}

func (tc *OperationRelateng_TopologyComputer) addLineEndOnLine(isLineA bool, locLineEnd, locLine int, pt *Geom_Coordinate) {
	// When a line end is in the EXTERIOR of a Line, some length of the source
	// Line INTERIOR is also in the target Line EXTERIOR. This works for
	// zero-length lines as well.
	if locLine == Geom_Location_Exterior {
		tc.updateDimAB(isLineA, Geom_Location_Interior, Geom_Location_Exterior, Geom_Dimension_L)
	}
}

func (tc *OperationRelateng_TopologyComputer) addLineEndOnArea(isLineA bool, locLineEnd, locArea int, pt *Geom_Coordinate) {
	if locArea != Geom_Location_Boundary {
		// When a line end is in an Area INTERIOR or EXTERIOR some length of the
		// source Line Interior AND the Exterior of the line is also in that
		// location of the target.
		// NOTE: this assumes the line end is NOT also in an Area of a mixed-dim
		// GC.
		tc.updateDimAB(isLineA, Geom_Location_Interior, locArea, Geom_Dimension_L)
		tc.updateDimAB(isLineA, Geom_Location_Exterior, locArea, Geom_Dimension_A)
	}
}

// AddAreaVertex adds topology for an area vertex interaction with a target
// geometry element. Assumes the target geometry element has highest dimension
// (i.e. if the point lies on two elements of different dimension, the location
// on the higher dimension element is provided. This is the semantic provided by
// RelatePointLocator).
//
// Note that in a GeometryCollection containing overlapping or adjacent
// polygons, the area vertex location may be INTERIOR instead of BOUNDARY.
func (tc *OperationRelateng_TopologyComputer) AddAreaVertex(isAreaA bool, locArea, locTarget, dimTarget int, pt *Geom_Coordinate) {
	if locTarget == Geom_Location_Exterior {
		tc.updateDimAB(isAreaA, Geom_Location_Interior, Geom_Location_Exterior, Geom_Dimension_A)
		// If area vertex is on Boundary further topology can be deduced from the
		// neighbourhood around the boundary vertex. This is always the case for
		// polygonal geometries. For GCs, the vertex may be either on boundary or
		// in interior (i.e. of overlapping or adjacent polygons).
		if locArea == Geom_Location_Boundary {
			tc.updateDimAB(isAreaA, Geom_Location_Boundary, Geom_Location_Exterior, Geom_Dimension_L)
			tc.updateDimAB(isAreaA, Geom_Location_Exterior, Geom_Location_Exterior, Geom_Dimension_A)
		}
		return
	}
	switch dimTarget {
	case Geom_Dimension_P:
		tc.addAreaVertexOnPoint(isAreaA, locArea, pt)
		return
	case Geom_Dimension_L:
		tc.addAreaVertexOnLine(isAreaA, locArea, locTarget, pt)
		return
	case Geom_Dimension_A:
		tc.addAreaVertexOnArea(isAreaA, locArea, locTarget, pt)
		return
	}
	panic("Unknown target dimension")
}

// addAreaVertexOnPoint updates topology for an area vertex (in Interior or on
// Boundary) intersecting a point. Note that because the largest dimension of
// intersecting target is determined, the intersecting point is not part of any
// other target geometry, and hence its neighbourhood is in the Exterior of the
// target.
func (tc *OperationRelateng_TopologyComputer) addAreaVertexOnPoint(isAreaA bool, locArea int, pt *Geom_Coordinate) {
	// Assert: locArea != EXTERIOR
	// Assert: locTarget == INTERIOR
	// The vertex location intersects the Point.
	tc.updateDimAB(isAreaA, locArea, Geom_Location_Interior, Geom_Dimension_P)
	// The area interior intersects the point's exterior neighbourhood.
	tc.updateDimAB(isAreaA, Geom_Location_Interior, Geom_Location_Exterior, Geom_Dimension_A)
	// If the area vertex is on the boundary, the area boundary and exterior
	// intersect the point's exterior neighbourhood.
	if locArea == Geom_Location_Boundary {
		tc.updateDimAB(isAreaA, Geom_Location_Boundary, Geom_Location_Exterior, Geom_Dimension_L)
		tc.updateDimAB(isAreaA, Geom_Location_Exterior, Geom_Location_Exterior, Geom_Dimension_A)
	}
}

func (tc *OperationRelateng_TopologyComputer) addAreaVertexOnLine(isAreaA bool, locArea, locTarget int, pt *Geom_Coordinate) {
	// Assert: locArea != EXTERIOR
	// If an area vertex intersects a line, all we know is the intersection at
	// that point. e.g. the line may or may not be collinear with the area
	// boundary, and the line may or may not intersect the area interior. Full
	// topology is determined later by node analysis.
	tc.updateDimAB(isAreaA, locArea, locTarget, Geom_Dimension_P)
	if locArea == Geom_Location_Interior {
		// The area interior intersects the line's exterior neighbourhood.
		tc.updateDimAB(isAreaA, Geom_Location_Interior, Geom_Location_Exterior, Geom_Dimension_A)
	}
}

func (tc *OperationRelateng_TopologyComputer) addAreaVertexOnArea(isAreaA bool, locArea, locTarget int, pt *Geom_Coordinate) {
	if locTarget == Geom_Location_Boundary {
		if locArea == Geom_Location_Boundary {
			// B/B topology is fully computed later by node analysis.
			tc.updateDimAB(isAreaA, Geom_Location_Boundary, Geom_Location_Boundary, Geom_Dimension_P)
		} else {
			// locArea == INTERIOR
			tc.updateDimAB(isAreaA, Geom_Location_Interior, Geom_Location_Interior, Geom_Dimension_A)
			tc.updateDimAB(isAreaA, Geom_Location_Interior, Geom_Location_Boundary, Geom_Dimension_L)
			tc.updateDimAB(isAreaA, Geom_Location_Interior, Geom_Location_Exterior, Geom_Dimension_A)
		}
	} else {
		// locTarget is INTERIOR or EXTERIOR.
		tc.updateDimAB(isAreaA, Geom_Location_Interior, locTarget, Geom_Dimension_A)
		// If area vertex is on Boundary further topology can be deduced from the
		// neighbourhood around the boundary vertex. This is always the case for
		// polygonal geometries. For GCs, the vertex may be either on boundary or
		// in interior (i.e. of overlapping or adjacent polygons).
		if locArea == Geom_Location_Boundary {
			tc.updateDimAB(isAreaA, Geom_Location_Boundary, locTarget, Geom_Dimension_L)
			tc.updateDimAB(isAreaA, Geom_Location_Exterior, locTarget, Geom_Dimension_A)
		}
	}
}

// EvaluateNodes evaluates the topology at all intersection nodes.
func (tc *OperationRelateng_TopologyComputer) EvaluateNodes() {
	for _, nodeSections := range tc.nodeMap {
		if nodeSections.HasInteractionAB() {
			tc.evaluateNode(nodeSections)
			if tc.IsResultKnown() {
				return
			}
		}
	}
}

func (tc *OperationRelateng_TopologyComputer) evaluateNode(nodeSections *OperationRelateng_NodeSections) {
	p := nodeSections.GetCoordinate()
	node := nodeSections.CreateNode()
	// Node must have edges for geom, but may also be in interior of a
	// overlapping GC.
	isAreaInteriorA := tc.geomA.IsNodeInArea(p, nodeSections.GetPolygonal(OperationRelateng_RelateGeometry_GEOM_A))
	isAreaInteriorB := tc.geomB.IsNodeInArea(p, nodeSections.GetPolygonal(OperationRelateng_RelateGeometry_GEOM_B))
	node.Finish(isAreaInteriorA, isAreaInteriorB)
	tc.evaluateNodeEdges(node)
}

func (tc *OperationRelateng_TopologyComputer) evaluateNodeEdges(node *OperationRelateng_RelateNode) {
	for _, e := range node.GetEdges() {
		// An optimization to avoid updates for cases with a linear geometry.
		if tc.IsAreaArea() {
			tc.updateDim(e.Location(OperationRelateng_RelateGeometry_GEOM_A, Geom_Position_Left),
				e.Location(OperationRelateng_RelateGeometry_GEOM_B, Geom_Position_Left), Geom_Dimension_A)
			tc.updateDim(e.Location(OperationRelateng_RelateGeometry_GEOM_A, Geom_Position_Right),
				e.Location(OperationRelateng_RelateGeometry_GEOM_B, Geom_Position_Right), Geom_Dimension_A)
		}
		tc.updateDim(e.Location(OperationRelateng_RelateGeometry_GEOM_A, Geom_Position_On),
			e.Location(OperationRelateng_RelateGeometry_GEOM_B, Geom_Position_On), Geom_Dimension_L)
	}
}
