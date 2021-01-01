package redblack

// Delete removes a key from the set. If the key doesn't exist, then nothing
// happens.
func (t *Tree) Delete(key int, cmp Compare) {
	if t.root != nil {
		t.root = del(t.root, key, cmp)
	}
	if t.root != nil {
		t.root.colour = black
	}
}

func moveRedLeft(h *node) *node {
	colourFlip(h)
	if isRed(h.right.left) {
		h.right = rotateRight(h.right)
		h = rotateLeft(h)
		colourFlip(h)
	}
	return h
}

func moveRedRight(h *node) *node {
	colourFlip(h)
	if isRed(h.left.left) {
		h = rotateRight(h)
		colourFlip(h)
	}
	return h
}

func del(h *node, key int, cmp Compare) *node {
	if cmp(key, h.key) < 0 {
		if !isRed(h.left) && !isRed(h.left.left) {
			h = moveRedLeft(h)
		}
		h.left = del(h.left, key, cmp)
	} else {
		if isRed(h.left) {
			h = rotateRight(h)
		}
		if cmp(key, h.key) == 0 && h.right == nil {
			return nil
		}
		if !isRed(h.right) && !isRed(h.right.left) {
			h = moveRedRight(h)
		}
		if cmp(key, h.key) == 0 {
			h.key = min(h.right).key
			h.right = deleteMin(h.right)
		} else {
			h.right = del(h.right, key, cmp)
		}
	}
	return fixUp(h)
}

func min(h *node) *node {
	for h.left != nil {
		h = h.left
	}
	return h
}

func deleteMin(h *node) *node {
	if h.left == nil {
		return nil
	}
	if !isRed(h.left) && !isRed(h.left.left) {
		h = moveRedLeft(h)
	}
	h.left = deleteMin(h.left)
	return fixUp(h)
}
