package geom

import (
	"math"
)

type nodeSet struct {
	bucketWidth float64
	nodes       map[nodeBucket]XY
}

type nodeBucket struct {
	x, y int
}

func newNodeSet(ulp float64, sizeHint int) nodeSet {
	return nodeSet{
		ulp * 0x1000,
		make(map[nodeBucket]XY, sizeHint),
	}
}

func (s nodeSet) insertOrGet(xy XY) XY {
	half := 0.5 * s.bucketWidth
	b := nodeBucket{
		x: int(math.Floor((xy.X + half) / s.bucketWidth)),
		y: int(math.Floor((xy.Y + half) / s.bucketWidth)),
	}
	if node, ok := s.nodes[b]; ok {
		return node
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
