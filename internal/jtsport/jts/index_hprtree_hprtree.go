package jts

import "math"

const indexHprtree_HPRtree_ENV_SIZE = 4
const indexHprtree_HPRtree_HILBERT_LEVEL = 12
const indexHprtree_HPRtree_DEFAULT_NODE_CAPACITY = 16

// IndexHprtree_HPRtree is a Hilbert-Packed R-tree. This is a static R-tree
// which is packed by using the Hilbert ordering of the tree items.
//
// The tree is constructed by sorting the items by the Hilbert code of the
// midpoint of their envelope. Then, a set of internal layers is created
// recursively as follows:
//   - The items/nodes of the previous are partitioned into blocks of size
//     nodeCapacity
//   - For each block a layer node is created with range equal to the envelope
//     of the items/nodes in the block
//
// The internal layers are stored using an array to store the node bounds.
// The link between a node and its children is stored implicitly in the indexes
// of the array. For efficiency, the offsets to the layers within the node
// array are pre-computed and stored.
//
// NOTE: Based on performance testing, the HPRtree is somewhat faster than the
// STRtree. It should also be more memory-efficient, due to fewer object
// allocations. However, it is not clear whether this will produce a
// significant improvement for use in JTS operations.
type IndexHprtree_HPRtree struct {
	itemsToLoad     []*IndexHprtree_Item
	nodeCapacity    int
	numItems        int
	totalExtent     *Geom_Envelope
	layerStartIndex []int
	nodeBounds      []float64
	itemBounds      []float64
	itemValues      []any
	isBuilt         bool
}

// IndexHprtree_NewHPRtree creates a new index with the default node capacity.
func IndexHprtree_NewHPRtree() *IndexHprtree_HPRtree {
	return IndexHprtree_NewHPRtreeWithCapacity(indexHprtree_HPRtree_DEFAULT_NODE_CAPACITY)
}

// IndexHprtree_NewHPRtreeWithCapacity creates a new index with the given node
// capacity.
func IndexHprtree_NewHPRtreeWithCapacity(nodeCapacity int) *IndexHprtree_HPRtree {
	return &IndexHprtree_HPRtree{
		itemsToLoad:  make([]*IndexHprtree_Item, 0),
		nodeCapacity: nodeCapacity,
		numItems:     0,
		totalExtent:  Geom_NewEnvelope(),
		isBuilt:      false,
	}
}

// Size gets the number of items in the index.
func (t *IndexHprtree_HPRtree) Size() int {
	return t.numItems
}

// Insert inserts an item with the given envelope into the index.
func (t *IndexHprtree_HPRtree) Insert(itemEnv *Geom_Envelope, item any) {
	if t.isBuilt {
		panic("Cannot insert items after tree is built.")
	}
	t.numItems++
	t.itemsToLoad = append(t.itemsToLoad, IndexHprtree_NewItem(itemEnv, item))
	t.totalExtent.ExpandToIncludeEnvelope(itemEnv)
}

// Query returns all items whose envelopes intersect the given search envelope.
func (t *IndexHprtree_HPRtree) Query(searchEnv *Geom_Envelope) []any {
	t.Build()

	if !t.totalExtent.IntersectsEnvelope(searchEnv) {
		return []any{}
	}

	visitor := Index_NewArrayListVisitor()
	t.QueryWithVisitor(searchEnv, visitor)
	return visitor.GetItems()
}

// QueryWithVisitor visits all items whose envelopes intersect the given search
// envelope using the provided visitor.
func (t *IndexHprtree_HPRtree) QueryWithVisitor(searchEnv *Geom_Envelope, visitor Index_ItemVisitor) {
	t.Build()
	if !t.totalExtent.IntersectsEnvelope(searchEnv) {
		return
	}
	if t.layerStartIndex == nil {
		t.queryItems(0, searchEnv, visitor)
	} else {
		t.queryTopLayer(searchEnv, visitor)
	}
}

func (t *IndexHprtree_HPRtree) queryTopLayer(searchEnv *Geom_Envelope, visitor Index_ItemVisitor) {
	layerIndex := len(t.layerStartIndex) - 2
	layerSize := t.layerSize(layerIndex)
	// query each node in layer
	for i := 0; i < layerSize; i += indexHprtree_HPRtree_ENV_SIZE {
		t.queryNode(layerIndex, i, searchEnv, visitor)
	}
}

func (t *IndexHprtree_HPRtree) queryNode(layerIndex, nodeOffset int, searchEnv *Geom_Envelope, visitor Index_ItemVisitor) {
	layerStart := t.layerStartIndex[layerIndex]
	nodeIndex := layerStart + nodeOffset
	if !indexHprtree_HPRtree_intersects(t.nodeBounds, nodeIndex, searchEnv) {
		return
	}
	if layerIndex == 0 {
		childNodesOffset := nodeOffset / indexHprtree_HPRtree_ENV_SIZE * t.nodeCapacity
		t.queryItems(childNodesOffset, searchEnv, visitor)
	} else {
		childNodesOffset := nodeOffset * t.nodeCapacity
		t.queryNodeChildren(layerIndex-1, childNodesOffset, searchEnv, visitor)
	}
}

