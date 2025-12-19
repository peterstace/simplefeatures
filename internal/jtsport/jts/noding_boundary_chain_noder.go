package jts

// Noding_BoundaryChainNoder is a noder which extracts chains of boundary
// segments as SegmentStrings from a polygonal coverage. Boundary segments are
// those which are not duplicated in the input polygonal coverage. Extracting
// chains of segments minimizes the number of segment strings created, which
// produces a more efficient topological graph structure.
//
// This enables fast overlay of polygonal coverages in CoverageUnion. Using
// this noder is faster than SegmentExtractingNoder and BoundarySegmentNoder.
//
// No precision reduction is carried out. If that is required, another noder
// must be used (such as a snap-rounding noder), or the input must be
// precision-reduced beforehand.
type Noding_BoundaryChainNoder struct {
	chainList []Noding_SegmentString
}

var _ Noding_Noder = (*Noding_BoundaryChainNoder)(nil)

func (bcn *Noding_BoundaryChainNoder) IsNoding_Noder() {}

// Noding_NewBoundaryChainNoder creates a new boundary-extracting noder.
func Noding_NewBoundaryChainNoder() *Noding_BoundaryChainNoder {
	return &Noding_BoundaryChainNoder{}
}

// ComputeNodes computes the boundary chains from the input segment strings.
func (bcn *Noding_BoundaryChainNoder) ComputeNodes(segStrings []Noding_SegmentString) {
	// segSet maps normalized segment keys to their segment data. When a
	// duplicate segment is found (same coordinates), it is removed from the
	// map. Segments remaining in the map are boundary segments.
	segSet := make(map[noding_boundaryChainNoder_segKey]*noding_boundaryChainNoder_segment)
	boundaryChains := make([]*noding_boundaryChainMap, len(segStrings))
	bcn.addSegments(segStrings, segSet, boundaryChains)
	bcn.markBoundarySegments(segSet)
	bcn.chainList = bcn.extractChains(boundaryChains)
}

func (bcn *Noding_BoundaryChainNoder) addSegments(
	segStrings []Noding_SegmentString,
	segSet map[noding_boundaryChainNoder_segKey]*noding_boundaryChainNoder_segment,
	boundaryChains []*noding_boundaryChainMap,
) {
	for i, ss := range segStrings {
		chainMap := noding_newBoundaryChainMap(ss)
		boundaryChains[i] = chainMap
		bcn.addSegmentsFrom(ss, chainMap, segSet)
	}
}

func (bcn *Noding_BoundaryChainNoder) addSegmentsFrom(
	segString Noding_SegmentString,
	chainMap *noding_boundaryChainMap,
	segSet map[noding_boundaryChainNoder_segKey]*noding_boundaryChainNoder_segment,
) {
	for i := 0; i < segString.Size()-1; i++ {
		p0 := segString.GetCoordinate(i)
		p1 := segString.GetCoordinate(i + 1)
		seg := noding_newBoundaryChainNoder_segment(p0, p1, chainMap, i)
		key := seg.key()
		if _, exists := segSet[key]; exists {
			delete(segSet, key)
		} else {
			segSet[key] = seg
		}
	}
}

func (bcn *Noding_BoundaryChainNoder) markBoundarySegments(segSet map[noding_boundaryChainNoder_segKey]*noding_boundaryChainNoder_segment) {
	for _, seg := range segSet {
		seg.markBoundary()
	}
}

func (bcn *Noding_BoundaryChainNoder) extractChains(boundaryChains []*noding_boundaryChainMap) []Noding_SegmentString {
	chainList := make([]Noding_SegmentString, 0)
	for _, chainMap := range boundaryChains {
		chainMap.createChains(&chainList)
	}
	return chainList
}

// GetNodedSubstrings returns the boundary chain segment strings.
func (bcn *Noding_BoundaryChainNoder) GetNodedSubstrings() []Noding_SegmentString {
	return bcn.chainList
}

