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
func (t *RTree) Insert(box Box, recordID uint64) {
	if len(t.nodes) == 0 {
		t.nodes = append(t.nodes, node{isLeaf: true, entries: nil, parent: -1})
		t.rootIndex = 0
	}

	leaf := t.chooseLeafNode(box)
	t.nodes[leaf].entries = append(t.nodes[leaf].entries, entry{box: box, recordID: recordID})

	current := leaf
	for current != t.rootIndex {
		parent := t.nodes[current].parent
		for i := range t.nodes[parent].entries {
			e := &t.nodes[parent].entries[i]
			if e.child == current {
				e.box = combine(e.box, box)
				break
			}
		}
		current = parent
	}

	if len(t.nodes[leaf].entries) <= maxChildren {
		return
	}

	newNode := t.splitNode(leaf)
	root1, root2 := t.adjustTree(leaf, newNode)

	if root2 != -1 {
		t.joinRoots(root1, root2)
	}
}

func (t *RTree) joinRoots(r1, r2 int) {
	t.nodes = append(t.nodes, node{
		isLeaf: false,
		entries: []entry{
			entry{
				box:   t.calculateBound(r1),
				child: r1,
			},
			entry{
				box:   t.calculateBound(r2),
				child: r2,
			},
		},
		parent: -1,
	})
	t.rootIndex = len(t.nodes) - 1
	t.nodes[r1].parent = len(t.nodes) - 1
	t.nodes[r2].parent = len(t.nodes) - 1
}

func (t *RTree) adjustTree(n, nn int) (int, int) {
	for {
		if n == t.rootIndex {
			return n, nn
		}
		parent := t.nodes[n].parent
		parentEntry := -1
		for i, entry := range t.nodes[parent].entries {
			if entry.child == n {
				parentEntry = i
				break
			}
		}
		t.nodes[parent].entries[parentEntry].box = t.calculateBound(n)

		// AT4
		pp := -1
		if nn != -1 {
			newEntry := entry{
				box:   t.calculateBound(nn),
				child: nn,
			}
			t.nodes[parent].entries = append(t.nodes[parent].entries, newEntry)
			t.nodes[nn].parent = parent
			if len(t.nodes[parent].entries) > maxChildren {
				pp = t.splitNode(parent)
			}
		}

		n, nn = parent, pp
	}
}

// splitNode splits node with index n into two nodes. The first node replaces
// n, and the second node is newly created. The return value is the index of
// the new node.
func (t *RTree) splitNode(n int) int {
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
		maxSplit = uint64((1 << (len(t.nodes[n].entries) - 1)) - 1)
	)
	bestArea := math.Inf(+1)
	var bestSplit uint64
	for split := minSplit; split <= maxSplit; split++ {
		if bits.OnesCount64(split) < minChildren {
			continue
		}
		var boxA, boxB Box
		var hasA, hasB bool
		for i, entry := range t.nodes[n].entries {
			if split&(1<<i) == 0 {
				if hasA {
					boxA = combine(boxA, entry.box)
				} else {
					boxA = entry.box
				}
			} else {
				if hasB {
					boxB = combine(boxB, entry.box)
				} else {
					boxB = entry.box
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
	for i, entry := range t.nodes[n].entries {
		if bestSplit&(1<<i) == 0 {
			entriesA = append(entriesA, entry)
		} else {
			entriesB = append(entriesB, entry)
		}
	}

	// Use the existing node for A, and create a new node for B.
	t.nodes[n].entries = entriesA
	t.nodes = append(t.nodes, node{
		isLeaf:  t.nodes[n].isLeaf,
		entries: entriesB,
		parent:  -1,
	})
	if !t.nodes[n].isLeaf {
		for _, entry := range entriesB {
			t.nodes[entry.child].parent = len(t.nodes) - 1
		}
	}
	return len(t.nodes) - 1
}

func (t *RTree) chooseLeafNode(box Box) int {
	node := t.rootIndex

	for {
		if t.nodes[node].isLeaf {
			return node
		}
		bestDelta := enlargement(box, t.nodes[node].entries[0].box)
		bestEntry := 0
		for i, entry := range t.nodes[node].entries[1:] {
			delta := enlargement(box, entry.box)
			if delta < bestDelta {
				bestDelta = delta
				bestEntry = i
			} else if delta == bestDelta && area(entry.box) < area(t.nodes[node].entries[bestEntry].box) {
				// Area is used as a tie breaking if the enlargements are the same.
				bestEntry = i
			}
		}
		node = t.nodes[node].entries[bestEntry].child
	}
}
