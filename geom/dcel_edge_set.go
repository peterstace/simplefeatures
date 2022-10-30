package geom

// edgeSet represents a set of edges in a DCEL. It makes use of assumptions
// around proper noding in conjunction with interaction points.
//
// Implementation detail: the map key is 2 XYs. The first is the start point of
// the edge, the second is the first intermediate point or the start point of
// the next edge if there are no intermediate points.
type edgeSet map[[2]XY]*halfEdgeRecord

func (s edgeSet) containsStartIntermediateEnd(start XY, intermediate []XY, end XY) bool {
	_, ok := s[s.key(start, intermediate, end)]
	return ok
}

func (s edgeSet) insertEdge(e *halfEdgeRecord) {
	k := s.key(e.origin.coords, e.intermediate, e.next.origin.coords)
	s[k] = e
}

func (s edgeSet) insertStartIntermediateEnd(start XY, intermediate []XY, end XY, e *halfEdgeRecord) {
	k := s.key(start, intermediate, end)
	s[k] = e
}

func (s edgeSet) lookupEdge(e *halfEdgeRecord) (*halfEdgeRecord, bool) {
	k := s.key(e.origin.coords, e.intermediate, e.next.origin.coords)
	e, ok := s[k]
	return e, ok
}

func (edgeSet) key(start XY, intermediate []XY, end XY) [2]XY {
	if len(intermediate) == 0 {
		return [2]XY{start, end}
	}
	return [2]XY{start, intermediate[0]}
}
