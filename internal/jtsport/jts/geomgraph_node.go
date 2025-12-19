package jts

import (
	"fmt"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// Geomgraph_Node represents a node in the topology graph.
type Geomgraph_Node struct {
	*Geomgraph_GraphComponent
	child java.Polymorphic

	coord *Geom_Coordinate // Only non-null if this node is precise.
	edges *Geomgraph_EdgeEndStar
}

// GetChild returns the immediate child in the type hierarchy chain.
func (n *Geomgraph_Node) GetChild() java.Polymorphic {
	return n.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (n *Geomgraph_Node) GetParent() java.Polymorphic {
	return n.Geomgraph_GraphComponent
}

// Geomgraph_NewNode creates a new Node with the given coordinate and edges.
func Geomgraph_NewNode(coord *Geom_Coordinate, edges *Geomgraph_EdgeEndStar) *Geomgraph_Node {
	gc := Geomgraph_NewGraphComponent()
	node := &Geomgraph_Node{
		Geomgraph_GraphComponent: gc,
		coord:                   coord,
		edges:                   edges,
	}
	gc.child = node
	gc.label = Geomgraph_NewLabelGeomOn(0, Geom_Location_None)
	return node
}

// GetCoordinate_BODY returns the coordinate of this node.
func (n *Geomgraph_Node) GetCoordinate_BODY() *Geom_Coordinate {
	return n.coord
}

// GetEdges returns the EdgeEndStar for this node.
func (n *Geomgraph_Node) GetEdges() *Geomgraph_EdgeEndStar {
	return n.edges
}

// IsIncidentEdgeInResult tests whether any incident edge is flagged as being
// in the result. This test can be used to determine if the node is in the
// result, since if any incident edge is in the result, the node must be in the
// result as well.
func (n *Geomgraph_Node) IsIncidentEdgeInResult() bool {
	for _, ee := range n.GetEdges().GetEdges() {
		de := java.Cast[*Geomgraph_DirectedEdge](ee)
		if de.GetEdge().IsInResult() {
			return true
		}
	}
	return false
}

// IsIsolated_BODY returns true if this is an isolated node.
func (n *Geomgraph_Node) IsIsolated_BODY() bool {
	return n.label.GetGeometryCount() == 1
}

// ComputeIM_BODY computes the contribution to an IM for this component. Basic
// nodes do not compute IMs.
func (n *Geomgraph_Node) ComputeIM_BODY(im *Geom_IntersectionMatrix) {
	// Basic nodes do not compute IMs.
}

// Add adds the edge to the list of edges at this node.
func (n *Geomgraph_Node) Add(e *Geomgraph_EdgeEnd) {
	// Assert: start pt of e is equal to node point.
	n.edges.Insert(e)
	e.SetNode(n)
}

// MergeLabel merges the label from another node.
func (n *Geomgraph_Node) MergeLabel(other *Geomgraph_Node) {
	n.MergeLabelFromLabel(other.label)
}

// MergeLabelFromLabel merges the label from another label. To merge labels for
// two nodes, the merged location for each LabelElement is computed. The
// location for the corresponding node LabelElement is set to the result, as
// long as the location is non-null.
func (n *Geomgraph_Node) MergeLabelFromLabel(label2 *Geomgraph_Label) {
	for i := 0; i < 2; i++ {
		loc := n.computeMergedLocation(label2, i)
		thisLoc := n.label.GetLocationOn(i)
		if thisLoc == Geom_Location_None {
			n.label.SetLocationOn(i, loc)
		}
	}
}

// SetLabelAt sets the label at the given argument index to the given location.
func (n *Geomgraph_Node) SetLabelAt(argIndex, onLocation int) {
	if n.label == nil {
		n.label = Geomgraph_NewLabelGeomOn(argIndex, onLocation)
	} else {
		n.label.SetLocationOn(argIndex, onLocation)
	}
}

// SetLabelBoundary updates the label of a node to BOUNDARY, obeying the mod-2
// boundary determination rule.
func (n *Geomgraph_Node) SetLabelBoundary(argIndex int) {
	if n.label == nil {
		return
	}

	// Determine the current location for the point (if any).
	loc := Geom_Location_None
	if n.label != nil {
		loc = n.label.GetLocationOn(argIndex)
	}
	// Flip the loc.
	var newLoc int
	switch loc {
	case Geom_Location_Boundary:
		newLoc = Geom_Location_Interior
	case Geom_Location_Interior:
		newLoc = Geom_Location_Boundary
	default:
		newLoc = Geom_Location_Boundary
	}
	n.label.SetLocationOn(argIndex, newLoc)
}

// computeMergedLocation computes the merged location for a given element
// index. The location for a given eltIndex for a node will be one of { null,
// INTERIOR, BOUNDARY }. A node may be on both the boundary and the interior of
// a geometry; in this case, the rule is that the node is considered to be in
// the boundary. The merged location is the maximum of the two input values.
func (n *Geomgraph_Node) computeMergedLocation(label2 *Geomgraph_Label, eltIndex int) int {
	loc := n.label.GetLocationOn(eltIndex)
	if !label2.IsNull(eltIndex) {
		nLoc := label2.GetLocationOn(eltIndex)
		if loc != Geom_Location_Boundary {
			loc = nLoc
		}
	}
	return loc
}

// String returns a string representation of this Node.
func (n *Geomgraph_Node) String() string {
	return fmt.Sprintf("Node(%v, %v)", n.coord.GetX(), n.coord.GetY())
}
