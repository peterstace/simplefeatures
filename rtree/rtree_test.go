package rtree

import (
	"fmt"
	"hash/crc64"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func testBulkLoad(rnd *rand.Rand, pop int, maxStart, maxWidth float64) (*RTree, []Box) {
	boxes := make([]Box, pop)
	for i := range boxes {
		boxes[i] = randomBox(rnd, maxStart, maxWidth)
	}
	inserts := make([]BulkItem, len(boxes))
	for i := range inserts {
		inserts[i].Box = boxes[i]
		inserts[i].RecordID = i
	}
	return BulkLoad(inserts), boxes
}

func testPopulations(manditory, max int, mult float64) []int {
	var pops []int
	for i := 0; i < manditory; i++ {
		pops = append(pops, i)
	}
	for pop := float64(manditory); pop < float64(max); pop *= mult {
		pops = append(pops, int(pop))
	}
	return pops
}

func TestRandom(t *testing.T) {
	for _, population := range testPopulations(66, 1000, 1.2) {
		t.Run(fmt.Sprintf("bulk_%d", population), func(t *testing.T) {
			rnd := rand.New(rand.NewSource(0))
			rt, boxes := testBulkLoad(rnd, population, 0.9, 0.1)
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

			rt := new(RTree)
			for i, box := range boxes {
				rt.Insert(box, i)
				checkInvariants(t, rt, boxes[:i+1])
			}

			checkSearch(t, rt, boxes, rnd)
		})
	}
}

func TestDelete(t *testing.T) {
	for _, population := range testPopulations(66, 1000, 1.5) {
		t.Run(fmt.Sprintf("pop=%d", population), func(t *testing.T) {
			rnd := rand.New(rand.NewSource(0))
			rt, boxes := testBulkLoad(rnd, population, 0.9, 0.1)
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

func TestBulkLoadGolden(t *testing.T) {
	for _, tt := range []struct {
		pop  int
		want uint64
	}{
		// Test data is 'golden'. We don't really care what the values are,
		// just that they remain stable over time. If they unexpectedly change,
		// then that's an indication that the structure of the bulkloaded tree
		// has changed (which may or may not be ok depending on the reason for
		// the change).
		{pop: 1, want: 4796333603149578240},
		{pop: 3, want: 4729504678986907648},
		{pop: 4, want: 4616912695452668560},
		{pop: 5, want: 4329441588449081019},
		{pop: 6, want: 2189616554920753830},
		{pop: 7, want: 18175851834761875554},
		{pop: 10, want: 3134134291419311046},
		{pop: 15, want: 4266143990412453129},
		{pop: 16, want: 3347339997226441897},
		{pop: 17, want: 492585592469164258},
		{pop: 100, want: 14966132007611414076},
		{pop: 1000, want: 6535743965441116214},
		{pop: 10_000, want: 7407227297016950893},
		{pop: 100_000, want: 14387712569501511190},
	} {
		t.Run(fmt.Sprintf("n=%d", tt.pop), func(t *testing.T) {
			rnd := rand.New(rand.NewSource(0))
			rt, _ := testBulkLoad(rnd, tt.pop, 0.9, 0.1)
			got := checksum(rt.root)
			if got != tt.want {
				t.Errorf("got=%d want=%d", got, tt.want)
			}
		})
	}
}

func checksum(n *node) uint64 {
	var entries []string
	for i := 0; i < n.numEntries; i++ {
		var entry string
		if n.isLeaf {
			entry = strconv.Itoa(n.entries[i].recordID)
		} else {
			entry = strconv.FormatUint(checksum(n.entries[i].child), 10)
		}
		entries = append(entries, entry)
	}
	sort.Strings(entries)
	return crc64.Checksum([]byte(strings.Join(entries, ",")), crc64.MakeTable(crc64.ISO))
}

func checkSearch(t *testing.T, rt *RTree, boxes []Box, rnd *rand.Rand) {
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

func checkInvariants(t *testing.T, rt *RTree, boxes []Box) {
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

	leafLevel := -1
	var check func(n *node, level int)
	check = func(current *node, level int) {
		if current.isLeaf {
			if leafLevel == -1 {
				leafLevel = level
			} else if leafLevel != level {
				t.Fatalf("inconsistent leaf level: %d vs %d", leafLevel, level)
			}

			for i := 0; i < current.numEntries; i++ {
				e := current.entries[i]
				if e.child != nil {
					t.Fatalf("leaf node has child (entry %d)", i)
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
					t.Fatalf("entry box doesn't match smallest box enclosing children")
				}
				check(e.child, level+1)
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
		check(rt.root, 0)
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
