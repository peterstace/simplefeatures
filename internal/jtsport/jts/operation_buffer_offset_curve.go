package jts

import (
	"math"

	"github.com/peterstace/simplefeatures/internal/jtsport/java"
)

// operationBuffer_offsetCurve_MATCH_DISTANCE_FACTOR is the nearness tolerance
// for matching the raw offset linework and the buffer curve.
const operationBuffer_offsetCurve_MATCH_DISTANCE_FACTOR = 10000

// operationBuffer_offsetCurve_MIN_QUADRANT_SEGMENTS is a QuadSegs minimum value
// that will prevent generating unwanted offset curve artifacts near end caps.
const operationBuffer_offsetCurve_MIN_QUADRANT_SEGMENTS = 8

// OperationBuffer_OffsetCurve computes an offset curve from a geometry.
// An offset curve is a linear geometry which is offset a given distance
// from the input.
// If the offset distance is positive the curve lies on the left side of the input;
// if it is negative the curve is on the right side.
// The curve(s) have the same direction as the input line(s).
// The result for a zero offset distance is a copy of the input linework.
//
// The offset curve is based on the boundary of the buffer for the geometry
// at the offset distance (see BufferOp).
// The normal mode of operation is to return the sections of the buffer boundary
// which lie on the raw offset curve
// (obtained via RawOffset).
// The offset curve will contain multiple sections
// if the input self-intersects or has close approaches.
// The computed sections are ordered along the raw offset curve.
// Sections are disjoint. They never self-intersect, but may be rings.
//
// For a LineString the offset curve is a linear geometry
// (LineString or MultiLineString).
// For a Point or MultiPoint the offset curve is an empty LineString.
// For a Polygon the offset curve is the boundary of the polygon buffer (which
// may be a MultiLineString).
// For a collection the output is a MultiLineString containing the offset curves of the elements.
//
// In "joined" mode (see SetJoined)
// the sections computed for each input line are joined into a single offset curve line.
// The joined curve may self-intersect.
// At larger offset distances the curve may contain "flat-line" artifacts
// in places where the input self-intersects.
//
// Offset curves support setting the number of quadrant segments,
// the join style, and the mitre limit (if applicable) via
// the BufferParameters.
type OperationBuffer_OffsetCurve struct {
	inputGeom     *Geom_Geometry
	distance      float64
	isJoined      bool
	bufferParams  *OperationBuffer_BufferParameters
	matchDistance float64
	geomFactory   *Geom_GeometryFactory
}

// OperationBuffer_OffsetCurve_GetCurve computes the offset curve of a geometry at a given distance.
func OperationBuffer_OffsetCurve_GetCurve(geom *Geom_Geometry, distance float64) *Geom_Geometry {
	oc := OperationBuffer_NewOffsetCurve(geom, distance)
	return oc.GetCurve()
}

// OperationBuffer_OffsetCurve_GetCurveWithParams computes the offset curve of a geometry at a given distance,
// with specified quadrant segments, join style and mitre limit.
func OperationBuffer_OffsetCurve_GetCurveWithParams(geom *Geom_Geometry, distance float64, quadSegs, joinStyle int, mitreLimit float64) *Geom_Geometry {
	bufferParams := OperationBuffer_NewBufferParameters()
	if quadSegs >= 0 {
		bufferParams.SetQuadrantSegments(quadSegs)
	}
	if joinStyle >= 0 {
		bufferParams.SetJoinStyle(joinStyle)
	}
	if mitreLimit >= 0 {
		bufferParams.SetMitreLimit(mitreLimit)
	}
	oc := OperationBuffer_NewOffsetCurveWithParams(geom, distance, bufferParams)
	return oc.GetCurve()
}

// OperationBuffer_OffsetCurve_GetCurveJoined computes the offset curve of a geometry at a given distance,
// joining curve sections into a single line for each input line.
func OperationBuffer_OffsetCurve_GetCurveJoined(geom *Geom_Geometry, distance float64) *Geom_Geometry {
	oc := OperationBuffer_NewOffsetCurve(geom, distance)
	oc.SetJoined(true)
	return oc.GetCurve()
}

