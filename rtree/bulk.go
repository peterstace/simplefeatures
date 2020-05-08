package rtree

import "sort"

// BulkItem is an item that can be inserted for bulk loading.
type BulkItem struct {
	Box      Box
	RecordID int
}

// BulkLoad bulk loads multiple items into a new R-Tree. The bulk load
// operation is optimised for creating R-Trees with minimal node overlap. This
// allows for fast searching.
func BulkLoad(items []BulkItem) RTree {
	var tr RTree
	n := tr.bulkInsert(items)
	tr.root = n
	return tr
}

func (t *RTree) bulkInsert(items []BulkItem) *node {
	if len(items) == 0 {
		return nil
	}
	if len(items) <= 2 {
		root := &node{isLeaf: true, numEntries: len(items)}
		for i, item := range items {
			root.entries[i] = entry{
				box:      item.Box,
				recordID: item.RecordID,
			}
		}
		return root
	}

	box := items[0].Box
	for _, item := range items[1:] {
		box = combine(box, item.Box)
	}

	horizontal := box.MaxX-box.MinX > box.MaxY-box.MinY
	sort.Slice(items, func(i, j int) bool {
		bi := items[i].Box
		bj := items[j].Box
		if horizontal {
			return bi.MinX+bi.MaxX < bj.MinX+bj.MaxX
		} else {
			return bi.MinY+bi.MaxY < bj.MinY+bj.MaxY
		}
	})

	split := len(items) / 2
	childA := t.bulkInsert(items[:split])
	childB := t.bulkInsert(items[split:])

	root := &node{
		entries: [1 + maxChildren]entry{
			entry{box: calculateBound(childA), child: childA},
			entry{box: calculateBound(childB), child: childB},
		},
		numEntries: 2,
		parent:     nil,
		isLeaf:     false,
	}
	childA.parent = root
	childB.parent = root

	return root
}
