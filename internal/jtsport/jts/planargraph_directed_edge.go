package jts

import (
	"fmt"
	"io"
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Planargraph_DirectedEdge represents a directed edge in a PlanarGraph. A DirectedEdge may or
// may not have a reference to a parent Edge (some applications of
// planar graphs may not require explicit Edge objects to be created). Usually
// a client using a PlanarGraph will subclass DirectedEdge
// to add its own application-specific data and methods.
type Planargraph_DirectedEdge struct {
	*Planargraph_GraphComponent
	child         java.Polymorphic
	parentEdge    *Planargraph_Edge
	from          *Planargraph_Node
	to            *Planargraph_Node
	p0            *Geom_Coordinate
	p1            *Geom_Coordinate
	sym           *Planargraph_DirectedEdge
	edgeDirection bool
	quadrant      int
	angle         float64
}

func (de *Planargraph_DirectedEdge) GetChild() java.Polymorphic {
	return de.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (de *Planargraph_DirectedEdge) GetParent() java.Polymorphic {
	return de.Planargraph_GraphComponent
}

// Planargraph_DirectedEdge_ToEdges returns a slice containing the parent Edge (possibly nil) for each of the given
// DirectedEdges.
func Planargraph_DirectedEdge_ToEdges(dirEdges []*Planargraph_DirectedEdge) []*Planargraph_Edge {
	edges := make([]*Planargraph_Edge, len(dirEdges))
	for i, de := range dirEdges {
		edges[i] = de.parentEdge
	}
	return edges
}

// Planargraph_NewDirectedEdge constructs a DirectedEdge connecting the from node to the
// to node.
//
// directionPt specifies this DirectedEdge's direction vector
// (determined by the vector from the from node to directionPt).
//
// edgeDirection indicates whether this DirectedEdge's direction is the same as or
// opposite to that of the parent Edge (if any).
func Planargraph_NewDirectedEdge(from, to *Planargraph_Node, directionPt *Geom_Coordinate, edgeDirection bool) *Planargraph_DirectedEdge {
	gc := &Planargraph_GraphComponent{}
	de := &Planargraph_DirectedEdge{
		Planargraph_GraphComponent: gc,
		from:                      from,
		to:                        to,
		edgeDirection:             edgeDirection,
		p0:                        from.GetCoordinate(),
		p1:                        directionPt,
	}
	gc.child = de
	dx := de.p1.GetX() - de.p0.GetX()
	dy := de.p1.GetY() - de.p0.GetY()
	de.quadrant = Geom_Quadrant_QuadrantFromDeltas(dx, dy)
	de.angle = math.Atan2(dy, dx)
	return de
}

// GetEdge returns this DirectedEdge's parent Edge, or nil if it has none.
func (de *Planargraph_DirectedEdge) GetEdge() *Planargraph_Edge {
	return de.parentEdge
}

// SetEdge associates this DirectedEdge with an Edge (possibly nil, indicating no associated Edge).
func (de *Planargraph_DirectedEdge) SetEdge(parentEdge *Planargraph_Edge) {
	de.parentEdge = parentEdge
}

// GetQuadrant returns 0, 1, 2, or 3, indicating the quadrant in which this DirectedEdge's
// orientation lies.
func (de *Planargraph_DirectedEdge) GetQuadrant() int {
	return de.quadrant
}

// GetDirectionPt returns a point to which an imaginary line is drawn from the from-node to
// specify this DirectedEdge's orientation.
func (de *Planargraph_DirectedEdge) GetDirectionPt() *Geom_Coordinate {
	return de.p1
}

// GetEdgeDirection returns whether the direction of the parent Edge (if any) is the same as that
// of this Directed Edge.
func (de *Planargraph_DirectedEdge) GetEdgeDirection() bool {
	return de.edgeDirection
}

// GetFromNode returns the node from which this DirectedEdge leaves.
func (de *Planargraph_DirectedEdge) GetFromNode() *Planargraph_Node {
	return de.from
}

// GetToNode returns the node to which this DirectedEdge goes.
func (de *Planargraph_DirectedEdge) GetToNode() *Planargraph_Node {
	return de.to
}

// GetCoordinate returns the coordinate of the from-node.
func (de *Planargraph_DirectedEdge) GetCoordinate() *Geom_Coordinate {
	return de.from.GetCoordinate()
}

// GetAngle returns the angle that the start of this DirectedEdge makes with the
// positive x-axis, in radians.
func (de *Planargraph_DirectedEdge) GetAngle() float64 {
	return de.angle
}

// GetSym returns the symmetric DirectedEdge -- the other DirectedEdge associated with
// this DirectedEdge's parent Edge.
func (de *Planargraph_DirectedEdge) GetSym() *Planargraph_DirectedEdge {
	return de.sym
}

// SetSym sets this DirectedEdge's symmetric DirectedEdge, which runs in the opposite direction.
func (de *Planargraph_DirectedEdge) SetSym(sym *Planargraph_DirectedEdge) {
	de.sym = sym
}

// remove removes this directed edge from its containing graph.
func (de *Planargraph_DirectedEdge) remove() {
	de.sym = nil
	de.parentEdge = nil
}

// IsRemoved_BODY tests whether this directed edge has been removed from its containing graph.
func (de *Planargraph_DirectedEdge) IsRemoved_BODY() bool {
	return de.parentEdge == nil
}

// CompareTo returns 1 if this DirectedEdge has a greater angle with the
// positive x-axis than other, 0 if the DirectedEdges are collinear, and -1 otherwise.
//
// Using the obvious algorithm of simply computing the angle is not robust,
// since the angle calculation is susceptible to roundoff. A robust algorithm
// is:
//   - first compare the quadrants. If the quadrants are different, it is
//     trivial to determine which vector is "greater".
//   - if the vectors lie in the same quadrant, the robust
//     Orientation.Index function can be used to decide the relative orientation of the vectors.
func (de *Planargraph_DirectedEdge) CompareTo(other *Planargraph_DirectedEdge) int {
	return de.CompareDirection(other)
}

// CompareDirection returns 1 if this DirectedEdge has a greater angle with the
// positive x-axis than e, 0 if the DirectedEdges are collinear, and -1 otherwise.
func (de *Planargraph_DirectedEdge) CompareDirection(e *Planargraph_DirectedEdge) int {
	// If the rays are in different quadrants, determining the ordering is trivial.
	if de.quadrant > e.quadrant {
		return 1
	}
	if de.quadrant < e.quadrant {
		return -1
	}
	// Vectors are in the same quadrant - check relative orientation of direction vectors.
	// This is > e if it is CCW of e.
	return Algorithm_Orientation_Index(e.p0, e.p1, de.p1)
}

// Print prints a detailed string representation of this DirectedEdge to the given writer.
func (de *Planargraph_DirectedEdge) Print(out io.Writer) {
	fmt.Fprintf(out, "  Planargraph_DirectedEdge: %v - %v %d:%v", de.p0, de.p1, de.quadrant, de.angle)
}
