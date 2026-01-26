package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// IndexIntervalrtree_IntervalRTreeLeafNode is a leaf node in an interval
// R-tree that stores an item.
type IndexIntervalrtree_IntervalRTreeLeafNode struct {
	*IndexIntervalrtree_IntervalRTreeNode
	item  any
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (n *IndexIntervalrtree_IntervalRTreeLeafNode) GetChild() java.Polymorphic {
	return n.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (n *IndexIntervalrtree_IntervalRTreeLeafNode) GetParent() java.Polymorphic {
	return n.IndexIntervalrtree_IntervalRTreeNode
}

// IndexIntervalrtree_NewIntervalRTreeLeafNode creates a new leaf node with the
// given interval and item.
func IndexIntervalrtree_NewIntervalRTreeLeafNode(min, max float64, item any) *IndexIntervalrtree_IntervalRTreeLeafNode {
	base := IndexIntervalrtree_NewIntervalRTreeNode()
	base.min = min
	base.max = max
	leaf := &IndexIntervalrtree_IntervalRTreeLeafNode{
		IndexIntervalrtree_IntervalRTreeNode: base,
		item:                                item,
	}
	base.child = leaf
	return leaf
}

// Query_BODY visits this leaf's item if the leaf's interval intersects the
// query interval.
func (n *IndexIntervalrtree_IntervalRTreeLeafNode) Query_BODY(queryMin, queryMax float64, visitor Index_ItemVisitor) {
	if !n.intersects(queryMin, queryMax) {
		return
	}
	visitor.VisitItem(n.item)
}
