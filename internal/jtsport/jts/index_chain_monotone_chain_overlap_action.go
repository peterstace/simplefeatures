package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// IndexChain_MonotoneChainOverlapAction is the action for the internal iterator
// for performing overlap queries on a MonotoneChain.
type IndexChain_MonotoneChainOverlapAction struct {
	child java.Polymorphic

	OverlapSeg1 *Geom_LineSegment
	OverlapSeg2 *Geom_LineSegment
}

// GetChild returns the immediate child in the type hierarchy chain.
func (mco *IndexChain_MonotoneChainOverlapAction) GetChild() java.Polymorphic {
	return mco.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (mco *IndexChain_MonotoneChainOverlapAction) GetParent() java.Polymorphic {
	return nil
}

// IndexChain_NewMonotoneChainOverlapAction creates a new MonotoneChainOverlapAction.
func IndexChain_NewMonotoneChainOverlapAction() *IndexChain_MonotoneChainOverlapAction {
	return &IndexChain_MonotoneChainOverlapAction{
		OverlapSeg1: Geom_NewLineSegment(),
		OverlapSeg2: Geom_NewLineSegment(),
	}
}

// Overlap processes overlapping segments from two monotone chains.
// This function can be overridden if the original chains are needed.
func (mco *IndexChain_MonotoneChainOverlapAction) Overlap(mc1 *IndexChain_MonotoneChain, start1 int, mc2 *IndexChain_MonotoneChain, start2 int) {
	if impl, ok := java.GetLeaf(mco).(interface {
		Overlap_BODY(*IndexChain_MonotoneChain, int, *IndexChain_MonotoneChain, int)
	}); ok {
		impl.Overlap_BODY(mc1, start1, mc2, start2)
		return
	}
	mco.Overlap_BODY(mc1, start1, mc2, start2)
}

// Overlap_BODY is the implementation of Overlap.
func (mco *IndexChain_MonotoneChainOverlapAction) Overlap_BODY(mc1 *IndexChain_MonotoneChain, start1 int, mc2 *IndexChain_MonotoneChain, start2 int) {
	mc1.GetLineSegment(start1, mco.OverlapSeg1)
	mc2.GetLineSegment(start2, mco.OverlapSeg2)
	mco.OverlapSegments(mco.OverlapSeg1, mco.OverlapSeg2)
}

// OverlapSegments processes the actual line segments which overlap.
// This is a convenience function which can be overridden.
func (mco *IndexChain_MonotoneChainOverlapAction) OverlapSegments(seg1, seg2 *Geom_LineSegment) {
	if impl, ok := java.GetLeaf(mco).(interface {
		OverlapSegments_BODY(*Geom_LineSegment, *Geom_LineSegment)
	}); ok {
		impl.OverlapSegments_BODY(seg1, seg2)
		return
	}
	mco.OverlapSegments_BODY(seg1, seg2)
}

// OverlapSegments_BODY is the implementation of OverlapSegments.
func (mco *IndexChain_MonotoneChainOverlapAction) OverlapSegments_BODY(seg1, seg2 *Geom_LineSegment) {
	// Empty default implementation.
}
