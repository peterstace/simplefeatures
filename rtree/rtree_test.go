package rtree

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestRandom(t *testing.T) {
	for pop := 0.0; pop < 1000; pop = (pop + 1) * 1.2 {
		population := int(pop)

		t.Run(fmt.Sprintf("bulk_%d", population), func(t *testing.T) {
			rnd := rand.New(rand.NewSource(0))
			boxes := make([]Box, population)
			for i := range boxes {
				boxes[i] = randomBox(rnd, 0.9, 0.1)
			}

			inserts := make([]BulkItem, len(boxes))
			for i := range inserts {
				inserts[i].Box = boxes[i]
				inserts[i].RecordID = i
			}
			rt := BulkLoad(inserts)

			checkInvariants(t, rt, boxes)
			checkSearch(t, rt, boxes, rnd)
		})
		name := fmt.Sprintf("insert_%d", population)
		t.Run(name, func(t *testing.T) {
			rnd := rand.New(rand.NewSource(0))
			boxes := make([]Box, population)
			for i := range boxes {
				boxes[i] = randomBox(rnd, 0.9, 0.1)
			}

			var rt RTree
			for i, box := range boxes {
				rt.Insert(box, i)
				checkInvariants(t, rt, boxes[:i+1])
			}

			checkSearch(t, rt, boxes, rnd)
		})
	}
}

func TestDelete(t *testing.T) {
	for pop := 0.0; pop < 1000; pop = (pop + 1) * 1.5 {
		population := int(pop)

		t.Run(fmt.Sprintf("pop=%d", population), func(t *testing.T) {
			rnd := rand.New(rand.NewSource(0))
			boxes := make([]Box, population)
			for i := range boxes {
				boxes[i] = randomBox(rnd, 0.9, 0.1)
			}
			inserts := make([]BulkItem, len(boxes))
			for i := range inserts {
				inserts[i].Box = boxes[i]
				inserts[i].RecordID = i
			}
			rt := BulkLoad(inserts)
			checkInvariants(t, rt, boxes)

			for i := len(boxes) - 1; i >= 0; i-- {
				t.Logf("deleting recordID %d", i)
				rt.Delete(boxes[i], i)
				checkInvariants(t, rt, boxes[:i])
				checkSearch(t, rt, boxes[:i], rnd)
			}
		})
	}
}

func checkSearch(t *testing.T, rt RTree, boxes []Box, rnd *rand.Rand) {
	for i := 0; i < 10; i++ {
		searchBB := randomBox(rnd, 0.5, 0.5)
		var got []int
		rt.RangeSearch(searchBB, func(idx int) error {
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

func checkInvariants(t *testing.T, rt RTree, boxes []Box) {
	var recurse func(*node, string)
	recurse = func(current *node, indent string) {
		t.Logf("%sNode addr=%p leaf=%t numEntries=%d", indent, current, current.isLeaf, current.numEntries)
		indent += "\t"
		if current.isLeaf {
			for i := 0; i < current.numEntries; i++ {
				e := current.entries[i]
				t.Logf("%sEntry[%d] recordID=%d box=%v", indent, i, e.recordID, e.box)
			}
		} else {
			for i := 0; i < current.numEntries; i++ {
				e := &current.entries[i]
				t.Logf("%sEntry[%d] box=%v", indent, i, e.box)
				recurse(e.child, indent+"\t")
			}
		}
	}
	t.Log("---")
	if rt.root != nil {
		recurse(rt.root, "")
	} else {
		t.Log("Root is nil")
	}
	t.Log("---")

	unfound := make(map[int]struct{})
	for i := range boxes {
		unfound[i] = struct{}{}
	}

	var check func(*node)
	check = func(current *node) {
		if current.isLeaf {
			for i := 0; i < current.numEntries; i++ {
				e := current.entries[i]
				if e.child != nil {
					t.Fatal("leaf node has child")
				}
				if _, ok := unfound[e.recordID]; !ok {
					t.Fatal("record ID found in tree but wasn't in unfound map")
				}
				delete(unfound, e.recordID)
			}
		} else {
			for i := 0; i < current.numEntries; i++ {
				e := &current.entries[i]
				if e.recordID != 0 {
					t.Fatal("non-leaf has recordID")
				}
				if e.child.parent != current {
					t.Fatalf("node %p has wrong parent", e.child)
				}
				box := e.child.entries[0].box
				for j := 1; j < e.child.numEntries; j++ {
					box = combine(box, e.child.entries[j].box)
				}
				if box != e.box {
					t.Fatalf("entry box doesn't match smallest box enclosing child")
				}
				check(e.child)
			}
		}
		for i := current.numEntries; i < len(current.entries); i++ {
			e := current.entries[i]
			if e.box != (Box{}) || e.child != nil || e.recordID != 0 {
				t.Fatal("entry past numEntries is not the zero value")
			}
		}
		if current.numEntries > maxChildren || (current != rt.root && current.numEntries < minChildren) {
			t.Fatalf("%p: unexpected number of entries", current)
		}
	}
	if rt.root != nil {
		check(rt.root)
		if rt.root.parent != nil {
			t.Fatalf("root parent should be nil, but is %p", rt.root.parent)
		}
	}

	if len(unfound) != 0 {
		t.Fatalf("there were still unfound record IDs after traversing tree")
	}

	gotExtent, hasExtent := rt.Extent()
	if len(boxes) == 0 {
		if hasExtent {
			t.Fatal("expected not to get an extent, but got one")
		}
	} else {
		if !hasExtent {
			t.Fatalf("expected to get an extent, but didn't")
		}
		wantExtent := boxes[0]
		for _, b := range boxes[1:] {
			wantExtent = combine(wantExtent, b)
		}
		if wantExtent != gotExtent {
			t.Fatalf("unexpected bounding box: want=%v got=%v", wantExtent, gotExtent)
		}
	}
}