// OperationBuffer_NewOffsetCurve creates a new instance for computing an offset curve
// for a geometry at a given distance with default quadrant segments
// (BufferParameters.DEFAULT_QUADRANT_SEGMENTS) and join style (BufferParameters.JOIN_STYLE).
func OperationBuffer_NewOffsetCurve(geom *Geom_Geometry, distance float64) *OperationBuffer_OffsetCurve {
	return OperationBuffer_NewOffsetCurveWithParams(geom, distance, nil)
}

// OperationBuffer_NewOffsetCurveWithParams creates a new instance for computing an offset curve
// for a geometry at a given distance, setting the quadrant segments and join style
// and mitre limit via BufferParameters.
func OperationBuffer_NewOffsetCurveWithParams(geom *Geom_Geometry, distance float64, bufParams *OperationBuffer_BufferParameters) *OperationBuffer_OffsetCurve {
	oc := &OperationBuffer_OffsetCurve{
		inputGeom:     geom,
		distance:      distance,
		matchDistance: math.Abs(distance) / operationBuffer_offsetCurve_MATCH_DISTANCE_FACTOR,
		geomFactory:   geom.GetFactory(),
	}

	//-- make new buffer params since the end cap style must be the default
	oc.bufferParams = OperationBuffer_NewBufferParameters()
	if bufParams != nil {
		// Prevent using a very small QuadSegs value, to avoid
		// offset curve artifacts near the end caps.
		quadSegs := bufParams.GetQuadrantSegments()
		if quadSegs < operationBuffer_offsetCurve_MIN_QUADRANT_SEGMENTS {
			quadSegs = operationBuffer_offsetCurve_MIN_QUADRANT_SEGMENTS
		}
		oc.bufferParams.SetQuadrantSegments(quadSegs)
		oc.bufferParams.SetJoinStyle(bufParams.GetJoinStyle())
		oc.bufferParams.SetMitreLimit(bufParams.GetMitreLimit())
	}

	return oc
}

// SetJoined computes a single curve line for each input linear component,
// by joining curve sections in order along the raw offset curve.
// The default mode is to compute separate curve sections.
func (oc *OperationBuffer_OffsetCurve) SetJoined(isJoined bool) {
	oc.isJoined = isJoined
}

// GetCurve gets the computed offset curve lines.
func (oc *OperationBuffer_OffsetCurve) GetCurve() *Geom_Geometry {
	return GeomUtil_GeometryMapper_FlatMap(oc.inputGeom, 1, &operationBuffer_offsetCurveMapOp{oc: oc})
}

type operationBuffer_offsetCurveMapOp struct {
	oc *OperationBuffer_OffsetCurve
}

func (op *operationBuffer_offsetCurveMapOp) Map(geom *Geom_Geometry) *Geom_Geometry {
	if java.InstanceOf[*Geom_Point](geom) {
		return nil
	}
	if java.InstanceOf[*Geom_Polygon](geom) {
		return op.oc.toLineString(geom.Buffer(op.oc.distance).GetBoundary())
	}
	return op.oc.computeCurve(java.Cast[*Geom_LineString](geom), op.oc.distance)
}

// toLineString forces LinearRings to be LineStrings.
func (oc *OperationBuffer_OffsetCurve) toLineString(geom *Geom_Geometry) *Geom_Geometry {
	if java.InstanceOf[*Geom_LinearRing](geom) {
		ring := java.Cast[*Geom_LinearRing](geom)
		return geom.GetFactory().CreateLineStringFromCoordinateSequence(ring.GetCoordinateSequence()).Geom_Geometry
	}
	return geom
}

