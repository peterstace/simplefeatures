package jts

import (
	"sort"
	"sync"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

const IndexStrtree_AbstractSTRtree_DEFAULT_NODE_CAPACITY = 10

// IndexStrtree_IntersectsOp is a test for intersection between two bounds,
// necessary because subclasses of AbstractSTRtree have different
// implementations of bounds.
type IndexStrtree_IntersectsOp interface {
	// Intersects tests whether two bounds intersect. For STRtrees, the bounds
	// will be Envelopes; for SIRtrees, Intervals; for other subclasses of
	// AbstractSTRtree, some other class.
	Intersects(aBounds, bBounds any) bool
}

// IndexStrtree_AbstractSTRtree is the base class for STRtree and SIRtree.
// STR-packed R-trees are described in: P. Rigaux, Michel Scholl and Agnes
// Voisard. Spatial Databases With Application To GIS. Morgan Kaufmann, San
// Francisco, 2002.
//
// This implementation is based on Boundables rather than AbstractNodes, because
// the STR algorithm operates on both nodes and data, both of which are treated
// as Boundables.
//
// This class is thread-safe. Building the tree is synchronized, and querying is
// stateless.
type IndexStrtree_AbstractSTRtree struct {
	child          java.Polymorphic
	root           *IndexStrtree_AbstractNode
	built          bool
	itemBoundables []IndexStrtree_Boundable
	nodeCapacity   int
	mu             sync.Mutex
}

// IndexStrtree_NewAbstractSTRtree constructs an AbstractSTRtree with the default
// node capacity.
func IndexStrtree_NewAbstractSTRtree() *IndexStrtree_AbstractSTRtree {
	return IndexStrtree_NewAbstractSTRtreeWithCapacity(IndexStrtree_AbstractSTRtree_DEFAULT_NODE_CAPACITY)
}

// IndexStrtree_NewAbstractSTRtreeWithCapacity constructs an AbstractSTRtree with
// the specified maximum number of child nodes that a node may have.
func IndexStrtree_NewAbstractSTRtreeWithCapacity(nodeCapacity int) *IndexStrtree_AbstractSTRtree {
	Util_Assert_IsTrueWithMessage(nodeCapacity > 1, "Node capacity must be greater than 1")
	return &IndexStrtree_AbstractSTRtree{
		itemBoundables: make([]IndexStrtree_Boundable, 0),
		nodeCapacity:   nodeCapacity,
	}
}

// IndexStrtree_NewAbstractSTRtreeWithCapacityAndRoot constructs an AbstractSTRtree
// with the specified maximum number of child nodes that a node may have, and
// the root node.
func IndexStrtree_NewAbstractSTRtreeWithCapacityAndRoot(nodeCapacity int, root *IndexStrtree_AbstractNode) *IndexStrtree_AbstractSTRtree {
	t := IndexStrtree_NewAbstractSTRtreeWithCapacity(nodeCapacity)
	t.built = true
	t.root = root
	t.itemBoundables = nil
	return t
}

// IndexStrtree_NewAbstractSTRtreeWithCapacityAndItems constructs an AbstractSTRtree
// with the specified maximum number of child nodes that a node may have, and
// all leaf nodes in the tree.
func IndexStrtree_NewAbstractSTRtreeWithCapacityAndItems(nodeCapacity int, itemBoundables []IndexStrtree_Boundable) *IndexStrtree_AbstractSTRtree {
	t := IndexStrtree_NewAbstractSTRtreeWithCapacity(nodeCapacity)
	t.itemBoundables = itemBoundables
	return t
}

// GetChild returns the immediate child in the type hierarchy chain.
func (t *IndexStrtree_AbstractSTRtree) GetChild() java.Polymorphic {
	return t.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (t *IndexStrtree_AbstractSTRtree) GetParent() java.Polymorphic {
	return nil
}

// Build creates parent nodes, grandparent nodes, and so forth up to the root
// node, for the data that has been inserted into the tree. Can only be called
// once, and thus can be called only after all of the data has been inserted
// into the tree.
func (t *IndexStrtree_AbstractSTRtree) Build() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.built {
		return
	}
	if len(t.itemBoundables) == 0 {
		t.root = t.CreateNode(0)
	} else {
		t.root = t.createHigherLevels(t.itemBoundables, -1)
	}
	// The item list is no longer needed.
	t.itemBoundables = nil
	t.built = true
}

// CreateNode creates a node at the given level.
func (t *IndexStrtree_AbstractSTRtree) CreateNode(level int) *IndexStrtree_AbstractNode {
	if impl, ok := java.GetLeaf(t).(interface {
		CreateNode_BODY(int) *IndexStrtree_AbstractNode
	}); ok {
		return impl.CreateNode_BODY(level)
	}
	panic("abstract method called")
}

// CreateParentBoundables sorts the childBoundables then divides them into
// groups of size M, where M is the node capacity.
func (t *IndexStrtree_AbstractSTRtree) CreateParentBoundables(childBoundables []IndexStrtree_Boundable, newLevel int) []IndexStrtree_Boundable {
	Util_Assert_IsTrue(len(childBoundables) > 0)
	parentBoundables := make([]IndexStrtree_Boundable, 0)
	parentBoundables = append(parentBoundables, t.CreateNode(newLevel))

	sortedChildBoundables := make([]IndexStrtree_Boundable, len(childBoundables))
	copy(sortedChildBoundables, childBoundables)
	comparator := t.GetComparator()
	sort.Slice(sortedChildBoundables, func(i, j int) bool {
		return comparator(sortedChildBoundables[i], sortedChildBoundables[j]) < 0
	})

	for _, childBoundable := range sortedChildBoundables {
		lastNode := t.lastNode(parentBoundables)
		if len(lastNode.GetChildBoundables()) == t.GetNodeCapacity() {
			parentBoundables = append(parentBoundables, t.CreateNode(newLevel))
		}
		t.lastNode(parentBoundables).AddChildBoundable(childBoundable)
	}
	return parentBoundables
}

func (t *IndexStrtree_AbstractSTRtree) lastNode(nodes []IndexStrtree_Boundable) *IndexStrtree_AbstractNode {
	return nodes[len(nodes)-1].(*IndexStrtree_AbstractNode)
}

// IndexStrtree_AbstractSTRtree_CompareDoubles compares two double values.
func IndexStrtree_AbstractSTRtree_CompareDoubles(a, b float64) int {
	if a > b {
		return 1
	}
	if a < b {
		return -1
	}
	return 0
}

// createHigherLevels creates the levels higher than the given level.
func (t *IndexStrtree_AbstractSTRtree) createHigherLevels(boundablesOfALevel []IndexStrtree_Boundable, level int) *IndexStrtree_AbstractNode {
	Util_Assert_IsTrue(len(boundablesOfALevel) > 0)
	parentBoundables := t.CreateParentBoundables(boundablesOfALevel, level+1)
	if len(parentBoundables) == 1 {
		return parentBoundables[0].(*IndexStrtree_AbstractNode)
	}
	return t.createHigherLevels(parentBoundables, level+1)
}

// GetRoot gets the root node of the tree.
func (t *IndexStrtree_AbstractSTRtree) GetRoot() *IndexStrtree_AbstractNode {
	t.Build()
	return t.root
}

// GetNodeCapacity returns the maximum number of child nodes that a node may have.
func (t *IndexStrtree_AbstractSTRtree) GetNodeCapacity() int {
	return t.nodeCapacity
}

// IsEmpty tests whether the index contains any items. This method does not
// build the index, so items can still be inserted after it has been called.
func (t *IndexStrtree_AbstractSTRtree) IsEmpty() bool {
	if !t.built {
		return len(t.itemBoundables) == 0
	}
	return t.root.IsEmpty()
}

// Size returns the number of items in the tree.
func (t *IndexStrtree_AbstractSTRtree) Size() int {
	if t.IsEmpty() {
		return 0
	}
	t.Build()
	return t.sizeNode(t.root)
}

func (t *IndexStrtree_AbstractSTRtree) sizeNode(node *IndexStrtree_AbstractNode) int {
	size := 0
	for _, childBoundable := range node.GetChildBoundables() {
		if childNode, ok := childBoundable.(*IndexStrtree_AbstractNode); ok {
			size += t.sizeNode(childNode)
		} else if _, ok := childBoundable.(*IndexStrtree_ItemBoundable); ok {
			size++
		}
	}
	return size
}

// Depth returns the depth of the tree.
func (t *IndexStrtree_AbstractSTRtree) Depth() int {
	if t.IsEmpty() {
		return 0
	}
	t.Build()
	return t.depthNode(t.root)
}

func (t *IndexStrtree_AbstractSTRtree) depthNode(node *IndexStrtree_AbstractNode) int {
	maxChildDepth := 0
	for _, childBoundable := range node.GetChildBoundables() {
		if childNode, ok := childBoundable.(*IndexStrtree_AbstractNode); ok {
			childDepth := t.depthNode(childNode)
			if childDepth > maxChildDepth {
				maxChildDepth = childDepth
			}
		}
	}
	return maxChildDepth + 1
}

// Insert inserts an item with the given bounds into the tree.
func (t *IndexStrtree_AbstractSTRtree) Insert(bounds, item any) {
	Util_Assert_IsTrueWithMessage(!t.built, "Cannot insert items into an STR packed R-tree after it has been built.")
	t.itemBoundables = append(t.itemBoundables, IndexStrtree_NewItemBoundable(bounds, item))
}

// Query returns items whose bounds intersect the given search bounds.
// Also builds the tree, if necessary.
func (t *IndexStrtree_AbstractSTRtree) Query(searchBounds any) []any {
	t.Build()
	matches := make([]any, 0)
	if t.IsEmpty() {
		return matches
	}
	if t.GetIntersectsOp().Intersects(t.root.GetBounds(), searchBounds) {
		t.queryInternal(searchBounds, t.root, &matches)
	}
	return matches
}

// QueryWithVisitor queries the tree using a visitor.
// Also builds the tree, if necessary.
func (t *IndexStrtree_AbstractSTRtree) QueryWithVisitor(searchBounds any, visitor Index_ItemVisitor) {
	t.Build()
	if t.IsEmpty() {
		return
	}
	if t.GetIntersectsOp().Intersects(t.root.GetBounds(), searchBounds) {
		t.queryInternalWithVisitor(searchBounds, t.root, visitor)
	}
}

// GetIntersectsOp returns a test for intersection between two bounds.
func (t *IndexStrtree_AbstractSTRtree) GetIntersectsOp() IndexStrtree_IntersectsOp {
	if impl, ok := java.GetLeaf(t).(interface {
		GetIntersectsOp_BODY() IndexStrtree_IntersectsOp
	}); ok {
		return impl.GetIntersectsOp_BODY()
	}
	panic("abstract method called")
}

func (t *IndexStrtree_AbstractSTRtree) queryInternal(searchBounds any, node *IndexStrtree_AbstractNode, matches *[]any) {
	childBoundables := node.GetChildBoundables()
	for i := 0; i < len(childBoundables); i++ {
		childBoundable := childBoundables[i]
		if !t.GetIntersectsOp().Intersects(childBoundable.GetBounds(), searchBounds) {
			continue
		}
		if childNode, ok := childBoundable.(*IndexStrtree_AbstractNode); ok {
			t.queryInternal(searchBounds, childNode, matches)
		} else if itemBoundable, ok := childBoundable.(*IndexStrtree_ItemBoundable); ok {
			*matches = append(*matches, itemBoundable.GetItem())
		} else {
			Util_Assert_ShouldNeverReachHere()
		}
	}
}

func (t *IndexStrtree_AbstractSTRtree) queryInternalWithVisitor(searchBounds any, node *IndexStrtree_AbstractNode, visitor Index_ItemVisitor) {
	childBoundables := node.GetChildBoundables()
	for i := 0; i < len(childBoundables); i++ {
		childBoundable := childBoundables[i]
		if !t.GetIntersectsOp().Intersects(childBoundable.GetBounds(), searchBounds) {
			continue
		}
		if childNode, ok := childBoundable.(*IndexStrtree_AbstractNode); ok {
			t.queryInternalWithVisitor(searchBounds, childNode, visitor)
		} else if itemBoundable, ok := childBoundable.(*IndexStrtree_ItemBoundable); ok {
			visitor.VisitItem(itemBoundable.GetItem())
		} else {
			Util_Assert_ShouldNeverReachHere()
		}
	}
}

// ItemsTree gets a tree structure (as a nested list) corresponding to the
// structure of the items and nodes in this tree.
//
// The returned slices contain either Object items, or slices which correspond
// to subtrees of the tree. Subtrees which do not contain any items are not
// included.
//
// Builds the tree if necessary.
func (t *IndexStrtree_AbstractSTRtree) ItemsTree() []any {
	t.Build()
	valuesTree := t.itemsTreeNode(t.root)
	if valuesTree == nil {
		return make([]any, 0)
	}
	return valuesTree
}

func (t *IndexStrtree_AbstractSTRtree) itemsTreeNode(node *IndexStrtree_AbstractNode) []any {
	valuesTreeForNode := make([]any, 0)
	for _, childBoundable := range node.GetChildBoundables() {
		if childNode, ok := childBoundable.(*IndexStrtree_AbstractNode); ok {
			valuesTreeForChild := t.itemsTreeNode(childNode)
			// Only add if not nil (which indicates an item somewhere in this tree).
			if valuesTreeForChild != nil {
				valuesTreeForNode = append(valuesTreeForNode, valuesTreeForChild)
			}
		} else if itemBoundable, ok := childBoundable.(*IndexStrtree_ItemBoundable); ok {
			valuesTreeForNode = append(valuesTreeForNode, itemBoundable.GetItem())
		} else {
			Util_Assert_ShouldNeverReachHere()
		}
	}
	if len(valuesTreeForNode) <= 0 {
		return nil
	}
	return valuesTreeForNode
}

// Remove removes an item from the tree. Builds the tree, if necessary.
func (t *IndexStrtree_AbstractSTRtree) Remove(searchBounds, item any) bool {
	t.Build()
	if t.GetIntersectsOp().Intersects(t.root.GetBounds(), searchBounds) {
		return t.remove(searchBounds, t.root, item)
	}
	return false
}

func (t *IndexStrtree_AbstractSTRtree) removeItem(node *IndexStrtree_AbstractNode, item any) bool {
	childToRemoveIdx := -1
	for i, childBoundable := range node.GetChildBoundables() {
		if itemBoundable, ok := childBoundable.(*IndexStrtree_ItemBoundable); ok {
			if itemBoundable.GetItem() == item {
				childToRemoveIdx = i
				break
			}
		}
	}
	if childToRemoveIdx >= 0 {
		// Remove element at index by replacing with last element.
		children := node.GetChildBoundables()
		children[childToRemoveIdx] = children[len(children)-1]
		node.SetChildBoundables(children[:len(children)-1])
		return true
	}
	return false
}

func (t *IndexStrtree_AbstractSTRtree) remove(searchBounds any, node *IndexStrtree_AbstractNode, item any) bool {
	// First try removing item from this node.
	found := t.removeItem(node, item)
	if found {
		return true
	}

	var childToPruneIdx int = -1
	// Next try removing item from lower nodes.
	for i, childBoundable := range node.GetChildBoundables() {
		if !t.GetIntersectsOp().Intersects(childBoundable.GetBounds(), searchBounds) {
			continue
		}
		if childNode, ok := childBoundable.(*IndexStrtree_AbstractNode); ok {
			found = t.remove(searchBounds, childNode, item)
			// If found, record child for pruning and exit.
			if found {
				childToPruneIdx = i
				break
			}
		}
	}
	// Prune child if possible.
	if childToPruneIdx >= 0 {
		childToPrune := node.GetChildBoundables()[childToPruneIdx].(*IndexStrtree_AbstractNode)
		if childToPrune.IsEmpty() {
			children := node.GetChildBoundables()
			children[childToPruneIdx] = children[len(children)-1]
			node.SetChildBoundables(children[:len(children)-1])
		}
	}
	return found
}

// BoundablesAtLevel returns all boundables at the specified level.
func (t *IndexStrtree_AbstractSTRtree) BoundablesAtLevel(level int) []IndexStrtree_Boundable {
	boundables := make([]IndexStrtree_Boundable, 0)
	t.boundablesAtLevel(level, t.root, &boundables)
	return boundables
}

// boundablesAtLevel collects boundables at the specified level.
// Level -1 gets items.
func (t *IndexStrtree_AbstractSTRtree) boundablesAtLevel(level int, top *IndexStrtree_AbstractNode, boundables *[]IndexStrtree_Boundable) {
	Util_Assert_IsTrue(level > -2)
	if top.GetLevel() == level {
		*boundables = append(*boundables, top)
		return
	}
	for _, boundable := range top.GetChildBoundables() {
		if childNode, ok := boundable.(*IndexStrtree_AbstractNode); ok {
			t.boundablesAtLevel(level, childNode, boundables)
		} else {
			Util_Assert_IsTrue(func() bool {
				_, ok := boundable.(*IndexStrtree_ItemBoundable)
				return ok
			}())
			if level == -1 {
				*boundables = append(*boundables, boundable)
			}
		}
	}
}

// GetComparator returns the comparator used to sort boundables.
func (t *IndexStrtree_AbstractSTRtree) GetComparator() func(a, b IndexStrtree_Boundable) int {
	if impl, ok := java.GetLeaf(t).(interface {
		GetComparator_BODY() func(a, b IndexStrtree_Boundable) int
	}); ok {
		return impl.GetComparator_BODY()
	}
	panic("abstract method called")
}

// GetItemBoundables returns the item boundables (internal use).
func (t *IndexStrtree_AbstractSTRtree) GetItemBoundables() []IndexStrtree_Boundable {
	return t.itemBoundables
}