func indexHprtree_HPRtree_intersects(bounds []float64, nodeIndex int, env *Geom_Envelope) bool {
	isBeyond := (env.GetMaxX() < bounds[nodeIndex]) ||
		(env.GetMaxY() < bounds[nodeIndex+1]) ||
		(env.GetMinX() > bounds[nodeIndex+2]) ||
		(env.GetMinY() > bounds[nodeIndex+3])
	return !isBeyond
}

func (t *IndexHprtree_HPRtree) queryNodeChildren(layerIndex, blockOffset int, searchEnv *Geom_Envelope, visitor Index_ItemVisitor) {
	layerStart := t.layerStartIndex[layerIndex]
	layerEnd := t.layerStartIndex[layerIndex+1]
	for i := 0; i < t.nodeCapacity; i++ {
		nodeOffset := blockOffset + indexHprtree_HPRtree_ENV_SIZE*i
		// don't query past layer end
		if layerStart+nodeOffset >= layerEnd {
			break
		}
		t.queryNode(layerIndex, nodeOffset, searchEnv, visitor)
	}
}

func (t *IndexHprtree_HPRtree) queryItems(blockStart int, searchEnv *Geom_Envelope, visitor Index_ItemVisitor) {
	for i := 0; i < t.nodeCapacity; i++ {
		itemIndex := blockStart + i
		// don't query past end of items
		if itemIndex >= t.numItems {
			break
		}
		if indexHprtree_HPRtree_intersects(t.itemBounds, itemIndex*indexHprtree_HPRtree_ENV_SIZE, searchEnv) {
			visitor.VisitItem(t.itemValues[itemIndex])
		}
	}
}

func (t *IndexHprtree_HPRtree) layerSize(layerIndex int) int {
	layerStart := t.layerStartIndex[layerIndex]
	layerEnd := t.layerStartIndex[layerIndex+1]
	return layerEnd - layerStart
}

// Remove removes an item from the index.
// Note: HPRtree does not support removal.
func (t *IndexHprtree_HPRtree) Remove(itemEnv *Geom_Envelope, item any) bool {
	return false
}

// Build builds the index, if not already built.
func (t *IndexHprtree_HPRtree) Build() {
	if !t.isBuilt {
		t.prepareIndex()
		t.prepareItems()
		t.isBuilt = true
	}
}

func (t *IndexHprtree_HPRtree) prepareIndex() {
	// don't need to build an empty or very small tree
	if len(t.itemsToLoad) <= t.nodeCapacity {
		return
	}

	t.sortItems()

	t.layerStartIndex = indexHprtree_HPRtree_computeLayerIndices(t.numItems, t.nodeCapacity)
	// allocate storage
	nodeCount := t.layerStartIndex[len(t.layerStartIndex)-1] / 4
	t.nodeBounds = indexHprtree_HPRtree_createBoundsArray(nodeCount)

	// compute tree nodes
	t.computeLeafNodes(t.layerStartIndex[1])
	for i := 1; i < len(t.layerStartIndex)-1; i++ {
		t.computeLayerNodes(i)
	}
}

func (t *IndexHprtree_HPRtree) prepareItems() {
	// copy item contents out to arrays for querying
	boundsIndex := 0
	valueIndex := 0
	t.itemBounds = make([]float64, len(t.itemsToLoad)*4)
	t.itemValues = make([]any, len(t.itemsToLoad))
	for _, item := range t.itemsToLoad {
		envelope := item.GetEnvelope()
		t.itemBounds[boundsIndex] = envelope.GetMinX()
		boundsIndex++
		t.itemBounds[boundsIndex] = envelope.GetMinY()
		boundsIndex++
		t.itemBounds[boundsIndex] = envelope.GetMaxX()
		boundsIndex++
		t.itemBounds[boundsIndex] = envelope.GetMaxY()
		boundsIndex++
		t.itemValues[valueIndex] = item.GetItem()
		valueIndex++
	}
	// and let GC free the original list
	t.itemsToLoad = nil
}

func indexHprtree_HPRtree_createBoundsArray(size int) []float64 {
	a := make([]float64, 4*size)
	for i := 0; i < size; i++ {
		index := 4 * i
		a[index] = math.MaxFloat64
		a[index+1] = math.MaxFloat64
		a[index+2] = -math.MaxFloat64
		a[index+3] = -math.MaxFloat64
	}
	return a
}

func (t *IndexHprtree_HPRtree) computeLayerNodes(layerIndex int) {
	layerStart := t.layerStartIndex[layerIndex]
	childLayerStart := t.layerStartIndex[layerIndex-1]
	layerSize := t.layerSize(layerIndex)
	childLayerEnd := layerStart
	for i := 0; i < layerSize; i += indexHprtree_HPRtree_ENV_SIZE {
		childStart := childLayerStart + t.nodeCapacity*i
		t.computeNodeBounds(layerStart+i, childStart, childLayerEnd)
	}
}

