package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// IndexChain_MonotoneChainSelectAction is the action for the internal iterator
// for performing envelope select queries on a MonotoneChain.
type IndexChain_MonotoneChainSelectAction struct {
	child java.Polymorphic

	// SelectedSegment is used during the MonotoneChain search process.
	SelectedSegment *Geom_LineSegment
}

// GetChild returns the immediate child in the type hierarchy chain.
func (mcs *IndexChain_MonotoneChainSelectAction) GetChild() java.Polymorphic {
	return mcs.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (mcs *IndexChain_MonotoneChainSelectAction) GetParent() java.Polymorphic {
	return nil
}

// IndexChain_NewMonotoneChainSelectAction creates a new MonotoneChainSelectAction.
func IndexChain_NewMonotoneChainSelectAction() *IndexChain_MonotoneChainSelectAction {
	return &IndexChain_MonotoneChainSelectAction{
		SelectedSegment: Geom_NewLineSegment(),
	}
}

// Select processes a segment in the context of the parent chain.
// This method is overridden to process a segment in the context of the parent
// chain.
func (mcs *IndexChain_MonotoneChainSelectAction) Select(mc *IndexChain_MonotoneChain, startIndex int) {
	if impl, ok := java.GetLeaf(mcs).(interface {
		Select_BODY(*IndexChain_MonotoneChain, int)
	}); ok {
		impl.Select_BODY(mc, startIndex)
		return
	}
	mcs.Select_BODY(mc, startIndex)
}

// Select_BODY is the implementation of Select.
func (mcs *IndexChain_MonotoneChainSelectAction) Select_BODY(mc *IndexChain_MonotoneChain, startIndex int) {
	mc.GetLineSegment(startIndex, mcs.SelectedSegment)
	// Call this routine in case SelectSegment was overridden.
	mcs.SelectSegment(mcs.SelectedSegment)
}

// SelectSegment processes the actual line segment which is selected.
// This is a convenience method which can be overridden.
func (mcs *IndexChain_MonotoneChainSelectAction) SelectSegment(seg *Geom_LineSegment) {
	if impl, ok := java.GetLeaf(mcs).(interface {
		SelectSegment_BODY(*Geom_LineSegment)
	}); ok {
		impl.SelectSegment_BODY(seg)
		return
	}
	mcs.SelectSegment_BODY(seg)
}

// SelectSegment_BODY is the implementation of SelectSegment.
func (mcs *IndexChain_MonotoneChainSelectAction) SelectSegment_BODY(seg *Geom_LineSegment) {
	// Empty default implementation.
}
