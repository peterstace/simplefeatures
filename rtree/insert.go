package rtree

import (
	"math"
	"math/bits"
)

const (
	minChildren = 2
	maxChildren = 4
)

// Insert adds a new record to the RTree.
func (t *RTree) Insert(box Box, recordID int) {
	if t.root == nil {
		t.root = &node{isLeaf: true}
	}

	leaf := t.chooseLeafNode(box)
	leaf.appendRecord(box, recordID)

	current := leaf
	for current != t.root {
		parent := current.parent
		for i := 0; i < parent.numEntries; i++ {
			e := &parent.entries[i]
			if e.child == current {
				e.box = combine(e.box, box)
			}
		}
		current = parent
	}

	if leaf.numEntries <= maxChildren {
		return
	}

	newNode := t.splitNode(leaf)
	root1, root2 := t.adjustTree(leaf, newNode)

	if root2 != nil {
		t.joinRoots(root1, root2)
	}
}

func (t *RTree) joinRoots(r1, r2 *node) {
	newRoot := &node{
		entries: [1 + maxChildren]entry{
			entry{box: calculateBound(r1), child: r1},
			entry{box: calculateBound(r2), child: r2},
		},
		numEntries: 2,
		parent:     nil,
		isLeaf:     false,
	}
	r1.parent = newRoot
	r2.parent = newRoot
	t.root = newRoot
}

// TODO: rename n and nn to leaf and newNode
func (t *RTree) adjustTree(n, nn *node) (*node, *node) {
	for {
		if n == t.root {
			return n, nn
		}
		parent := n.parent
		for i := 0; i < parent.numEntries; i++ {
			if parent.entries[i].child == n {
				parent.entries[i].box = calculateBound(n)
				break
			}
		}

		// AT4
		var pp *node
		if nn != nil {
			parent.appendChild(calculateBound(nn), nn)
			nn.parent = parent
			if parent.numEntries > maxChildren {
				pp = t.splitNode(parent)
			}
		}

		n, nn = parent, pp
	}
}

// splitNode splits node with index n into two nodes. The first node replaces
// n, and the second node is newly created. The return value is the index of
// the new node.
func (t *RTree) splitNode(n *node) *node {
	var (
		// All zeros would not be valid split, so start at 1.
		minSplit = uint64(1)
		// The MSB should always be 0, to remove duplicates from inverting the
		// bit pattern. So we raise 2 to the power of one less than the number
		// of entries rather than the number of entries.
		//
		// E.g. for 4 entries, we want the following bit patterns:
		// 0001, 0010, 0011, 0100, 0101, 0110, 0111.
		//
		// (1 << (4 - 1)) - 1 == 0111, so the maths checks out.
		maxSplit = uint64((1 << (n.numEntries - 1)) - 1)
	)
	bestArea := math.Inf(+1)
	var bestSplit uint64
	for split := minSplit; split <= maxSplit; split++ {
		if bits.OnesCount64(split) < minChildren {
			continue
		}
		var boxA, boxB Box
		var hasA, hasB bool
		for i := 0; i < n.numEntries; i++ {
			entryBox := n.entries[i].box
			if split&(1<<i) == 0 {
				if hasA {
					boxA = combine(boxA, entryBox)
				} else {
					boxA = entryBox
				}
			} else {
				if hasB {
					boxB = combine(boxB, entryBox)
				} else {
					boxB = entryBox
				}
			}
		}
		combinedArea := area(boxA) + area(boxB)
		if combinedArea < bestArea {
			bestArea = combinedArea
			bestSplit = split
		}
	}

	var entriesA, entriesB []entry
	for i := 0; i < n.numEntries; i++ {
		entry := n.entries[i]
		if bestSplit&(1<<i) == 0 {
			entriesA = append(entriesA, entry)
		} else {
			entriesB = append(entriesB, entry)
		}
	}

	// Use the existing node for A...
	copy(n.entries[:], entriesA)
	n.numEntries = len(entriesA)

	// And create a new node for B.
	newNode := &node{
		numEntries: len(entriesB),
		parent:     nil,
		isLeaf:     n.isLeaf,
	}
	copy(newNode.entries[:], entriesB)
	if !n.isLeaf {
		for i := 0; i < newNode.numEntries; i++ {
			newNode.entries[i].child.parent = newNode
		}
	}
	return newNode
}

func (t *RTree) chooseLeafNode(box Box) *node {
	node := t.root
	for {
		if node.isLeaf {
			return node
		}
		bestDelta := enlargement(box, node.entries[0].box)
		bestEntry := 0
		for i := 1; i < node.numEntries; i++ {
			entryBox := node.entries[i].box
			delta := enlargement(box, entryBox)
			if delta < bestDelta {
				bestDelta = delta
				bestEntry = i
			} else if delta == bestDelta && area(entryBox) < area(node.entries[bestEntry].box) {
				// Area is used as a tie breaking if the enlargements are the same.
				bestEntry = i
			}
		}
		node = node.entries[bestEntry].child
	}
}
