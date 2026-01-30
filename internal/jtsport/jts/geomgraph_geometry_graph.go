package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Geomgraph_GeometryGraph_DetermineBoundary determines whether a point is on
// the boundary based on the boundary node rule and boundary count.
func Geomgraph_GeometryGraph_DetermineBoundary(boundaryNodeRule Algorithm_BoundaryNodeRule, boundaryCount int) int {
	if boundaryNodeRule.IsInBoundary(boundaryCount) {
		return Geom_Location_Boundary
	}
	return Geom_Location_Interior
}

// Geomgraph_GeometryGraph is a graph that models a given Geometry.
type Geomgraph_GeometryGraph struct {
	*Geomgraph_PlanarGraph
	child java.Polymorphic

	parentGeom *Geom_Geometry

	// lineEdgeMap is a map of the linestring components of the parentGeometry
	// to the edges which are derived from them.
	lineEdgeMap map[*Geom_LineString]*Geomgraph_Edge

	boundaryNodeRule Algorithm_BoundaryNodeRule

	// If this flag is true, the Boundary Determination Rule will be used when
	// deciding whether nodes are in the boundary or not.
	useBoundaryDeterminationRule bool
	argIndex                     int // The index of this geometry as an argument to a spatial function.
	boundaryNodes                []*Geomgraph_Node
	hasTooFewPoints              bool
	invalidPoint                 *Geom_Coordinate

	areaPtLocator *AlgorithmLocate_IndexedPointInAreaLocator
	ptLocator     *Algorithm_PointLocator
}