// noding_boundaryChainMap tracks which segments in a SegmentString are
// boundary segments.
type noding_boundaryChainMap struct {
	segString  Noding_SegmentString
	isBoundary []bool
}

func noding_newBoundaryChainMap(ss Noding_SegmentString) *noding_boundaryChainMap {
	return &noding_boundaryChainMap{
		segString:  ss,
		isBoundary: make([]bool, ss.Size()-1),
	}
}

func (bcm *noding_boundaryChainMap) setBoundarySegment(index int) {
	bcm.isBoundary[index] = true
}

func (bcm *noding_boundaryChainMap) createChains(chainList *[]Noding_SegmentString) {
	endIndex := 0
	for {
		startIndex := bcm.findChainStart(endIndex)
		if startIndex >= bcm.segString.Size()-1 {
			break
		}
		endIndex = bcm.findChainEnd(startIndex)
		ss := bcm.createChain(bcm.segString, startIndex, endIndex)
		*chainList = append(*chainList, ss)
	}
}

func (bcm *noding_boundaryChainMap) createChain(segString Noding_SegmentString, startIndex, endIndex int) Noding_SegmentString {
	pts := make([]*Geom_Coordinate, endIndex-startIndex+1)
	ipts := 0
	for i := startIndex; i < endIndex+1; i++ {
		pts[ipts] = segString.GetCoordinate(i).Copy()
		ipts++
	}
	bss := Noding_NewBasicSegmentString(pts, segString.GetData())
	return bss
}

func (bcm *noding_boundaryChainMap) findChainStart(index int) int {
	for index < len(bcm.isBoundary) && !bcm.isBoundary[index] {
		index++
	}
	return index
}

func (bcm *noding_boundaryChainMap) findChainEnd(index int) int {
	index++
	for index < len(bcm.isBoundary) && bcm.isBoundary[index] {
		index++
	}
	return index
}

// noding_boundaryChainNoder_segKey is a normalized segment key used only for
// map lookups. It contains only the coordinate values, ensuring that segments
// with the same coordinates are considered equal regardless of which polygon
// they came from. This mirrors the Java behavior where LineSegment.equals()
// and hashCode() only compare coordinates.
type noding_boundaryChainNoder_segKey struct {
	// Normalized coordinates (p0 < p1 lexicographically).
	p0x, p0y, p1x, p1y float64
}

// noding_boundaryChainNoder_segment represents a segment with associated
// marking data. The key() method extracts the coordinate-only key for map
// lookups.
type noding_boundaryChainNoder_segment struct {
	// Normalized coordinates (p0 < p1 lexicographically).
	p0x, p0y, p1x, p1y float64
	// Original segment information for marking.
	segMap *noding_boundaryChainMap
	index  int
}

func noding_newBoundaryChainNoder_segment(p0, p1 *Geom_Coordinate, segMap *noding_boundaryChainMap, index int) *noding_boundaryChainNoder_segment {
	seg := &noding_boundaryChainNoder_segment{
		segMap: segMap,
		index:  index,
	}
	// Normalize: ensure p0 <= p1 lexicographically.
	if p0.CompareTo(p1) <= 0 {
		seg.p0x = p0.GetX()
		seg.p0y = p0.GetY()
		seg.p1x = p1.GetX()
		seg.p1y = p1.GetY()
	} else {
		seg.p0x = p1.GetX()
		seg.p0y = p1.GetY()
		seg.p1x = p0.GetX()
		seg.p1y = p0.GetY()
	}
	return seg
}

func (seg *noding_boundaryChainNoder_segment) key() noding_boundaryChainNoder_segKey {
	return noding_boundaryChainNoder_segKey{
		p0x: seg.p0x,
		p0y: seg.p0y,
		p1x: seg.p1x,
		p1y: seg.p1y,
	}
}

func (seg *noding_boundaryChainNoder_segment) markBoundary() {
	seg.segMap.setBoundarySegment(seg.index)
}
