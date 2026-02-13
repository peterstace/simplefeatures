package jts

// operationBuffer_SegmentMCIndex is a spatial index over a segment sequence
// using MonotoneChains.
type operationBuffer_SegmentMCIndex struct {
	index *IndexStrtree_STRtree
}

// operationBuffer_newSegmentMCIndex creates a new SegmentMCIndex for the given coordinates.
func operationBuffer_newSegmentMCIndex(segs []*Geom_Coordinate) *operationBuffer_SegmentMCIndex {
	smi := &operationBuffer_SegmentMCIndex{}
	smi.index = smi.buildIndex(segs)
	return smi
}

func (smi *operationBuffer_SegmentMCIndex) buildIndex(segs []*Geom_Coordinate) *IndexStrtree_STRtree {
	index := IndexStrtree_NewSTRtree()
	segChains := IndexChain_MonotoneChainBuilder_GetChainsWithContext(segs, segs)
	for _, mc := range segChains {
		index.Insert(mc.GetEnvelope(), mc)
	}
	return index
}

// Query queries the index with an envelope and calls the action for each matching segment.
func (smi *operationBuffer_SegmentMCIndex) Query(env *Geom_Envelope, action *IndexChain_MonotoneChainSelectAction) {
	smi.index.QueryWithVisitor(env, Index_NewItemVisitorFunc(func(item any) {
		testChain := item.(*IndexChain_MonotoneChain)
		testChain.Select(env, action)
	}))
}
