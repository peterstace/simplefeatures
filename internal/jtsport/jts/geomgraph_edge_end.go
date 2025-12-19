package jts

import (
	"fmt"
	"io"
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geomgraph_EdgeEnd models the end of an edge incident on a node. EdgeEnds
// have a direction determined by the direction of the ray from the initial
// point to the next point. EdgeEnds are comparable under the ordering "a has a
// greater angle with the x-axis than b". This ordering is used to sort
// EdgeEnds around a node.
type Geomgraph_EdgeEnd struct {
	child java.Polymorphic

	edge  *Geomgraph_Edge  // The parent edge of this edge end.
	label *Geomgraph_Label // Label for this edge end.

	node     *Geomgraph_Node // The node this edge end originates at.
	p0       *Geom_Coordinate
	p1       *Geom_Coordinate
	dx       float64 // The direction vector for this edge from its starting point.
	dy       float64
	quadrant int
}

// GetChild returns the immediate child in the type hierarchy chain.
func (ee *Geomgraph_EdgeEnd) GetChild() java.Polymorphic {
	return ee.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (ee *Geomgraph_EdgeEnd) GetParent() java.Polymorphic {
	return nil
}

// Geomgraph_NewEdgeEndFromEdge creates a new EdgeEnd with just the parent
// edge.
func Geomgraph_NewEdgeEndFromEdge(edge *Geomgraph_Edge) *Geomgraph_EdgeEnd {
	return &Geomgraph_EdgeEnd{
		edge: edge,
	}
}

// Geomgraph_NewEdgeEnd creates a new EdgeEnd with the given edge and points.
func Geomgraph_NewEdgeEnd(edge *Geomgraph_Edge, p0, p1 *Geom_Coordinate) *Geomgraph_EdgeEnd {
	return Geomgraph_NewEdgeEndWithLabel(edge, p0, p1, nil)
}

// Geomgraph_NewEdgeEndWithLabel creates a new EdgeEnd with the given edge,
// points, and label.
func Geomgraph_NewEdgeEndWithLabel(edge *Geomgraph_Edge, p0, p1 *Geom_Coordinate, label *Geomgraph_Label) *Geomgraph_EdgeEnd {
	ee := Geomgraph_NewEdgeEndFromEdge(edge)
	ee.init(p0, p1)
	ee.label = label
	return ee
}

func (ee *Geomgraph_EdgeEnd) init(p0, p1 *Geom_Coordinate) {
	ee.p0 = p0
	ee.p1 = p1
	ee.dx = p1.GetX() - p0.GetX()
	ee.dy = p1.GetY() - p0.GetY()
	ee.quadrant = Geom_Quadrant_QuadrantFromDeltas(ee.dx, ee.dy)
	Util_Assert_IsTrueWithMessage(!(ee.dx == 0 && ee.dy == 0), "EdgeEnd with identical endpoints found")
}

// GetEdge returns the parent edge of this edge end.
func (ee *Geomgraph_EdgeEnd) GetEdge() *Geomgraph_Edge {
	return ee.edge
}

// GetLabel returns the label for this edge end.
func (ee *Geomgraph_EdgeEnd) GetLabel() *Geomgraph_Label {
	return ee.label
}

// GetCoordinate returns the starting point of this edge end.
func (ee *Geomgraph_EdgeEnd) GetCoordinate() *Geom_Coordinate {
	return ee.p0
}

// GetDirectedCoordinate returns the direction point of this edge end.
func (ee *Geomgraph_EdgeEnd) GetDirectedCoordinate() *Geom_Coordinate {
	return ee.p1
}

// GetQuadrant returns the quadrant of this edge end.
func (ee *Geomgraph_EdgeEnd) GetQuadrant() int {
	return ee.quadrant
}

// GetDx returns the x component of the direction vector.
func (ee *Geomgraph_EdgeEnd) GetDx() float64 {
	return ee.dx
}

// GetDy returns the y component of the direction vector.
func (ee *Geomgraph_EdgeEnd) GetDy() float64 {
	return ee.dy
}

// SetNode sets the node this edge end originates at.
func (ee *Geomgraph_EdgeEnd) SetNode(node *Geomgraph_Node) {
	ee.node = node
}

// GetNode returns the node this edge end originates at.
func (ee *Geomgraph_EdgeEnd) GetNode() *Geomgraph_Node {
	return ee.node
}

// CompareTo compares this EdgeEnd to another for ordering.
func (ee *Geomgraph_EdgeEnd) CompareTo(other *Geomgraph_EdgeEnd) int {
	return ee.CompareDirection(other)
}

// CompareDirection implements the total order relation: a has a greater angle
// with the positive x-axis than b.
//
// Using the obvious algorithm of simply computing the angle is not robust,
// since the angle calculation is obviously susceptible to roundoff. A robust
// algorithm is:
//   - first compare the quadrant. If the quadrants are different, it is trivial
//     to determine which vector is "greater".
//   - if the vectors lie in the same quadrant, the computeOrientation function
//     can be used to decide the relative orientation of the vectors.
func (ee *Geomgraph_EdgeEnd) CompareDirection(e *Geomgraph_EdgeEnd) int {
	if ee.dx == e.dx && ee.dy == e.dy {
		return 0
	}
	// If the rays are in different quadrants, determining the ordering is
	// trivial.
	if ee.quadrant > e.quadrant {
		return 1
	}
	if ee.quadrant < e.quadrant {
		return -1
	}
	// Vectors are in the same quadrant - check relative orientation of
	// direction vectors. This is > e if it is CCW of e.
	return Algorithm_Orientation_Index(e.p0, e.p1, ee.p1)
}

// ComputeLabel computes the label for this edge end. Subclasses should override
// this if they are using labels.
func (ee *Geomgraph_EdgeEnd) ComputeLabel(boundaryNodeRule Algorithm_BoundaryNodeRule) {
	if impl, ok := java.GetLeaf(ee).(interface {
		ComputeLabel_BODY(Algorithm_BoundaryNodeRule)
	}); ok {
		impl.ComputeLabel_BODY(boundaryNodeRule)
		return
	}
	// Default implementation does nothing.
}

// String returns a string representation of this EdgeEnd.
func (ee *Geomgraph_EdgeEnd) String() string {
	angle := math.Atan2(ee.dy, ee.dx)
	return fmt.Sprintf("  EdgeEnd: %v - %v %d:%v   %v", ee.p0, ee.p1, ee.quadrant, angle, ee.label)
}

// Print writes a representation of this EdgeEnd to the given writer.
func (ee *Geomgraph_EdgeEnd) Print(out io.Writer) {
	angle := math.Atan2(ee.dy, ee.dx)
	fmt.Fprintf(out, "  EdgeEnd: %v - %v %d:%v   %v", ee.p0, ee.p1, ee.quadrant, angle, ee.label)
}
