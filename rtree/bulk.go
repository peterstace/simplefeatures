package rtree

// BulkItem is an item that can be inserted for bulk loading.
type BulkItem struct {
	Box      Box
	RecordID int
}

// BulkLoad bulk loads multiple items into a new R-Tree. The bulk load
// operation is optimised for creating R-Trees with minimal node overlap. This
// allows for fast searching.
func BulkLoad(items []BulkItem) *RTree {
	if len(items) == 0 {
		return &RTree{}
	}
	root := bulkInsert(items)
	return &RTree{root, len(items)}
}

func bulkInsert(items []BulkItem) *node {
	if len(items) == 0 {
		panic("should not have recursed into bulkInsert without any items")
	}

	// NOTE: bulk loading is hardcoded around the fact that the min and max
	// node cardinalities are 2 and 4.

	// 4 or fewer items can fit into a single node.
	if len(items) <= 4 {
		n := &node{numEntries: len(items)}
		for i, item := range items {
			n.entries[i] = entry{
				box:      item.Box,
				recordID: item.RecordID,
			}
		}
		return n
	}

	// 5 to 8 items are put into 3 nodes (one intermediate
	// node and two child nodes).
	if len(items) <= 8 {
		firstHalf, secondHalf := splitBulkItems2Ways(items)
		return bulkNode(firstHalf, secondHalf)
	}

	// 9 or more items are split into 4 groups, completely filling the
	// intermediate node.
	firstHalf, secondHalf := splitBulkItems2Ways(items)
	firstQuarter, secondQuarter := splitBulkItems2Ways(firstHalf)
	thirdQuarter, fourthQuarter := splitBulkItems2Ways(secondHalf)
	return bulkNode(firstQuarter, secondQuarter, thirdQuarter, fourthQuarter)
}

func bulkNode(parts ...[]BulkItem) *node {
	root := &node{numEntries: len(parts)}
	for i, part := range parts {
		child := bulkInsert(part)
		root.entries[i].child = child
		root.entries[i].box = calculateBound(child)
	}
	return root
}

func splitBulkItems2Ways(items []BulkItem) ([]BulkItem, []BulkItem) {
	horizontal := itemsAreHorizontal(items)
	split := len(items) / 2
	quickPartition(items, split, horizontal)
	return items[:split], items[split:]
}

// quickPartition performs a partial in-place sort on the items slice. The
// partial sort is such that items 0 through k-1 are less than or equal to item
// k, and items k+1 through n-1 are greater than or equal to item k.
func quickPartition(items []BulkItem, k int, horizontal bool) {
	// Use a custom linear congruential random number generator. This is used
	// because we don't need high quality random numbers. Using a regular
	// rand.Rand generator causes a significant bottleneck due to the reliance
	// on random numbers in this algorithm.
	var rndState uint32
	rnd := func(n int) int {
		rndState = 1664525*rndState + 1013904223
		return int((uint64(rndState) * uint64(n)) >> 32)
	}

	less := func(i, j int) bool {
		bi := items[i].Box
		bj := items[j].Box
		if horizontal {
			return bi.MinX+bi.MaxX < bj.MinX+bj.MaxX
		}
		return bi.MinY+bi.MaxY < bj.MinY+bj.MaxY
	}
	swap := func(i, j int) {
		items[i], items[j] = items[j], items[i]
	}

	left, right := 0, len(items)-1
	for {
		// For the case where there are 2 or 3 items remaining, we can use
		// special case logic to reduce the number of comparisons and swaps.
		switch right - left {
		case 1:
			if less(right, left) {
				swap(right, left)
			}
			return
		case 2:
			if less(left+1, left) {
				swap(left+1, left)
			}
			if less(left+2, left+1) {
				swap(left+2, left+1)
				if less(left+1, left) {
					swap(left+1, left)
				}
			}
			return
		}

		// Select pivot and store it at the end.
		pivot := left + rnd(right-left+1)
		if pivot != right {
			swap(pivot, right)
		}

		// Partition the left and right sides of the pivot.
		j := left
		for i := left; i < right; i++ {
			if less(i, right) {
				swap(i, j)
				j++
			}
		}

		// Restore the pivot to the middle position between the two partitions.
		swap(right, j)

		// Repeat on either the left or right parts depending on which contains
		// the kth element.
		switch {
		case j-left < k:
			k -= j - left + 1
			left = j + 1
		case j-left > k:
			right = j - 1
		default:
			return
		}
	}
}

func itemsAreHorizontal(items []BulkItem) bool {
	box := items[0].Box
	for _, item := range items[1:] {
		box = combine(box, item.Box)
	}
	return box.MaxX-box.MinX > box.MaxY-box.MinY
}

// fastMin is a faster but not functionally identical version of math.Min.
func fastMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// fastMax is a faster but not functionally identical version of math.Max.
func fastMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
