package geom

import (
	"fmt"
	"math"
)

func newNodeSet(maxULPSize float64, sizeHint int) nodeSet {
	// The appropriate multiplication factor to use to calculate bucket size is
	// a bit of a guess.
	bucketSize := maxULPSize * 0x200
	return nodeSet{
		bucketSize,
		make(map[nodeBucket]XY, sizeHint),
		make(map[nodeBucket]XY, sizeHint),
	}
}

// nodeSet is a set of XY values (nodes). If an XY value is inserted, but it is
// "close" to an existing XY in the set, then the original XY is returned (and
// the new XY _not_ inserted). The two XYs essentially merge together.
type nodeSet struct {
	bucketWidth float64

	// Keep track of the nodes via two maps. The maps have identical structure,
	// but the buckets of the second node map are offset by half a bucket
	// width. This is to detect cases where two XYs are in adjacent buckets,
	// but still very close together.
	//
	// It's an invariant that entries for overlapping buckets in nodesA and
	// nodesB have the same node value.
	nodesA map[nodeBucket]XY
	nodesB map[nodeBucket]XY
}

type nodeBucket struct {
	x, y int
}

func (s nodeSet) insertOrGet(xy XY) XY {
	bucketA := nodeBucket{
		int(math.Floor(xy.X / s.bucketWidth)),
		int(math.Floor(xy.Y / s.bucketWidth)),
	}
	bucketB := nodeBucket{
		int(math.Floor((xy.X + s.bucketWidth/2) / s.bucketWidth)),
		int(math.Floor((xy.Y + s.bucketWidth/2) / s.bucketWidth)),
	}

	nodeA, okA := s.nodesA[bucketA]
	nodeB, okB := s.nodesB[bucketB]

	if okA && okB {
		if nodeA != nodeB {
			panic(fmt.Sprintf("nodeA != nodeB: %v vs %v", nodeA, nodeB))
		}
		return nodeA
	}

	if okA && !okB {
		s.nodesB[bucketB] = nodeA
		return nodeA
	}
	if okB && !okA {
		s.nodesA[bucketA] = nodeB
		return nodeB
	}

	s.nodesA[bucketA] = xy
	s.nodesB[bucketB] = xy
	return xy
}
