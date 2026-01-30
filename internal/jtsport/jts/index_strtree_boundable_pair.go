package jts

import "container/heap"

// IndexStrtree_BoundablePair is a pair of Boundables, whose leaf items support a
// distance metric between them. Used to compute the distance between the
// members, and to expand a member relative to the other in order to produce new
// branches of the Branch-and-Bound evaluation tree. Provides an ordering based
// on the distance between the members, which allows building a priority queue
// by minimum distance.
type IndexStrtree_BoundablePair struct {
	boundable1   IndexStrtree_Boundable
	boundable2   IndexStrtree_Boundable
	distance     float64
	itemDistance IndexStrtree_ItemDistance
}

// IndexStrtree_NewBoundablePair creates a new BoundablePair.
func IndexStrtree_NewBoundablePair(boundable1, boundable2 IndexStrtree_Boundable, itemDistance IndexStrtree_ItemDistance) *IndexStrtree_BoundablePair {
	bp := &IndexStrtree_BoundablePair{
		boundable1:   boundable1,
		boundable2:   boundable2,
		itemDistance: itemDistance,
	}
	bp.distance = bp.computeDistance()
	return bp
}

// GetBoundable gets one of the member Boundables in the pair (indexed by [0, 1]).
func (bp *IndexStrtree_BoundablePair) GetBoundable(i int) IndexStrtree_Boundable {
	if i == 0 {
		return bp.boundable1
	}
	return bp.boundable2
}

// MaximumDistance computes the maximum distance between any two items in the
// pair of nodes.
func (bp *IndexStrtree_BoundablePair) MaximumDistance() float64 {
	return IndexStrtree_EnvelopeDistance_MaximumDistance(
		bp.boundable1.GetBounds().(*Geom_Envelope),
		bp.boundable2.GetBounds().(*Geom_Envelope),
	)
}

// computeDistance computes the distance between the Boundables in this pair.
// The boundables are either composites or leaves. If either is composite, the
// distance is computed as the minimum distance between the bounds. If both are
// leaves, the distance is computed by itemDistance.
func (bp *IndexStrtree_BoundablePair) computeDistance() float64 {
	// If items, compute exact distance.
	if bp.IsLeaves() {
		return bp.itemDistance.Distance(
			bp.boundable1.(*IndexStrtree_ItemBoundable),
			bp.boundable2.(*IndexStrtree_ItemBoundable),
		)
	}
	// Otherwise compute distance between bounds of boundables.
	return bp.boundable1.GetBounds().(*Geom_Envelope).Distance(
		bp.boundable2.GetBounds().(*Geom_Envelope),
	)
}

// GetDistance gets the minimum possible distance between the Boundables in this
// pair. If the members are both items, this will be the exact distance between
// them. Otherwise, this distance will be a lower bound on the distances between
// the items in the members.
func (bp *IndexStrtree_BoundablePair) GetDistance() float64 {
	return bp.distance
}

// CompareTo compares two pairs based on their minimum distances.
func (bp *IndexStrtree_BoundablePair) CompareTo(other *IndexStrtree_BoundablePair) int {
	if bp.distance < other.distance {
		return -1
	}
	if bp.distance > other.distance {
		return 1
	}
	return 0
}

// IsLeaves tests if both elements of the pair are leaf nodes.
func (bp *IndexStrtree_BoundablePair) IsLeaves() bool {
	return !IndexStrtree_BoundablePair_IsComposite(bp.boundable1) && !IndexStrtree_BoundablePair_IsComposite(bp.boundable2)
}

// IndexStrtree_BoundablePair_IsComposite tests if the item is a composite (node)
// rather than a leaf.
func IndexStrtree_BoundablePair_IsComposite(item any) bool {
	_, ok := item.(*IndexStrtree_AbstractNode)
	return ok
}

func indexStrtree_BoundablePair_area(b IndexStrtree_Boundable) float64 {
	return b.GetBounds().(*Geom_Envelope).GetArea()
}

// ExpandToQueue expands the pair for a pair which is not a leaf (i.e. has at
// least one composite boundable), computes a list of new pairs from the
// expansion of the larger boundable with distance less than minDistance and
// adds them to a priority queue.
//
// Note that expanded pairs may contain the same item/node on both sides. This
// must be allowed to support distance functions which have non-zero distances
// between the item and itself (non-zero reflexive distance).
func (bp *IndexStrtree_BoundablePair) ExpandToQueue(priQ indexStrtree_BoundablePairQueue, minDistance float64) {
	isComp1 := IndexStrtree_BoundablePair_IsComposite(bp.boundable1)
	isComp2 := IndexStrtree_BoundablePair_IsComposite(bp.boundable2)

	// HEURISTIC: If both boundables are composite, choose the one with largest
	// area to expand. Otherwise, simply expand whichever is composite.
	if isComp1 && isComp2 {
		if indexStrtree_BoundablePair_area(bp.boundable1) > indexStrtree_BoundablePair_area(bp.boundable2) {
			bp.expand(bp.boundable1, bp.boundable2, false, priQ, minDistance)
			return
		}
		bp.expand(bp.boundable2, bp.boundable1, true, priQ, minDistance)
		return
	} else if isComp1 {
		bp.expand(bp.boundable1, bp.boundable2, false, priQ, minDistance)
		return
	} else if isComp2 {
		bp.expand(bp.boundable2, bp.boundable1, true, priQ, minDistance)
		return
	}

	panic("neither boundable is composite")
}

