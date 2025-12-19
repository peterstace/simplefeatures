package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationRelate_RelateComputer computes the topological relationship between
// two Geometries.
//
// RelateComputer does not need to build a complete graph structure to compute
// the IntersectionMatrix. The relationship between the geometries can be
// computed by simply examining the labelling of edges incident on each node.
//
// RelateComputer does not currently support arbitrary GeometryCollections. This
// is because GeometryCollections can contain overlapping Polygons. In order to
// correct compute relate on overlapping Polygons, they would first need to be
// noded and merged (if not explicitly, at least implicitly).
type OperationRelate_RelateComputer struct {
	child java.Polymorphic

	li            *Algorithm_LineIntersector
	ptLocator     *Algorithm_PointLocator
	arg           []*Geomgraph_GeometryGraph // The arg(s) of the operation.
	nodes         *Geomgraph_NodeMap
	im            *Geom_IntersectionMatrix
	isolatedEdges []*Geomgraph_Edge

	// The intersection point found (if any).
	invalidPoint *Geom_Coordinate
}

// GetChild returns the immediate child in the type hierarchy chain.
func (rc *OperationRelate_RelateComputer) GetChild() java.Polymorphic {
	return rc.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (rc *OperationRelate_RelateComputer) GetParent() java.Polymorphic {
	return nil
}

// OperationRelate_NewRelateComputer creates a new RelateComputer with the given
// geometry graphs.
func OperationRelate_NewRelateComputer(arg []*Geomgraph_GeometryGraph) *OperationRelate_RelateComputer {
	return &OperationRelate_RelateComputer{
		li:            Algorithm_NewRobustLineIntersector().Algorithm_LineIntersector,
		ptLocator:     Algorithm_NewPointLocator(),
		arg:           arg,
		nodes:         Geomgraph_NewNodeMap(OperationRelate_NewRelateNodeFactory().Geomgraph_NodeFactory),
		isolatedEdges: make([]*Geomgraph_Edge, 0),
	}
}

// ComputeIM computes the IntersectionMatrix for the relate operation.
func (rc *OperationRelate_RelateComputer) ComputeIM() *Geom_IntersectionMatrix {
	im := Geom_NewIntersectionMatrix()
	// Since Geometries are finite and embedded in a 2-D space, the EE element
	// must always be 2.
	im.Set(Geom_Location_Exterior, Geom_Location_Exterior, 2)

	// If the Geometries don't overlap there is nothing to do.
	if !rc.arg[0].GetGeometry().GetEnvelopeInternal().IntersectsEnvelope(
		rc.arg[1].GetGeometry().GetEnvelopeInternal()) {
		rc.computeDisjointIM(im, rc.arg[0].GetBoundaryNodeRule())
		return im
	}

	rc.arg[0].ComputeSelfNodes(rc.li, false)
	rc.arg[1].ComputeSelfNodes(rc.li, false)

	// Compute intersections between edges of the two input geometries.
	intersector := rc.arg[0].ComputeEdgeIntersections(rc.arg[1], rc.li, false)

	rc.computeIntersectionNodes(0)
	rc.computeIntersectionNodes(1)

	// Copy the labelling for the nodes in the parent Geometries. These override
	// any labels determined by intersections between the geometries.
	rc.copyNodesAndLabels(0)
	rc.copyNodesAndLabels(1)

	// Complete the labelling for any nodes which only have a label for a single
	// geometry.
	rc.labelIsolatedNodes()

	// If a proper intersection was found, we can set a lower bound on the IM.
	rc.computeProperIntersectionIM(intersector, im)

	// Now process improper intersections (e.g. where one or other of the
	// geometries has a vertex at the intersection point). We need to compute
	// the edge graph at all nodes to determine the IM.

	// Build EdgeEnds for all intersections.
	eeBuilder := OperationRelate_NewEdgeEndBuilder()
	ee0 := eeBuilder.ComputeEdgeEndsFromIterator(rc.arg[0].GetEdgeIterator())
	rc.insertEdgeEnds(ee0)
	ee1 := eeBuilder.ComputeEdgeEndsFromIterator(rc.arg[1].GetEdgeIterator())
	rc.insertEdgeEnds(ee1)

	rc.labelNodeEdges()

	// Compute the labeling for isolated components. Isolated components are
	// components that do not touch any other components in the graph. They can
	// be identified by the fact that they will contain labels containing ONLY a
	// single element, the one for their parent geometry. We only need to check
	// components contained in the input graphs, since isolated components will
	// not have been replaced by new components formed by intersections.
	rc.labelIsolatedEdges(0, 1)
	rc.labelIsolatedEdges(1, 0)

	// Update the IM from all components.
	rc.updateIM(im)
	return im
}

func (rc *OperationRelate_RelateComputer) insertEdgeEnds(ee []*Geomgraph_EdgeEnd) {
	for _, e := range ee {
		rc.nodes.Add(e)
	}
}

func (rc *OperationRelate_RelateComputer) computeProperIntersectionIM(intersector *GeomgraphIndex_SegmentIntersector, im *Geom_IntersectionMatrix) {
	// If a proper intersection is found, we can set a lower bound on the IM.
	dimA := rc.arg[0].GetGeometry().GetDimension()
	dimB := rc.arg[1].GetGeometry().GetDimension()
	hasProper := intersector.HasProperIntersection()
	hasProperInterior := intersector.HasProperInteriorIntersection()

	// For Geometry's of dim 0 there can never be proper intersections.

	// If edge segments of Areas properly intersect, the areas must properly
	// overlap.
	if dimA == 2 && dimB == 2 {
		if hasProper {
			im.SetAtLeastFromString("212101212")
		}
	} else if dimA == 2 && dimB == 1 {
		// If an Line segment properly intersects an edge segment of an Area, it
		// follows that the Interior of the Line intersects the Boundary of the
		// Area. If the intersection is a proper *interior* intersection, then
		// there is an Interior-Interior intersection too. Note that it does not
		// follow that the Interior of the Line intersects the Exterior of the
		// Area, since there may be another Area component which contains the
		// rest of the Line.
		if hasProper {
			im.SetAtLeastFromString("FFF0FFFF2")
		}
		if hasProperInterior {
			im.SetAtLeastFromString("1FFFFF1FF")
		}
	} else if dimA == 1 && dimB == 2 {
		if hasProper {
			im.SetAtLeastFromString("F0FFFFFF2")
		}
		if hasProperInterior {
			im.SetAtLeastFromString("1F1FFFFFF")
		}
	} else if dimA == 1 && dimB == 1 {
		// If edges of LineStrings properly intersect *in an interior point*, all
		// we can deduce is that the interiors intersect. (We can NOT deduce that
		// the exteriors intersect, since some other segments in the geometries
		// might cover the points in the neighbourhood of the intersection.) It
		// is important that the point be known to be an interior point of both
		// Geometries, since it is possible in a self-intersecting geometry to
		// have a proper intersection on one segment that is also a boundary
		// point of another segment.
		if hasProperInterior {
			im.SetAtLeastFromString("0FFFFFFFF")
		}
	}
}

// copyNodesAndLabels copies all nodes from an arg geometry into this graph.
// The node label in the arg geometry overrides any previously computed label
// for that argIndex. (E.g. a node may be an intersection node with a computed
// label of BOUNDARY, but in the original arg Geometry it is actually in the
// interior due to the Boundary Determination Rule).
func (rc *OperationRelate_RelateComputer) copyNodesAndLabels(argIndex int) {
	for _, graphNode := range rc.arg[argIndex].GetNodeIterator() {
		newNode := rc.nodes.AddNodeFromCoord(graphNode.GetCoordinate())
		newNode.SetLabelAt(argIndex, graphNode.GetLabel().GetLocationOn(argIndex))
	}
}

// computeIntersectionNodes inserts nodes for all intersections on the edges of
// a Geometry. Label the created nodes the same as the edge label if they do not
// already have a label. This allows nodes created by either self-intersections
// or mutual intersections to be labelled. Endpoint nodes will already be
// labelled from when they were inserted.
func (rc *OperationRelate_RelateComputer) computeIntersectionNodes(argIndex int) {
	for _, e := range rc.arg[argIndex].GetEdgeIterator() {
		eLoc := e.GetLabel().GetLocationOn(argIndex)
		for _, ei := range e.GetEdgeIntersectionList().Iterator() {
			n := rc.nodes.AddNodeFromCoord(ei.Coord)
			rn := java.Cast[*OperationRelate_RelateNode](n)
			if eLoc == Geom_Location_Boundary {
				rn.SetLabelBoundary(argIndex)
			} else {
				if rn.GetLabel().IsNull(argIndex) {
					rn.SetLabelAt(argIndex, Geom_Location_Interior)
				}
			}
		}
	}
}

// computeDisjointIM fills in the IM when the Geometries are disjoint. We need
// to enter their dimension and boundary dimension in the Ext rows in the IM.
func (rc *OperationRelate_RelateComputer) computeDisjointIM(im *Geom_IntersectionMatrix, boundaryNodeRule Algorithm_BoundaryNodeRule) {
	ga := rc.arg[0].GetGeometry()
	if !ga.IsEmpty() {
		im.Set(Geom_Location_Interior, Geom_Location_Exterior, ga.GetDimension())
		im.Set(Geom_Location_Boundary, Geom_Location_Exterior, rc.getBoundaryDim(ga, boundaryNodeRule))
	}
	gb := rc.arg[1].GetGeometry()
	if !gb.IsEmpty() {
		im.Set(Geom_Location_Exterior, Geom_Location_Interior, gb.GetDimension())
		im.Set(Geom_Location_Exterior, Geom_Location_Boundary, rc.getBoundaryDim(gb, boundaryNodeRule))
	}
}

// getBoundaryDim computes the IM entry for the intersection of the boundary of
// a geometry with the Exterior. This is the nominal dimension of the boundary
// unless the boundary is empty, in which case it is Dimension.FALSE. For linear
// geometries the Boundary Node Rule determines whether the boundary is empty.
func (rc *OperationRelate_RelateComputer) getBoundaryDim(geom *Geom_Geometry, boundaryNodeRule Algorithm_BoundaryNodeRule) int {
	// If the geometry has a non-empty boundary the intersection is the nominal
	// dimension.
	if Operation_BoundaryOp_HasBoundary(geom, boundaryNodeRule) {
		// Special case for lines, since Geometry.getBoundaryDimension is not
		// aware of Boundary Node Rule.
		if geom.GetDimension() == 1 {
			return Geom_Dimension_P
		}
		return geom.GetBoundaryDimension()
	}
	// Otherwise intersection is F.
	return Geom_Dimension_False
}

func (rc *OperationRelate_RelateComputer) labelNodeEdges() {
	for _, node := range rc.nodes.Values() {
		rn := java.Cast[*OperationRelate_RelateNode](node)
		rn.GetEdges().ComputeLabelling(rc.arg)
	}
}

// updateIM updates the IM with the sum of the IMs for each component.
func (rc *OperationRelate_RelateComputer) updateIM(im *Geom_IntersectionMatrix) {
	for _, e := range rc.isolatedEdges {
		e.UpdateIM(im)
	}
	for _, node := range rc.nodes.Values() {
		rn := java.Cast[*OperationRelate_RelateNode](node)
		rn.UpdateIM(im)
		rn.UpdateIMFromEdges(im)
	}
}

// labelIsolatedEdges processes isolated edges by computing their labelling and
// adding them to the isolated edges list. Isolated edges are guaranteed not to
// touch the boundary of the target (since if they did, they would have caused
// an intersection to be computed and hence would not be isolated).
func (rc *OperationRelate_RelateComputer) labelIsolatedEdges(thisIndex, targetIndex int) {
	for _, e := range rc.arg[thisIndex].GetEdgeIterator() {
		if e.IsIsolated() {
			rc.labelIsolatedEdge(e, targetIndex, rc.arg[targetIndex].GetGeometry())
			rc.isolatedEdges = append(rc.isolatedEdges, e)
		}
	}
}

// labelIsolatedEdge labels an isolated edge of a graph with its relationship
// to the target geometry. If the target has dim 2 or 1, the edge can either be
// in the interior or the exterior. If the target has dim 0, the edge must be in
// the exterior.
func (rc *OperationRelate_RelateComputer) labelIsolatedEdge(e *Geomgraph_Edge, targetIndex int, target *Geom_Geometry) {
	// This won't work for GeometryCollections with both dim 2 and 1 geoms.
	if target.GetDimension() > 0 {
		// Since edge is not in boundary, may not need the full generality of
		// PointLocator? Possibly should use ptInArea locator instead? We
		// probably know here that the edge does not touch the bdy of the target
		// Geometry.
		loc := rc.ptLocator.Locate(e.GetCoordinate(), target)
		e.GetLabel().SetAllLocations(targetIndex, loc)
	} else {
		e.GetLabel().SetAllLocations(targetIndex, Geom_Location_Exterior)
	}
}

// labelIsolatedNodes labels isolated nodes (nodes whose labels are incomplete,
// e.g. the location for one Geometry is null). This is the case because nodes
// in one graph which don't intersect nodes in the other are not completely
// labelled by the initial process of adding nodes to the nodeList. To complete
// the labelling we need to check for nodes that lie in the interior of edges,
// and in the interior of areas.
func (rc *OperationRelate_RelateComputer) labelIsolatedNodes() {
	for _, n := range rc.nodes.Values() {
		label := n.GetLabel()
		// Isolated nodes should always have at least one geometry in their label.
		Util_Assert_IsTrueWithMessage(label.GetGeometryCount() > 0, "node with empty label found")
		if n.IsIsolated() {
			if label.IsNull(0) {
				rc.labelIsolatedNode(n, 0)
			} else {
				rc.labelIsolatedNode(n, 1)
			}
		}
	}
}

// labelIsolatedNode labels an isolated node with its relationship to the target
// geometry.
func (rc *OperationRelate_RelateComputer) labelIsolatedNode(n *Geomgraph_Node, targetIndex int) {
	loc := rc.ptLocator.Locate(n.GetCoordinate(), rc.arg[targetIndex].GetGeometry())
	n.GetLabel().SetAllLocations(targetIndex, loc)
}
