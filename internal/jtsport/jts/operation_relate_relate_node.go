package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationRelate_RelateNode represents a node in the topological graph used
// to compute spatial relationships.
type OperationRelate_RelateNode struct {
	*Geomgraph_Node
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (rn *OperationRelate_RelateNode) GetChild() java.Polymorphic {
	return rn.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (rn *OperationRelate_RelateNode) GetParent() java.Polymorphic {
	return rn.Geomgraph_Node
}

// OperationRelate_NewRelateNode creates a new RelateNode with the given
// coordinate and edges.
func OperationRelate_NewRelateNode(coord *Geom_Coordinate, edges *Geomgraph_EdgeEndStar) *OperationRelate_RelateNode {
	node := Geomgraph_NewNode(coord, edges)
	rn := &OperationRelate_RelateNode{
		Geomgraph_Node: node,
	}
	node.child = rn
	return rn
}

// ComputeIM_BODY updates the IM with the contribution for this component. A
// component only contributes if it has a labelling for both parent geometries.
func (rn *OperationRelate_RelateNode) ComputeIM_BODY(im *Geom_IntersectionMatrix) {
	im.SetAtLeastIfValid(rn.label.GetLocationOn(0), rn.label.GetLocationOn(1), 0)
}

// UpdateIMFromEdges updates the IM with the contribution for the EdgeEnds
// incident on this node.
func (rn *OperationRelate_RelateNode) UpdateIMFromEdges(im *Geom_IntersectionMatrix) {
	if eebs, ok := rn.edges.GetChild().(*OperationRelate_EdgeEndBundleStar); ok {
		eebs.UpdateIM(im)
	}
}
