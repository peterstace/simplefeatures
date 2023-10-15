package geom

import (
	"strconv"
	"testing"
)

func TestGraphHasCycle(t *testing.T) {
	type edge struct {
		a, b int
	}
	for i, tt := range []struct {
		edges    []edge
		hasCycle bool
	}{
		// 0 edges
		{nil, false},

		// 1 edge
		{[]edge{{1, 2}}, false},

		// 2 edges
		{[]edge{{1, 2}, {2, 3}}, false},
		{[]edge{{1, 2}, {3, 4}}, false},

		// 3 edges
		{[]edge{{1, 2}, {2, 3}, {3, 4}}, false},
		{[]edge{{1, 2}, {2, 3}, {4, 5}}, false},
		{[]edge{{1, 2}, {3, 4}, {5, 6}}, false},
		{[]edge{{1, 2}, {2, 3}, {3, 1}}, true},

		// 4 edeges
		{[]edge{{1, 2}, {2, 3}, {3, 4}, {4, 5}}, false},
		{[]edge{{1, 2}, {2, 3}, {3, 4}, {4, 2}}, true},
		{[]edge{{1, 2}, {2, 3}, {3, 4}, {4, 1}}, true},
		{[]edge{{1, 2}, {2, 3}, {3, 1}, {4, 5}}, true},
		{[]edge{{1, 2}, {2, 3}, {4, 5}, {6, 7}}, false},
		{[]edge{{1, 2}, {2, 3}, {4, 5}, {5, 6}}, false},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			graph := newGraph()
			for _, e := range tt.edges {
				graph.addEdge(e.a, e.b)
			}
			got := graph.hasCycle()
			if got != tt.hasCycle {
				t.Log(graph)
				t.Errorf("got=%v want=%v", got, tt.hasCycle)
			}
		})
	}
}
