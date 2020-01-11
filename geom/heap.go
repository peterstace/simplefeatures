package geom

type intHeap struct {
	less func(i, j int) bool
	data []int
}

func (h *intHeap) push(elem int) {
	h.data = append(h.data, elem)
	i := len(h.data) - 1
	for i > 0 {
		parent := (i - 1) / 2
		if h.less(parent, i) {
			break
		}
		h.data[parent], h.data[i] = h.data[i], h.data[parent]
		i = parent
	}
}

func (h *intHeap) pop() {
	h.data[0] = h.data[len(h.data)-1]
	h.data = h.data[:len(h.data)-1]
	i := 0
	for {
		swapWith := -1
		childA := 2*i + 1
		childB := 2*i + 2
		switch {
		case childA < len(h.data) && childB < len(h.data):
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
		case childA < len(h.data):
			if h.less(childA, i) {
				swapWith = childA
			}
		case childB < len(h.data):
			if h.less(childB, i) {
				swapWith = childB
			}
		}
		if swapWith == -1 {
			break
		}
		h.data[swapWith], h.data[i] = h.data[i], h.data[swapWith]
		i = swapWith
	}
}
