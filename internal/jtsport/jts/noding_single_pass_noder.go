package jts

// Noding_SinglePassNoder is a base class for Noders which make a single pass
// to find intersections. This allows using a custom SegmentIntersector (which
// for instance may simply identify intersections, rather than insert them).
type Noding_SinglePassNoder struct {
	segInt Noding_SegmentIntersector
}

// Noding_NewSinglePassNoder creates a new SinglePassNoder with no segment
// intersector.
func Noding_NewSinglePassNoder() *Noding_SinglePassNoder {
	return &Noding_SinglePassNoder{}
}

// Noding_NewSinglePassNoderWithIntersector creates a new SinglePassNoder with
// the given segment intersector.
func Noding_NewSinglePassNoderWithIntersector(segInt Noding_SegmentIntersector) *Noding_SinglePassNoder {
	return &Noding_SinglePassNoder{segInt: segInt}
}

// SetSegmentIntersector sets the SegmentIntersector to use with this noder. A
// SegmentIntersector will normally add intersection nodes to the input segment
// strings, but it may not - it may simply record the presence of
// intersections. However, some Noders may require that intersections be added.
func (n *Noding_SinglePassNoder) SetSegmentIntersector(segInt Noding_SegmentIntersector) {
	n.segInt = segInt
}
