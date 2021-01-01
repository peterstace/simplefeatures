package redblack

// Insert adds a key to the set. If it already exists, then nothing happens.
func (t *Tree) Insert(key int, cmp Compare) {
	t.root = insert(t.root, key, cmp)
	t.root.colour = black
}

func insert(h *node, key int, cmp Compare) *node {
	if h == nil {
		return &node{key: key /*, colour: red*/}
	}

	c := cmp(key, h.key)
	if c < 0 {
		h.left = insert(h.left, key, cmp)
	} else if c > 0 {
		h.right = insert(h.right, key, cmp)
	} else {
		// Do nothing for the c == 0 case.
	}

	return fixUp(h)
}

func fixUp(h *node) *node {
	// Rotate-left right-leaning reds.
	if isRed(h.right) {
		h = rotateLeft(h)
	}

	// Rotate-right red-red pairs.
	if isRed(h.left) && isRed(h.left.left) {
		h = rotateRight(h)
	}

	// Split 4-nodes.
	if isRed(h.left) && isRed(h.right) {
		colourFlip(h)
	}
	return h
}
