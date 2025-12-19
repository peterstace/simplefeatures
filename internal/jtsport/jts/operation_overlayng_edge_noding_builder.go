package jts

// OperationOverlayng_EdgeNodingBuilder builds a set of noded, unique, labelled
// Edges from the edges of the two input geometries.
//
// It performs the following steps:
//   - Extracts input edges, and attaches topological information
//   - if clipping is enabled, handles clipping or limiting input geometry
//   - chooses a Noder based on provided precision model, unless a custom one
//     is supplied
//   - calls the chosen Noder, with precision model
//   - removes any fully collapsed noded edges
//   - builds Edges and merges them
type OperationOverlayng_EdgeNodingBuilder struct {
	pm          *Geom_PrecisionModel
	inputEdges  []*Noding_NodedSegmentString
	customNoder Noding_Noder

	clipEnv *Geom_Envelope
	clipper *OperationOverlayng_RingClipper
	limiter *OperationOverlayng_LineLimiter

	hasEdges [2]bool
}

const (
	// Limiting is skipped for Lines with few vertices, to avoid additional
	// copying.
	operationOverlayng_EdgeNodingBuilder_MIN_LIMIT_PTS = 20

	// Indicates whether floating precision noder output is validated.
	operationOverlayng_EdgeNodingBuilder_IS_NODING_VALIDATED = true
)

func operationOverlayng_EdgeNodingBuilder_createFixedPrecisionNoder(pm *Geom_PrecisionModel) Noding_Noder {
	return NodingSnapround_NewSnapRoundingNoder(pm)
}

func operationOverlayng_EdgeNodingBuilder_createFloatingPrecisionNoder(doValidation bool) Noding_Noder {
	mcNoder := Noding_NewMCIndexNoder()
	li := Algorithm_NewRobustLineIntersector()
	mcNoder.SetSegmentIntersector(Noding_NewIntersectionAdder(li.Algorithm_LineIntersector))

	var noder Noding_Noder = mcNoder
	if doValidation {
		noder = Noding_NewValidatingNoder(mcNoder)
	}
	return noder
}

// OperationOverlayng_NewEdgeNodingBuilder creates a new builder, with an
// optional custom noder. If the noder is not provided, a suitable one will be
// used based on the supplied precision model.
func OperationOverlayng_NewEdgeNodingBuilder(pm *Geom_PrecisionModel, noder Noding_Noder) *OperationOverlayng_EdgeNodingBuilder {
	return &OperationOverlayng_EdgeNodingBuilder{
		pm:          pm,
		customNoder: noder,
		inputEdges:  make([]*Noding_NodedSegmentString, 0),
	}
}

// getNoder gets a noder appropriate for the precision model supplied. This is
// one of:
//   - Fixed precision: a snap-rounding noder (which should be fully robust)
//   - Floating precision: a conventional noder (which may be non-robust). In
//     this case, a validation step is applied to the output from the noder.
func (enb *OperationOverlayng_EdgeNodingBuilder) getNoder() Noding_Noder {
	if enb.customNoder != nil {
		return enb.customNoder
	}
	if OperationOverlayng_OverlayUtil_IsFloating(enb.pm) {
		return operationOverlayng_EdgeNodingBuilder_createFloatingPrecisionNoder(operationOverlayng_EdgeNodingBuilder_IS_NODING_VALIDATED)
	}
	return operationOverlayng_EdgeNodingBuilder_createFixedPrecisionNoder(enb.pm)
}

// SetClipEnvelope sets the clip envelope for clipping or limiting input geometry.
func (enb *OperationOverlayng_EdgeNodingBuilder) SetClipEnvelope(clipEnv *Geom_Envelope) {
	enb.clipEnv = clipEnv
	enb.clipper = OperationOverlayng_NewRingClipper(clipEnv)
	enb.limiter = OperationOverlayng_NewLineLimiter(clipEnv)
}

// HasEdgesFor reports whether there are noded edges for the given input
// geometry. If there are none, this indicates that either the geometry was
// empty, or has completely collapsed (because it is smaller than the noding
// precision).
func (enb *OperationOverlayng_EdgeNodingBuilder) HasEdgesFor(geomIndex int) bool {
	return enb.hasEdges[geomIndex]
}

// Build creates a set of labelled Edges representing the fully noded edges of
// the input geometries. Coincident edges (from the same or both geometries)
// are merged along with their labels into a single unique, fully labelled edge.
func (enb *OperationOverlayng_EdgeNodingBuilder) Build(geom0, geom1 *Geom_Geometry) []*OperationOverlayng_Edge {
	enb.add(geom0, 0)
	enb.add(geom1, 1)
	nodedEdges := enb.node(enb.inputEdges)

	// Merge the noded edges to eliminate duplicates. Labels are combined.
	mergedEdges := OperationOverlayng_EdgeMerger_Merge(nodedEdges)
	return mergedEdges
}

