package rtree

import (
	"errors"
)

const (
	minEntries = 2
	maxEntries = 4
)

// node is a node in an R-Tree, holding user record IDs and/or links to deeper
// nodes in the tree.
type node struct {
	entries    [maxEntries]entry
	numEntries int
}

// entry is an entry contained inside a node. An entry can either hold a user
// record ID, or point to a deeper node in the tree (but not both). Because 0
// is a valid record ID, the child pointer should be used to distinguish
// between the two types of entries.
type entry struct {
	box      Box
	child    *node
	recordID int
}

// RTree is an in-memory R-Tree data structure. It holds record ID and bounding
// box pairs (the actual records aren't stored in the tree; the user is
// responsible for storing their own records). Its zero value is an empty
// R-Tree.
type RTree struct {
	root  *node
	count int
}

// Stop is a special sentinel error that can be used to stop a search operation
// without any error.
var Stop = errors.New("stop") //nolint:stylecheck,revive

// RangeSearch looks for any items in the tree that overlap with the given
// bounding box. The callback is called with the record ID for each found item.
// If an error is returned from the callback then the search is terminated
// early.  Any error returned from the callback is returned by RangeSearch,
// except for the case where the special Stop sentinel error is returned (in
// which case nil will be returned from RangeSearch). Stop may be wrapped.
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
			if entry.child == nil {
				if err := callback(entry.recordID); errors.Is(err, Stop) {
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
