package rtree

import "container/heap"

// Nearest finds the record in the RTree that is the closest to the input box
// as measured by the Euclidean metric. Note that there may be multiple records
// that are equidistant from the input box, in which case one is chosen
// arbitrarily. If the RTree is empty, then false is returned.
func (t *RTree) Nearest(box Box) (recordID int, found bool) {
	t.PrioritySearch(box, func(rid int) error {
		recordID = rid
		found = true
		return Stop
	})
	return recordID, found
}

// PrioritySearch iterates over the records in the RTree in priority order of
// distance from the input box (shortest distance first using the Euclidean
// metric).  The callback is called for every element iterated over. If an
// error is returned from the callback, then iteration stops immediately. Any
// error returned from the callback is returned by PrioritySearch, except for
// the case where the special Stop sentinel error is returned (in which case
// nil will be returned from PrioritySearch).
func (t *RTree) PrioritySearch(box Box, callback func(recordID int) error) error {
	if len(t.nodes) == 0 {
		return nil
	}

	queue := entriesQueue{origin: box}
	equeueNode := func(idx int) {
		n := &t.nodes[idx]
		for i := 0; i < n.numEntries; i++ {
			heap.Push(&queue, entriesQueueItem{
				entry:  &n.entries[i],
				isLeaf: n.isLeaf,
			})
		}
	}

	equeueNode(0)
	for len(queue.entries) > 0 {
		nearest := heap.Pop(&queue).(entriesQueueItem)
		if nearest.isLeaf {
			recordID := nearest.entry.data
			if err := callback(recordID); err != nil {
				if err == Stop {
					return nil
				}
				return err
			}
		} else {
			child := nearest.entry.data
			equeueNode(child)
		}
	}
	return nil
}

type entriesQueueItem struct {
	entry  *entry
	isLeaf bool
}

type entriesQueue struct {
	entries []entriesQueueItem
	origin  Box
}

func (q *entriesQueue) Len() int {
	return len(q.entries)
}

func (q *entriesQueue) Less(i int, j int) bool {
	d1 := squaredEuclideanDistance(q.entries[i].entry.box, q.origin)
	d2 := squaredEuclideanDistance(q.entries[j].entry.box, q.origin)
	return d1 < d2
}

func (q *entriesQueue) Swap(i int, j int) {
	q.entries[i], q.entries[j] = q.entries[j], q.entries[i]
}

func (q *entriesQueue) Push(x interface{}) {
	q.entries = append(q.entries, x.(entriesQueueItem))
}

func (q *entriesQueue) Pop() interface{} {
	e := q.entries[len(q.entries)-1]
	q.entries = q.entries[:len(q.entries)-1]
	return e
}