func (t *IndexHprtree_HPRtree) computeNodeBounds(nodeIndex, blockStart, nodeMaxIndex int) {
	for i := 0; i <= t.nodeCapacity; i++ {
		index := blockStart + 4*i
		if index >= nodeMaxIndex {
			break
		}
		t.updateNodeBounds(nodeIndex, t.nodeBounds[index], t.nodeBounds[index+1], t.nodeBounds[index+2], t.nodeBounds[index+3])
	}
}

func (t *IndexHprtree_HPRtree) computeLeafNodes(layerSize int) {
	for i := 0; i < layerSize; i += indexHprtree_HPRtree_ENV_SIZE {
		t.computeLeafNodeBounds(i, t.nodeCapacity*i/4)
	}
}

func (t *IndexHprtree_HPRtree) computeLeafNodeBounds(nodeIndex, blockStart int) {
	for i := 0; i <= t.nodeCapacity; i++ {
		itemIndex := blockStart + i
		if itemIndex >= len(t.itemsToLoad) {
			break
		}
		env := t.itemsToLoad[itemIndex].GetEnvelope()
		t.updateNodeBounds(nodeIndex, env.GetMinX(), env.GetMinY(), env.GetMaxX(), env.GetMaxY())
	}
}

func (t *IndexHprtree_HPRtree) updateNodeBounds(nodeIndex int, minX, minY, maxX, maxY float64) {
	if minX < t.nodeBounds[nodeIndex] {
		t.nodeBounds[nodeIndex] = minX
	}
	if minY < t.nodeBounds[nodeIndex+1] {
		t.nodeBounds[nodeIndex+1] = minY
	}
	if maxX > t.nodeBounds[nodeIndex+2] {
		t.nodeBounds[nodeIndex+2] = maxX
	}
	if maxY > t.nodeBounds[nodeIndex+3] {
		t.nodeBounds[nodeIndex+3] = maxY
	}
}

func indexHprtree_HPRtree_computeLayerIndices(itemSize, nodeCapacity int) []int {
	layerIndexList := Util_NewIntArrayList()
	layerSize := itemSize
	index := 0
	for {
		layerIndexList.Add(index)
		layerSize = indexHprtree_HPRtree_numNodesToCover(layerSize, nodeCapacity)
		index += indexHprtree_HPRtree_ENV_SIZE * layerSize
		if layerSize <= 1 {
			break
		}
	}
	return layerIndexList.ToArray()
}

// indexHprtree_HPRtree_numNodesToCover computes the number of blocks (nodes)
// required to cover a given number of children.
func indexHprtree_HPRtree_numNodesToCover(nChild, nodeCapacity int) int {
	mult := nChild / nodeCapacity
	total := mult * nodeCapacity
	if total == nChild {
		return mult
	}
	return mult + 1
}

// GetBounds gets the extents of the internal index nodes.
func (t *IndexHprtree_HPRtree) GetBounds() []*Geom_Envelope {
	numNodes := len(t.nodeBounds) / 4
	bounds := make([]*Geom_Envelope, numNodes)
	// create from largest to smallest
	for i := numNodes - 1; i >= 0; i-- {
		boundIndex := 4 * i
		bounds[i] = Geom_NewEnvelopeFromXY(
			t.nodeBounds[boundIndex], t.nodeBounds[boundIndex+2],
			t.nodeBounds[boundIndex+1], t.nodeBounds[boundIndex+3])
	}
	return bounds
}

func (t *IndexHprtree_HPRtree) sortItems() {
	encoder := IndexHprtree_NewHilbertEncoder(indexHprtree_HPRtree_HILBERT_LEVEL, t.totalExtent)
	hilbertValues := make([]int, len(t.itemsToLoad))
	for pos, item := range t.itemsToLoad {
		hilbertValues[pos] = encoder.Encode(item.GetEnvelope())
	}
	t.quickSortItemsIntoNodes(hilbertValues, 0, len(t.itemsToLoad)-1)
}

func (t *IndexHprtree_HPRtree) quickSortItemsIntoNodes(values []int, lo, hi int) {
	// stop sorting when left/right pointers are within the same node
	// because queryItems just searches through them all sequentially
	if lo/t.nodeCapacity < hi/t.nodeCapacity {
		pivot := t.hoarePartition(values, lo, hi)
		t.quickSortItemsIntoNodes(values, lo, pivot)
		t.quickSortItemsIntoNodes(values, pivot+1, hi)
	}
}

func (t *IndexHprtree_HPRtree) hoarePartition(values []int, lo, hi int) int {
	pivot := values[(lo+hi)>>1]
	i := lo - 1
	j := hi + 1

	for {
		for {
			i++
			if values[i] >= pivot {
				break
			}
		}
		for {
			j--
			if values[j] <= pivot {
				break
			}
		}
		if i >= j {
			return j
		}
		t.swapItems(values, i, j)
	}
}

func (t *IndexHprtree_HPRtree) swapItems(values []int, i, j int) {
	tmpItem := t.itemsToLoad[i]
	t.itemsToLoad[i] = t.itemsToLoad[j]
	t.itemsToLoad[j] = tmpItem

	tmpValue := values[i]
	values[i] = values[j]
	values[j] = tmpValue
}
