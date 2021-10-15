package exact

import "sort"

type IntersectionReport struct {
	Indexes   [2]int
	Dimension int // 0 or 1
}

func BentlyOttmann(
	segments []Segment,
	callback func(IntersectionReport) bool,
) {
	var queue eventQueue
	for range segments {
		queue.push(event{})
	}
	// TODO
}

type event struct {
	// TODO
}

func (e event) cmp(o event) int {
	// TODO
	return -1
}

// TODO: the event queue implemented here is inefficient and just a
// placeholder.
type eventQueue struct {
	events []event
}

//func (q *eventQueue) empty() bool {
//	return len(q.events) == 0
//}

func (q *eventQueue) push(e event) {
	q.events = append(q.events, e)
	sort.Slice(q.events, func(i, j int) bool {
		return q.events[i].cmp(q.events[j]) > 0
	})
}

//func (q *eventQueue) pop() event {
//	e := q.events[len(q.events)-1]
//	q.events = q.events[:len(q.events)-1]
//	return e
//}
