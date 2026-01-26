package jts

// NodingSnapround_SnapRoundingNoder uses Snap Rounding to compute a rounded,
// fully noded arrangement from a set of SegmentStrings, in a performant way,
// and avoiding unnecessary noding.
//
// Implements the Snap Rounding technique described in the papers by Hobby,
// Guibas & Marimont, and Goodrich et al. Snap Rounding enforces that all
// output vertices lie on a uniform grid, which is determined by the provided
// PrecisionModel.
//
// Input vertices do not have to be rounded to the grid beforehand; this is
// done during the snap-rounding process. In fact, rounding cannot be done a
// priori, since rounding vertices by themselves can distort the rounded
// topology of the arrangement (i.e. by moving segments away from hot pixels
// that would otherwise intersect them, or by moving vertices across segments).
//
// To minimize the number of introduced nodes, the Snap-Rounding Noder avoids
// creating nodes at edge vertices if there is no intersection or snap at that
// location. However, if two different input edges contain identical segments,
// each of the segment vertices will be noded. This still provides fully-noded
// output.
type NodingSnapround_SnapRoundingNoder struct {
	pm            *Geom_PrecisionModel
	pixelIndex    *NodingSnapround_HotPixelIndex
	snappedResult []*Noding_NodedSegmentString
}

// Compile-time check that NodingSnapround_SnapRoundingNoder implements Noding_Noder.
var _ Noding_Noder = (*NodingSnapround_SnapRoundingNoder)(nil)

// IsNoding_Noder is a marker method for interface identification.
func (srn *NodingSnapround_SnapRoundingNoder) IsNoding_Noder() {}

// The division factor used to determine nearness distance tolerance for
// intersection detection.
const nodingSnapround_SnapRoundingNoder_NEARNESS_FACTOR = 100

// NodingSnapround_NewSnapRoundingNoder creates a new SnapRoundingNoder with
// the given precision model.
func NodingSnapround_NewSnapRoundingNoder(pm *Geom_PrecisionModel) *NodingSnapround_SnapRoundingNoder {
	return &NodingSnapround_SnapRoundingNoder{
		pm:         pm,
		pixelIndex: NodingSnapround_NewHotPixelIndex(pm),
	}
}

// GetNodedSubstrings returns a Collection of NodedSegmentStrings
// representing the substrings.
func (srn *NodingSnapround_SnapRoundingNoder) GetNodedSubstrings() []Noding_SegmentString {
	nodedResult := Noding_NodedSegmentString_GetNodedSubstrings(srn.snappedResult)
	result := make([]Noding_SegmentString, len(nodedResult))
	for i, nss := range nodedResult {
		result[i] = nss
	}
	return result
}

// ComputeNodes computes the nodes in the snap-rounding line arrangement.
// The nodes are added to the NodedSegmentStrings provided as the input.
func (srn *NodingSnapround_SnapRoundingNoder) ComputeNodes(inputSegmentStrings []Noding_SegmentString) {
	// Convert to NodedSegmentString slice.
	nssSlice := make([]*Noding_NodedSegmentString, len(inputSegmentStrings))
	for i, ss := range inputSegmentStrings {
		nssSlice[i] = ss.(*Noding_NodedSegmentString)
	}
	srn.snappedResult = srn.snapRound(nssSlice)
}

func (srn *NodingSnapround_SnapRoundingNoder) snapRound(segStrings []*Noding_NodedSegmentString) []*Noding_NodedSegmentString {
	// Determine hot pixels for intersections and vertices. This is done BEFORE
	// the input lines are rounded, to avoid distorting the line arrangement
	// (rounding can cause vertices to move across edges).
	srn.addIntersectionPixels(segStrings)
	srn.addVertexPixels(segStrings)

	snapped := srn.computeSnaps(segStrings)
	return snapped
}

// addIntersectionPixels detects interior intersections in the collection of
// SegmentStrings, and adds nodes for them to the segment strings. Also creates
// HotPixel nodes for the intersection points.
func (srn *NodingSnapround_SnapRoundingNoder) addIntersectionPixels(segStrings []*Noding_NodedSegmentString) {
	// Nearness tolerance is a small fraction of the grid size.
	snapGridSize := 1.0 / srn.pm.GetScale()
	nearnessTol := snapGridSize / nodingSnapround_SnapRoundingNoder_NEARNESS_FACTOR

	intAdder := NodingSnapround_NewSnapRoundingIntersectionAdder(nearnessTol)
	noder := Noding_NewMCIndexNoderWithIntersectorAndTolerance(intAdder, nearnessTol)

	// Convert to SegmentString slice for noder.
	ssSlice := make([]Noding_SegmentString, len(segStrings))
	for i, nss := range segStrings {
		ssSlice[i] = nss
	}

	noder.ComputeNodes(ssSlice)
	intPts := intAdder.GetIntersections()
	srn.pixelIndex.AddNodes(intPts)
}

// addVertexPixels creates HotPixels for each vertex in the input segStrings.
// The HotPixels are not marked as nodes, since they will only be nodes in the
// final line arrangement if they interact with other segments (or they are
// already created as intersection nodes).
func (srn *NodingSnapround_SnapRoundingNoder) addVertexPixels(segStrings []*Noding_NodedSegmentString) {
	for _, nss := range segStrings {
		pts := nss.GetCoordinates()
		srn.pixelIndex.Add(pts)
	}
}

func (srn *NodingSnapround_SnapRoundingNoder) round(pt *Geom_Coordinate) *Geom_Coordinate {
	p2 := pt.Copy()
	srn.pm.MakePreciseCoordinate(p2)
	return p2
}

