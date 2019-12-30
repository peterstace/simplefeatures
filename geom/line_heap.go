package geom

type lineHeap struct {
	less func(a, b Line) bool
	data []Line
}

func (h *lineHeap) peek() Line {
	return h.data[0]
}

func (h *lineHeap) empty() bool {
	return len(h.data) == 0
}

func (h *lineHeap) push(ln Line) {
	h.data = append(h.data, ln)
	i := len(h.data) - 1
	for i > 0 {
		parent := (i - 1) / 2
		if h.lt(parent, i) {
			break
		}
		h.data[parent], h.data[i] = h.data[i], h.data[parent]
		i = parent
	}
}

func (h *lineHeap) pop() {
	h.data[0] = h.data[len(h.data)-1]
	h.data = h.data[:len(h.data)-1]
	i := 0
	for {
		swapWith := -1
		childA := 2*i + 1
		childB := 2*i + 2
		switch {
		case childA < len(h.data) && childB < len(h.data):
			swapWith = childA
			if h.lt(childB, childA) {
				swapWith = childB
			}
		case childA < len(h.data):
			if h.lt(childA, i) {
				swapWith = childA
			}
		case childB < len(h.data):
			if h.lt(childB, i) {
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

func (h *lineHeap) lt(i, j int) bool {
	return h.less(h.data[i], h.data[j])
}
