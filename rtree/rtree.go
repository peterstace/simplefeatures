package rtree

import (
	"errors"
)

const (
	minEntries = 2
	maxEntries = 4
)

// node is a node in an R-Tree, holding user records and/or links to deeper
// nodes in the tree.
type node[T any] struct {
	entries    [maxEntries]entry[T]
	numEntries int
}

// entry is an entry contained inside a node. An entry can either hold a user
// record, or point to a deeper node in the tree (but not both). The child
// pointer should be used to distinguish between the two types of entries.
type entry[T any] struct {
	box    Box
	child  *node[T]
	record T
}

// RTree is an in-memory R-Tree data structure. It holds records of type T
// along with their bounding boxes. Its zero value is an empty R-Tree.
type RTree[T any] struct {
	root  *node[T]
	count int
}

// Stop is a special sentinel error that can be used to stop a search operation
// without any error.
var Stop = errors.New("stop") //nolint:stylecheck,revive

// RangeSearch looks for any items in the tree that overlap with the given
// bounding box. The callback is called with each found item's record. If an
// error is returned from the callback then the search is terminated early.
// Any error returned from the callback is returned by RangeSearch, except for
// the case where the special Stop sentinel error is returned (in which case
// nil will be returned from RangeSearch). Stop may be wrapped.
func (t *RTree[T]) RangeSearch(box Box, callback func(record T) error) error {
	if t.root == nil {
		return nil
	}
	var recurse func(*node[T]) error
	recurse = func(n *node[T]) error {
		for i := 0; i < n.numEntries; i++ {
			entry := n.entries[i]
			if !overlap(entry.box, box) {
				continue
			}
			if entry.child == nil {
				if err := callback(entry.record); errors.Is(err, Stop) {
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
func (t *RTree[T]) Extent() (Box, bool) {
	if t.root == nil || t.root.numEntries == 0 {
		return Box{}, false
	}
	return calculateBound(t.root), true
}

// Count gives the number of entries in the RTree.
func (t *RTree[T]) Count() int {
	return t.count
}
