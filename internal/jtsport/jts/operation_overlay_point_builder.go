package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationOverlay_PointBuilder constructs Points from the nodes of an overlay
// graph.
type OperationOverlay_PointBuilder struct {
	child java.Polymorphic

	op              *OperationOverlay_OverlayOp
	geometryFactory *Geom_GeometryFactory
	resultPointList []*Geom_Point
}

// GetChild returns the immediate child in the type hierarchy chain.
func (pb *OperationOverlay_PointBuilder) GetChild() java.Polymorphic {
	return pb.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (pb *OperationOverlay_PointBuilder) GetParent() java.Polymorphic {
	return nil
}

// OperationOverlay_NewPointBuilder creates a new PointBuilder.
func OperationOverlay_NewPointBuilder(op *OperationOverlay_OverlayOp, geometryFactory *Geom_GeometryFactory, ptLocator *Algorithm_PointLocator) *OperationOverlay_PointBuilder {
	// ptLocator is never used in this class.
	return &OperationOverlay_PointBuilder{
		op:              op,
		geometryFactory: geometryFactory,
	}
}

// Build computes the Point geometries which will appear in the result, given
// the specified overlay operation.
func (pb *OperationOverlay_PointBuilder) Build(opCode int) []*Geom_Point {
	pb.extractNonCoveredResultNodes(opCode)
	return pb.resultPointList
}

// extractNonCoveredResultNodes determines nodes which are in the result, and
// creates Points for them. This method determines nodes which are candidates
// for the result via their labelling and their graph topology.
func (pb *OperationOverlay_PointBuilder) extractNonCoveredResultNodes(opCode int) {
	for _, n := range pb.op.GetGraph().GetNodes() {
		// Filter out nodes which are known to be in the result.
		if n.IsInResult() {
			continue
		}
		// If an incident edge is in the result, then the node coordinate is
		// included already.
		if n.IsIncidentEdgeInResult() {
			continue
		}
		if n.GetEdges().GetDegree() == 0 || opCode == OperationOverlay_OverlayOp_Intersection {
			// For nodes on edges, only INTERSECTION can result in edge nodes
			// being included even if none of their incident edges are included.
			label := n.GetLabel()
			if OperationOverlay_OverlayOp_IsResultOfOpLabel(label, opCode) {
				pb.filterCoveredNodeToPoint(n)
			}
		}
	}
}

// filterCoveredNodeToPoint converts non-covered nodes to Point objects and
// adds them to the result. A node is covered if it is contained in another
// element Geometry with higher dimension (e.g. a node point might be contained
// in a polygon, in which case the point can be eliminated from the result).
func (pb *OperationOverlay_PointBuilder) filterCoveredNodeToPoint(n *Geomgraph_Node) {
	coord := n.GetCoordinate()
	if !pb.op.IsCoveredByLA(coord) {
		pt := pb.geometryFactory.CreatePointFromCoordinate(coord)
		pb.resultPointList = append(pb.resultPointList, pt)
	}
}