// GetChild returns the immediate child in the type hierarchy chain.
func (gg *Geomgraph_GeometryGraph) GetChild() java.Polymorphic {
	return gg.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (gg *Geomgraph_GeometryGraph) GetParent() java.Polymorphic {
	return gg.Geomgraph_PlanarGraph
}

func (gg *Geomgraph_GeometryGraph) createEdgeSetIntersector() *GeomgraphIndex_SimpleMCSweepLineIntersector {
	return GeomgraphIndex_NewSimpleMCSweepLineIntersector()
}

// Geomgraph_NewGeometryGraph creates a new GeometryGraph with the default boundary rule.
func Geomgraph_NewGeometryGraph(argIndex int, parentGeom *Geom_Geometry) *Geomgraph_GeometryGraph {
	return Geomgraph_NewGeometryGraphWithBoundaryNodeRule(argIndex, parentGeom, Algorithm_BoundaryNodeRule_OGC_SFS_BOUNDARY_RULE)
}

// Geomgraph_NewGeometryGraphWithBoundaryNodeRule creates a new GeometryGraph with the given boundary rule.
func Geomgraph_NewGeometryGraphWithBoundaryNodeRule(argIndex int, parentGeom *Geom_Geometry, boundaryNodeRule Algorithm_BoundaryNodeRule) *Geomgraph_GeometryGraph {
	pg := Geomgraph_NewPlanarGraphDefault()
	gg := &Geomgraph_GeometryGraph{
		Geomgraph_PlanarGraph:         pg,
		argIndex:                     argIndex,
		parentGeom:                   parentGeom,
		boundaryNodeRule:             boundaryNodeRule,
		lineEdgeMap:                  make(map[*Geom_LineString]*Geomgraph_Edge),
		useBoundaryDeterminationRule: true,
		ptLocator:                    Algorithm_NewPointLocator(),
	}
	pg.child = gg
	if parentGeom != nil {
		gg.add(parentGeom)
	}
	return gg
}

// HasTooFewPoints returns true if too few points were found.
func (gg *Geomgraph_GeometryGraph) HasTooFewPoints() bool {
	return gg.hasTooFewPoints
}

// GetInvalidPoint returns the invalid point if HasTooFewPoints is true.
func (gg *Geomgraph_GeometryGraph) GetInvalidPoint() *Geom_Coordinate {
	return gg.invalidPoint
}

// GetGeometry returns the parent geometry.
func (gg *Geomgraph_GeometryGraph) GetGeometry() *Geom_Geometry {
	return gg.parentGeom
}

// GetBoundaryNodeRule returns the boundary node rule.
func (gg *Geomgraph_GeometryGraph) GetBoundaryNodeRule() Algorithm_BoundaryNodeRule {
	return gg.boundaryNodeRule
}

// GetBoundaryNodes returns the boundary nodes.
func (gg *Geomgraph_GeometryGraph) GetBoundaryNodes() []*Geomgraph_Node {
	if gg.boundaryNodes == nil {
		gg.boundaryNodes = gg.nodes.GetBoundaryNodes(gg.argIndex)
	}
	return gg.boundaryNodes
}

// GetBoundaryPoints returns the coordinates of the boundary nodes.
func (gg *Geomgraph_GeometryGraph) GetBoundaryPoints() []*Geom_Coordinate {
	coll := gg.GetBoundaryNodes()
	pts := make([]*Geom_Coordinate, len(coll))
	for i, node := range coll {
		pts[i] = node.GetCoordinate().Copy()
	}
	return pts
}

// FindEdgeFromLine returns the edge derived from the given line.
func (gg *Geomgraph_GeometryGraph) FindEdgeFromLine(line *Geom_LineString) *Geomgraph_Edge {
	return gg.lineEdgeMap[line]
}

// ComputeSplitEdges computes the split edges from the edges in this graph.
func (gg *Geomgraph_GeometryGraph) ComputeSplitEdges(edgelist *[]*Geomgraph_Edge) {
	for _, e := range gg.edges {
		e.GetEdgeIntersectionList().AddSplitEdges(edgelist)
	}
}

func (gg *Geomgraph_GeometryGraph) add(g *Geom_Geometry) {
	if g.IsEmpty() {
		return
	}

	// Check if this Geometry should obey the Boundary Determination Rule.
	// All collections except MultiPolygons obey the rule.
	if java.InstanceOf[*Geom_MultiPolygon](g) {
		gg.useBoundaryDeterminationRule = false
	}

	switch {
	case java.InstanceOf[*Geom_Polygon](g):
		gg.addPolygon(java.Cast[*Geom_Polygon](g))
	case java.InstanceOf[*Geom_LineString](g):
		gg.addLineString(java.Cast[*Geom_LineString](g))
	case java.InstanceOf[*Geom_Point](g):
		gg.addPoint(java.Cast[*Geom_Point](g))
	case java.InstanceOf[*Geom_MultiPoint](g):
		gg.addCollection(java.Cast[*Geom_MultiPoint](g).Geom_GeometryCollection)
	case java.InstanceOf[*Geom_MultiLineString](g):
		gg.addCollection(java.Cast[*Geom_MultiLineString](g).Geom_GeometryCollection)
	case java.InstanceOf[*Geom_MultiPolygon](g):
		gg.addCollection(java.Cast[*Geom_MultiPolygon](g).Geom_GeometryCollection)
	case java.InstanceOf[*Geom_GeometryCollection](g):
		gg.addCollection(java.Cast[*Geom_GeometryCollection](g))
	default:
		panic("unsupported geometry type")
	}
}

func (gg *Geomgraph_GeometryGraph) addCollection(gc *Geom_GeometryCollection) {
	for i := 0; i < gc.GetNumGeometries(); i++ {
		gg.add(gc.GetGeometryN(i))
	}
}

func (gg *Geomgraph_GeometryGraph) addPoint(p *Geom_Point) {
	coord := p.GetCoordinate()
	gg.insertPoint(gg.argIndex, coord, Geom_Location_Interior)
}

// addPolygonRing adds a polygon ring to the graph. Empty rings are ignored.
// The left and right topological location arguments assume that the ring is
// oriented CW. If the ring is in the opposite orientation, the left and right
// locations must be interchanged.
func (gg *Geomgraph_GeometryGraph) addPolygonRing(lr *Geom_LinearRing, cwLeft, cwRight int) {
	// Don't bother adding empty holes.
	if lr.IsEmpty() {
		return
	}

	coord := Geom_CoordinateArrays_RemoveRepeatedPoints(lr.GetCoordinates())

	if len(coord) < 4 {
		gg.hasTooFewPoints = true
		gg.invalidPoint = coord[0]
		return
	}

	left := cwLeft
	right := cwRight
	if Algorithm_Orientation_IsCCW(coord) {
		left = cwRight
		right = cwLeft
	}
	e := Geomgraph_NewEdge(coord, Geomgraph_NewLabelGeomOnLeftRight(gg.argIndex, Geom_Location_Boundary, left, right))
	gg.lineEdgeMap[lr.Geom_LineString] = e

	gg.InsertEdge(e)
	// Insert the endpoint as a node, to mark that it is on the boundary.
	gg.insertPoint(gg.argIndex, coord[0], Geom_Location_Boundary)
}

func (gg *Geomgraph_GeometryGraph) addPolygon(p *Geom_Polygon) {
	gg.addPolygonRing(p.GetExteriorRing(), Geom_Location_Exterior, Geom_Location_Interior)

	for i := 0; i < p.GetNumInteriorRing(); i++ {
		hole := p.GetInteriorRingN(i)
		// Holes are topologically labelled opposite to the shell, since
		// the interior of the polygon lies on their opposite side
		// (on the left, if the hole is oriented CW).
		gg.addPolygonRing(hole, Geom_Location_Interior, Geom_Location_Exterior)
	}
}

func (gg *Geomgraph_GeometryGraph) addLineString(line *Geom_LineString) {
	coord := Geom_CoordinateArrays_RemoveRepeatedPoints(line.GetCoordinates())

	if len(coord) < 2 {
		gg.hasTooFewPoints = true
		gg.invalidPoint = coord[0]
		return
	}

	// Add the edge for the LineString.
	// Line edges do not have locations for their left and right sides.
	e := Geomgraph_NewEdge(coord, Geomgraph_NewLabelGeomOn(gg.argIndex, Geom_Location_Interior))
	gg.lineEdgeMap[line] = e
	gg.InsertEdge(e)

	// Add the boundary points of the LineString, if any.
	// Even if the LineString is closed, add both points as if they were endpoints.
	// This allows for the case that the node already exists and is a boundary point.
	Util_Assert_IsTrueWithMessage(len(coord) >= 2, "found LineString with single point")
	gg.insertBoundaryPoint(gg.argIndex, coord[0])
	gg.insertBoundaryPoint(gg.argIndex, coord[len(coord)-1])
}

// AddEdge adds an Edge computed externally. The label on the Edge is assumed
// to be correct.
func (gg *Geomgraph_GeometryGraph) AddEdge(e *Geomgraph_Edge) {
	gg.InsertEdge(e)
	coord := e.GetCoordinates()
	// Insert the endpoint as a node, to mark that it is on the boundary.
	gg.insertPoint(gg.argIndex, coord[0], Geom_Location_Boundary)
	gg.insertPoint(gg.argIndex, coord[len(coord)-1], Geom_Location_Boundary)
}

// AddPointFromCoord adds a point computed externally. The point is assumed to be
// a Point Geometry part, which has a location of INTERIOR.
func (gg *Geomgraph_GeometryGraph) AddPointFromCoord(pt *Geom_Coordinate) {
	gg.insertPoint(gg.argIndex, pt, Geom_Location_Interior)
}

// ComputeSelfNodes computes self-nodes, taking advantage of the Geometry type to
// minimize the number of intersection tests. (E.g. rings are not tested for
// self-intersection, since they are assumed to be valid).
func (gg *Geomgraph_GeometryGraph) ComputeSelfNodes(li *Algorithm_LineIntersector, computeRingSelfNodes bool) *GeomgraphIndex_SegmentIntersector {
	si := GeomgraphIndex_NewSegmentIntersector(li, true, false)
	esi := gg.createEdgeSetIntersector()
	// Optimize intersection search for valid Polygons and LinearRings.
	isRings := java.InstanceOf[*Geom_LinearRing](gg.parentGeom) ||
		java.InstanceOf[*Geom_Polygon](gg.parentGeom) ||
		java.InstanceOf[*Geom_MultiPolygon](gg.parentGeom)
	computeAllSegments := computeRingSelfNodes || !isRings
	esi.ComputeIntersectionsSingleList(gg.edges, si, computeAllSegments)

	gg.addSelfIntersectionNodes(gg.argIndex)
	return si
}

// ComputeEdgeIntersections computes intersections between this graph's edges
// and another graph's edges.
func (gg *Geomgraph_GeometryGraph) ComputeEdgeIntersections(g *Geomgraph_GeometryGraph, li *Algorithm_LineIntersector, includeProper bool) *GeomgraphIndex_SegmentIntersector {
	si := GeomgraphIndex_NewSegmentIntersector(li, includeProper, true)
	si.SetBoundaryNodes(gg.GetBoundaryNodes(), g.GetBoundaryNodes())

	esi := gg.createEdgeSetIntersector()
	esi.ComputeIntersectionsTwoLists(gg.edges, g.edges, si)
	return si
}

func (gg *Geomgraph_GeometryGraph) insertPoint(argIndex int, coord *Geom_Coordinate, onLocation int) {
	n := gg.nodes.AddNodeFromCoord(coord)
	lbl := n.GetLabel()
	if lbl == nil {
		n.label = Geomgraph_NewLabelGeomOn(argIndex, onLocation)
	} else {
		lbl.SetLocationOn(argIndex, onLocation)
	}
}

// insertBoundaryPoint adds candidate boundary points using the current
// BoundaryNodeRule. This is used to add the boundary points of dim-1
// geometries (Curves/MultiCurves).
func (gg *Geomgraph_GeometryGraph) insertBoundaryPoint(argIndex int, coord *Geom_Coordinate) {
	n := gg.nodes.AddNodeFromCoord(coord)
	// Nodes always have labels.
	lbl := n.GetLabel()
	// The new point to insert is on a boundary.
	boundaryCount := 1
	// Determine the current location for the point (if any).
	loc := lbl.GetLocation(argIndex, Geom_Position_On)
	if loc == Geom_Location_Boundary {
		boundaryCount++
	}

	// Determine the boundary status of the point according to the Boundary Determination Rule.
	newLoc := Geomgraph_GeometryGraph_DetermineBoundary(gg.boundaryNodeRule, boundaryCount)
	lbl.SetLocationOn(argIndex, newLoc)
}

func (gg *Geomgraph_GeometryGraph) addSelfIntersectionNodes(argIndex int) {
	for _, e := range gg.edges {
		eLoc := e.GetLabel().GetLocationOn(argIndex)
		for _, ei := range e.GetEdgeIntersectionList().Iterator() {
			gg.addSelfIntersectionNode(argIndex, ei.Coord, eLoc)
		}
	}
}

// addSelfIntersectionNode adds a node for a self-intersection. If the node is
// a potential boundary node (e.g. came from an edge which is a boundary) then
// insert it as a potential boundary node. Otherwise, just add it as a regular
// node.
func (gg *Geomgraph_GeometryGraph) addSelfIntersectionNode(argIndex int, coord *Geom_Coordinate, loc int) {
	// If this node is already a boundary node, don't change it.
	if gg.IsBoundaryNode(argIndex, coord) {
		return
	}
	if loc == Geom_Location_Boundary && gg.useBoundaryDeterminationRule {
		gg.insertBoundaryPoint(argIndex, coord)
	} else {
		gg.insertPoint(argIndex, coord, loc)
	}
}

// Locate determines the Location of the given Coordinate in this geometry.
func (gg *Geomgraph_GeometryGraph) Locate(pt *Geom_Coordinate) int {
	if java.InstanceOf[*Geom_Polygonal](gg.parentGeom) && gg.parentGeom.GetNumGeometries() > 50 {
		// Lazily init point locator.
		if gg.areaPtLocator == nil {
			gg.areaPtLocator = AlgorithmLocate_NewIndexedPointInAreaLocator(gg.parentGeom)
		}
		return gg.areaPtLocator.Locate(pt)
	}
	return gg.ptLocator.Locate(pt, gg.parentGeom)
}
