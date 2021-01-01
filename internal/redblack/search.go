package redblack

// Contains searches for key in the Tree to see if it exists.
func (t *Tree) Contains(key int, cmp Compare) bool {
	h := t.root
	for h != nil {
		c := cmp(key, h.key)
		switch {
		case c < 0:
			h = h.left
		case c > 0:
			h = h.right
		default:
			return true
		}
	}
	return false
}
