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

	levels := calculateLevels(len(items))
	return &RTree{bulkInsert(items, levels), len(items)}
}

func calculateLevels(numItems int) int {
	// We could theoretically do this calculation using math.Log. However,
	// float precision issues can cause off-by-one errors in some scenarios.
	// Instead, we calculate the number of levels using integer arithmetic
	// only. This will be fast anyway, since the calculation only requires
	// logarithmic time.
	levels := 1
	count := maxChildren
	for count < numItems {
		count *= maxChildren
		levels++
	}
	return levels
}

func bulkInsert(items []BulkItem, levels int) *node {
	if levels == 1 {
		root := &node{isLeaf: true, numEntries: len(items)}
		for i, item := range items {
			root.entries[i] = entry{
				box:      item.Box,
				recordID: item.RecordID,
			}
		}
		return root
	}

	// NOTE: bulk loading is hardcoded around the fact that the min and max
	// node cardinalities are 2 and 4.

	// 6 is the first number of items that can be split into 3 nodes while
	// respecting the minimum node cardinality, i.e. 6 = 2 + 2 + 2. Anything
	// less than 6 must instead be split into 2 nodes.
	if len(items) < 6 {
		firstHalf, secondHalf := splitBulkItems2Ways(items)
		return bulkNode(levels, firstHalf, secondHalf)
	}

	// 8 is the first number of items that can be split into 4 nodes while
	// respecting the minimum node cardinality, i.e. 8 = 2 + 2 + 2 + 2.
	// Anything less that 8 must instead be split into 3 nodes.
	if len(items) < 8 {
		firstThird, secondThird, thirdThird := splitBulkItems3Ways(items)
		return bulkNode(levels, firstThird, secondThird, thirdThird)
	}

	// 4-way split:
	firstHalf, secondHalf := splitBulkItems2Ways(items)
	firstQuarter, secondQuarter := splitBulkItems2Ways(firstHalf)
	thirdQuarter, fourthQuarter := splitBulkItems2Ways(secondHalf)
	return bulkNode(levels, firstQuarter, secondQuarter, thirdQuarter, fourthQuarter)
}

func bulkNode(levels int, parts ...[]BulkItem) *node {
	root := &node{
		numEntries: len(parts),
		parent:     nil,
		isLeaf:     false,
	}
	for i, part := range parts {
		child := bulkInsert(part, levels-1)
		child.parent = root
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

func splitBulkItems3Ways(items []BulkItem) ([]BulkItem, []BulkItem, []BulkItem) {
	// We only need to split 3 ways when we have 6 or 7 elements. By making use
	// of that assumption, we greatly simplify the logic in this function
	// compared to if we were to implement a 3 way split in the general case.
	if ln := len(items); ln != 6 && ln != 7 {
		panic(len(items))
	}

	horizontal := itemsAreHorizontal(items)
	quickPartition(items, 2, horizontal)
	quickPartition(items[3:], 1, horizontal)

	return items[:2], items[2:4], items[4:]
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
	// fastMin and fastMax are used rather than Box's combine method to avoid
	// math.Min and math.Max calls (which are more expensive).
	minX := items[0].Box.MinX
	maxX := items[0].Box.MaxX
	minY := items[0].Box.MinY
	maxY := items[0].Box.MaxY
	for _, item := range items[1:] {
		box := item.Box
		minX = fastMin(minX, box.MinX)
		maxX = fastMax(maxX, box.MaxX)
		minY = fastMin(minY, box.MinY)
		maxY = fastMax(maxY, box.MaxY)
	}
	return maxX-minX > maxY-minY
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
