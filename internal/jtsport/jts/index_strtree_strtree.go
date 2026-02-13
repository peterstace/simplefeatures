package jts

import (
	"math"
	"sort"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

const IndexStrtree_STRtree_DEFAULT_NODE_CAPACITY = 10

// IndexStrtree_STRtreeNode is a node of an STRtree.
type IndexStrtree_STRtreeNode struct {
	*IndexStrtree_AbstractNode
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (n *IndexStrtree_STRtreeNode) GetChild() java.Polymorphic {
	return n.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (n *IndexStrtree_STRtreeNode) GetParent() java.Polymorphic {
	return n.IndexStrtree_AbstractNode
}

// IndexStrtree_NewSTRtreeNode creates a new STRtreeNode at the given level.
func IndexStrtree_NewSTRtreeNode(level int) *IndexStrtree_STRtreeNode {
	base := IndexStrtree_NewAbstractNode(level)
	node := &IndexStrtree_STRtreeNode{
		IndexStrtree_AbstractNode: base,
	}
	base.child = node
	return node
}

// ComputeBounds_BODY computes the bounds of this node by expanding an envelope
// to include the bounds of all child boundables.
func (n *IndexStrtree_STRtreeNode) ComputeBounds_BODY() any {
	var bounds *Geom_Envelope
	for _, childBoundable := range n.GetChildBoundables() {
		childBounds := childBoundable.GetBounds().(*Geom_Envelope)
		if bounds == nil {
			bounds = Geom_NewEnvelopeFromEnvelope(childBounds)
		} else {
			bounds.ExpandToIncludeEnvelope(childBounds)
		}
	}
	return bounds
}

// IndexStrtree_STRtree_xComparator compares boundables by X coordinate of
// envelope centre.
func IndexStrtree_STRtree_xComparator(o1, o2 IndexStrtree_Boundable) int {
	return IndexStrtree_AbstractSTRtree_CompareDoubles(
		indexStrtree_STRtree_centreX(o1.GetBounds().(*Geom_Envelope)),
		indexStrtree_STRtree_centreX(o2.GetBounds().(*Geom_Envelope)),
	)
}

// IndexStrtree_STRtree_yComparator compares boundables by Y coordinate of
// envelope centre.
func IndexStrtree_STRtree_yComparator(o1, o2 IndexStrtree_Boundable) int {
	return IndexStrtree_AbstractSTRtree_CompareDoubles(
		indexStrtree_STRtree_centreY(o1.GetBounds().(*Geom_Envelope)),
		indexStrtree_STRtree_centreY(o2.GetBounds().(*Geom_Envelope)),
	)
}

func indexStrtree_STRtree_centreX(e *Geom_Envelope) float64 {
	return indexStrtree_STRtree_avg(e.GetMinX(), e.GetMaxX())
}

func indexStrtree_STRtree_centreY(e *Geom_Envelope) float64 {
	return indexStrtree_STRtree_avg(e.GetMinY(), e.GetMaxY())
}

func indexStrtree_STRtree_avg(a, b float64) float64 {
	return (a + b) / 2.0
}

// indexStrtree_STRtree_intersectsOp tests whether two Envelopes intersect.
type indexStrtree_STRtree_intersectsOp struct{}

func (op *indexStrtree_STRtree_intersectsOp) Intersects(aBounds, bBounds any) bool {
	return aBounds.(*Geom_Envelope).IntersectsEnvelope(bBounds.(*Geom_Envelope))
}

var indexStrtree_STRtree_IntersectsOpInstance = &indexStrtree_STRtree_intersectsOp{}

// IndexStrtree_STRtree is a query-only R-tree created using the Sort-Tile-
// Recursive (STR) algorithm. For two-dimensional spatial data.
//
// The STR packed R-tree is simple to implement and maximizes space utilization;
// that is, as many leaves as possible are filled to capacity. Overlap between
// nodes is far less than in a basic R-tree. However, the index is semi-static;
// once the tree has been built (which happens automatically upon the first
// query), items may not be added. Items may be removed from the tree using
// Remove(Envelope, Object).
//
// Described in: P. Rigaux, Michel Scholl and Agnes Voisard. Spatial Databases
// With Application To GIS. Morgan Kaufmann, San Francisco, 2002.
//
// Note that inserting items into a tree is not thread-safe. Inserting performed
// on more than one thread must be synchronized externally.
//
// Querying a tree is thread-safe. The building phase is done synchronously, and
// querying is stateless.
type IndexStrtree_STRtree struct {
	*IndexStrtree_AbstractSTRtree
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (t *IndexStrtree_STRtree) GetChild() java.Polymorphic {
	return t.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (t *IndexStrtree_STRtree) GetParent() java.Polymorphic {
	return t.IndexStrtree_AbstractSTRtree
}

// Compile-time check that IndexStrtree_STRtree implements Index_SpatialIndex.
var _ Index_SpatialIndex = (*IndexStrtree_STRtree)(nil)

// IsIndex_SpatialIndex is a marker method for interface identification.
func (t *IndexStrtree_STRtree) IsIndex_SpatialIndex() {}

// IndexStrtree_NewSTRtree constructs an STRtree with the default node capacity.
func IndexStrtree_NewSTRtree() *IndexStrtree_STRtree {
	return IndexStrtree_NewSTRtreeWithCapacity(IndexStrtree_STRtree_DEFAULT_NODE_CAPACITY)
}

// IndexStrtree_NewSTRtreeWithCapacity constructs an STRtree with the given
// maximum number of child nodes that a node may have.
//
// The minimum recommended capacity setting is 4.
func IndexStrtree_NewSTRtreeWithCapacity(nodeCapacity int) *IndexStrtree_STRtree {
	base := IndexStrtree_NewAbstractSTRtreeWithCapacity(nodeCapacity)
	t := &IndexStrtree_STRtree{
		IndexStrtree_AbstractSTRtree: base,
	}
	base.child = t
	return t
}

// IndexStrtree_NewSTRtreeWithCapacityAndRoot constructs an STRtree with the
// given maximum number of child nodes that a node may have, and the root that
// links to all other nodes.
//
// The minimum recommended capacity setting is 4.
func IndexStrtree_NewSTRtreeWithCapacityAndRoot(nodeCapacity int, root *IndexStrtree_STRtreeNode) *IndexStrtree_STRtree {
	base := IndexStrtree_NewAbstractSTRtreeWithCapacityAndRoot(nodeCapacity, root.IndexStrtree_AbstractNode)
	t := &IndexStrtree_STRtree{
		IndexStrtree_AbstractSTRtree: base,
	}
	base.child = t
	return t
}

// IndexStrtree_NewSTRtreeWithCapacityAndItems constructs an STRtree with the
// given maximum number of child nodes that a node may have, and all leaf nodes
// in the tree.
//
// The minimum recommended capacity setting is 4.
func IndexStrtree_NewSTRtreeWithCapacityAndItems(nodeCapacity int, itemBoundables []IndexStrtree_Boundable) *IndexStrtree_STRtree {
	base := IndexStrtree_NewAbstractSTRtreeWithCapacityAndItems(nodeCapacity, itemBoundables)
	t := &IndexStrtree_STRtree{
		IndexStrtree_AbstractSTRtree: base,
	}
	base.child = t
	return t
}

// CreateNode_BODY creates a new STRtreeNode at the given level.
func (t *IndexStrtree_STRtree) CreateNode_BODY(level int) *IndexStrtree_AbstractNode {
	return IndexStrtree_NewSTRtreeNode(level).IndexStrtree_AbstractNode
}

// GetIntersectsOp_BODY returns the intersects operation for STRtree.
func (t *IndexStrtree_STRtree) GetIntersectsOp_BODY() IndexStrtree_IntersectsOp {
	return indexStrtree_STRtree_IntersectsOpInstance
}

// CreateParentBoundables creates the parent level for the given child level.
// First, orders the items by the x-values of the midpoints, and groups them
// into vertical slices. For each slice, orders the items by the y-values of the
// midpoints, and group them into runs of size M (the node capacity). For each
// run, creates a new (parent) node.
func (t *IndexStrtree_STRtree) CreateParentBoundables(childBoundables []IndexStrtree_Boundable, newLevel int) []IndexStrtree_Boundable {
	Util_Assert_IsTrue(len(childBoundables) > 0)
	minLeafCount := int(math.Ceil(float64(len(childBoundables)) / float64(t.GetNodeCapacity())))
	sortedChildBoundables := make([]IndexStrtree_Boundable, len(childBoundables))
	copy(sortedChildBoundables, childBoundables)
	sort.Slice(sortedChildBoundables, func(i, j int) bool {
		return IndexStrtree_STRtree_xComparator(sortedChildBoundables[i], sortedChildBoundables[j]) < 0
	})
	verticalSlices := t.verticalSlices(sortedChildBoundables, int(math.Ceil(math.Sqrt(float64(minLeafCount)))))
	return t.createParentBoundablesFromVerticalSlices(verticalSlices, newLevel)
}

func (t *IndexStrtree_STRtree) createParentBoundablesFromVerticalSlices(verticalSlices [][]IndexStrtree_Boundable, newLevel int) []IndexStrtree_Boundable {
	Util_Assert_IsTrue(len(verticalSlices) > 0)
	parentBoundables := make([]IndexStrtree_Boundable, 0)
	for i := 0; i < len(verticalSlices); i++ {
		parentBoundables = append(parentBoundables, t.createParentBoundablesFromVerticalSlice(verticalSlices[i], newLevel)...)
	}
	return parentBoundables
}

func (t *IndexStrtree_STRtree) createParentBoundablesFromVerticalSlice(childBoundables []IndexStrtree_Boundable, newLevel int) []IndexStrtree_Boundable {
	return t.IndexStrtree_AbstractSTRtree.CreateParentBoundables(childBoundables, newLevel)
}

// verticalSlices divides childBoundables into vertical slices.
func (t *IndexStrtree_STRtree) verticalSlices(childBoundables []IndexStrtree_Boundable, sliceCount int) [][]IndexStrtree_Boundable {
	sliceCapacity := int(math.Ceil(float64(len(childBoundables)) / float64(sliceCount)))
	slices := make([][]IndexStrtree_Boundable, sliceCount)
	idx := 0
	for j := 0; j < sliceCount; j++ {
		slices[j] = make([]IndexStrtree_Boundable, 0)
		boundablesAddedToSlice := 0
		for idx < len(childBoundables) && boundablesAddedToSlice < sliceCapacity {
			slices[j] = append(slices[j], childBoundables[idx])
			idx++
			boundablesAddedToSlice++
		}
	}
	return slices
}

// Insert inserts an item having the given bounds into the tree.
func (t *IndexStrtree_STRtree) Insert(itemEnv *Geom_Envelope, item any) {
	if itemEnv.IsNull() {
		return
	}
	t.IndexStrtree_AbstractSTRtree.Insert(itemEnv, item)
}

// Query returns items whose bounds intersect the given envelope.
func (t *IndexStrtree_STRtree) Query(searchEnv *Geom_Envelope) []any {
	return t.IndexStrtree_AbstractSTRtree.Query(searchEnv)
}

// QueryWithVisitor returns items whose bounds intersect the given envelope.
func (t *IndexStrtree_STRtree) QueryWithVisitor(searchEnv *Geom_Envelope, visitor Index_ItemVisitor) {
	t.IndexStrtree_AbstractSTRtree.QueryWithVisitor(searchEnv, visitor)
}

// Remove removes a single item from the tree.
func (t *IndexStrtree_STRtree) Remove(itemEnv *Geom_Envelope, item any) bool {
	return t.IndexStrtree_AbstractSTRtree.Remove(itemEnv, item)
}

// Size returns the number of items in the tree.
func (t *IndexStrtree_STRtree) Size() int {
	return t.IndexStrtree_AbstractSTRtree.Size()
}

// Depth returns the number of levels in the tree.
func (t *IndexStrtree_STRtree) Depth() int {
	return t.IndexStrtree_AbstractSTRtree.Depth()
}

// GetComparator_BODY returns the comparator used to sort boundables.
func (t *IndexStrtree_STRtree) GetComparator_BODY() func(a, b IndexStrtree_Boundable) int {
	return IndexStrtree_STRtree_yComparator
}

// NearestNeighbour finds the two nearest items in the tree, using ItemDistance
// as the distance metric. A Branch-and-Bound tree traversal algorithm is used
// to provide an efficient search.
//
// If the tree is empty, the return value is nil. If the tree contains only one
// item, the return value is a pair containing that item.
//
// If it is required to find only pairs of distinct items, the ItemDistance
// function must be anti-reflexive.
func (t *IndexStrtree_STRtree) NearestNeighbour(itemDist IndexStrtree_ItemDistance) []any {
	if t.IsEmpty() {
		return nil
	}
	// If tree has only one item this will return nil.
	bp := IndexStrtree_NewBoundablePair(t.GetRoot(), t.GetRoot(), itemDist)
	return t.nearestNeighbour(bp)
}

// NearestNeighbourWithEnvelope finds the item in this tree which is nearest to
// the given Object, using ItemDistance as the distance metric. A
// Branch-and-Bound tree traversal algorithm is used to provide an efficient
// search.
//
// The query object does not have to be contained in the tree, but it does have
// to be compatible with the itemDist distance metric.
func (t *IndexStrtree_STRtree) NearestNeighbourWithEnvelope(env *Geom_Envelope, item any, itemDist IndexStrtree_ItemDistance) any {
	if t.IsEmpty() {
		return nil
	}
	bnd := IndexStrtree_NewItemBoundable(env, item)
	bp := IndexStrtree_NewBoundablePair(t.GetRoot(), bnd, itemDist)
	return t.nearestNeighbour(bp)[0]
}

// NearestNeighbourFromTree finds the two nearest items from this tree and
// another tree, using ItemDistance as the distance metric. A Branch-and-Bound
// tree traversal algorithm is used to provide an efficient search. The result
// value is a pair of items, the first from this tree and the second from the
// argument tree.
func (t *IndexStrtree_STRtree) NearestNeighbourFromTree(tree *IndexStrtree_STRtree, itemDist IndexStrtree_ItemDistance) []any {
	if t.IsEmpty() || tree.IsEmpty() {
		return nil
	}
	bp := IndexStrtree_NewBoundablePair(t.GetRoot(), tree.GetRoot(), itemDist)
	return t.nearestNeighbour(bp)
}

func (t *IndexStrtree_STRtree) nearestNeighbour(initBndPair *IndexStrtree_BoundablePair) []any {
	distanceLowerBound := math.Inf(1)
	var minPair *IndexStrtree_BoundablePair

	// Initialize search queue.
	priQ := &indexStrtree_BoundablePairPriorityQueue{}

	priQ.Add(initBndPair)

	for !priQ.IsEmpty() && distanceLowerBound > 0.0 {
		// Pop head of queue and expand one side of pair.
		bndPair := priQ.Poll()
		pairDistance := bndPair.GetDistance()

		// If the distance for the first pair in the queue is >= current minimum
		// distance, other nodes in the queue must also have a greater distance.
		// So the current minDistance must be the true minimum, and we are done.
		if pairDistance >= distanceLowerBound {
			break
		}

		// If the pair members are leaves then their distance is the exact lower
		// bound. Update the distanceLowerBound to reflect this (which must be
		// smaller, due to the test immediately prior to this).
		if bndPair.IsLeaves() {
			distanceLowerBound = pairDistance
			minPair = bndPair
		} else {
			// Otherwise, expand one side of the pair, and insert the expanded
			// pairs into the queue. The choice of which side to expand is
			// determined heuristically.
			bndPair.ExpandToQueue(priQ, distanceLowerBound)
		}
	}
	if minPair == nil {
		return nil
	}
	// Done - return items with min distance.
	return []any{
		minPair.GetBoundable(0).(*IndexStrtree_ItemBoundable).GetItem(),
		minPair.GetBoundable(1).(*IndexStrtree_ItemBoundable).GetItem(),
	}
}

// IsWithinDistance tests whether some two items from this tree and another tree
// lie within a given distance. ItemDistance is used as the distance metric. A
// Branch-and-Bound tree traversal algorithm is used to provide an efficient
// search.
func (t *IndexStrtree_STRtree) IsWithinDistance(tree *IndexStrtree_STRtree, itemDist IndexStrtree_ItemDistance, maxDistance float64) bool {
	bp := IndexStrtree_NewBoundablePair(t.GetRoot(), tree.GetRoot(), itemDist)
	return t.isWithinDistance(bp, maxDistance)
}

func (t *IndexStrtree_STRtree) isWithinDistance(initBndPair *IndexStrtree_BoundablePair, maxDistance float64) bool {
	distanceUpperBound := math.Inf(1)
	_ = distanceUpperBound // Used in the algorithm but set, not always read.

	// Initialize search queue.
	priQ := &indexStrtree_BoundablePairPriorityQueue{}
	priQ.Add(initBndPair)

	for !priQ.IsEmpty() {
		// Pop head of queue and expand one side of pair.
		bndPair := priQ.Poll()
		pairDistance := bndPair.GetDistance()

		// If the distance for the first pair in the queue is > maxDistance, all
		// other pairs in the queue must have a greater distance as well. So can
		// conclude no items are within the distance and terminate with result =
		// false.
		if pairDistance > maxDistance {
			return false
		}

		// If the maximum distance between the nodes is less than the
		// maxDistance, than all items in the nodes must be closer than the max
		// distance. Then can terminate with result = true.
		if bndPair.MaximumDistance() <= maxDistance {
			return true
		}

		// If the pair items are leaves then their actual distance is an upper
		// bound. Update the distanceUpperBound to reflect this.
		if bndPair.IsLeaves() {
			distanceUpperBound = pairDistance

			// If the items are closer than maxDistance can terminate with
			// result = true.
			if distanceUpperBound <= maxDistance {
				return true
			}
		} else {
			// Otherwise, expand one side of the pair, and insert the expanded
			// pairs into the queue. The choice of which side to expand is
			// determined heuristically.
			bndPair.ExpandToQueue(priQ, distanceUpperBound)
		}
	}
	return false
}

// NearestNeighbourK finds up to k items in this tree which are the nearest
// neighbors to the given item, using itemDist as the distance metric. A
// Branch-and-Bound tree traversal algorithm is used to provide an efficient
// search.
//
// The query item does not have to be contained in the tree, but it does have to
// be compatible with the itemDist distance metric.
//
// If the tree size is smaller than k fewer items will be returned. If the tree
// is empty an array of size 0 is returned.
func (t *IndexStrtree_STRtree) NearestNeighbourK(env *Geom_Envelope, item any, itemDist IndexStrtree_ItemDistance, k int) []any {
	if t.IsEmpty() {
		return make([]any, 0)
	}
	bnd := IndexStrtree_NewItemBoundable(env, item)
	bp := IndexStrtree_NewBoundablePair(t.GetRoot(), bnd, itemDist)
	return t.nearestNeighbourK(bp, k)
}

func (t *IndexStrtree_STRtree) nearestNeighbourK(initBndPair *IndexStrtree_BoundablePair, k int) []any {
	return t.nearestNeighbourKWithMaxDistance(initBndPair, math.Inf(1), k)
}

func (t *IndexStrtree_STRtree) nearestNeighbourKWithMaxDistance(initBndPair *IndexStrtree_BoundablePair, maxDistance float64, k int) []any {
	distanceLowerBound := maxDistance

	// Initialize internal structures.
	priQ := &indexStrtree_BoundablePairPriorityQueue{}

	// Initialize queue.
	priQ.Add(initBndPair)

	kNearestNeighbors := &indexStrtree_BoundablePairMaxPriorityQueue{}

	for !priQ.IsEmpty() && distanceLowerBound >= 0.0 {
		// Pop head of queue and expand one side of pair.
		bndPair := priQ.Poll()
		pairDistance := bndPair.GetDistance()

		// If the distance for the first node in the queue is >= the current
		// maximum distance in the k queue, all other nodes in the queue must
		// also have a greater distance. So the current minDistance must be the
		// true minimum, and we are done.
		if pairDistance >= distanceLowerBound {
			break
		}

		// If the pair members are leaves then their distance is the exact lower
		// bound. Update the distanceLowerBound to reflect this (which must be
		// smaller, due to the test immediately prior to this).
		if bndPair.IsLeaves() {
			if kNearestNeighbors.Size() < k {
				kNearestNeighbors.Add(bndPair)
			} else {
				bp1 := kNearestNeighbors.Peek()
				if bp1.GetDistance() > pairDistance {
					kNearestNeighbors.Poll()
					kNearestNeighbors.Add(bndPair)
				}
				// minDistance should be the farthest point in the K nearest
				// neighbor queue.
				bp2 := kNearestNeighbors.Peek()
				distanceLowerBound = bp2.GetDistance()
			}
		} else {
			// Otherwise, expand one side of the pair, (the choice of which side
			// to expand is heuristically determined) and insert the new
			// expanded pairs into the queue.
			bndPair.ExpandToQueue(priQ, distanceLowerBound)
		}
	}
	// Done - return items with min distance.
	return t.getItems(kNearestNeighbors)
}

func (t *IndexStrtree_STRtree) getItems(kNearestNeighbors *indexStrtree_BoundablePairMaxPriorityQueue) []any {
	// Iterate the K Nearest Neighbour Queue and retrieve the item from each
	// BoundablePair in this queue.
	items := make([]any, kNearestNeighbors.Size())
	count := 0
	for kNearestNeighbors.Size() > 0 {
		bp := kNearestNeighbors.Poll()
		items[count] = bp.GetBoundable(0).(*IndexStrtree_ItemBoundable).GetItem()
		count++
	}
	return items
}

// Insert_BODY implements the SpatialIndex interface.
func (t *IndexStrtree_STRtree) Insert_BODY(itemEnv *Geom_Envelope, item any) {
	t.Insert(itemEnv, item)
}

// Query_BODY implements the SpatialIndex interface.
func (t *IndexStrtree_STRtree) Query_BODY(searchEnv *Geom_Envelope) []any {
	return t.Query(searchEnv)
}

// QueryWithVisitor_BODY implements the SpatialIndex interface.
func (t *IndexStrtree_STRtree) QueryWithVisitor_BODY(searchEnv *Geom_Envelope, visitor Index_ItemVisitor) {
	t.QueryWithVisitor(searchEnv, visitor)
}

// Remove_BODY implements the SpatialIndex interface.
func (t *IndexStrtree_STRtree) Remove_BODY(itemEnv *Geom_Envelope, item any) bool {
	return t.Remove(itemEnv, item)
}
