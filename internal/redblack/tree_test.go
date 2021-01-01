package redblack_test

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/peterstace/simplefeatures/internal/redblack"
)

func generateKeys(population int) []int {
	keys := make([]int, population)
	for i := 0; i < population; i++ {
		keys[i] = 1000 + i
	}
	rng := rand.New(rand.NewSource(0))
	rng.Shuffle(population, func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})
	return keys
}

type MapTracker struct {
	t        *testing.T
	cmp      redblack.Compare
	tree     *redblack.Tree
	set      map[int]bool
	universe []int
}

func NewMapTracker(t *testing.T, cmp redblack.Compare, universe []int) *MapTracker {
	return &MapTracker{
		t:        t,
		cmp:      cmp,
		tree:     new(redblack.Tree),
		set:      make(map[int]bool),
		universe: universe,
	}
}

func (m *MapTracker) Insert(key int) {
	m.tree.Insert(key, m.cmp)
	m.set[key] = true
}

func (m *MapTracker) Delete(key int) {
	m.tree.Delete(key, m.cmp)
	delete(m.set, key)
}

func (m *MapTracker) CheckConsistency() {
	m.checkFwdIteration()
	m.checkRevIteration()
	m.checkContains()
	m.checkSeek()
}

func (m *MapTracker) checkFwdIteration() {
	var keyList []int
	keySet := make(map[int]bool)
	iter := m.tree.Begin()
	for iter.Next() {
		key := iter.Key()
		keyList = append(keyList, key)
		keySet[key] = true
	}

	if len(keyList) != len(keySet) {
		m.t.Fatalf("keyList length doesn't match keySet length: %v vs %v", len(keyList), len(keySet))
	}
	if !sort.SliceIsSorted(keyList, func(i, j int) bool {
		return m.cmp(keyList[i], keyList[j]) < 0
	}) {
		m.t.Fatalf("keyList is not sorted: %v", keyList)
	}
}

func (m *MapTracker) checkRevIteration() {
	var keyList []int
	keySet := make(map[int]bool)
	iter := m.tree.End()
	for iter.Prev() {
		key := iter.Key()
		keyList = append(keyList, key)
		keySet[key] = true
	}

	if len(keyList) != len(keySet) {
		m.t.Fatalf("keyList length doesn't match keySet length: %v vs %v", len(keyList), len(keySet))
	}
	if !sort.SliceIsSorted(keyList, func(i, j int) bool {
		return m.cmp(keyList[i], keyList[j]) > 0
	}) {
		m.t.Fatalf("keyList is not sorted: %v", keyList)
	}
}

func (m *MapTracker) checkContains() {
	for _, k := range m.universe {
		want := m.set[k]
		got := m.tree.Contains(k, m.cmp)
		if want != got {
			m.t.Fatalf("wrong contains result for key %d: want=%v got=%v", k, want, got)
		}
	}
}

func (m *MapTracker) checkSeek() {
	iter := m.tree.Begin()
	for _, k := range m.universe {
		wantOK := m.set[k]
		gotOK := iter.Seek(k, m.cmp)
		if wantOK != gotOK {
			m.t.Fatalf("wrong ok result for seek %d: want=%v got=%v", k, wantOK, gotOK)
		}
		if gotOK {
			if iter.Key() != k {
				m.t.Fatalf("wrong iterator key: %v vs %v", k, iter.Key())
			}

			set := make(map[int]bool)
			var count int
			for iter.Next() {
				count++
				set[iter.Key()] = true
			}
			for i := 0; i <= count; i++ {
				iter.Prev()
			}
			if iter.Key() != k {
				m.t.Fatalf("unexpected iter key value: %v vs %v", k, iter.Key())
			}
			for iter.Prev() {
				set[iter.Key()] = true
			}
			if !reflect.DeepEqual(set, m.set) {
			}
		}

	}
}

func TestSimulation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		cmp    func(key1, key2 int) int
		maxPop int
	}{
		{"fwd", func(key1, key2 int) int { return key1 - key2 }, 100},
		{"rev", func(key1, key2 int) int { return key2 - key1 }, 100},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for pop := 0; pop < tc.maxPop; pop++ {
				t.Run(fmt.Sprintf("pop=%d", pop), func(t *testing.T) {
					keys := generateKeys(pop)
					tracker := NewMapTracker(t, tc.cmp, keys)
					for _, k := range keys {
						tracker.Insert(k)
						tracker.CheckConsistency()
					}
					for _, k := range keys {
						tracker.Delete(k)
						tracker.CheckConsistency()
					}
				})
			}
		})
	}
}

func TestDeleteOnEmptyTree(t *testing.T) {
	tr := new(redblack.Tree)
	cmp := func(i, j int) int { return i - j }
	tr.Delete(0, cmp)
}

func TestDeleteOnTreeWithOneEntry(t *testing.T) {
	tr := new(redblack.Tree)
	less := func(i, j int) int { return i - j }
	tr.Insert(0, less)
	tr.Delete(0, less)
}
