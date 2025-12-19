package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Noding_MCIndexSegmentSetMutualIntersector intersects two sets of
// SegmentStrings using an index based on MonotoneChains and a SpatialIndex.
//
// Thread-safe and immutable.
type Noding_MCIndexSegmentSetMutualIntersector struct {
	index            *IndexStrtree_STRtree
	overlapTolerance float64
	envelope         *Geom_Envelope
}

// Noding_NewMCIndexSegmentSetMutualIntersector constructs a new intersector for
// a given set of SegmentStrings.
func Noding_NewMCIndexSegmentSetMutualIntersector(baseSegStrings []Noding_SegmentString) *Noding_MCIndexSegmentSetMutualIntersector {
	intersector := &Noding_MCIndexSegmentSetMutualIntersector{
		index: IndexStrtree_NewSTRtree(),
	}
	intersector.initBaseSegments(baseSegStrings)
	return intersector
}

// Noding_NewMCIndexSegmentSetMutualIntersectorWithEnvelope constructs a new
// intersector for a given set of SegmentStrings and an envelope filter.
func Noding_NewMCIndexSegmentSetMutualIntersectorWithEnvelope(baseSegStrings []Noding_SegmentString, env *Geom_Envelope) *Noding_MCIndexSegmentSetMutualIntersector {
	intersector := &Noding_MCIndexSegmentSetMutualIntersector{
		index:    IndexStrtree_NewSTRtree(),
		envelope: env,
	}
	intersector.initBaseSegments(baseSegStrings)
	return intersector
}

// Noding_NewMCIndexSegmentSetMutualIntersectorWithTolerance constructs a new
// intersector for a given set of SegmentStrings and an overlap tolerance.
func Noding_NewMCIndexSegmentSetMutualIntersectorWithTolerance(baseSegStrings []Noding_SegmentString, overlapTolerance float64) *Noding_MCIndexSegmentSetMutualIntersector {
	intersector := &Noding_MCIndexSegmentSetMutualIntersector{
		index:            IndexStrtree_NewSTRtree(),
		overlapTolerance: overlapTolerance,
	}
	intersector.initBaseSegments(baseSegStrings)
	return intersector
}

// GetIndex gets the index constructed over the base segment strings.
//
// NOTE: To retain thread-safety, treat returned value as immutable!
func (i *Noding_MCIndexSegmentSetMutualIntersector) GetIndex() *IndexStrtree_STRtree {
	return i.index
}

func (i *Noding_MCIndexSegmentSetMutualIntersector) initBaseSegments(segStrings []Noding_SegmentString) {
	for _, ss := range segStrings {
		if ss.Size() == 0 {
			continue
		}
		i.addToIndex(ss)
	}
	// Build index to ensure thread-safety.
	i.index.Build()
}

func (i *Noding_MCIndexSegmentSetMutualIntersector) addToIndex(segStr Noding_SegmentString) {
	segChains := IndexChain_MonotoneChainBuilder_GetChainsWithContext(segStr.GetCoordinates(), segStr)
	for _, mc := range segChains {
		if i.envelope == nil || i.envelope.IntersectsEnvelope(mc.GetEnvelope()) {
			i.index.Insert(mc.GetEnvelopeWithExpansion(i.overlapTolerance), mc)
		}
	}
}

// Process calls SegmentIntersector.ProcessIntersections for all candidate
// intersections between the given collection of SegmentStrings and the set of
// indexed segments.
func (i *Noding_MCIndexSegmentSetMutualIntersector) Process(segStrings []Noding_SegmentString, segInt Noding_SegmentIntersector) {
	var monoChains []*IndexChain_MonotoneChain
	for _, ss := range segStrings {
		monoChains = i.addToMonoChains(ss, monoChains)
	}
	i.intersectChains(monoChains, segInt)
}

func (i *Noding_MCIndexSegmentSetMutualIntersector) addToMonoChains(segStr Noding_SegmentString, monoChains []*IndexChain_MonotoneChain) []*IndexChain_MonotoneChain {
	if segStr.Size() == 0 {
		return monoChains
	}
	segChains := IndexChain_MonotoneChainBuilder_GetChainsWithContext(segStr.GetCoordinates(), segStr)
	for _, mc := range segChains {
		if i.envelope == nil || i.envelope.IntersectsEnvelope(mc.GetEnvelope()) {
			monoChains = append(monoChains, mc)
		}
	}
	return monoChains
}

func (i *Noding_MCIndexSegmentSetMutualIntersector) intersectChains(monoChains []*IndexChain_MonotoneChain, segInt Noding_SegmentIntersector) {
	overlapAction := Noding_NewSegmentOverlapAction(segInt)

	for _, queryChain := range monoChains {
		queryEnv := queryChain.GetEnvelopeWithExpansion(i.overlapTolerance)
		overlapChains := i.index.Query(queryEnv)
		for _, item := range overlapChains {
			testChain := item.(*IndexChain_MonotoneChain)
			queryChain.ComputeOverlapsWithTolerance(testChain, i.overlapTolerance, overlapAction.IndexChain_MonotoneChainOverlapAction)
			if segInt.IsDone() {
				return
			}
		}
	}
}

// Noding_SegmentOverlapAction is the MonotoneChainOverlapAction for processing
// segment overlaps.
type Noding_SegmentOverlapAction struct {
	*IndexChain_MonotoneChainOverlapAction
	child java.Polymorphic
	si    Noding_SegmentIntersector
}

// GetChild returns the immediate child in the type hierarchy chain.
func (a *Noding_SegmentOverlapAction) GetChild() java.Polymorphic {
	return a.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (a *Noding_SegmentOverlapAction) GetParent() java.Polymorphic {
	return a.IndexChain_MonotoneChainOverlapAction
}

// Noding_NewSegmentOverlapAction creates a new SegmentOverlapAction.
func Noding_NewSegmentOverlapAction(si Noding_SegmentIntersector) *Noding_SegmentOverlapAction {
	base := IndexChain_NewMonotoneChainOverlapAction()
	action := &Noding_SegmentOverlapAction{
		IndexChain_MonotoneChainOverlapAction: base,
		si:                                   si,
	}
	base.child = action
	return action
}

// Overlap_BODY processes overlapping segments from two monotone chains.
func (a *Noding_SegmentOverlapAction) Overlap_BODY(mc1 *IndexChain_MonotoneChain, start1 int, mc2 *IndexChain_MonotoneChain, start2 int) {
	ss1 := mc1.GetContext().(Noding_SegmentString)
	ss2 := mc2.GetContext().(Noding_SegmentString)
	a.si.ProcessIntersections(ss1, start1, ss2, start2)
}