// OperationBuffer_OffsetCurve_RawOffsetWithParams gets the raw offset curve for a line at a given distance.
// The quadrant segments, join style and mitre limit can be specified
// via BufferParameters.
//
// The raw offset line may contain loops and other artifacts which are
// not present in the true offset curve.
func OperationBuffer_OffsetCurve_RawOffsetWithParams(line *Geom_LineString, distance float64, bufParams *OperationBuffer_BufferParameters) []*Geom_Coordinate {
	pts := line.GetCoordinates()
	cleanPts := Geom_CoordinateArrays_RemoveRepeatedOrInvalidPoints(pts)
	ocb := OperationBuffer_NewOffsetCurveBuilder(
		line.GetFactory().GetPrecisionModel(), bufParams,
	)
	rawPts := ocb.GetOffsetCurve(cleanPts, distance)
	return rawPts
}

// OperationBuffer_OffsetCurve_RawOffset gets the raw offset curve for a line at a given distance,
// with default buffer parameters.
func OperationBuffer_OffsetCurve_RawOffset(line *Geom_LineString, distance float64) []*Geom_Coordinate {
	return OperationBuffer_OffsetCurve_RawOffsetWithParams(line, distance, OperationBuffer_NewBufferParameters())
}

func (oc *OperationBuffer_OffsetCurve) computeCurve(lineGeom *Geom_LineString, distance float64) *Geom_Geometry {
	//-- first handle simple cases
	//-- empty or single-point line
	if lineGeom.GetNumPoints() < 2 || lineGeom.GetLength() == 0.0 {
		return oc.geomFactory.CreateLineString().Geom_Geometry
	}
	//-- zero offset distance
	if distance == 0 {
		return lineGeom.Copy()
	}
	//-- two-point line
	if lineGeom.GetNumPoints() == 2 {
		return oc.offsetSegment(lineGeom.GetCoordinates(), distance)
	}

	sections := oc.computeSections(lineGeom, distance)

	var offsetCurve *Geom_Geometry
	if oc.isJoined {
		offsetCurve = OperationBuffer_OffsetCurveSection_ToLine(sections, oc.geomFactory)
	} else {
		offsetCurve = OperationBuffer_OffsetCurveSection_ToGeometry(sections, oc.geomFactory)
	}
	return offsetCurve
}

func (oc *OperationBuffer_OffsetCurve) computeSections(lineGeom *Geom_LineString, distance float64) []*OperationBuffer_OffsetCurveSection {
	rawCurve := OperationBuffer_OffsetCurve_RawOffsetWithParams(lineGeom, distance, oc.bufferParams)
	sections := make([]*OperationBuffer_OffsetCurveSection, 0)
	if len(rawCurve) == 0 {
		return sections
	}

	// Note: If the raw offset curve has no
	// narrow concave angles or self-intersections it could be returned as is.
	// However, this is likely to be a less frequent situation,
	// and testing indicates little performance advantage,
	// so not doing this.

	bufferPoly := operationBuffer_offsetCurve_getBufferOriented(lineGeom, distance, oc.bufferParams)

	//-- first extract offset curve sections from shell
	shell := bufferPoly.GetExteriorRing().GetCoordinates()
	oc.computeCurveSections(shell, rawCurve, &sections)

	//-- extract offset curve sections from holes
	for i := 0; i < bufferPoly.GetNumInteriorRing(); i++ {
		hole := bufferPoly.GetInteriorRingN(i).GetCoordinates()
		oc.computeCurveSections(hole, rawCurve, &sections)
	}
	return sections
}

func (oc *OperationBuffer_OffsetCurve) offsetSegment(pts []*Geom_Coordinate, distance float64) *Geom_Geometry {
	offsetSeg := Geom_NewLineSegmentFromCoordinates(pts[0], pts[1]).Offset(distance)
	return oc.geomFactory.CreateLineStringFromCoordinates([]*Geom_Coordinate{offsetSeg.P0, offsetSeg.P1}).Geom_Geometry
}

