package geom

// disjointSet implements a standard disjoint-set data structure (also known as
// a union-find structure or a merge-find structure). It stores a collection of
// disjoint (non-overlapping) sets. The set elements are integers (which may be
// mapped externally to more complicated types).
type disjointSet struct {
	parent []int // self reference indicates root
	rank   []int
}

// newDistointSet creates a new disjoint set containing n sets, each with a
// single item. The items are 0 (inclusive) through to n (exclusive).
func newDisjointSet(n int) disjointSet {
	set := disjointSet{make([]int, n), make([]int, n)}
	for i := range set.parent {
		set.parent[i] = i
	}
	return set
}

// find searches for the representative for the set containing x. All elements
// of the set containing x will have the same representative. To find out if
// two elements are in the same set, find can be used on each element and the
// representatives compared.
func (s disjointSet) find(x int) int {
	root := x
	for s.parent[root] != root {
		root = s.parent[root]
	}
	for s.parent[x] != root {
		parent := s.parent[x]
		s.parent[x] = root
		x = parent
	}
	return root
}

// union merges the set containing x with the set containing y.
func (s disjointSet) union(x, y int) {
	x = s.find(x)
	y = s.find(y)
	if x == y {
		return
	}

	if s.rank[x] < s.rank[y] {
		x, y = y, x
	}

	s.parent[y] = x
	if s.rank[x] == s.rank[y] {
		s.rank[x]++
	}
}
