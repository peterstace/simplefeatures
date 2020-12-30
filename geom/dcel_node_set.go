package geom

import "math"

func newNodeSet(maxULPSize float64, sizeHint int) nodeSet {
	// The appropriate multiplication factor to use to calculate bucket size is
	// a bit of a guess.
	bucketSize := maxULPSize * 0xff
	return nodeSet{bucketSize, make(map[nodeBucket]XY, sizeHint)}
}

// nodeSet is a set of XY values (nodes). If an XY value is inserted, but it is
// "close" to an existing XY in the set, then the original XY is returned (and
// the new XY _not_ inserted). The two XYs essentially merge together.
type nodeSet struct {
	bucketSize float64
	nodes      map[nodeBucket]XY
}

type nodeBucket struct {
	x, y int
}

func (s nodeSet) insertOrGet(xy XY) XY {
	bucket := nodeBucket{
		int(math.Floor(xy.X / s.bucketSize)),
		int(math.Floor(xy.Y / s.bucketSize)),
	}
	xNext := bucket.x + 1
	xPrev := bucket.x - 1
	yNext := bucket.y + 1
	yPrev := bucket.y - 1

	for _, bucket := range [...]nodeBucket{
		bucket, // the original bucket goes first, since it's the most likely entry
		nodeBucket{bucket.x, yNext},
		nodeBucket{bucket.x, yPrev},
		nodeBucket{xPrev, yPrev},
		nodeBucket{xPrev, bucket.y},
		nodeBucket{xPrev, yNext},
		nodeBucket{xNext, yPrev},
		nodeBucket{xNext, bucket.y},
		nodeBucket{xNext, yNext},
	} {
		node, ok := s.nodes[bucket]
		if ok {
			return node
		}
	}
	s.nodes[bucket] = xy
	return xy
}