func (bp *IndexStrtree_BoundablePair) expand(bndComposite, bndOther IndexStrtree_Boundable, isFlipped bool, priQ indexStrtree_BoundablePairQueue, minDistance float64) {
	children := bndComposite.(*IndexStrtree_AbstractNode).GetChildBoundables()
	for _, child := range children {
		var newBp *IndexStrtree_BoundablePair
		if isFlipped {
			newBp = IndexStrtree_NewBoundablePair(bndOther, child, bp.itemDistance)
		} else {
			newBp = IndexStrtree_NewBoundablePair(child, bndOther, bp.itemDistance)
		}
		// Only add to queue if this pair might contain the closest points.
		if newBp.GetDistance() < minDistance {
			priQ.Add(newBp)
		}
	}
}

// indexStrtree_BoundablePairQueue is an interface for priority queues that
// hold BoundablePairs.
type indexStrtree_BoundablePairQueue interface {
	Add(bp *IndexStrtree_BoundablePair)
	Poll() *IndexStrtree_BoundablePair
	IsEmpty() bool
}

// indexStrtree_BoundablePairPriorityQueue is a min-heap priority queue for
// BoundablePairs, ordered by distance.
type indexStrtree_BoundablePairPriorityQueue struct {
	heap boundablePairMinHeap
}

func (pq *indexStrtree_BoundablePairPriorityQueue) Add(bp *IndexStrtree_BoundablePair) {
	heap.Push(&pq.heap, bp)
}

func (pq *indexStrtree_BoundablePairPriorityQueue) Poll() *IndexStrtree_BoundablePair {
	return heap.Pop(&pq.heap).(*IndexStrtree_BoundablePair)
}

func (pq *indexStrtree_BoundablePairPriorityQueue) IsEmpty() bool {
	return len(pq.heap) == 0
}

// boundablePairMinHeap implements heap.Interface for min-heap ordering.
type boundablePairMinHeap []*IndexStrtree_BoundablePair

func (h boundablePairMinHeap) Len() int           { return len(h) }
func (h boundablePairMinHeap) Less(i, j int) bool { return h[i].distance < h[j].distance }
func (h boundablePairMinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *boundablePairMinHeap) Push(x any) {
	*h = append(*h, x.(*IndexStrtree_BoundablePair))
}

func (h *boundablePairMinHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // Avoid memory leak.
	*h = old[0 : n-1]
	return item
}

// indexStrtree_BoundablePairMaxPriorityQueue is a max-heap priority queue for
// BoundablePairs, ordered by distance (largest distance at top).
type indexStrtree_BoundablePairMaxPriorityQueue struct {
	heap boundablePairMaxHeap
}

func (pq *indexStrtree_BoundablePairMaxPriorityQueue) Add(bp *IndexStrtree_BoundablePair) {
	heap.Push(&pq.heap, bp)
}

func (pq *indexStrtree_BoundablePairMaxPriorityQueue) Poll() *IndexStrtree_BoundablePair {
	return heap.Pop(&pq.heap).(*IndexStrtree_BoundablePair)
}

func (pq *indexStrtree_BoundablePairMaxPriorityQueue) Peek() *IndexStrtree_BoundablePair {
	if len(pq.heap) == 0 {
		return nil
	}
	return pq.heap[0]
}

func (pq *indexStrtree_BoundablePairMaxPriorityQueue) Size() int {
	return len(pq.heap)
}

// boundablePairMaxHeap implements heap.Interface for max-heap ordering.
type boundablePairMaxHeap []*IndexStrtree_BoundablePair

func (h boundablePairMaxHeap) Len() int           { return len(h) }
func (h boundablePairMaxHeap) Less(i, j int) bool { return h[i].distance > h[j].distance }
func (h boundablePairMaxHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *boundablePairMaxHeap) Push(x any) {
	*h = append(*h, x.(*IndexStrtree_BoundablePair))
}

func (h *boundablePairMaxHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // Avoid memory leak.
	*h = old[0 : n-1]
	return item
}
