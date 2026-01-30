package jts

// NodingSnapround_HotPixelIndex is an index which creates unique HotPixels for
// provided points, and performs range queries on them. The points passed to
// the index do not need to be rounded to the specified scale factor; this is
// done internally when creating the HotPixels for them.
type NodingSnapround_HotPixelIndex struct {
	precModel   *Geom_PrecisionModel
	scaleFactor float64
	// Use a kd-tree to index the pixel centers for optimum performance. Since
	// HotPixels have an extent, range queries to the index must enlarge the
	// query range by a suitable value (using the pixel width is safest).
	index *IndexKdtree_KdTree
}

// NodingSnapround_NewHotPixelIndex creates a new HotPixelIndex with the given
// precision model.
func NodingSnapround_NewHotPixelIndex(pm *Geom_PrecisionModel) *NodingSnapround_HotPixelIndex {
	return &NodingSnapround_HotPixelIndex{
		precModel:   pm,
		scaleFactor: pm.GetScale(),
		index:       IndexKdtree_NewKdTree(),
	}
}

// Add adds a list of points as non-node pixels.
func (hpi *NodingSnapround_HotPixelIndex) Add(pts []*Geom_Coordinate) {
	// Shuffle the points before adding. This avoids having long monotonic runs
	// of points causing an unbalanced KD-tree, which would create performance
	// and robustness issues.
	shuffler := nodingSnapround_newCoordinateShuffler(pts)
	for shuffler.HasNext() {
		hpi.AddPoint(shuffler.Next())
	}
}

// AddNodes adds a list of points as node pixels.
func (hpi *NodingSnapround_HotPixelIndex) AddNodes(pts []*Geom_Coordinate) {
	// Node points are not shuffled, since they are added after the vertex
	// points, and hence the KD-tree should be reasonably balanced already.
	for _, pt := range pts {
		hp := hpi.AddPoint(pt)
		hp.SetToNode()
	}
}

// AddPoint adds a point as a Hot Pixel. If the point has been added already,
// it is marked as a node.
func (hpi *NodingSnapround_HotPixelIndex) AddPoint(p *Geom_Coordinate) *NodingSnapround_HotPixel {
	pRound := hpi.round(p)

	hp := hpi.find(pRound)
	// Hot Pixels which are added more than once must have more than one vertex
	// in them and thus must be nodes.
	if hp != nil {
		hp.SetToNode()
		return hp
	}

	// A pixel containing the point was not found, so create a new one. It is
	// initially set to NOT be a node (but may become one later on).
	hp = NodingSnapround_NewHotPixel(pRound, hpi.scaleFactor)
	hpi.index.InsertWithData(hp.GetCoordinate(), hp)
	return hp
}

func (hpi *NodingSnapround_HotPixelIndex) find(pixelPt *Geom_Coordinate) *NodingSnapround_HotPixel {
	kdNode := hpi.index.QueryPoint(pixelPt)
	if kdNode == nil {
		return nil
	}
	return kdNode.GetData().(*NodingSnapround_HotPixel)
}

func (hpi *NodingSnapround_HotPixelIndex) round(pt *Geom_Coordinate) *Geom_Coordinate {
	p2 := pt.Copy()
	hpi.precModel.MakePreciseCoordinate(p2)
	return p2
}

// Query visits all the hot pixels which may intersect a segment (p0-p1). The
// visitor must determine whether each hot pixel actually intersects the
// segment.
func (hpi *NodingSnapround_HotPixelIndex) Query(p0, p1 *Geom_Coordinate, visitor IndexKdtree_KdNodeVisitor) {
	queryEnv := Geom_NewEnvelopeFromCoordinates(p0, p1)
	// Expand query range to account for HotPixel extent. Expand by full width
	// of one pixel to be safe.
	queryEnv.ExpandBy(1.0 / hpi.scaleFactor)
	hpi.index.QueryEnvelopeVisitor(queryEnv, visitor)
}

// nodingSnapround_coordinateShuffler shuffles coordinates using the
// Fisher-Yates shuffle algorithm.
type nodingSnapround_coordinateShuffler struct {
	coordinates []*Geom_Coordinate
	indices     []int
	index       int
	rnd         *math_lcgRandom
}

func nodingSnapround_newCoordinateShuffler(pts []*Geom_Coordinate) *nodingSnapround_coordinateShuffler {
	indices := make([]int, len(pts))
	for i := range pts {
		indices[i] = i
	}
	return &nodingSnapround_coordinateShuffler{
		coordinates: pts,
		indices:     indices,
		index:       len(pts) - 1,
		rnd:         &math_lcgRandom{state: 13},
	}
}

func (cs *nodingSnapround_coordinateShuffler) HasNext() bool {
	return cs.index >= 0
}

func (cs *nodingSnapround_coordinateShuffler) Next() *Geom_Coordinate {
	j := cs.rnd.nextInt(cs.index + 1)
	res := cs.coordinates[cs.indices[j]]
	cs.indices[j] = cs.indices[cs.index]
	cs.index--
	return res
}
