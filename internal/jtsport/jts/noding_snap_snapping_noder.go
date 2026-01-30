package jts

var _ Noding_Noder = (*NodingSnap_SnappingNoder)(nil)

// NodingSnap_SnappingNoder nodes a set of segment strings snapping vertices
// and intersection points together if they lie within the given snap tolerance
// distance. Vertices take priority over intersection points for snapping.
// Input segment strings are generally only split at true node points (i.e. the
// output segment strings are of maximal length in the output arrangement).
//
// The snap tolerance should be chosen to be as small as possible while still
// producing a correct result. It probably only needs to be small enough to
// eliminate "nearly-coincident" segments, for which intersection points cannot
// be computed accurately. This implies a factor of about 10e-12 smaller than
// the magnitude of the segment coordinates.
//
// With an appropriate snap tolerance this algorithm appears to be very robust.
// So far no failure cases have been found, given a small enough snap
// tolerance.
//
// The correctness of the output is not verified by this noder. If required
// this can be done by ValidatingNoder.
type NodingSnap_SnappingNoder struct {
	snapIndex     *NodingSnap_SnappingPointIndex
	snapTolerance float64
	nodedResult   []Noding_SegmentString
}

// IsNoding_Noder is a marker method for interface identification.
func (sn *NodingSnap_SnappingNoder) IsNoding_Noder() {}

// NodingSnap_NewSnappingNoder creates a snapping noder using the given snap
// distance tolerance.
func NodingSnap_NewSnappingNoder(snapTolerance float64) *NodingSnap_SnappingNoder {
	return &NodingSnap_SnappingNoder{
		snapTolerance: snapTolerance,
		snapIndex:     NodingSnap_NewSnappingPointIndex(snapTolerance),
	}
}

// GetNodedSubstrings gets the noded result.
func (sn *NodingSnap_SnappingNoder) GetNodedSubstrings() []Noding_SegmentString {
	return sn.nodedResult
}

// ComputeNodes computes the noding of a set of SegmentStrings.
func (sn *NodingSnap_SnappingNoder) ComputeNodes(inputSegStrings []Noding_SegmentString) {
	snappedSS := sn.snapVertices(inputSegStrings)
	sn.nodedResult = sn.snapIntersections(snappedSS)
}

func (sn *NodingSnap_SnappingNoder) snapVertices(segStrings []Noding_SegmentString) []*Noding_NodedSegmentString {
	sn.seedSnapIndex(segStrings)

	nodedStrings := make([]*Noding_NodedSegmentString, 0, len(segStrings))
	for _, ss := range segStrings {
		nodedStrings = append(nodedStrings, sn.snapVerticesForSS(ss))
	}
	return nodedStrings
}

// seedSnapIndex seeds the snap index with a small set of vertices chosen
// quasi-randomly using a low-discrepancy sequence. Seeding the snap index
// KdTree induces a more balanced tree. This prevents monotonic runs of
// vertices unbalancing the tree and causing poor query performance.
func (sn *NodingSnap_SnappingNoder) seedSnapIndex(segStrings []Noding_SegmentString) {
	const seedSizeFactor = 100

	for _, ss := range segStrings {
		pts := ss.GetCoordinates()
		numPtsToLoad := len(pts) / seedSizeFactor
		rand := 0.0
		for i := 0; i < numPtsToLoad; i++ {
			rand = Math_MathUtil_Quasirandom(rand)
			index := int(float64(len(pts)) * rand)
			sn.snapIndex.Snap(pts[index])
		}
	}
}

func (sn *NodingSnap_SnappingNoder) snapVerticesForSS(ss Noding_SegmentString) *Noding_NodedSegmentString {
	snapCoords := sn.snap(ss.GetCoordinates())
	return Noding_NewNodedSegmentString(snapCoords, ss.GetData())
}

func (sn *NodingSnap_SnappingNoder) snap(coords []*Geom_Coordinate) []*Geom_Coordinate {
	snapCoords := Geom_NewCoordinateList()
	for _, coord := range coords {
		pt := sn.snapIndex.Snap(coord)
		snapCoords.AddCoordinate(pt, false)
	}
	return snapCoords.ToCoordinateArray()
}

// snapIntersections computes all interior intersections in the collection of
// SegmentStrings, and returns their noded substrings. Also adds the
// intersection nodes to the segments.
func (sn *NodingSnap_SnappingNoder) snapIntersections(inputSS []*Noding_NodedSegmentString) []Noding_SegmentString {
	intAdder := NodingSnap_NewSnappingIntersectionAdder(sn.snapTolerance, sn.snapIndex)
	// Use an overlap tolerance to ensure all possible snapped intersections
	// are found.
	noder := Noding_NewMCIndexNoderWithIntersectorAndTolerance(intAdder, 2*sn.snapTolerance)

	// Convert to SegmentString slice for noder.
	ssSlice := make([]Noding_SegmentString, len(inputSS))
	for i, nss := range inputSS {
		ssSlice[i] = nss
	}

	noder.ComputeNodes(ssSlice)
	return noder.GetNodedSubstrings()
}
