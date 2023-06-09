package rtree

import (
	"errors"
	"fmt"
	"unsafe"
)

const (
	minChildren = 2
	maxChildren = 4
)

// node is a node in an R-Tree. nodes can either be leaf nodes holding entries
// for terminal items, or intermediate nodes holding entries for more nodes.
type node struct {
	entries [maxChildren]entry

	// TODO: could save some memory here by using uint8 and bool.
	numEntries int
	isLeaf     bool
}

func init() {
	// 208 to start with
	fmt.Println("DEBUG rtree/rtree.go:23 unsafe.Sizeof(node{})", unsafe.Sizeof(node{})) // XXX
}

// entry is an entry under a node, leading either to terminal items, or more nodes.
type entry struct {
	box Box

	// For leaf nodes, data is the user's recordID. For non-leaf nodes, data is
	// an index to a child node.
	data int
}

// RTree is an in-memory R-Tree data structure. It holds record ID and bounding
// box pairs (the actual records aren't stored in the tree; the user is
// responsible for storing their own records). Its zero value is an empty
// R-Tree.
type RTree struct {
	// TODO: the root node will be the last node in the slice. Could things be
	// restructured to make it the first node in the slice instead? That feels
	// a little bit more sane.
	nodes []node
	count int
}

// Stop is a special sentinel error that can be used to stop a search operation
// without any error.
var Stop = errors.New("stop")

// RangeSearch looks for any items in the tree that overlap with the given
// bounding box. The callback is called with the record ID for each found item.
// If an error is returned from the callback then the search is terminated
// early.  Any error returned from the callback is returned by RangeSearch,
// except for the case where the special Stop sentinal error is returned (in
// which case nil will be returned from RangeSearch).
func (t *RTree) RangeSearch(box Box, callback func(recordID int) error) error {
	if len(t.nodes) == 0 {
		return nil
	}
	var recurse func(int) error
	recurse = func(nodeIdx int) error {
		n := t.nodes[nodeIdx]
		for i := 0; i < n.numEntries; i++ {
			entry := n.entries[i]
			if !overlap(entry.box, box) {
				continue
			}
			if n.isLeaf {
				if err := callback(entry.data); err == Stop {
					return nil
				} else if err != nil {
					return err
				}
			} else {
				if err := recurse(entry.data); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return recurse(0)
}

// Extent gives the Box that most closely bounds the RTree. If the RTree is
// empty, then false is returned.
func (t *RTree) Extent() (Box, bool) {
	if len(t.nodes) == 0 {
		return Box{}, false
	}
	if t.nodes[0].numEntries == 0 {
		return Box{}, false
	}
	return t.calculateBound(0), true
}

// Count gives the number of entries in the RTree.
func (t *RTree) Count() int {
	return t.count
}
