package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// IndexStrtree_AbstractNode is a node of an AbstractSTRtree. A node is one of:
//   - empty
//   - an interior node containing child AbstractNodes
//   - a leaf node containing data items (ItemBoundables).
//
// A node stores the bounds of its children, and its level within the index tree.
type IndexStrtree_AbstractNode struct {
	child           java.Polymorphic
	childBoundables []IndexStrtree_Boundable
	bounds          any
	level           int
}

// GetChild returns the immediate child in the type hierarchy chain.
func (n *IndexStrtree_AbstractNode) GetChild() java.Polymorphic {
	return n.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (n *IndexStrtree_AbstractNode) GetParent() java.Polymorphic {
	return nil
}

// IndexStrtree_NewAbstractNode constructs an AbstractNode at the given level in
// the tree. Level is 0 if this node is a leaf, 1 if a parent of a leaf, and so
// on; the root node will have the highest level.
func IndexStrtree_NewAbstractNode(level int) *IndexStrtree_AbstractNode {
	return &IndexStrtree_AbstractNode{
		childBoundables: make([]IndexStrtree_Boundable, 0),
		level:           level,
	}
}

// GetChildBoundables returns either child AbstractNodes, or if this is a leaf
// node, real data (wrapped in ItemBoundables).
func (n *IndexStrtree_AbstractNode) GetChildBoundables() []IndexStrtree_Boundable {
	return n.childBoundables
}

// ComputeBounds returns a representation of space that encloses this Boundable,
// preferably not much bigger than this Boundable's boundary yet fast to test
// for intersection with the bounds of other Boundables. The class of object
// returned depends on the subclass of AbstractSTRtree.
func (n *IndexStrtree_AbstractNode) ComputeBounds() any {
	if impl, ok := java.GetLeaf(n).(interface{ ComputeBounds_BODY() any }); ok {
		return impl.ComputeBounds_BODY()
	}
	panic("abstract method called")
}

// GetBounds gets the bounds of this node.
func (n *IndexStrtree_AbstractNode) GetBounds() any {
	if n.bounds == nil {
		n.bounds = n.ComputeBounds()
	}
	return n.bounds
}

// TRANSLITERATION NOTE: Marker method for Boundable interface. Not present in
// Java source.
func (n *IndexStrtree_AbstractNode) IsIndexStrtree_Boundable() {}

// GetLevel returns 0 if this node is a leaf, 1 if a parent of a leaf, and so
// on; the root node will have the highest level.
func (n *IndexStrtree_AbstractNode) GetLevel() int {
	return n.level
}

// Size gets the count of the Boundables at this node.
func (n *IndexStrtree_AbstractNode) Size() int {
	return len(n.childBoundables)
}

// IsEmpty tests whether there are any Boundables at this node.
func (n *IndexStrtree_AbstractNode) IsEmpty() bool {
	return len(n.childBoundables) == 0
}

// AddChildBoundable adds either an AbstractNode, or if this is a leaf node, a
// data object (wrapped in an ItemBoundable).
func (n *IndexStrtree_AbstractNode) AddChildBoundable(childBoundable IndexStrtree_Boundable) {
	Util_Assert_IsTrue(n.bounds == nil)
	n.childBoundables = append(n.childBoundables, childBoundable)
}

// SetChildBoundables sets the child boundables list.
func (n *IndexStrtree_AbstractNode) SetChildBoundables(childBoundables []IndexStrtree_Boundable) {
	n.childBoundables = childBoundables
}
