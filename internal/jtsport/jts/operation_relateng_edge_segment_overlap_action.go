package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// OperationRelateng_EdgeSegmentOverlapAction is the action for processing
// overlapping monotone chain segments during RelateNG computation.
type OperationRelateng_EdgeSegmentOverlapAction struct {
	*IndexChain_MonotoneChainOverlapAction
	child java.Polymorphic
	si    Noding_SegmentIntersector
}

// GetChild returns the immediate child in the type hierarchy chain.
func (esoa *OperationRelateng_EdgeSegmentOverlapAction) GetChild() java.Polymorphic {
	return esoa.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (esoa *OperationRelateng_EdgeSegmentOverlapAction) GetParent() java.Polymorphic {
	return esoa.IndexChain_MonotoneChainOverlapAction
}

// OperationRelateng_NewEdgeSegmentOverlapAction creates a new
// EdgeSegmentOverlapAction.
func OperationRelateng_NewEdgeSegmentOverlapAction(si Noding_SegmentIntersector) *OperationRelateng_EdgeSegmentOverlapAction {
	parent := IndexChain_NewMonotoneChainOverlapAction()
	esoa := &OperationRelateng_EdgeSegmentOverlapAction{
		IndexChain_MonotoneChainOverlapAction: parent,
		si:                                   si,
	}
	parent.child = esoa
	return esoa
}

// Overlap_BODY processes overlapping segments from two monotone chains.
func (esoa *OperationRelateng_EdgeSegmentOverlapAction) Overlap_BODY(mc1 *IndexChain_MonotoneChain, start1 int, mc2 *IndexChain_MonotoneChain, start2 int) {
	ss1 := mc1.GetContext().(*OperationRelateng_RelateSegmentString)
	ss2 := mc2.GetContext().(*OperationRelateng_RelateSegmentString)
	esoa.si.ProcessIntersections(ss1, start1, ss2, start2)
}
