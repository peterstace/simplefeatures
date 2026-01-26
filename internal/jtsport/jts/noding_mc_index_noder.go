package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Noding_MCIndexNoder nodes a set of SegmentStrings using a index based on
// MonotoneChains and a SpatialIndex. The SpatialIndex used should be something
// that supports envelope (range) queries efficiently (such as HPRtree, which is
// the default index provided).
//
// The noder supports using an overlap tolerance distance. This allows
// determining segment intersection using a buffer for uses involving snapping
// with a distance tolerance.
type Noding_MCIndexNoder struct {
	singlePassNoder  *Noding_SinglePassNoder
	monoChains       []*IndexChain_MonotoneChain
	index            *IndexHprtree_HPRtree
	idCounter        int
	nodedSegStrings  []Noding_SegmentString
	nOverlaps        int
	overlapTolerance float64
}

var _ Noding_Noder = (*Noding_MCIndexNoder)(nil)

func (n *Noding_MCIndexNoder) IsNoding_Noder() {}

// Noding_NewMCIndexNoder creates a new MCIndexNoder.
func Noding_NewMCIndexNoder() *Noding_MCIndexNoder {
	return &Noding_MCIndexNoder{
		singlePassNoder: Noding_NewSinglePassNoder(),
		monoChains:      make([]*IndexChain_MonotoneChain, 0),
		index:           IndexHprtree_NewHPRtree(),
		idCounter:       0,
	}
}

// Noding_NewMCIndexNoderWithIntersector creates a new MCIndexNoder with the
// given segment intersector.
func Noding_NewMCIndexNoderWithIntersector(si Noding_SegmentIntersector) *Noding_MCIndexNoder {
	return &Noding_MCIndexNoder{
		singlePassNoder: Noding_NewSinglePassNoderWithIntersector(si),
		monoChains:      make([]*IndexChain_MonotoneChain, 0),
		index:           IndexHprtree_NewHPRtree(),
		idCounter:       0,
	}
}

// Noding_NewMCIndexNoderWithIntersectorAndTolerance creates a new MCIndexNoder
// with the given segment intersector and overlap tolerance.
func Noding_NewMCIndexNoderWithIntersectorAndTolerance(si Noding_SegmentIntersector, overlapTolerance float64) *Noding_MCIndexNoder {
	return &Noding_MCIndexNoder{
		singlePassNoder:  Noding_NewSinglePassNoderWithIntersector(si),
		monoChains:       make([]*IndexChain_MonotoneChain, 0),
		index:            IndexHprtree_NewHPRtree(),
		idCounter:        0,
		overlapTolerance: overlapTolerance,
	}
}

// GetMonotoneChains returns the monotone chains.
func (n *Noding_MCIndexNoder) GetMonotoneChains() []*IndexChain_MonotoneChain {
	return n.monoChains
}

// GetIndex returns the spatial index.
func (n *Noding_MCIndexNoder) GetIndex() *IndexHprtree_HPRtree {
	return n.index
}

// SetSegmentIntersector sets the SegmentIntersector to use with this noder.
func (n *Noding_MCIndexNoder) SetSegmentIntersector(segInt Noding_SegmentIntersector) {
	n.singlePassNoder.SetSegmentIntersector(segInt)
}

// GetNodedSubstrings returns a collection of fully noded SegmentStrings.
func (n *Noding_MCIndexNoder) GetNodedSubstrings() []Noding_SegmentString {
	// Convert nodedSegStrings to NodedSegmentString slice.
	nssSlice := make([]*Noding_NodedSegmentString, len(n.nodedSegStrings))
	for i, ss := range n.nodedSegStrings {
		// The segment strings should already be NodedSegmentStrings.
		nssSlice[i] = ss.(*Noding_NodedSegmentString)
	}
	nodedResult := Noding_NodedSegmentString_GetNodedSubstrings(nssSlice)
	// Convert back to SegmentString slice.
	result := make([]Noding_SegmentString, len(nodedResult))
	for i, nss := range nodedResult {
		result[i] = nss
	}
	return result
}

// ComputeNodes computes the noding for a collection of SegmentStrings.
func (n *Noding_MCIndexNoder) ComputeNodes(inputSegStrings []Noding_SegmentString) {
	n.nodedSegStrings = inputSegStrings
	for _, segStr := range inputSegStrings {
		n.add(segStr)
	}
	n.intersectChains()
}

func (n *Noding_MCIndexNoder) intersectChains() {
	overlapAction := noding_NewSegmentOverlapAction(n.singlePassNoder.segInt)

	for _, queryChain := range n.monoChains {
		queryEnv := queryChain.GetEnvelopeWithExpansion(n.overlapTolerance)
		overlapChains := n.index.Query(queryEnv)
		for _, item := range overlapChains {
			testChain := item.(*IndexChain_MonotoneChain)
			// following test makes sure we only compare each pair of chains once
			// and that we don't compare a chain to itself
			if testChain.GetId() > queryChain.GetId() {
				queryChain.ComputeOverlapsWithTolerance(testChain, n.overlapTolerance, overlapAction.IndexChain_MonotoneChainOverlapAction)
				n.nOverlaps++
			}
			// short-circuit if possible
			if n.singlePassNoder.segInt.IsDone() {
				return
			}
		}
	}
}

func (n *Noding_MCIndexNoder) add(segStr Noding_SegmentString) {
	segChains := IndexChain_MonotoneChainBuilder_GetChainsWithContext(segStr.GetCoordinates(), segStr)
	for _, mc := range segChains {
		mc.SetId(n.idCounter)
		n.idCounter++
		n.index.Insert(mc.GetEnvelopeWithExpansion(n.overlapTolerance), mc)
		n.monoChains = append(n.monoChains, mc)
	}
}

// noding_SegmentOverlapAction is the overlap action for MCIndexNoder.
type noding_SegmentOverlapAction struct {
	*IndexChain_MonotoneChainOverlapAction
	child java.Polymorphic
	si    Noding_SegmentIntersector
}

// GetChild returns the immediate child in the type hierarchy chain.
func (a *noding_SegmentOverlapAction) GetChild() java.Polymorphic {
	return a.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (a *noding_SegmentOverlapAction) GetParent() java.Polymorphic {
	return a.IndexChain_MonotoneChainOverlapAction
}

func noding_NewSegmentOverlapAction(si Noding_SegmentIntersector) *noding_SegmentOverlapAction {
	parent := &IndexChain_MonotoneChainOverlapAction{}
	soa := &noding_SegmentOverlapAction{
		IndexChain_MonotoneChainOverlapAction: parent,
		si:                                   si,
	}
	parent.child = soa
	return soa
}

// Overlap_BODY handles overlap between two monotone chains.
func (a *noding_SegmentOverlapAction) Overlap_BODY(mc1 *IndexChain_MonotoneChain, start1 int, mc2 *IndexChain_MonotoneChain, start2 int) {
	ss1 := mc1.GetContext().(Noding_SegmentString)
	ss2 := mc2.GetContext().(Noding_SegmentString)
	a.si.ProcessIntersections(ss1, start1, ss2, start2)
}
