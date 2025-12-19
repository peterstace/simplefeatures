package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

func geomgraphIndex_MonotoneChainIndexer_toIntArray(list []int) []int {
	array := make([]int, len(list))
	copy(array, list)
	return array
}

// GeomgraphIndex_MonotoneChainIndexer provides methods to compute monotone chains
// for a sequence of points.
//
// MonotoneChains are a way of partitioning the segments of an edge to allow for
// fast searching of intersections. Specifically, a sequence of contiguous line
// segments is a monotone chain if all the vectors defined by the oriented
// segments lies in the same quadrant.
//
// Monotone Chains have the following useful properties:
//  1. the segments within a monotone chain will never intersect each other
//  2. the envelope of any contiguous subset of the segments in a monotone chain
//     is simply the envelope of the endpoints of the subset.
//
// Property 1 means that there is no need to test pairs of segments from within
// the same monotone chain for intersection. Property 2 allows binary search to
// be used to find the intersection points of two monotone chains. For many
// types of real-world data, these properties eliminate a large number of
// segment comparisons, producing substantial speed gains.
//
// Note that due to the efficient intersection test, there is no need to limit
// the size of chains to obtain fast performance.
type GeomgraphIndex_MonotoneChainIndexer struct {
	child java.Polymorphic
}

// GetChild returns the immediate child in the type hierarchy chain.
func (mci *GeomgraphIndex_MonotoneChainIndexer) GetChild() java.Polymorphic {
	return mci.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (mci *GeomgraphIndex_MonotoneChainIndexer) GetParent() java.Polymorphic {
	return nil
}

// GeomgraphIndex_NewMonotoneChainIndexer creates a new MonotoneChainIndexer.
func GeomgraphIndex_NewMonotoneChainIndexer() *GeomgraphIndex_MonotoneChainIndexer {
	return &GeomgraphIndex_MonotoneChainIndexer{}
}

// GetChainStartIndices finds the start (and end) indices of all monotone chains
// in the given coordinate array.
func (mci *GeomgraphIndex_MonotoneChainIndexer) GetChainStartIndices(pts []*Geom_Coordinate) []int {
	start := 0
	// Use heuristic to size initial slice.
	startIndexList := make([]int, 0, len(pts)/2)
	startIndexList = append(startIndexList, start)
	for {
		last := mci.findChainEnd(pts, start)
		startIndexList = append(startIndexList, last)
		start = last
		if start >= len(pts)-1 {
			break
		}
	}
	return startIndexList
}

// OLDgetChainStartIndices is an old version of GetChainStartIndices.
func (mci *GeomgraphIndex_MonotoneChainIndexer) OLDgetChainStartIndices(pts []*Geom_Coordinate) []int {
	start := 0
	startIndexList := make([]int, 0)
	startIndexList = append(startIndexList, start)
	for {
		last := mci.findChainEnd(pts, start)
		startIndexList = append(startIndexList, last)
		start = last
		if start >= len(pts)-1 {
			break
		}
	}
	return startIndexList
}

// findChainEnd returns the index of the last point in the monotone chain
// starting at the given index.
func (mci *GeomgraphIndex_MonotoneChainIndexer) findChainEnd(pts []*Geom_Coordinate, start int) int {
	// Determine quadrant for chain.
	chainQuad := Geom_Quadrant_QuadrantFromCoords(pts[start], pts[start+1])
	last := start + 1
	for last < len(pts) {
		// Compute quadrant for next possible segment in chain.
		quad := Geom_Quadrant_QuadrantFromCoords(pts[last-1], pts[last])
		if quad != chainQuad {
			break
		}
		last++
	}
	return last - 1
}
