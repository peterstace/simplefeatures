package geom

// edgeSet represents a set of edges in a DCEL. It makes use of assumptions
// around proper noding in conjunction with interaction points.
//
// Implementation detail: the map key is 2 XYs. The first is the start point of
// the edge, the second is the first intermediate point or the start point of
// the next edge if there are no intermediate points.
type edgeSet map[[2]XY]*halfEdgeRecord

func (s edgeSet) addOrGet(segment Sequence) (*halfEdgeRecord, bool) {
	k := s.key(segment)
	e, ok := s[k]
	if !ok {
		e = new(halfEdgeRecord)
		s[k] = e
	}
	return e, !ok
}

//func (s edgeSet) lookupEdge(e *halfEdgeRecord) (*halfEdgeRecord, bool) {
//	k := s.key(e.seq)
//	e, ok := s[k]
//	return e, ok
//}

func (edgeSet) key(seq Sequence) [2]XY {
	return [2]XY{seq.GetXY(0), seq.GetXY(1)}
}
