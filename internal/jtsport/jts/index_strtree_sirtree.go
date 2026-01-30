package jts

import (
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// IndexStrtree_SIRtreeNode is a node of an SIRtree.
type IndexStrtree_SIRtreeNode struct {
	*IndexStrtree_AbstractNode
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (n *IndexStrtree_SIRtreeNode) GetChild() java.Polymorphic {
	return n.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (n *IndexStrtree_SIRtreeNode) GetParent() java.Polymorphic {
	return n.IndexStrtree_AbstractNode
}

// IndexStrtree_NewSIRtreeNode creates a new SIRtreeNode at the given level.
func IndexStrtree_NewSIRtreeNode(level int) *IndexStrtree_SIRtreeNode {
	base := IndexStrtree_NewAbstractNode(level)
	node := &IndexStrtree_SIRtreeNode{
		IndexStrtree_AbstractNode: base,
	}
	base.child = node
	return node
}

// ComputeBounds_BODY computes the bounds of this node by expanding an interval
// to include the bounds of all child boundables.
func (n *IndexStrtree_SIRtreeNode) ComputeBounds_BODY() any {
	var bounds *IndexStrtree_Interval
	for _, childBoundable := range n.GetChildBoundables() {
		childBounds := childBoundable.GetBounds().(*IndexStrtree_Interval)
		if bounds == nil {
			bounds = IndexStrtree_NewIntervalFromInterval(childBounds)
		} else {
			bounds.ExpandToInclude(childBounds)
		}
	}
	return bounds
}

// IndexStrtree_SIRtree_comparator compares boundables by centre of interval bounds.
func IndexStrtree_SIRtree_comparator(o1, o2 IndexStrtree_Boundable) int {
	return IndexStrtree_AbstractSTRtree_CompareDoubles(
		o1.GetBounds().(*IndexStrtree_Interval).GetCentre(),
		o2.GetBounds().(*IndexStrtree_Interval).GetCentre(),
	)
}

// indexStrtree_SIRtree_intersectsOp tests whether two Intervals intersect.
type indexStrtree_SIRtree_intersectsOp struct{}

func (op *indexStrtree_SIRtree_intersectsOp) Intersects(aBounds, bBounds any) bool {
	return aBounds.(*IndexStrtree_Interval).Intersects(bBounds.(*IndexStrtree_Interval))
}

var indexStrtree_SIRtree_IntersectsOpInstance = &indexStrtree_SIRtree_intersectsOp{}

// IndexStrtree_SIRtree is a one-dimensional version of an STR-packed R-tree. SIR
// stands for "Sort-Interval-Recursive". STR-packed R-trees are described in: P.
// Rigaux, Michel Scholl and Agnes Voisard. Spatial Databases With Application
// To GIS. Morgan Kaufmann, San Francisco, 2002.
//
// This class is thread-safe. Building the tree is synchronized, and querying is
// stateless.
type IndexStrtree_SIRtree struct {
	*IndexStrtree_AbstractSTRtree
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (t *IndexStrtree_SIRtree) GetChild() java.Polymorphic {
	return t.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (t *IndexStrtree_SIRtree) GetParent() java.Polymorphic {
	return t.IndexStrtree_AbstractSTRtree
}

// IndexStrtree_NewSIRtree constructs an SIRtree with the default node capacity.
func IndexStrtree_NewSIRtree() *IndexStrtree_SIRtree {
	return IndexStrtree_NewSIRtreeWithCapacity(10)
}

// IndexStrtree_NewSIRtreeWithCapacity constructs an SIRtree with the given
// maximum number of child nodes that a node may have.
func IndexStrtree_NewSIRtreeWithCapacity(nodeCapacity int) *IndexStrtree_SIRtree {
	base := IndexStrtree_NewAbstractSTRtreeWithCapacity(nodeCapacity)
	t := &IndexStrtree_SIRtree{
		IndexStrtree_AbstractSTRtree: base,
	}
	base.child = t
	return t
}

// CreateNode_BODY creates a new SIRtreeNode at the given level.
func (t *IndexStrtree_SIRtree) CreateNode_BODY(level int) *IndexStrtree_AbstractNode {
	return IndexStrtree_NewSIRtreeNode(level).IndexStrtree_AbstractNode
}

// Insert inserts an item having the given bounds into the tree.
func (t *IndexStrtree_SIRtree) Insert(x1, x2 float64, item any) {
	t.IndexStrtree_AbstractSTRtree.Insert(IndexStrtree_NewInterval(math.Min(x1, x2), math.Max(x1, x2)), item)
}

// QueryPoint returns items whose bounds intersect the given value.
func (t *IndexStrtree_SIRtree) QueryPoint(x float64) []any {
	return t.QueryRange(x, x)
}

// QueryRange returns items whose bounds intersect the given bounds.
func (t *IndexStrtree_SIRtree) QueryRange(x1, x2 float64) []any {
	return t.IndexStrtree_AbstractSTRtree.Query(IndexStrtree_NewInterval(math.Min(x1, x2), math.Max(x1, x2)))
}

// GetIntersectsOp_BODY returns the intersects operation for SIRtree.
func (t *IndexStrtree_SIRtree) GetIntersectsOp_BODY() IndexStrtree_IntersectsOp {
	return indexStrtree_SIRtree_IntersectsOpInstance
}

// GetComparator_BODY returns the comparator used to sort boundables.
func (t *IndexStrtree_SIRtree) GetComparator_BODY() func(a, b IndexStrtree_Boundable) int {
	return IndexStrtree_SIRtree_comparator
}
