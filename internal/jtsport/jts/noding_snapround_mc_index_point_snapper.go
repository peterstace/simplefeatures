package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

const nodingSnapround_MCIndexPointSnapper_SAFE_ENV_EXPANSION_FACTOR = 0.75

// NodingSnapround_MCIndexPointSnapper "snaps" all SegmentStrings in a
// SpatialIndex containing MonotoneChains to a given HotPixel.
type NodingSnapround_MCIndexPointSnapper struct {
	index *IndexHprtree_HPRtree
}

// NodingSnapround_NewMCIndexPointSnapper creates a new MCIndexPointSnapper
// with the given spatial index.
func NodingSnapround_NewMCIndexPointSnapper(index *IndexHprtree_HPRtree) *NodingSnapround_MCIndexPointSnapper {
	return &NodingSnapround_MCIndexPointSnapper{
		index: index,
	}
}

// Snap snaps (nodes) all interacting segments to this hot pixel. The hot pixel
// may represent a vertex of an edge, in which case this routine uses the
// optimization of not noding the vertex itself.
func (ps *NodingSnapround_MCIndexPointSnapper) Snap(hotPixel *NodingSnapround_HotPixel, parentEdge Noding_SegmentString, hotPixelVertexIndex int) bool {
	pixelEnv := ps.getSafeEnvelope(hotPixel)
	hotPixelSnapAction := nodingSnapround_newHotPixelSnapAction(hotPixel, parentEdge, hotPixelVertexIndex)

	items := ps.index.Query(pixelEnv)
	for _, item := range items {
		testChain := item.(*IndexChain_MonotoneChain)
		testChain.Select(pixelEnv, hotPixelSnapAction.IndexChain_MonotoneChainSelectAction)
	}
	return hotPixelSnapAction.isNodeAdded
}

// SnapSimple snaps a hot pixel without a parent edge.
func (ps *NodingSnapround_MCIndexPointSnapper) SnapSimple(hotPixel *NodingSnapround_HotPixel) bool {
	return ps.Snap(hotPixel, nil, -1)
}

// getSafeEnvelope returns a "safe" envelope that is guaranteed to contain the
// hot pixel. The envelope returned is larger than the exact envelope of the
// pixel by a safe margin.
func (ps *NodingSnapround_MCIndexPointSnapper) getSafeEnvelope(hp *NodingSnapround_HotPixel) *Geom_Envelope {
	safeTolerance := nodingSnapround_MCIndexPointSnapper_SAFE_ENV_EXPANSION_FACTOR / hp.GetScaleFactor()
	safeEnv := Geom_NewEnvelopeFromCoordinate(hp.GetCoordinate())
	safeEnv.ExpandBy(safeTolerance)
	return safeEnv
}

// nodingSnapround_HotPixelSnapAction is the select action for
// MCIndexPointSnapper.
type nodingSnapround_HotPixelSnapAction struct {
	*IndexChain_MonotoneChainSelectAction
	child               java.Polymorphic
	hotPixel            *NodingSnapround_HotPixel
	parentEdge          Noding_SegmentString
	hotPixelVertexIndex int
	isNodeAdded         bool
}

func (a *nodingSnapround_HotPixelSnapAction) GetChild() java.Polymorphic {
	return a.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (a *nodingSnapround_HotPixelSnapAction) GetParent() java.Polymorphic {
	return a.IndexChain_MonotoneChainSelectAction
}

func nodingSnapround_newHotPixelSnapAction(hotPixel *NodingSnapround_HotPixel, parentEdge Noding_SegmentString, hotPixelVertexIndex int) *nodingSnapround_HotPixelSnapAction {
	parent := &IndexChain_MonotoneChainSelectAction{}
	hpsa := &nodingSnapround_HotPixelSnapAction{
		IndexChain_MonotoneChainSelectAction: parent,
		hotPixel:                            hotPixel,
		parentEdge:                          parentEdge,
		hotPixelVertexIndex:                 hotPixelVertexIndex,
	}
	parent.child = hpsa
	return hpsa
}

// Select_BODY checks if a segment of the monotone chain intersects the hot
// pixel vertex and introduces a snap node if so. Optimized to avoid noding
// segments which contain the vertex (which otherwise would cause every vertex
// to be noded).
func (a *nodingSnapround_HotPixelSnapAction) Select_BODY(mc *IndexChain_MonotoneChain, startIndex int) {
	ss := mc.GetContext().(Noding_SegmentString)
	// Check to avoid snapping a hotPixel vertex to its original vertex. This
	// method is called on segments which intersect the hot pixel. If either
	// end of the segment is equal to the hot pixel do not snap.
	if a.parentEdge != nil && ss == a.parentEdge {
		if startIndex == a.hotPixelVertexIndex || startIndex+1 == a.hotPixelVertexIndex {
			return
		}
	}
	// Records if this HotPixel caused any node to be added.
	a.isNodeAdded = a.addSnappedNode(a.hotPixel, ss, startIndex) || a.isNodeAdded
}

// addSnappedNode adds a new node (equal to the snap pt) to the specified
// segment if the segment passes through the hot pixel.
func (a *nodingSnapround_HotPixelSnapAction) addSnappedNode(hotPixel *NodingSnapround_HotPixel, segStr Noding_SegmentString, segIndex int) bool {
	p0 := segStr.GetCoordinate(segIndex)
	p1 := segStr.GetCoordinate(segIndex + 1)

	if hotPixel.IntersectsSegment(p0, p1) {
		nss := segStr.(*Noding_NodedSegmentString)
		nss.AddIntersection(hotPixel.GetCoordinate(), segIndex)
		return true
	}
	return false
}
