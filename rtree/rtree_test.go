package rtree

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestRandom(t *testing.T) {
	for population := 0; population < 200; population++ {
		t.Run(fmt.Sprintf("bulk_%d", population), func(t *testing.T) {
			rnd := rand.New(rand.NewSource(0))
			boxes := make([]Box, population)
			for i := range boxes {
				boxes[i] = randomBox(rnd, 0.9, 0.1)
			}

			inserts := make([]BulkItem, len(boxes))
			for i := range inserts {
				inserts[i].Box = boxes[i]
				inserts[i].DataIndex = i
			}
			rt := BulkLoad(inserts)

			checkInvariants(t, rt)
			checkSearch(t, rt, boxes, rnd)
		})
		name := fmt.Sprintf("pop_%d", population)
		t.Run(name, func(t *testing.T) {
			rnd := rand.New(rand.NewSource(0))
			boxes := make([]Box, population)
			for i := range boxes {
				boxes[i] = randomBox(rnd, 0.9, 0.1)
			}

			var rt RTree
			for i, box := range boxes {
				rt.Insert(box, i)
				checkInvariants(t, rt)
			}

			checkSearch(t, rt, boxes, rnd)
		})
	}
}

func checkSearch(t *testing.T, rt RTree, boxes []Box, rnd *rand.Rand) {
	for i := 0; i < 10; i++ {
		searchBB := randomBox(rnd, 0.5, 0.5)
		var got []int
		rt.Search(searchBB, func(idx int) error {
			got = append(got, idx)
			return nil
		})

		var want []int
		for i, box := range boxes {
			if overlap(box, searchBB) {
				want = append(want, i)
			}
		}

		sort.Ints(want)
		sort.Ints(got)

		if !reflect.DeepEqual(want, got) {
			t.Logf("search box: %v", searchBB)
			t.Errorf("search failed, got: %v want: %v", got, want)
		}
	}
}

func randomBox(rnd *rand.Rand, maxStart, maxWidth float64) Box {
	box := Box{
		MinX: rnd.Float64() * maxStart,
		MinY: rnd.Float64() * maxStart,
	}
	box.MaxX = box.MinX + rnd.Float64()*maxWidth
	box.MaxY = box.MinY + rnd.Float64()*maxWidth

	box.MinX = float64(int(box.MinX*100)) / 100
	box.MinY = float64(int(box.MinY*100)) / 100
	box.MaxX = float64(int(box.MaxX*100)) / 100
	box.MaxY = float64(int(box.MaxY*100)) / 100
	return box
}

func checkInvariants(t *testing.T, rt RTree) {
	t.Logf("")
	t.Logf("RTree description:")
	t.Logf("node_count=%v, root=%d", len(rt.nodes), rt.rootIndex)
	for i, n := range rt.nodes {
		t.Logf("%d: leaf=%t numentries=%d parent=%d", i, n.isLeaf, len(n.entries), n.parent)
		for j, e := range n.entries {
			t.Logf("\t%d: index=%d box=%v", j, e.index, e.box)
		}
	}

	// Each node has the correct parent set.
	for i, node := range rt.nodes {
		if i == rt.rootIndex {
			if node.parent != -1 {
				t.Fatalf("expected root to have parent -1, but has %d", node.parent)
			}
			continue
		}
		if node.parent == -1 {
			t.Fatalf("expected parent for non-root not to be -1, but was -1")
		}

		var matchingChildren int
		for _, entry := range rt.nodes[node.parent].entries {
			if entry.index == i {
				matchingChildren++
			}
		}
		if matchingChildren != 1 {
			t.Fatalf("expected parent to have 1 matching child, but has %d", matchingChildren)
		}
	}

	// For each non-leaf node, its entries should have the smallest bounding boxes that cover its children.
	for i, parentNode := range rt.nodes {
		if parentNode.isLeaf {
			continue
		}
		for j, parentEntry := range parentNode.entries {
			childNode := rt.nodes[parentEntry.index]
			union := childNode.entries[0].box
			for _, childEntry := range childNode.entries[1:] {
				union = combine(childEntry.box, union)
			}
			if union != parentEntry.box {
				t.Fatalf("expected parent to have smallest box that covers its children (node=%d, entry=%d)", i, j)
			}
		}
	}

	// Each leaf should be reached exactly once from the root. This implies
	// that the tree has no loops, and there are no orphan leafs. Also checks
	// that each non-leaf is visited at least once (i.e. no orphan non-leaves).
	leafCount := make(map[int]int)
	visited := make(map[int]bool)
	var recurse func(int)
	recurse = func(n int) {
		visited[n] = true
		node := &rt.nodes[n]
		if node.isLeaf {
			leafCount[n]++
			return
		}
		for _, entry := range node.entries {
			recurse(entry.index)
		}
	}
	recurse(rt.rootIndex)
	for leaf, count := range leafCount {
		if count != 1 {
			t.Fatalf("leaf %d visited %d times", leaf, count)
		}
	}
	for i := range rt.nodes {
		if !visited[i] {
			t.Fatalf("node %d was not visited", i)
		}
	}
}
