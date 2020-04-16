package rtree

import (
	"errors"
)

// node is a node in an R-Tree. nodes can either be leaf nodes holding entries
// for terminal items, or intermediate nodes holding entries for more nodes.
type node struct {
	isLeaf  bool
	entries []entry
	parent  int
}

// entry is an entry under a node, leading either to terminal items, or more nodes.
type entry struct {
	box Box

	// For leaf nodes, this is a record ID. For non-leaf nodes, this is the
	// index of the child node.
	index int
}

// RTree is an in-memory R-Tree data structure. It holds record ID and bounding
// box pairs (the actual records aren't stored in the tree; the user is
// responsible for storing their own records). Its zero value is an empty
// R-Tree.
type RTree struct {
	rootIndex int
	nodes     []node
}

// Stop is a special sentinal error that can be used to stop a search operation
// without any error.
var Stop = errors.New("stop")

// Search looks for any items in the tree that overlap with the given bounding
// box. The callback is called with the record ID for each found item. If an
// error is returned from the callback then the search is terminated early.
// Any error returned from the callback is returned by Search, except for the
// case where the special Stop sentinal error is returned (in which case nil
// will be returned from Search).
func (t *RTree) Search(box Box, callback func(recordID int) error) error {
	if len(t.nodes) == 0 {
		return nil
	}
	var recurse func(*node) error
	recurse = func(n *node) error {
		for _, entry := range n.entries {
			if !overlap(entry.box, box) {
				continue
			}
			if n.isLeaf {
				if err := callback(entry.index); err == Stop {
					return nil
				} else if err != nil {
					return err
				}
			} else {
				if err := recurse(&t.nodes[entry.index]); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return recurse(&t.nodes[t.rootIndex])
}
