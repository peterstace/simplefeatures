package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// GeomgraphIndex_MonotoneChain wraps a MonotoneChainEdge with a specific chain index.
type GeomgraphIndex_MonotoneChain struct {
	child java.Polymorphic

	mce        *GeomgraphIndex_MonotoneChainEdge
	chainIndex int
}

// GetChild returns the immediate child in the type hierarchy chain.
func (mc *GeomgraphIndex_MonotoneChain) GetChild() java.Polymorphic {
	return mc.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (mc *GeomgraphIndex_MonotoneChain) GetParent() java.Polymorphic {
	return nil
}

// GeomgraphIndex_NewMonotoneChain creates a new MonotoneChain.
func GeomgraphIndex_NewMonotoneChain(mce *GeomgraphIndex_MonotoneChainEdge, chainIndex int) *GeomgraphIndex_MonotoneChain {
	return &GeomgraphIndex_MonotoneChain{
		mce:        mce,
		chainIndex: chainIndex,
	}
}

// ComputeIntersections computes intersections between this chain and another.
func (mc *GeomgraphIndex_MonotoneChain) ComputeIntersections(other *GeomgraphIndex_MonotoneChain, si *GeomgraphIndex_SegmentIntersector) {
	mc.mce.ComputeIntersectsForChain(mc.chainIndex, other.mce, other.chainIndex, si)
}
