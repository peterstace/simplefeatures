package simplefeatures

// graph is an adjacency list representing an undirected simple graph.
type graph map[int]map[int]struct{}

func newGraph() graph {
	return make(map[int]map[int]struct{})
}

func (g graph) addEdge(u, v int) {
	if g[u] == nil {
		g[u] = make(map[int]struct{})
	}
	if g[v] == nil {
		g[v] = make(map[int]struct{})
	}
	g[u][v] = struct{}{}
	g[v][u] = struct{}{}
}

func (g graph) hasCycle() bool {
	unvisited := make(map[int]struct{})
	for v := range g {
		unvisited[v] = struct{}{}
	}
	for v := range unvisited {
		visited := make(map[int]struct{})
		if g.dfsHasCycle(-1, v, visited, unvisited) {
			return true
		}
	}
	return false
}

func (g graph) dfsHasCycle(parent, v int, visited, unvisited map[int]struct{}) bool {
	visited[v] = struct{}{}
	delete(unvisited, v)
	for neighbour := range g[v] {
		if neighbour == parent {
			continue
		}
		if _, have := visited[neighbour]; have {
			return true
		}
		if g.dfsHasCycle(v, neighbour, visited, unvisited) {
			return true
		}
	}
	delete(visited, v)
	return false
}
