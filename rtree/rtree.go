package rtree

import (
	"errors"
)

// node is a node in an R-Tree. nodes can either be leaf nodes holding entries
// for terminal items, or intermediate nodes holding entries for more nodes.
type node struct {
	entries    [1 + maxChildren]entry
	numEntries int
	parent     *node
	isLeaf     bool
}

// entry is an entry under a node, leading either to terminal items, or more nodes.
type entry struct {
	box Box

	// For leaf nodes, recordID is populated. For non-leaf nodes, child is populated.
	child    *node
	recordID int
}

func (n *node) appendRecord(box Box, recordID int) {
	n.entries[n.numEntries] = entry{box: box, recordID: recordID}
	n.numEntries++
}

func (n *node) appendChild(box Box, child *node) {
	n.entries[n.numEntries] = entry{box: box, child: child}
	n.numEntries++
	child.parent = n
}

// depth calculates the number of layers of nodes in the subtree rooted at the node.
func (n *node) depth() int {
	var d = 1
	for !n.isLeaf {
		d++
		n = n.entries[0].child
	}
	return d
}

// RTree is an in-memory R-Tree data structure. It holds record ID and bounding
// box pairs (the actual records aren't stored in the tree; the user is
// responsible for storing their own records). Its zero value is an empty
// R-Tree.
type RTree struct {
	root  *node
	count int
}

// Stop is a special sentinal error that can be used to stop a search operation
// without any error.
var Stop = errors.New("stop")

// RangeSearch looks for any items in the tree that overlap with the given
// bounding box. The callback is called with the record ID for each found item.
// If an error is returned from the callback then the search is terminated
// early.  Any error returned from the callback is returned by RangeSearch,
// except for the case where the special Stop sentinal error is returned (in
// which case nil will be returned from RangeSearch).
func (t *RTree) RangeSearch(box Box, callback func(recordID int) error) error {
	if t.root == nil {
		return nil
	}
	var recurse func(*node) error
	recurse = func(n *node) error {
		for i := 0; i < n.numEntries; i++ {
			entry := n.entries[i]
			if !overlap(entry.box, box) {
				continue
			}
			if n.isLeaf {
				if err := callback(entry.recordID); err == Stop {
					return nil
				} else if err != nil {
					return err
				}
			} else {
				if err := recurse(entry.child); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return recurse(t.root)
}

// Extent gives the Box that most closely bounds the RTree. If the RTree is
// empty, then false is returned.
func (t *RTree) Extent() (Box, bool) {
	if t.root == nil || t.root.numEntries == 0 {
		return Box{}, false
	}
	return calculateBound(t.root), true
}

// Count gives the number of entries in the RTree.
func (t *RTree) Count() int {
	return t.count
}