func operationBuffer_offsetCurve_getBufferOriented(geom *Geom_LineString, distance float64, bufParams *OperationBuffer_BufferParameters) *Geom_Polygon {
	buffer := OperationBuffer_BufferOp_BufferOpWithParams(geom.Geom_Geometry, math.Abs(distance), bufParams)
	bufferPoly := operationBuffer_offsetCurve_extractMaxAreaPolygon(buffer)
	//-- for negative distances (Right of input) reverse buffer direction to match offset curve
	if distance < 0 {
		bufferPoly = bufferPoly.ReverseInternal()
	}
	return bufferPoly
}

// extractMaxAreaPolygon extracts the largest polygon by area from a geometry.
// Used here to avoid issues with non-robust buffer results
// which have spurious extra polygons.
func operationBuffer_offsetCurve_extractMaxAreaPolygon(geom *Geom_Geometry) *Geom_Polygon {
	if geom.GetNumGeometries() == 1 {
		return java.Cast[*Geom_Polygon](geom)
	}

	maxArea := 0.0
	var maxPoly *Geom_Polygon
	for i := 0; i < geom.GetNumGeometries(); i++ {
		poly := java.Cast[*Geom_Polygon](geom.GetGeometryN(i))
		area := poly.GetArea()
		if maxPoly == nil || area > maxArea {
			maxPoly = poly
			maxArea = area
		}
	}
	return maxPoly
}

const operationBuffer_offsetCurve_NOT_IN_CURVE = -1.0

func (oc *OperationBuffer_OffsetCurve) computeCurveSections(bufferRingPts []*Geom_Coordinate,
	rawCurve []*Geom_Coordinate, sections *[]*OperationBuffer_OffsetCurveSection) {
	rawPosition := make([]float64, len(bufferRingPts)-1)
	for i := 0; i < len(rawPosition); i++ {
		rawPosition[i] = operationBuffer_offsetCurve_NOT_IN_CURVE
	}
	bufferSegIndex := operationBuffer_newSegmentMCIndex(bufferRingPts)
	bufferFirstIndex := -1
	minRawPosition := -1.0
	for i := 0; i < len(rawCurve)-1; i++ {
		minBufferIndexForSeg := oc.matchSegments(
			rawCurve[i], rawCurve[i+1], i, bufferSegIndex, bufferRingPts, rawPosition)
		if minBufferIndexForSeg >= 0 {
			pos := rawPosition[minBufferIndexForSeg]
			if bufferFirstIndex < 0 || pos < minRawPosition {
				minRawPosition = pos
				bufferFirstIndex = minBufferIndexForSeg
			}
		}
	}
	//-- no matching sections found in this buffer ring
	if bufferFirstIndex < 0 {
		return
	}
	oc.extractSections(bufferRingPts, rawPosition, bufferFirstIndex, sections)
}

// matchSegments matches the segments in a buffer ring to the raw offset curve
// to obtain their match positions (if any).
func (oc *OperationBuffer_OffsetCurve) matchSegments(raw0, raw1 *Geom_Coordinate, rawCurveIndex int,
	bufferSegIndex *operationBuffer_SegmentMCIndex, bufferPts []*Geom_Coordinate,
	rawCurvePos []float64) int {
	matchEnv := Geom_NewEnvelopeFromCoordinates(raw0, raw1)
	matchEnv.ExpandBy(oc.matchDistance)
	matchAction := operationBuffer_newMatchCurveSegmentAction(raw0, raw1, rawCurveIndex, oc.matchDistance, bufferPts, rawCurvePos)
	bufferSegIndex.Query(matchEnv, matchAction.IndexChain_MonotoneChainSelectAction)
	return matchAction.GetBufferMinIndex()
}

