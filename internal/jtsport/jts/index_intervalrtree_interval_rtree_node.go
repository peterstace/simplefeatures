package jts

import (
	"fmt"
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// IndexIntervalrtree_IntervalRTreeNode is an abstract base class for nodes in
// an interval R-tree.
type IndexIntervalrtree_IntervalRTreeNode struct {
	child java.Polymorphic
	min   float64
	max   float64
}

// IndexIntervalrtree_NewIntervalRTreeNode creates a new IntervalRTreeNode with
// default min/max values.
func IndexIntervalrtree_NewIntervalRTreeNode() *IndexIntervalrtree_IntervalRTreeNode {
	return &IndexIntervalrtree_IntervalRTreeNode{
		min: math.Inf(1),
		max: math.Inf(-1),
	}
}

// GetChild returns the immediate child in the type hierarchy chain.
func (n *IndexIntervalrtree_IntervalRTreeNode) GetChild() java.Polymorphic {
	return n.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (n *IndexIntervalrtree_IntervalRTreeNode) GetParent() java.Polymorphic {
	return nil
}

// GetMin returns the minimum value of this node's interval.
func (n *IndexIntervalrtree_IntervalRTreeNode) GetMin() float64 {
	return n.min
}

// GetMax returns the maximum value of this node's interval.
func (n *IndexIntervalrtree_IntervalRTreeNode) GetMax() float64 {
	return n.max
}

// Query visits all items in this node whose intervals intersect the query
// interval.
func (n *IndexIntervalrtree_IntervalRTreeNode) Query(queryMin, queryMax float64, visitor Index_ItemVisitor) {
	if impl, ok := java.GetLeaf(n).(interface {
		Query_BODY(float64, float64, Index_ItemVisitor)
	}); ok {
		impl.Query_BODY(queryMin, queryMax, visitor)
		return
	}
	panic("abstract method called")
}

// intersects tests whether this node's interval intersects the query interval.
func (n *IndexIntervalrtree_IntervalRTreeNode) intersects(queryMin, queryMax float64) bool {
	if n.min > queryMax || n.max < queryMin {
		return false
	}
	return true
}

// String returns a WKT representation of this node's interval as a line
// segment.
func (n *IndexIntervalrtree_IntervalRTreeNode) String() string {
	return fmt.Sprintf("LINESTRING ( %v %v, %v %v )", n.min, 0.0, n.max, 0.0)
}

// IndexIntervalrtree_IntervalRTreeNode_NodeComparator compares
// IntervalRTreeNodes by the midpoint of their intervals.
type IndexIntervalrtree_IntervalRTreeNode_NodeComparator struct{}

// Compare compares two IntervalRTreeNodes by the midpoint of their intervals.
func (c *IndexIntervalrtree_IntervalRTreeNode_NodeComparator) Compare(o1, o2 any) int {
	n1 := o1.(*IndexIntervalrtree_IntervalRTreeNode)
	n2 := o2.(*IndexIntervalrtree_IntervalRTreeNode)
	mid1 := (n1.min + n1.max) / 2
	mid2 := (n2.min + n2.max) / 2
	if mid1 < mid2 {
		return -1
	}
	if mid1 > mid2 {
		return 1
	}
	return 0
}
