package geom

// edgeSet represents a set of edges in a DCEL. It makes use of assumptions
// around proper noding in conjunction with interaction points.
//
// Implementation detail: the map key is 2 XYs. The first is the start point of
// the edge, the second is the second point of the edge (which may or may not
// be the final point of the edge).
type edgeSet map[[2]XY]*halfEdgeRecord

func (s edgeSet) containsStartIntermediateEnd(segment Sequence) bool {
	_, ok := s[s.key(segment)]
	return ok
}

func (s edgeSet) insertEdge(e *halfEdgeRecord) {
	k := s.key(e.seq)
	s[k] = e
}

func (s edgeSet) insertStartIntermediateEnd(segment Sequence) {
	k := s.key(segment)
	s[k] = nil // TODO: this is a bit weird...
}

func (s edgeSet) lookupEdge(e *halfEdgeRecord) (*halfEdgeRecord, bool) {
	k := s.key(e.seq)
	e, ok := s[k]
	return e, ok
}

func (edgeSet) key(segment Sequence) [2]XY {
	return [2]XY{segment.GetXY(0), segment.GetXY(1)}
}
