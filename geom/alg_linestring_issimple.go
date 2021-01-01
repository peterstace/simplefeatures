package geom

import (
	"sort"

	"github.com/peterstace/simplefeatures/internal/redblack"
)

// IsSimple returns true if this geometry contains no anomalous geometry
// points, such as self intersection or self tangency. LineStrings are simple
// if and only if the curve defined by the LineString doesn't pass through the
// same point twice (with the exception of the two endpoints being coincident).
func (s LineString) IsSimple() bool {
	type event struct {
		XY
		idx   int
		start bool
	}
	var events []event
	lines := s.asLines()
	for i, ln := range lines {
		// Flip start and endpoints of segment so that the endpoint with the
		// smaller X is the start endpoint. For vertical lines, it is as though
		// they learn slightly to the right.
		if ln.a.X > ln.b.X || (ln.a.X == ln.b.X && ln.a.Y > ln.b.Y) {
			ln.a, ln.b = ln.b, ln.a
			lines[i] = ln
		}
		events = append(events,
			event{ln.a, i, true},
			event{ln.b, i, false},
		)
	}

	sort.Slice(events, func(i, j int) bool {
		ei := events[i]
		ej := events[j]
		if ei.X == ej.X {
			if ei.start != ej.start {
				// Start events come before end events.
				return ei.start
			}
			return ei.Y < ej.Y
		}
		return ei.X < ej.X
	})

	var tree redblack.Tree
	var currentX float64
	cmp := func(key1, key2 int) int {
		if key1 == key2 {
			return 0
		}
		// TODO: what if a line segment is vertical?
		ln1 := lines[key1]
		ln2 := lines[key2]

		if ln1.a == ln2.a {
			o := orientation(ln1.b, ln1.a, ln2.b)
			if o == rightTurn {
				return -1
			}
			return +1
		}
		if ln1.b == ln2.b {
			o := orientation(ln1.a, ln1.b, ln2.a)
			if o == leftTurn {
				return -1
			}
			return +1
		}
		if ln1.b == ln2.a {
			return -1
		}
		if ln1.a == ln2.b {
			return +1
		}

		y1 := yIntercept(currentX, ln1)
		y2 := yIntercept(currentX, ln2)
		if y1-y2 < 0 {
			return -1
		}
		return +1
	}

	adjKeys := func(key int) (int, bool, int, bool) {
		iter := tree.Begin()
		if !iter.Seek(key, cmp) {
			panic("could not seek key")
		}
		prevOK := iter.Prev()
		var prev int
		if prevOK {
			prev = iter.Key()
		}
		if !iter.Next() {
			panic("iterator not valid after advancing it")
		}
		nextOK := iter.Next()
		var next int
		if nextOK {
			next = iter.Key()
		}
		return prev, prevOK, next, nextOK
	}

	isClosed := s.IsClosed()
	intersect := func(i, j int) bool {
		inter := lines[i].intersectLine(lines[j])
		if inter.empty {
			// No intersections.
			return false
		}
		if (j+1 == i || j-1 == i) && inter.ptA == inter.ptB {
			// There is an intersection, but it's the joining XY between 2
			// adjacent segments.
			return false
		}
		if n := len(lines); isClosed && i+j == n-1 && j*i == 0 && inter.ptA == inter.ptB {
			// There is an intersection, but it's at the joining XY at the
			// start/end of a loop.
			return false
		}
		return true // Some other sort of intersection.
	}

	for _, e := range events {
		if e.start {
			currentX = e.X
			// Start of a line segment.
			i := e.idx
			tree.Insert(i, cmp)
			below, belowOK, above, aboveOK := adjKeys(i)
			if aboveOK && intersect(i, above) || belowOK && intersect(i, below) {
				return false
			}
		} else {
			// End of a line segment.
			currentX = e.XY.X
			i := e.idx
			above, aboveOK, below, belowOK := adjKeys(i)
			tree.Delete(i, cmp)
			if aboveOK && belowOK && intersect(above, below) {
				return false
			}
		}
	}
	return true
}

func keys(tree redblack.Tree) []int {
	var keys []int
	iter := tree.Begin()
	for iter.Next() {
		keys = append(keys, iter.Key())
	}
	return keys
}

func yIntercept(x float64, ln line) float64 {
	if ln.a.X == ln.b.X {
		// Pretend that vertical lines are leaning ever so slightly right.
		return ln.a.Y
	}
	numer := (x - ln.a.X) * (ln.b.Y - ln.a.Y)
	denom := ln.b.X - ln.a.X
	return ln.a.Y + numer/denom
}