// operationBuffer_MatchCurveSegmentAction is an action to match a raw offset curve segment
// to segments in a buffer ring and record the matched segment locations(s) along the raw curve.
type operationBuffer_MatchCurveSegmentAction struct {
	*IndexChain_MonotoneChainSelectAction
	child            java.Polymorphic
	raw0             *Geom_Coordinate
	raw1             *Geom_Coordinate
	rawLen           float64
	rawCurveIndex    int
	bufferRingPts    []*Geom_Coordinate
	matchDistance    float64
	rawCurveLoc      []float64
	minRawLocation   float64
	bufferRingMinIdx int
}

func (ma *operationBuffer_MatchCurveSegmentAction) GetChild() java.Polymorphic {
	return ma.child
}

func (ma *operationBuffer_MatchCurveSegmentAction) GetParent() java.Polymorphic {
	return ma.IndexChain_MonotoneChainSelectAction
}

func operationBuffer_newMatchCurveSegmentAction(raw0, raw1 *Geom_Coordinate, rawCurveIndex int,
	matchDistance float64, bufferRingPts []*Geom_Coordinate, rawCurveLoc []float64) *operationBuffer_MatchCurveSegmentAction {
	base := IndexChain_NewMonotoneChainSelectAction()
	action := &operationBuffer_MatchCurveSegmentAction{
		IndexChain_MonotoneChainSelectAction: base,
		raw0:                                 raw0,
		raw1:                                 raw1,
		rawLen:                               raw0.Distance(raw1),
		rawCurveIndex:                        rawCurveIndex,
		bufferRingPts:                        bufferRingPts,
		matchDistance:                        matchDistance,
		rawCurveLoc:                          rawCurveLoc,
		minRawLocation:                       -1,
		bufferRingMinIdx:                     -1,
	}
	base.child = action
	return action
}

func (ma *operationBuffer_MatchCurveSegmentAction) GetBufferMinIndex() int {
	return ma.bufferRingMinIdx
}

func (ma *operationBuffer_MatchCurveSegmentAction) Select_BODY(mc *IndexChain_MonotoneChain, segIndex int) {
	// Generally buffer segments are no longer than raw curve segments,
	// since the final buffer line likely has node points added.
	// So a buffer segment may match all or only a portion of a single raw segment.
	// There may be multiple buffer ring segs that match along the raw segment.
	//
	// HOWEVER, in some cases the buffer construction may contain
	// a matching buffer segment which is slightly longer than a raw curve segment.
	// Specifically, at the endpoint of a closed line with nearly parallel end segments
	// - the closing fillet line is very short so is heuristically removed in the buffer.
	// In this case, the buffer segment must still be matched.
	// This produces closed offset curves, which is technically
	// an anomaly, but only happens in rare cases.
	frac := ma.segmentMatchFrac(ma.bufferRingPts[segIndex], ma.bufferRingPts[segIndex+1],
		ma.raw0, ma.raw1, ma.matchDistance)
	//-- no match
	if frac < 0 {
		return
	}

	//-- location is used to sort segments along raw curve
	location := float64(ma.rawCurveIndex) + frac
	ma.rawCurveLoc[segIndex] = location
	//-- buffer seg index at lowest raw location is the curve start
	if ma.minRawLocation < 0 || location < ma.minRawLocation {
		ma.minRawLocation = location
		ma.bufferRingMinIdx = segIndex
	}
}

func (ma *operationBuffer_MatchCurveSegmentAction) segmentMatchFrac(buf0, buf1, raw0, raw1 *Geom_Coordinate, matchDistance float64) float64 {
	if !ma.isMatch(buf0, buf1, raw0, raw1, matchDistance) {
		return -1
	}

	//-- matched - determine location as fraction along raw segment
	seg := Geom_NewLineSegmentFromCoordinates(raw0, raw1)
	return seg.SegmentFraction(buf0)
}

