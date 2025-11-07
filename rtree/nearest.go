package rtree

import (
	"container/heap"
	"errors"
)

// Nearest finds the record in the RTree that is the closest to the input box
// as measured by the Euclidean metric. Note that there may be multiple records
// that are equidistant from the input box, in which case one is chosen
// arbitrarily. If the RTree is empty, then false is returned.
func (t *RTree[T]) Nearest(box Box) (record T, found bool) {
	_ = t.PrioritySearch(box, func(rec T) error {
		record = rec
		found = true
		return Stop
	})
	return record, found
}

// PrioritySearch iterates over the records in the RTree in priority order of
// distance from the input box (shortest distance first using the Euclidean
// metric).  The callback is called for every element iterated over. If an
// error is returned from the callback, then iteration stops immediately. Any
// error returned from the callback is returned by PrioritySearch, except for
// the case where the special Stop sentinel error is returned (in which case
// nil will be returned from PrioritySearch). Stop may be wrapped.
func (t *RTree[T]) PrioritySearch(box Box, callback func(record T) error) error {
	if t.root == nil {
		return nil
	}

	queue := entriesQueue[T]{origin: box}
	equeueNode := func(n *node[T]) {
		for i := 0; i < n.numEntries; i++ {
			heap.Push(&queue, &n.entries[i])
		}
	}

	equeueNode(t.root)
	for len(queue.entries) > 0 {
		nearest := heap.Pop(&queue).(*entry[T])
		if nearest.child == nil {
			if err := callback(nearest.record); err != nil {
				if errors.Is(err, Stop) {
					return nil
				}
				return err
			}
		} else {
			equeueNode(nearest.child)
		}
	}
	return nil
}

type entriesQueue[T any] struct {
	entries []*entry[T]
	origin  Box
}

func (q *entriesQueue[T]) Len() int {
	return len(q.entries)
}

func (q *entriesQueue[T]) Less(i int, j int) bool {
	d1 := squaredEuclideanDistance(q.entries[i].box, q.origin)
	d2 := squaredEuclideanDistance(q.entries[j].box, q.origin)
	return d1 < d2
}

func (q *entriesQueue[T]) Swap(i int, j int) {
	q.entries[i], q.entries[j] = q.entries[j], q.entries[i]
}

func (q *entriesQueue[T]) Push(x interface{}) {
	q.entries = append(q.entries, x.(*entry[T]))
}

func (q *entriesQueue[T]) Pop() interface{} {
	e := q.entries[len(q.entries)-1]
	q.entries = q.entries[:len(q.entries)-1]
	return e
}
