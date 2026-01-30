package jts

// OperationOverlayng_IntersectionPointBuilder extracts Point resultants from an
// overlay graph created by an Intersection operation between non-Point inputs.
// Points may be created during intersection if lines or areas touch one another
// at single points. Intersection is the only overlay operation which can result
// in Points from non-Point inputs.
type OperationOverlayng_IntersectionPointBuilder struct {
	geometryFactory      *Geom_GeometryFactory
	graph                *OperationOverlayng_OverlayGraph
	points               []*Geom_Point
	isAllowCollapseLines bool
}

// OperationOverlayng_NewIntersectionPointBuilder creates a new
// IntersectionPointBuilder.
func OperationOverlayng_NewIntersectionPointBuilder(graph *OperationOverlayng_OverlayGraph, geomFact *Geom_GeometryFactory) *OperationOverlayng_IntersectionPointBuilder {
	return &OperationOverlayng_IntersectionPointBuilder{
		graph:                graph,
		geometryFactory:      geomFact,
		points:               make([]*Geom_Point, 0),
		isAllowCollapseLines: !OperationOverlayng_OverlayNG_STRICT_MODE_DEFAULT,
	}
}

// SetStrictMode sets strict mode for the point builder.
func (ipb *OperationOverlayng_IntersectionPointBuilder) SetStrictMode(isStrictMode bool) {
	ipb.isAllowCollapseLines = !isStrictMode
}

// GetPoints returns the result points from the overlay graph.
func (ipb *OperationOverlayng_IntersectionPointBuilder) GetPoints() []*Geom_Point {
	ipb.addResultPoints()
	return ipb.points
}

func (ipb *OperationOverlayng_IntersectionPointBuilder) addResultPoints() {
	for _, nodeEdge := range ipb.graph.GetNodeEdges() {
		if ipb.isResultPoint(nodeEdge) {
			pt := ipb.geometryFactory.CreatePointFromCoordinate(nodeEdge.GetCoordinate().Copy())
			ipb.points = append(ipb.points, pt)
		}
	}
}

// isResultPoint tests if a node is a result point. This is the case if the
// node is incident on edges from both inputs, and none of the edges are
// themselves in the result.
func (ipb *OperationOverlayng_IntersectionPointBuilder) isResultPoint(nodeEdge *OperationOverlayng_OverlayEdge) bool {
	isEdgeOfA := false
	isEdgeOfB := false

	edge := nodeEdge
	for {
		if edge.IsInResult() {
			return false
		}
		label := edge.GetLabel()
		isEdgeOfA = isEdgeOfA || ipb.isEdgeOf(label, 0)
		isEdgeOfB = isEdgeOfB || ipb.isEdgeOf(label, 1)
		edge = edge.ONextOE()
		if edge == nodeEdge {
			break
		}
	}
	isNodeInBoth := isEdgeOfA && isEdgeOfB
	return isNodeInBoth
}

func (ipb *OperationOverlayng_IntersectionPointBuilder) isEdgeOf(label *OperationOverlayng_OverlayLabel, i int) bool {
	if !ipb.isAllowCollapseLines && label.IsBoundaryCollapse() {
		return false
	}
	return label.IsBoundary(i) || label.IsLineIndex(i)
}
