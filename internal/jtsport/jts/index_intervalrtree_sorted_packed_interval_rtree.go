package jts

import (
	"sort"
	"sync"
)

// IndexIntervalrtree_SortedPackedIntervalRTree is a static index on a set of
// 1-dimensional intervals, using an R-Tree packed based on the order of the
// interval midpoints. It supports range searching, where the range is an
// interval of the real line (which may be a single point). A common use is to
// index 1-dimensional intervals which are the projection of 2-D objects onto
// an axis of the coordinate system.
//
// This index structure is static - items cannot be added or removed once the
// first query has been made. The advantage of this characteristic is that the
// index performance can be optimized based on a fixed set of items.
type IndexIntervalrtree_SortedPackedIntervalRTree struct {
	leaves []*IndexIntervalrtree_IntervalRTreeLeafNode
	// If root is nil that indicates that the tree has not yet been built, OR
	// nothing has been added to the tree. In both cases, the tree is still
	// open for insertions.
	root *IndexIntervalrtree_IntervalRTreeNode
	mu   sync.Mutex
}

// IndexIntervalrtree_NewSortedPackedIntervalRTree creates a new empty interval
// R-tree.
func IndexIntervalrtree_NewSortedPackedIntervalRTree() *IndexIntervalrtree_SortedPackedIntervalRTree {
	return &IndexIntervalrtree_SortedPackedIntervalRTree{}
}

// Insert adds an item to the index which is associated with the given
// interval. Panics if the index has already been queried.
func (t *IndexIntervalrtree_SortedPackedIntervalRTree) Insert(min, max float64, item any) {
	if t.root != nil {
		panic("Index cannot be added to once it has been queried")
	}
	t.leaves = append(t.leaves, IndexIntervalrtree_NewIntervalRTreeLeafNode(min, max, item))
}

func (t *IndexIntervalrtree_SortedPackedIntervalRTree) init() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Already built.
	if t.root != nil {
		return
	}

	// If leaves is empty then nothing has been inserted. In this case it is
	// safe to leave the tree in an open state.
	if len(t.leaves) == 0 {
		return
	}

	t.buildRoot()
}

func (t *IndexIntervalrtree_SortedPackedIntervalRTree) buildRoot() {
	if t.root != nil {
		return
	}
	t.root = t.buildTree()
}

func (t *IndexIntervalrtree_SortedPackedIntervalRTree) buildTree() *IndexIntervalrtree_IntervalRTreeNode {
	// Sort the leaf nodes.
	comparator := &IndexIntervalrtree_IntervalRTreeNode_NodeComparator{}
	sort.Slice(t.leaves, func(i, j int) bool {
		return comparator.Compare(t.leaves[i].IndexIntervalrtree_IntervalRTreeNode, t.leaves[j].IndexIntervalrtree_IntervalRTreeNode) < 0
	})

	// Now group nodes into blocks of two and build tree up recursively.
	src := make([]*IndexIntervalrtree_IntervalRTreeNode, len(t.leaves))
	for i, leaf := range t.leaves {
		src[i] = leaf.IndexIntervalrtree_IntervalRTreeNode
	}
	var temp []*IndexIntervalrtree_IntervalRTreeNode
	dest := make([]*IndexIntervalrtree_IntervalRTreeNode, 0)

	for {
		dest = t.buildLevel(src, dest)
		if len(dest) == 1 {
			return dest[0]
		}

		temp = src
		src = dest
		dest = temp
	}
}

func (t *IndexIntervalrtree_SortedPackedIntervalRTree) buildLevel(
	src, dest []*IndexIntervalrtree_IntervalRTreeNode,
) []*IndexIntervalrtree_IntervalRTreeNode {
	dest = dest[:0]
	for i := 0; i < len(src); i += 2 {
		n1 := src[i]
		var n2 *IndexIntervalrtree_IntervalRTreeNode
		if i+1 < len(src) {
			// Note: The original Java code has a bug here - it uses src.get(i)
			// instead of src.get(i+1). We replicate the bug for behavioral
			// equivalence.
			n2 = src[i]
		}
		if n2 == nil {
			dest = append(dest, n1)
		} else {
			node := IndexIntervalrtree_NewIntervalRTreeBranchNode(src[i], src[i+1])
			dest = append(dest, node.IndexIntervalrtree_IntervalRTreeNode)
		}
	}
	return dest
}

// Query searches for intervals in the index which intersect the given closed
// interval and applies the visitor to them.
func (t *IndexIntervalrtree_SortedPackedIntervalRTree) Query(min, max float64, visitor Index_ItemVisitor) {
	t.init()

	// If root is nil tree must be empty.
	if t.root == nil {
		return
	}

	t.root.Query(min, max, visitor)
}
