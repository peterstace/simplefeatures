package jts

// OperationRelateng_EdgeSetIntersector indexes RelateSegmentStrings using
// monotone chains and an HPRtree for efficient intersection detection.
type OperationRelateng_EdgeSetIntersector struct {
	index      *IndexHprtree_HPRtree
	envelope   *Geom_Envelope
	monoChains []*IndexChain_MonotoneChain
	idCounter  int
}

// OperationRelateng_NewEdgeSetIntersector creates a new EdgeSetIntersector for
// the given edge sets and optional bounding envelope.
func OperationRelateng_NewEdgeSetIntersector(edgesA, edgesB []*OperationRelateng_RelateSegmentString, env *Geom_Envelope) *OperationRelateng_EdgeSetIntersector {
	esi := &OperationRelateng_EdgeSetIntersector{
		index:      IndexHprtree_NewHPRtree(),
		envelope:   env,
		monoChains: make([]*IndexChain_MonotoneChain, 0),
		idCounter:  0,
	}
	esi.addEdges(edgesA)
	esi.addEdges(edgesB)
	// Build index to ensure thread-safety.
	esi.index.Build()
	return esi
}

func (esi *OperationRelateng_EdgeSetIntersector) addEdges(segStrings []*OperationRelateng_RelateSegmentString) {
	for _, ss := range segStrings {
		esi.addToIndex(ss)
	}
}

func (esi *OperationRelateng_EdgeSetIntersector) addToIndex(segStr *OperationRelateng_RelateSegmentString) {
	segChains := IndexChain_MonotoneChainBuilder_GetChainsWithContext(segStr.GetCoordinates(), segStr)
	for _, mc := range segChains {
		if esi.envelope == nil || esi.envelope.IntersectsEnvelope(mc.GetEnvelope()) {
			mc.SetId(esi.idCounter)
			esi.idCounter++
			esi.index.Insert(mc.GetEnvelope(), mc)
			esi.monoChains = append(esi.monoChains, mc)
		}
	}
}

// Process processes all potential intersections using the given intersector.
func (esi *OperationRelateng_EdgeSetIntersector) Process(intersector *OperationRelateng_EdgeSegmentIntersector) {
	overlapAction := OperationRelateng_NewEdgeSegmentOverlapAction(intersector)

	for _, queryChain := range esi.monoChains {
		overlapChainsAny := esi.index.Query(queryChain.GetEnvelope())
		for _, testChainAny := range overlapChainsAny {
			testChain := testChainAny.(*IndexChain_MonotoneChain)
			// Following test makes sure we only compare each pair of chains once
			// and that we don't compare a chain to itself.
			if testChain.GetId() <= queryChain.GetId() {
				continue
			}

			testChain.ComputeOverlaps(queryChain, overlapAction.IndexChain_MonotoneChainOverlapAction)
			if intersector.IsDone() {
				return
			}
		}
	}
}