func (ma *operationBuffer_MatchCurveSegmentAction) isMatch(buf0, buf1, raw0, raw1 *Geom_Coordinate, matchDistance float64) bool {
	bufSegLen := buf0.Distance(buf1)
	if ma.rawLen <= bufSegLen {
		if matchDistance < Algorithm_Distance_PointToSegment(raw0, buf0, buf1) {
			return false
		}
		if matchDistance < Algorithm_Distance_PointToSegment(raw1, buf0, buf1) {
			return false
		}
	} else {
		//TODO: only match longer buf segs at raw curve end segs?
		if matchDistance < Algorithm_Distance_PointToSegment(buf0, raw0, raw1) {
			return false
		}
		if matchDistance < Algorithm_Distance_PointToSegment(buf1, raw0, raw1) {
			return false
		}
	}
	return true
}

// extractSections is only called when there is at least one ring segment matched
// (so rawCurvePos has at least one entry != NOT_IN_CURVE).
// The start index of the first section must be provided.
// This is intended to be the section with lowest position
// along the raw curve.
func (oc *OperationBuffer_OffsetCurve) extractSections(ringPts []*Geom_Coordinate, rawCurveLoc []float64,
	startIndex int, sections *[]*OperationBuffer_OffsetCurveSection) {
	sectionStart := startIndex
	sectionCount := 0
	var sectionEnd int
	for {
		sectionEnd = oc.findSectionEnd(rawCurveLoc, sectionStart, startIndex)
		location := rawCurveLoc[sectionStart]
		lastIndex := operationBuffer_offsetCurve_prev(sectionEnd, len(rawCurveLoc))
		lastLoc := rawCurveLoc[lastIndex]
		section := OperationBuffer_OffsetCurveSection_Create(ringPts, sectionStart, sectionEnd, location, lastLoc)
		*sections = append(*sections, section)
		sectionStart = oc.findSectionStart(rawCurveLoc, sectionEnd)

		//-- check for an abnormal state
		sectionCount++
		if sectionCount > len(ringPts) {
			Util_Assert_ShouldNeverReachHereWithMessage("Too many sections for ring - probable bug")
		}
		if !(sectionStart != startIndex && sectionEnd != startIndex) {
			break
		}
	}
}

func (oc *OperationBuffer_OffsetCurve) findSectionStart(loc []float64, end int) int {
	start := end
	for {
		next := operationBuffer_offsetCurve_next(start, len(loc))
		//-- skip ahead if segment is not in raw curve
		if loc[start] == operationBuffer_offsetCurve_NOT_IN_CURVE {
			start = next
			continue
		}
		prev := operationBuffer_offsetCurve_prev(start, len(loc))
		//-- if prev segment is not in raw curve then have found a start
		if loc[prev] == operationBuffer_offsetCurve_NOT_IN_CURVE {
			return start
		}
		if oc.isJoined {
			// Start section at next gap in raw curve.
			// Only needed for joined curve, since otherwise
			// contiguous buffer segments can be in same curve section.
			locDelta := math.Abs(loc[start] - loc[prev])
			if locDelta > 1 {
				return start
			}
		}
		start = next
		if start == end {
			break
		}
	}
	return start
}

func (oc *OperationBuffer_OffsetCurve) findSectionEnd(loc []float64, start, firstStartIndex int) int {
	// assert: pos[start] is IN CURVE
	end := start
	var next int
	for {
		next = operationBuffer_offsetCurve_next(end, len(loc))
		if loc[next] == operationBuffer_offsetCurve_NOT_IN_CURVE {
			return next
		}
		if oc.isJoined {
			// End section at gap in raw curve.
			// Only needed for joined curve, since otherwise
			// contiguous buffer segments can be in same section
			locDelta := math.Abs(loc[next] - loc[end])
			if locDelta > 1 {
				return next
			}
		}
		end = next
		if !(end != start && end != firstStartIndex) {
			break
		}
	}
	return end
}

func operationBuffer_offsetCurve_next(i, size int) int {
	i += 1
	if i < size {
		return i
	}
	return 0
}

func operationBuffer_offsetCurve_prev(i, size int) int {
	i -= 1
	if i < 0 {
		return size - 1
	}
	return i
}