// node nodes a set of segment strings and creates Edges from the result. The
// input segment strings each carry an EdgeSourceInfo object, which is used to
// provide source topology info to the constructed Edges (and then is discarded).
func (enb *OperationOverlayng_EdgeNodingBuilder) node(segStrings []*Noding_NodedSegmentString) []*OperationOverlayng_Edge {
	noder := enb.getNoder()
	noder.ComputeNodes(Noding_NodedSegmentStringsToSegmentStrings(segStrings))

	nodedSS := noder.GetNodedSubstrings()
	edges := enb.createEdges(nodedSS)
	return edges
}

func (enb *OperationOverlayng_EdgeNodingBuilder) createEdges(segStrings []Noding_SegmentString) []*OperationOverlayng_Edge {
	edges := make([]*OperationOverlayng_Edge, 0)
	for _, ss := range segStrings {
		pts := ss.GetCoordinates()

		// don't create edges from collapsed lines
		if OperationOverlayng_Edge_IsCollapsed(pts) {
			continue
		}

		info := ss.GetData().(*OperationOverlayng_EdgeSourceInfo)
		// Record that a non-collapsed edge exists for the parent geometry
		enb.hasEdges[info.GetIndex()] = true
		edges = append(edges, OperationOverlayng_NewEdge(ss.GetCoordinates(), info))
	}
	return edges
}

func (enb *OperationOverlayng_EdgeNodingBuilder) add(g *Geom_Geometry, geomIndex int) {
	if g == nil || g.IsEmpty() {
		return
	}

	if enb.isClippedCompletely(g.GetEnvelopeInternal()) {
		return
	}

	if polygon, ok := g.GetChild().(*Geom_Polygon); ok {
		enb.addPolygon(polygon, geomIndex)
	} else if lineString, ok := g.GetChild().(*Geom_LineString); ok {
		// LineString also handles LinearRings
		enb.addLine(lineString, geomIndex)
	} else if mls, ok := g.GetChild().(*Geom_MultiLineString); ok {
		enb.addCollection(mls.Geom_GeometryCollection, geomIndex)
	} else if mp, ok := g.GetChild().(*Geom_MultiPolygon); ok {
		enb.addCollection(mp.Geom_GeometryCollection, geomIndex)
	} else if gc, ok := g.GetChild().(*Geom_GeometryCollection); ok {
		enb.addGeometryCollection(gc, geomIndex, g.GetDimension())
	}
	// ignore Point geometries - they are handled elsewhere
}

func (enb *OperationOverlayng_EdgeNodingBuilder) addCollection(gc *Geom_GeometryCollection, geomIndex int) {
	for i := 0; i < gc.GetNumGeometries(); i++ {
		g := gc.GetGeometryN(i)
		enb.add(g, geomIndex)
	}
}

func (enb *OperationOverlayng_EdgeNodingBuilder) addGeometryCollection(gc *Geom_GeometryCollection, geomIndex, expectedDim int) {
	for i := 0; i < gc.GetNumGeometries(); i++ {
		g := gc.GetGeometryN(i)
		// check for mixed-dimension input, which is not supported
		if g.GetDimension() != expectedDim {
			panic("Overlay input is mixed-dimension")
		}
		enb.add(g, geomIndex)
	}
}

func (enb *OperationOverlayng_EdgeNodingBuilder) addPolygon(poly *Geom_Polygon, geomIndex int) {
	shell := poly.GetExteriorRing()
	enb.addPolygonRing(shell, false, geomIndex)

	for i := 0; i < poly.GetNumInteriorRing(); i++ {
		hole := poly.GetInteriorRingN(i)
		// Holes are topologically labelled opposite to the shell, since the
		// interior of the polygon lies on their opposite side (on the left, if
		// the hole is oriented CW)
		enb.addPolygonRing(hole, true, geomIndex)
	}
}

// addPolygonRing adds a polygon ring to the graph. Empty rings are ignored.
func (enb *OperationOverlayng_EdgeNodingBuilder) addPolygonRing(ring *Geom_LinearRing, isHole bool, index int) {
	// don't add empty rings
	if ring.IsEmpty() {
		return
	}

	if enb.isClippedCompletely(ring.GetEnvelopeInternal()) {
		return
	}

	pts := enb.clip(ring)

	// Don't add edges that collapse to a point
	if len(pts) < 2 {
		return
	}

	depthDelta := operationOverlayng_EdgeNodingBuilder_computeDepthDelta(ring, isHole)
	info := OperationOverlayng_NewEdgeSourceInfoForArea(index, depthDelta, isHole)
	enb.addEdge(pts, info)
}

