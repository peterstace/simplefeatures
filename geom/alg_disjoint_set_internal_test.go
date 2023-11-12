package geom

import (
	"strconv"
	"testing"
)

func TestDisjointSet(t *testing.T) {
	type intPair struct{ a, b int }
	for idx, tc := range []struct {
		size   int
		merges []intPair
	}{
		{
			size:   0,
			merges: nil,
		},
		{
			size:   1,
			merges: []intPair{{0, 0}},
		},
		{
			size:   2,
			merges: []intPair{{0, 1}},
		},
		{
			size:   2,
			merges: []intPair{{1, 0}},
		},
		{
			size:   2,
			merges: []intPair{{1, 1}},
		},
		{
			size: 3,
			merges: []intPair{
				{1, 0},
				{0, 2},
			},
		},
		{
			size: 4,
			merges: []intPair{
				{2, 1},
				{0, 3},
				{1, 0},
			},
		},
		{
			size: 5,
			merges: []intPair{
				{2, 1},
				{0, 3},
				{1, 0},
				{3, 4},
			},
		},
	} {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			t.Logf("num elements: %d", tc.size)
			simpleSet := newSimpleDisjointSet(tc.size)
			fastSet := newDisjointSet(tc.size)
			for _, m := range tc.merges {
				t.Logf("merging %d and %d", m.a, m.b)
				fastSet.union(m.a, m.b)
				simpleSet.union(m.a, m.b)
				for i := 0; i < tc.size; i++ {
					for j := 0; j < tc.size; j++ {
						gotFast := fastSet.find(i) == fastSet.find(j)
						gotSimple := simpleSet.find(i) == simpleSet.find(j)
						if gotFast != gotSimple {
							t.Errorf("mismatch between %d and %d in same set: fast=%v simple=%v", i, j, gotFast, gotSimple)
						}
					}
				}
			}
		})
	}
}

// simpleDisjointSet is a _simple_ but _inefficient_ implementation
// of a disjoint set data structure. It's used as a reference
// implementation for testing.
type simpleDisjointSet struct {
	// For the simple implementation, we store the set identifier
	// for each element directly. This results in very simple
	// operations, but linear union time complexity.
	set []int
}

func newSimpleDisjointSet(size int) simpleDisjointSet {
	items := make([]int, size)
	for i := range items {
		items[i] = i
	}
	return simpleDisjointSet{items}
}

func (s simpleDisjointSet) find(x int) int {
	return s.set[x]
}

func (s simpleDisjointSet) union(x, y int) {
	setX := s.find(x)
	setY := s.find(y)
	for i := range s.set {
		if s.set[i] == setY {
			s.set[i] = setX
		}
	}
}
