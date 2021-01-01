package redblack

// Tree is a Red-Black self balancing binary search tree. In particular, it's
// the Left-Leaning Red-Black variant, invented by Robert Sedgewick.
//
// Tree holds an ordered set of int keys. This differs from regular tree
// implementations, which typically hold an ordered set of generic key/value
// pairs. If values need to be stored along with the keys, then these should be
// stored externally by the user of Tree. This is to avoid the need for Tree to
// manage various key and value types generically.
type Tree struct {
	root *node
}

// Compare is a function that compares two keys. It should return a negative
// number if key1 is less than key2, zero if key1 is equal to key2, and a
// positive number of key1 is greater than key2. This should define a total
// order on the set of keys (i.e. it is a relation that is antisymmetric,
// transitive, and connex).
type Compare func(key1, key2 int) int

type colour bool

const (
	red   colour = false
	black colour = true
)

type node struct {
	key    int
	left   *node
	right  *node
	colour colour
}

func isRed(h *node) bool {
	return h != nil && h.colour == red
}

func colourFlip(h *node) {
	h.colour = !h.colour
	h.left.colour = !h.left.colour
	h.right.colour = !h.right.colour
}

func rotateLeft(h *node) *node {
	x := h.right
	h.right = x.left
	x.left = h
	x.colour = h.colour
	h.colour = red
	return x
}

func rotateRight(h *node) *node {
	x := h.left
	h.left = x.right
	x.right = h
	x.colour = h.colour
	h.colour = red
	return x
}

type childDirection bool

const (
	left  childDirection = false
	right childDirection = true
)

func (d childDirection) other() childDirection {
	return !d
}

func (h *node) child(dir childDirection) *node {
	if dir == left {
		return h.left
	}
	return h.right
}
