package geom

import (
	"math"
)

func newNodeSet(maxULPSize float64, sizeHint int) nodeSet {
	// The appropriate multiplication factor to use to calculate bucket size is
	// a bit of a guess.
	bucketSize := maxULPSize * 0x200
	return nodeSet{
		bucketSize,
		make(map[nodeBucket]XY, sizeHint),
	}
}

// nodeSet is a set of XY values (nodes). If an XY value is inserted, but it is
// "close" to an existing XY in the set, then the original XY is returned (and
// the new XY _not_ inserted). The two XYs essentially merge together.
type nodeSet struct {
	bucketWidth float64
	nodes       map[nodeBucket]XY
}

type nodeBucket struct {
	x, y int
}

func (s nodeSet) insertOrGet(xy XY) XY {
	b := nodeBucket{
		int(math.Floor(xy.X / s.bucketWidth)),
		int(math.Floor(xy.Y / s.bucketWidth)),
	}
	for _, offset := range [...]nodeBucket{
		b,
		{b.x - 1, b.y - 1},
		{b.x - 1, b.y},
		{b.x - 1, b.y + 1},
		{b.x, b.y - 1},
		{b.x, b.y + 1},
		{b.x + 1, b.y - 1},
		{b.x + 1, b.y},
		{b.x + 1, b.y + 1},
	} {
		if node, ok := s.nodes[offset]; ok {
			return node
		}
	}
	s.nodes[b] = xy
	return xy
}

func (s nodeSet) list() []XY {
	xys := make([]XY, 0, len(s.nodes))
	for _, xy := range s.nodes {
		xys = append(xys, xy)
	}
	return xys
}
