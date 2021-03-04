package rtree

// Delete removes a single record with a matching recordID from the RTree. The
// box specifies where to search in the RTree for the record (the search box
// must intersect with the box of the record for it to be found and deleted).
// The returned bool indicates whether or not the record could be found and
// thus removed from the RTree (true indicates success).
func (t *RTree) Delete(box Box, recordID int) bool {
	if t.root == nil {
		return false
	}

	// D1 [Find node containing record]
	var foundNode *node
	var foundEntryIndex int
	var recurse func(*node)
	recurse = func(n *node) {
		for i := 0; i < n.numEntries; i++ {
			entry := n.entries[i]
			if !overlap(entry.box, box) {
				continue
			}
			if !n.isLeaf {
				recurse(entry.child)
				if foundNode != nil {
					break
				}
			} else {
				if entry.recordID == recordID {
					foundNode = n
					foundEntryIndex = i
					break
				}
			}
		}
	}
	recurse(t.root)
	if foundNode == nil {
		return false
	}

	// D2 [Delete record]
	originalCount := t.count
	deleteEntry(foundNode, foundEntryIndex)

	// D3 [Propagate changes]
	t.condenseTree(foundNode)

	// D4 [Shorten tree]
	if !t.root.isLeaf && t.root.numEntries == 1 {
		t.root = t.root.entries[0].child
		t.root.parent = nil
	}

	t.count = originalCount - 1
	return true
}

func deleteEntry(n *node, entryIndex int) {
	n.entries[entryIndex] = n.entries[n.numEntries-1]
	n.numEntries--
	n.entries[n.numEntries] = entry{}
}

func (t *RTree) condenseTree(leaf *node) {
	// CT1 [Initialise]
	var eliminated []*node
	current := leaf

	for current != t.root {
		// CT2 [Find Parent Entry]
		parent := current.parent
		entryIdx := -1
		for i := 0; i < parent.numEntries; i++ {
			if parent.entries[i].child == current {
				entryIdx = i
				break
			}
		}

		// CT3 [Eliminate Under-Full Node]
		if current.numEntries < minChildren {
			eliminated = append(eliminated, current)
			deleteEntry(parent, entryIdx)
		} else {
			// CT4 [Adjust Covering Rectangle]
			newBox := current.entries[0].box
			for i := 1; i < current.numEntries; i++ {
				newBox = combine(newBox, current.entries[i].box)
			}
			parent.entries[entryIdx].box = newBox
		}

		// CT5 [Move Up One Level In Tree]
		current = parent
	}

	// CT6 [Reinsert orphaned entries]
	for _, node := range eliminated {
		if node.isLeaf {
			for i := 0; i < node.numEntries; i++ {
				e := node.entries[i]
				t.Insert(e.box, e.recordID)
			}
		} else {
			for i := 0; i < node.numEntries; i++ {
				t.reInsertNode(node.entries[i].child)
			}
		}
	}
}

// reInsertNode reinserts the subtree rooted at a node that was previously
// deleted from the tree.
func (t *RTree) reInsertNode(node *node) {
	box := calculateBound(node)
	treeDepth := t.root.depth()
	nodeDepth := node.depth()
	insNode := t.chooseBestNode(box, treeDepth-nodeDepth-1)

	insNode.appendChild(box, node)
	t.adjustBoxesUpwards(node, box)

	if insNode.numEntries <= maxChildren {
		return
	}

	newNode := t.splitNode(insNode)
	root1, root2 := t.adjustTree(insNode, newNode)
	if root2 != nil {
		t.joinRoots(root1, root2)
	}
}