// roundCoords gets a list of the rounded coordinates. Duplicate (collapsed)
// coordinates are removed.
func (srn *NodingSnapround_SnapRoundingNoder) roundCoords(pts []*Geom_Coordinate) []*Geom_Coordinate {
	roundPts := Geom_NewCoordinateList()
	for _, pt := range pts {
		roundPts.AddCoordinate(srn.round(pt), false)
	}
	return roundPts.ToCoordinateArray()
}

// computeSnaps computes new segment strings which are rounded and contain
// intersections added as a result of snapping segments to snap points (hot
// pixels).
func (srn *NodingSnapround_SnapRoundingNoder) computeSnaps(segStrings []*Noding_NodedSegmentString) []*Noding_NodedSegmentString {
	snapped := make([]*Noding_NodedSegmentString, 0)
	for _, ss := range segStrings {
		snappedSS := srn.computeSegmentSnaps(ss)
		if snappedSS != nil {
			snapped = append(snapped, snappedSS)
		}
	}
	// Some intersection hot pixels may have been marked as nodes in the
	// previous loop, so add nodes for them.
	for _, ss := range snapped {
		srn.addVertexNodeSnaps(ss)
	}
	return snapped
}

// computeSegmentSnaps adds snapped vertices to a segment string. If the
// segment string collapses completely due to rounding, nil is returned.
func (srn *NodingSnapround_SnapRoundingNoder) computeSegmentSnaps(ss *Noding_NodedSegmentString) *Noding_NodedSegmentString {
	// Get edge coordinates, including added intersection nodes. The
	// coordinates are now rounded to the grid, in preparation for snapping to
	// the Hot Pixels.
	pts := ss.GetNodedCoordinates()
	ptsRound := srn.roundCoords(pts)

	// If complete collapse this edge can be eliminated.
	if len(ptsRound) <= 1 {
		return nil
	}

	// Create new nodedSS to allow adding any hot pixel nodes.
	snapSS := Noding_NewNodedSegmentString(ptsRound, ss.GetData())

	snapSSindex := 0
	for i := 0; i < len(pts)-1; i++ {
		currSnap := snapSS.GetCoordinate(snapSSindex)

		// If the segment has collapsed completely, skip it.
		p1 := pts[i+1]
		p1Round := srn.round(p1)
		if p1Round.Equals2D(currSnap) {
			continue
		}

		p0 := pts[i]

		// Add any Hot Pixel intersections with *original* segment to rounded
		// segment. (It is important to check original segment because rounding
		// can move it enough to intersect other hot pixels not intersecting
		// original segment)
		srn.snapSegment(p0, p1, snapSS, snapSSindex)
		snapSSindex++
	}
	return snapSS
}

// snapSegment snaps a segment in a segmentString to HotPixels that it
// intersects.
func (srn *NodingSnapround_SnapRoundingNoder) snapSegment(p0, p1 *Geom_Coordinate, ss *Noding_NodedSegmentString, segIndex int) {
	srn.pixelIndex.Query(p0, p1, &nodingSnapround_snapSegmentVisitor{
		p0:       p0,
		p1:       p1,
		ss:       ss,
		segIndex: segIndex,
	})
}

type nodingSnapround_snapSegmentVisitor struct {
	p0       *Geom_Coordinate
	p1       *Geom_Coordinate
	ss       *Noding_NodedSegmentString
	segIndex int
}

func (v *nodingSnapround_snapSegmentVisitor) Visit(node *IndexKdtree_KdNode) {
	hp := node.GetData().(*NodingSnapround_HotPixel)

	// If the hot pixel is not a node, and it contains one of the segment
	// vertices, then that vertex is the source for the hot pixel. To avoid
	// over-noding a node is not added at this point. The hot pixel may be
	// subsequently marked as a node, in which case the intersection will be
	// added during the final vertex noding phase.
	if !hp.IsNode() {
		if hp.IntersectsPoint(v.p0) || hp.IntersectsPoint(v.p1) {
			return
		}
	}
	// Add a node if the segment intersects the pixel. Mark the HotPixel as a
	// node (since it may not have been one before). This ensures the vertex
	// for it is added as a node during the final vertex noding phase.
	if hp.IntersectsSegment(v.p0, v.p1) {
		v.ss.AddIntersection(hp.GetCoordinate(), v.segIndex)
		hp.SetToNode()
	}
}

// addVertexNodeSnaps adds nodes for any vertices in hot pixels that were
// added as nodes during segment noding.
func (srn *NodingSnapround_SnapRoundingNoder) addVertexNodeSnaps(ss *Noding_NodedSegmentString) {
	pts := ss.GetCoordinates()
	for i := 1; i < len(pts)-1; i++ {
		p0 := pts[i]
		srn.snapVertexNode(p0, ss, i)
	}
}

func (srn *NodingSnapround_SnapRoundingNoder) snapVertexNode(p0 *Geom_Coordinate, ss *Noding_NodedSegmentString, segIndex int) {
	srn.pixelIndex.Query(p0, p0, &nodingSnapround_vertexNodeVisitor{
		p0:       p0,
		ss:       ss,
		segIndex: segIndex,
	})
}

type nodingSnapround_vertexNodeVisitor struct {
	p0       *Geom_Coordinate
	ss       *Noding_NodedSegmentString
	segIndex int
}

func (v *nodingSnapround_vertexNodeVisitor) Visit(node *IndexKdtree_KdNode) {
	hp := node.GetData().(*NodingSnapround_HotPixel)
	// If vertex pixel is a node, add it.
	if hp.IsNode() && hp.GetCoordinate().Equals2D(v.p0) {
		v.ss.AddIntersection(v.p0, v.segIndex)
	}
}
