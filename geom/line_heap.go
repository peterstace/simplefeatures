package geom

// lineHeap is a binary heap data structure that contains Lines. The advantage
// of this implementation of a heap over the the standard container/heap
// package is that it doesn't use interface{} (and therefore doesn't allocate
// memory on each heap operation). The obvious disadvantage is that it is a
// non-trivial implementation of something that already exists. The trade off
// is worth it because the heap is used within tight loops.
type lineHeap []Line

func (h *lineHeap) push(ln Line) {
	*h = append(*h, ln)
	i := len(*h) - 1
	for i > 0 {
		parent := (i - 1) / 2
		if h.less(parent, i) {
			break
		}
		(*h)[parent], (*h)[i] = (*h)[i], (*h)[parent]
		i = parent
	}
}

func (h *lineHeap) pop() {
	(*h)[0] = (*h)[len((*h))-1]
	(*h) = (*h)[:len((*h))-1]
	i := 0
	for {
		swapWith := -1
		childA := 2*i + 1
		childB := 2*i + 2
		switch {
		case childA < len((*h)) && childB < len((*h)):
			if h.less(i, childA) {
				if h.less(childB, i) {
					swapWith = childB
				}
			} else {
				swapWith = childA
				if h.less(childB, childA) {
					swapWith = childB
				}
			}
		case childA < len((*h)):
			if h.less(childA, i) {
				swapWith = childA
			}
		case childB < len((*h)):
			if h.less(childB, i) {
				swapWith = childB
			}
		}
		if swapWith == -1 {
			break
		}
		(*h)[swapWith], (*h)[i] = (*h)[i], (*h)[swapWith]
		i = swapWith
	}
}

func (h *lineHeap) less(i, j int) bool {
	// If the lineHeap needs to be used in more than one place with a different
	// less function, then this will need to be made generic.
	return (*h)[i].EndPoint().XY().X < (*h)[j].EndPoint().XY().X
}
