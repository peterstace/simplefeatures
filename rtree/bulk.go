package rtree

import "sort"

// BulkItem is an item that can be inserted for bulk loading.
type BulkItem struct {
	Box      Box
	DataIndex int
}

// BulkLoad bulk loads multiple items into a new R-Tree. The bulk load
// operation is optimised for creating R-Trees with minimal node overlap. This
// allows for fast searching.
func BulkLoad(items []BulkItem) RTree {
	var tr RTree
	n := tr.bulkInsert(items)
	tr.rootIndex = n
	return tr
}

func (t *RTree) bulkInsert(items []BulkItem) int {
	if len(items) <= 2 {
		node := node{isLeaf: true, parent: -1}
		for _, item := range items {
			node.entries = append(node.entries, entry{
				box:  item.Box,
				index: item.DataIndex,
			})
		}
		t.nodes = append(t.nodes, node)
		return len(t.nodes) - 1
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
	n1 := t.bulkInsert(items[:split])
	n2 := t.bulkInsert(items[split:])

	parent := node{isLeaf: false, parent: -1, entries: []entry{
		entry{box: t.calculateBound(n1), index: n1},
		entry{box: t.calculateBound(n2), index: n2},
	}}
	t.nodes = append(t.nodes, parent)
	t.nodes[n1].parent = len(t.nodes) - 1
	t.nodes[n2].parent = len(t.nodes) - 1
	return len(t.nodes) - 1
}
