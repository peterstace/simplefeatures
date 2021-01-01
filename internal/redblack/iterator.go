package redblack

import (
	"fmt"
	"strconv"
	"strings"
)

// Iterator allows a Tree to be iterated over in a bidirectional manner.  No
// iterator invalidation guarantees are made (i.e. iterators may become invalid
// if the Tree is modified).
type Iterator struct {
	tree  *Tree
	state iterState
	nodes []*node
}

type iterState byte

const (
	begin iterState = iota + 1
	valid
	end
)

// Begin returns an iterator that is positioned just before the first element
// in the Tree.
func (t *Tree) Begin() *Iterator {
	return &Iterator{t, begin, nil}
}

// End returns an iterator that is positioned just after the last element in
// the Tree.
func (t *Tree) End() *Iterator {
	return &Iterator{t, end, nil}
}

// Next advances the iterator to the next position in the tree. If Next is
// called when the iterator is positioned just after the last element, then it
// will panic.
func (i *Iterator) Next() bool {
	return i.move(true)
}

// Prev reverses the iterator to the previous position in the tree. If Prev is
// called when the iterator is positioned just before the first element, then
// it will panic.
func (i *Iterator) Prev() bool {
	return i.move(false)
}

func (i *Iterator) move(fwd bool) bool {
	dir := left
	if !fwd {
		dir = right
	}
	switch {
	case fwd && i.state == begin || !fwd && i.state == end:
		i.state = valid
		i.pushNodes(i.tree.root, dir)
		return i.checkValid(fwd)
	case i.state == valid:
		child := i.current().child(dir.other())
		if child != nil {
			i.pushNodes(child, dir)
			return true
		}
		i.popUntilAscendDir(dir)
		return i.checkValid(fwd)
	case fwd && i.state == end || !fwd && i.state == begin:
		panic("move() called on iterator in wrong state")
	default:
		panic("invalid state: " + strconv.Itoa(int(i.state)))
	}
}

func (i *Iterator) checkValid(fwd bool) bool {
	if len(i.nodes) == 0 {
		if fwd {
			i.state = end
		} else {
			i.state = begin
		}
		return false
	}
	return true
}

func (i *Iterator) pushNodes(h *node, dir childDirection) {
	for h != nil {
		i.nodes = append(i.nodes, h)
		h = h.child(dir)
	}
}

func (i *Iterator) popUntilAscendDir(dir childDirection) {
	for len(i.nodes) > 0 {
		n := len(i.nodes)
		popped := i.nodes[n-1]
		i.nodes = i.nodes[:n-1]
		if len(i.nodes) == 0 || i.nodes[n-2].child(dir) == popped {
			break
		}
	}
}

// Key gives the key at the current iterator position. It panics if the
// iterator is not positioned at a Tree entry.
func (i *Iterator) Key() int {
	return i.current().key
}

func (i *Iterator) current() *node {
	if i.state != valid {
		panic("iterator in invalid state")
	}
	return i.nodes[len(i.nodes)-1]
}

// String gives a textual representation of the iterator state.
func (i *Iterator) String() string {
	var nodes strings.Builder
	nodes.WriteByte('(')
	for j, h := range i.nodes {
		if j != 0 {
			nodes.WriteString(", ")
		}
		nodes.WriteString(strconv.Itoa(h.key))
	}
	nodes.WriteByte(')')
	return fmt.Sprintf(
		"state=%s nodes=%s",
		map[iterState]string{
			begin: "begin",
			valid: "valid",
			end:   "end",
		}[i.state],
		nodes.String(),
	)
}

// Seek searches for key in the Tree, and positions the iterator to it (if it
// can be found). If it can't be found, then the iterator is positioned just
// before the first element in the tree. The bool return indicates if the key
// could be found.
func (i *Iterator) Seek(key int, cmp Compare) bool {
	i.state = begin
	i.nodes = nil
	h := i.tree.root
	for h != nil {
		i.nodes = append(i.nodes, h)
		c := cmp(key, h.key)
		switch {
		case c < 0:
			h = h.left
		case c > 0:
			h = h.right
		default:
			i.state = valid
			return true
		}
	}
	if i.state != valid {
		i.nodes = i.nodes[:0]
	}
	return false
}
