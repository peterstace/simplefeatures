package redblack

import (
	"fmt"
	"strconv"
)

func (t *Tree) Print() {
	printSubTree(t.root, "")
}

func printSubTree(h *node, indent string) {
	c := "BLACK"
	if isRed(h) {
		c = "RED"
	}
	key := "nil"
	if h != nil {
		key = strconv.Itoa(h.key)
	}
	fmt.Printf("%s %s key=%v\n", indent, c, key)
	if h != nil {
		indent = "  " + indent
		printSubTree(h.left, indent)
		printSubTree(h.right, indent)
	}
}
