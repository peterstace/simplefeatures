package jts

// NodingSnapround_MCIndexSnapRounder uses Snap Rounding to compute a rounded,
// fully noded arrangement from a set of SegmentStrings. Implements the Snap
// Rounding technique described in papers by Hobby, Guibas & Marimont, and
// Goodrich et al. Snap Rounding assumes that all vertices lie on a uniform
// grid; hence the precision model of the input must be fixed precision, and
// all the input vertices must be rounded to that precision.
//
// This implementation uses monotone chains and a spatial index to speed up the
// intersection tests.
//
// KNOWN BUGS: This implementation is not fully robust.
//
// Deprecated: Not robust. Use SnapRoundingNoder instead.
type NodingSnapround_MCIndexSnapRounder struct {
	pm              *Geom_PrecisionModel
	li              *Algorithm_LineIntersector
	scaleFactor     float64
	noder           *Noding_MCIndexNoder
	pointSnapper    *NodingSnapround_MCIndexPointSnapper
	nodedSegStrings []Noding_SegmentString
}

var _ Noding_Noder = (*NodingSnapround_MCIndexSnapRounder)(nil)

func (sr *NodingSnapround_MCIndexSnapRounder) IsNoding_Noder() {}

// NodingSnapround_NewMCIndexSnapRounder creates a new MCIndexSnapRounder with
// the given precision model.
func NodingSnapround_NewMCIndexSnapRounder(pm *Geom_PrecisionModel) *NodingSnapround_MCIndexSnapRounder {
	rli := Algorithm_NewRobustLineIntersector()
	rli.SetPrecisionModel(pm)
	return &NodingSnapround_MCIndexSnapRounder{
		pm:          pm,
		li:          rli.Algorithm_LineIntersector,
		scaleFactor: pm.GetScale(),
	}
}

// GetNodedSubstrings returns the noded substrings.
func (sr *NodingSnapround_MCIndexSnapRounder) GetNodedSubstrings() []Noding_SegmentString {
	// Convert to NodedSegmentString slice.
	nssSlice := make([]*Noding_NodedSegmentString, len(sr.nodedSegStrings))
	for i, ss := range sr.nodedSegStrings {
		nssSlice[i] = ss.(*Noding_NodedSegmentString)
	}
	nodedResult := Noding_NodedSegmentString_GetNodedSubstrings(nssSlice)
	result := make([]Noding_SegmentString, len(nodedResult))
	for i, nss := range nodedResult {
		result[i] = nss
	}
	return result
}

// ComputeNodes computes the noding.
func (sr *NodingSnapround_MCIndexSnapRounder) ComputeNodes(inputSegmentStrings []Noding_SegmentString) {
	sr.nodedSegStrings = inputSegmentStrings
	sr.noder = Noding_NewMCIndexNoder()
	sr.pointSnapper = NodingSnapround_NewMCIndexPointSnapper(sr.noder.GetIndex())
	sr.snapRound(inputSegmentStrings, sr.li)
}

func (sr *NodingSnapround_MCIndexSnapRounder) snapRound(segStrings []Noding_SegmentString, li *Algorithm_LineIntersector) {
	intersections := sr.findInteriorIntersections(segStrings, li)
	sr.computeIntersectionSnaps(intersections)
	sr.computeVertexSnaps(segStrings)
}

// findInteriorIntersections computes all interior intersections in the
// collection of SegmentStrings, and returns their Coordinates. Does NOT node
// the segStrings.
func (sr *NodingSnapround_MCIndexSnapRounder) findInteriorIntersections(segStrings []Noding_SegmentString, li *Algorithm_LineIntersector) []*Geom_Coordinate {
	intFinderAdder := Noding_NewInteriorIntersectionFinderAdder(li)
	sr.noder.SetSegmentIntersector(intFinderAdder)
	sr.noder.ComputeNodes(segStrings)
	return intFinderAdder.GetInteriorIntersections()
}

// computeIntersectionSnaps snaps segments to nodes created by segment
// intersections.
func (sr *NodingSnapround_MCIndexSnapRounder) computeIntersectionSnaps(snapPts []*Geom_Coordinate) {
	for _, snapPt := range snapPts {
		hotPixel := NodingSnapround_NewHotPixel(snapPt, sr.scaleFactor)
		sr.pointSnapper.SnapSimple(hotPixel)
	}
}

// computeVertexSnaps snaps segments to all vertices.
func (sr *NodingSnapround_MCIndexSnapRounder) computeVertexSnaps(edges []Noding_SegmentString) {
	for _, edge := range edges {
		nss := edge.(*Noding_NodedSegmentString)
		sr.computeVertexSnapsForEdge(nss)
	}
}

// computeVertexSnapsForEdge snaps segments to the vertices of a Segment
// String.
func (sr *NodingSnapround_MCIndexSnapRounder) computeVertexSnapsForEdge(e *Noding_NodedSegmentString) {
	pts0 := e.GetCoordinates()
	for i := 0; i < len(pts0); i++ {
		hotPixel := NodingSnapround_NewHotPixel(pts0[i], sr.scaleFactor)
		isNodeAdded := sr.pointSnapper.Snap(hotPixel, e, i)
		// If a node is created for a vertex, that vertex must be noded too.
		if isNodeAdded {
			e.AddIntersection(pts0[i], i)
		}
	}
}