// isClippedCompletely tests whether a geometry (represented by its envelope)
// lies completely outside the clip extent (if any).
func (enb *OperationOverlayng_EdgeNodingBuilder) isClippedCompletely(env *Geom_Envelope) bool {
	if enb.clipEnv == nil {
		return false
	}
	return enb.clipEnv.Disjoint(env)
}

// clip clips the line to the clip extent if a clipper is present, otherwise
// removes duplicate points from the ring.
func (enb *OperationOverlayng_EdgeNodingBuilder) clip(ring *Geom_LinearRing) []*Geom_Coordinate {
	pts := ring.GetCoordinates()
	env := ring.GetEnvelopeInternal()

	// If no clipper or ring is completely contained then no need to clip. But
	// repeated points must be removed to ensure correct noding.
	if enb.clipper == nil || enb.clipEnv.CoversEnvelope(env) {
		return operationOverlayng_EdgeNodingBuilder_removeRepeatedPoints(ring.Geom_LineString)
	}

	return enb.clipper.Clip(pts)
}

// removeRepeatedPoints removes any repeated points from a linear component.
// This is required so that noding can be computed correctly.
func operationOverlayng_EdgeNodingBuilder_removeRepeatedPoints(line *Geom_LineString) []*Geom_Coordinate {
	pts := line.GetCoordinates()
	return Geom_CoordinateArrays_RemoveRepeatedPoints(pts)
}

func operationOverlayng_EdgeNodingBuilder_computeDepthDelta(ring *Geom_LinearRing, isHole bool) int {
	// Compute the orientation of the ring, to allow assigning side
	// interior/exterior labels correctly. JTS canonical orientation is that
	// shells are CW, holes are CCW.
	//
	// It is important to compute orientation on the original ring, since
	// topology collapse can make the orientation computation give the wrong
	// answer.
	isCCW := Algorithm_Orientation_IsCCWSeq(ring.GetCoordinateSequence())

	// Compute whether ring is in canonical orientation or not. Canonical
	// orientation for the overlay process is Shells : CW, Holes: CCW
	isOriented := true
	if !isHole {
		isOriented = !isCCW
	} else {
		isOriented = isCCW
	}

	// Depth delta can now be computed. Canonical depth delta is 1 (Exterior on
	// L, Interior on R). It is flipped to -1 if the ring is oppositely oriented.
	depthDelta := 1
	if !isOriented {
		depthDelta = -1
	}
	return depthDelta
}

// addLine adds a line geometry, limiting it if enabled, and otherwise
// removing repeated points.
func (enb *OperationOverlayng_EdgeNodingBuilder) addLine(line *Geom_LineString, geomIndex int) {
	// don't add empty lines
	if line.IsEmpty() {
		return
	}

	if enb.isClippedCompletely(line.GetEnvelopeInternal()) {
		return
	}

	if enb.isToBeLimited(line) {
		sections := enb.limit(line)
		for _, pts := range sections {
			enb.addLineCoords(pts, geomIndex)
		}
	} else {
		ptsNoRepeat := operationOverlayng_EdgeNodingBuilder_removeRepeatedPoints(line)
		enb.addLineCoords(ptsNoRepeat, geomIndex)
	}
}

func (enb *OperationOverlayng_EdgeNodingBuilder) addLineCoords(pts []*Geom_Coordinate, geomIndex int) {
	// Don't add edges that collapse to a point
	if len(pts) < 2 {
		return
	}

	info := OperationOverlayng_NewEdgeSourceInfoForLine(geomIndex)
	enb.addEdge(pts, info)
}

func (enb *OperationOverlayng_EdgeNodingBuilder) addEdge(pts []*Geom_Coordinate, info *OperationOverlayng_EdgeSourceInfo) {
	ss := Noding_NewNodedSegmentString(pts, info)
	enb.inputEdges = append(enb.inputEdges, ss)
}

// isToBeLimited tests whether it is worth limiting a line. Lines that have
// few vertices or are covered by the clip extent do not need to be limited.
func (enb *OperationOverlayng_EdgeNodingBuilder) isToBeLimited(line *Geom_LineString) bool {
	pts := line.GetCoordinates()
	if enb.limiter == nil || len(pts) <= operationOverlayng_EdgeNodingBuilder_MIN_LIMIT_PTS {
		return false
	}
	env := line.GetEnvelopeInternal()
	// If line is completely contained then no need to limit
	if enb.clipEnv.CoversEnvelope(env) {
		return false
	}
	return true
}

// limit limits the line to the clip envelope if a limiter is provided.
func (enb *OperationOverlayng_EdgeNodingBuilder) limit(line *Geom_LineString) [][]*Geom_Coordinate {
	pts := line.GetCoordinates()
	return enb.limiter.Limit(pts)
}
