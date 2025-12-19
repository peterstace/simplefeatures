package jts

import (
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// IndexIntervalrtree_IntervalRTreeBranchNode is a branch node in an interval
// R-tree that contains two child nodes.
type IndexIntervalrtree_IntervalRTreeBranchNode struct {
	*IndexIntervalrtree_IntervalRTreeNode
	node1 *IndexIntervalrtree_IntervalRTreeNode
	node2 *IndexIntervalrtree_IntervalRTreeNode
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (n *IndexIntervalrtree_IntervalRTreeBranchNode) GetChild() java.Polymorphic {
	return n.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (n *IndexIntervalrtree_IntervalRTreeBranchNode) GetParent() java.Polymorphic {
	return n.IndexIntervalrtree_IntervalRTreeNode
}

// IndexIntervalrtree_NewIntervalRTreeBranchNode creates a new branch node with
// the given child nodes.
func IndexIntervalrtree_NewIntervalRTreeBranchNode(
	n1, n2 *IndexIntervalrtree_IntervalRTreeNode,
) *IndexIntervalrtree_IntervalRTreeBranchNode {
	base := IndexIntervalrtree_NewIntervalRTreeNode()
	branch := &IndexIntervalrtree_IntervalRTreeBranchNode{
		IndexIntervalrtree_IntervalRTreeNode: base,
		node1:                               n1,
		node2:                               n2,
	}
	branch.buildExtent(n1, n2)
	base.child = branch
	return branch
}

// buildExtent computes this node's extent from its children.
func (n *IndexIntervalrtree_IntervalRTreeBranchNode) buildExtent(
	n1, n2 *IndexIntervalrtree_IntervalRTreeNode,
) {
	n.min = math.Min(n1.min, n2.min)
	n.max = math.Max(n1.max, n2.max)
}

// Query_BODY queries both children if this branch's interval intersects the
// query interval.
func (n *IndexIntervalrtree_IntervalRTreeBranchNode) Query_BODY(queryMin, queryMax float64, visitor Index_ItemVisitor) {
	if !n.intersects(queryMin, queryMax) {
		return
	}
	if n.node1 != nil {
		n.node1.Query(queryMin, queryMax, visitor)
	}
	if n.node2 != nil {
		n.node2.Query(queryMin, queryMax, visitor)
	}
}
