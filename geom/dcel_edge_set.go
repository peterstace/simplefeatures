package geom

// edgeSet represents a set of edges in a DCEL. It makes use of assumptions
// around proper noding in conjunction with interaction points.
//
// Implementation detail: the map key is 3 XYs. The first is the start point of
// the edge, the second is the first intermediate point (or a repeat of the
// first XY if there are no intermediate points), and the third is the start
// point of the next edge.
type edgeSet map[[3]XY]struct{}

func (s edgeSet) containsEdge(e *halfEdgeRecord) bool {
	k := s.key(e.origin.coords, e.intermediate, e.next.origin.coords)
	_, ok := s[k]
	return ok
}

func (s edgeSet) containsLine(ln line) bool {
	_, ok := s[s.key(ln.a, nil, ln.b)]
	return ok
}

func (s edgeSet) containsStartIntermediateEnd(start XY, intermediate []XY, end XY) bool {
	_, ok := s[s.key(start, intermediate, end)]
	return ok
}

func (s edgeSet) insertEdge(e *halfEdgeRecord) {
	k := s.key(e.origin.coords, e.intermediate, e.next.origin.coords)
	s[k] = struct{}{}
}

func (s edgeSet) insertLine(ln line) {
	k := s.key(ln.a, nil, ln.b)
	s[k] = struct{}{}
}

func (s edgeSet) insertStartIntermediateEnd(start XY, intermediate []XY, end XY) {
	k := s.key(start, intermediate, end)
	s[k] = struct{}{}
}

func (edgeSet) key(start XY, intermediate []XY, end XY) [3]XY {
	key := [3]XY{start, start, end}
	if len(intermediate) > 0 {
		key[1] = intermediate[0]
	}
	return key
}
