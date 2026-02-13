package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// Noding_NodingIntersectionFinder finds non-noded intersections in a set of
// SegmentStrings, if any exist.
//
// Non-noded intersections include:
//   - Interior intersections which lie in the interior of a segment
//     (with another segment interior or with a vertex or endpoint)
//   - Vertex intersections which occur at vertices in the interior of
//     SegmentStrings (with a segment string endpoint or with another interior
//     vertex)
//
// The finder can be limited to finding only interior intersections by setting
// SetInteriorIntersectionsOnly.
//
// By default only the first intersection is found, but all can be found by
// setting SetFindAllIntersections.
type Noding_NodingIntersectionFinder struct {
	findAllIntersections        bool
	isCheckEndSegmentsOnly      bool
	keepIntersections           bool
	isInteriorIntersectionsOnly bool

	li                   *Algorithm_LineIntersector
	interiorIntersection *Geom_Coordinate
	intSegments          []*Geom_Coordinate
	intersections        []*Geom_Coordinate
	intersectionCount    int
}

var _ Noding_SegmentIntersector = (*Noding_NodingIntersectionFinder)(nil)

func (nif *Noding_NodingIntersectionFinder) IsNoding_SegmentIntersector() {}

// Noding_NodingIntersectionFinder_CreateAnyIntersectionFinder creates a finder
// which tests if there is at least one intersection. Uses short-circuiting for
// efficient performance. The intersection found is recorded.
func Noding_NodingIntersectionFinder_CreateAnyIntersectionFinder(li *Algorithm_LineIntersector) *Noding_NodingIntersectionFinder {
	return Noding_NewNodingIntersectionFinder(li)
}

// Noding_NodingIntersectionFinder_CreateAllIntersectionsFinder creates a finder
// which finds all intersections. The intersections are recorded for later
// inspection.
func Noding_NodingIntersectionFinder_CreateAllIntersectionsFinder(li *Algorithm_LineIntersector) *Noding_NodingIntersectionFinder {
	finder := Noding_NewNodingIntersectionFinder(li)
	finder.SetFindAllIntersections(true)
	return finder
}

// Noding_NodingIntersectionFinder_CreateInteriorIntersectionsFinder creates a
// finder which finds all interior intersections. The intersections are recorded
// for later inspection.
func Noding_NodingIntersectionFinder_CreateInteriorIntersectionsFinder(li *Algorithm_LineIntersector) *Noding_NodingIntersectionFinder {
	finder := Noding_NewNodingIntersectionFinder(li)
	finder.SetFindAllIntersections(true)
	finder.SetInteriorIntersectionsOnly(true)
	return finder
}

// Noding_NodingIntersectionFinder_CreateIntersectionCounter creates a finder
// which counts all intersections. The intersections are not recorded to reduce
// memory usage.
func Noding_NodingIntersectionFinder_CreateIntersectionCounter(li *Algorithm_LineIntersector) *Noding_NodingIntersectionFinder {
	finder := Noding_NewNodingIntersectionFinder(li)
	finder.SetFindAllIntersections(true)
	finder.SetKeepIntersections(false)
	return finder
}

// Noding_NodingIntersectionFinder_CreateInteriorIntersectionCounter creates a
// finder which counts all interior intersections. The intersections are not
// recorded to reduce memory usage.
func Noding_NodingIntersectionFinder_CreateInteriorIntersectionCounter(li *Algorithm_LineIntersector) *Noding_NodingIntersectionFinder {
	finder := Noding_NewNodingIntersectionFinder(li)
	finder.SetInteriorIntersectionsOnly(true)
	finder.SetFindAllIntersections(true)
	finder.SetKeepIntersections(false)
	return finder
}

// Noding_NewNodingIntersectionFinder creates an intersection finder which finds
// an intersection if one exists.
func Noding_NewNodingIntersectionFinder(li *Algorithm_LineIntersector) *Noding_NodingIntersectionFinder {
	return &Noding_NodingIntersectionFinder{
		li:                   li,
		interiorIntersection: nil,
		keepIntersections:    true,
		intersections:        make([]*Geom_Coordinate, 0),
	}
}

// SetFindAllIntersections sets whether all intersections should be computed.
// When this is false (the default value) the value of IsDone() is true after
// the first intersection is found.
//
// Default is false.
func (nif *Noding_NodingIntersectionFinder) SetFindAllIntersections(findAllIntersections bool) {
	nif.findAllIntersections = findAllIntersections
}

// SetInteriorIntersectionsOnly sets whether only interior (proper)
// intersections will be found.
func (nif *Noding_NodingIntersectionFinder) SetInteriorIntersectionsOnly(isInteriorIntersectionsOnly bool) {
	nif.isInteriorIntersectionsOnly = isInteriorIntersectionsOnly
}

// SetCheckEndSegmentsOnly sets whether only end segments should be tested for
// intersection. This is a performance optimization that may be used if the
// segments have been previously noded by an appropriate algorithm. It may be
// known that any potential noding failures will occur only in end segments.
func (nif *Noding_NodingIntersectionFinder) SetCheckEndSegmentsOnly(isCheckEndSegmentsOnly bool) {
	nif.isCheckEndSegmentsOnly = isCheckEndSegmentsOnly
}

// SetKeepIntersections sets whether intersection points are recorded. If the
// only need is to count intersection points, this can be set to false.
//
// Default is true.
func (nif *Noding_NodingIntersectionFinder) SetKeepIntersections(keepIntersections bool) {
	nif.keepIntersections = keepIntersections
}

// GetIntersections gets the intersections found.
func (nif *Noding_NodingIntersectionFinder) GetIntersections() []*Geom_Coordinate {
	return nif.intersections
}

// Count gets the count of intersections found.
func (nif *Noding_NodingIntersectionFinder) Count() int {
	return nif.intersectionCount
}

// HasIntersection tests whether an intersection was found.
func (nif *Noding_NodingIntersectionFinder) HasIntersection() bool {
	return nif.interiorIntersection != nil
}

// GetIntersection gets the computed location of the intersection. Due to
// round-off, the location may not be exact.
func (nif *Noding_NodingIntersectionFinder) GetIntersection() *Geom_Coordinate {
	return nif.interiorIntersection
}

// GetIntersectionSegments gets the endpoints of the intersecting segments.
func (nif *Noding_NodingIntersectionFinder) GetIntersectionSegments() []*Geom_Coordinate {
	return nif.intSegments
}

// ProcessIntersections is called by clients of the SegmentIntersector class to
// process intersections for two segments of the SegmentStrings being
// intersected. Note that some clients (such as MonotoneChains) may optimize
// away this call for segment pairs which they have determined do not intersect
// (e.g. by a disjoint envelope test).
func (nif *Noding_NodingIntersectionFinder) ProcessIntersections(
	e0 Noding_SegmentString, segIndex0 int,
	e1 Noding_SegmentString, segIndex1 int,
) {
	// short-circuit if intersection already found
	if !nif.findAllIntersections && nif.HasIntersection() {
		return
	}

	// don't bother intersecting a segment with itself
	isSameSegString := e0 == e1
	isSameSegment := isSameSegString && segIndex0 == segIndex1
	if isSameSegment {
		return
	}

	// If enabled, only test end segments (on either segString).
	if nif.isCheckEndSegmentsOnly {
		isEndSegPresent := noding_NodingIntersectionFinder_isEndSegment(e0, segIndex0) ||
			noding_NodingIntersectionFinder_isEndSegment(e1, segIndex1)
		if !isEndSegPresent {
			return
		}
	}

	p00 := e0.GetCoordinate(segIndex0)
	p01 := e0.GetCoordinate(segIndex0 + 1)
	p10 := e1.GetCoordinate(segIndex1)
	p11 := e1.GetCoordinate(segIndex1 + 1)
	isEnd00 := segIndex0 == 0
	isEnd01 := segIndex0+2 == e0.Size()
	isEnd10 := segIndex1 == 0
	isEnd11 := segIndex1+2 == e1.Size()

	nif.li.ComputeIntersection(p00, p01, p10, p11)
	// if (li.hasIntersection() && li.isProper()) Debug.println(li);

	// Check for an intersection in the interior of a segment
	isInteriorInt := nif.li.HasIntersection() && nif.li.IsInteriorIntersection()

	// Check for an intersection between two vertices which are not both endpoints.
	isInteriorVertexInt := false
	if !nif.isInteriorIntersectionsOnly {
		isAdjacentSegment := isSameSegString && java.AbsInt(segIndex1-segIndex0) <= 1
		isInteriorVertexInt = (!isAdjacentSegment) && noding_NodingIntersectionFinder_isInteriorVertexIntersection4(
			p00, p01, p10, p11,
			isEnd00, isEnd01, isEnd10, isEnd11)
	}

	if isInteriorInt || isInteriorVertexInt {
		// found an intersection!
		nif.intSegments = make([]*Geom_Coordinate, 4)
		nif.intSegments[0] = p00
		nif.intSegments[1] = p01
		nif.intSegments[2] = p10
		nif.intSegments[3] = p11

		// TODO: record endpoint intersection(s)
		nif.interiorIntersection = nif.li.GetIntersection(0)
		if nif.keepIntersections {
			nif.intersections = append(nif.intersections, nif.interiorIntersection)
		}
		nif.intersectionCount++
	}
}

// noding_NodingIntersectionFinder_isInteriorVertexIntersection4 tests if an
// intersection occurs between a segmentString interior vertex and another
// vertex. Note that intersections between two endpoint vertices are valid
// noding, and are not flagged.
func noding_NodingIntersectionFinder_isInteriorVertexIntersection4(
	p00, p01 *Geom_Coordinate,
	p10, p11 *Geom_Coordinate,
	isEnd00, isEnd01 bool,
	isEnd10, isEnd11 bool,
) bool {
	if noding_NodingIntersectionFinder_isInteriorVertexIntersection(p00, p10, isEnd00, isEnd10) {
		return true
	}
	if noding_NodingIntersectionFinder_isInteriorVertexIntersection(p00, p11, isEnd00, isEnd11) {
		return true
	}
	if noding_NodingIntersectionFinder_isInteriorVertexIntersection(p01, p10, isEnd01, isEnd10) {
		return true
	}
	if noding_NodingIntersectionFinder_isInteriorVertexIntersection(p01, p11, isEnd01, isEnd11) {
		return true
	}
	return false
}

// noding_NodingIntersectionFinder_isInteriorVertexIntersection tests if two
// vertices with at least one in a segmentString interior are equal.
func noding_NodingIntersectionFinder_isInteriorVertexIntersection(
	p0, p1 *Geom_Coordinate,
	isEnd0, isEnd1 bool,
) bool {
	// Intersections between endpoints are valid nodes, so not reported
	if isEnd0 && isEnd1 {
		return false
	}

	if p0.Equals2D(p1) {
		return true
	}
	return false
}

// noding_NodingIntersectionFinder_isEndSegment tests whether a segment in a
// SegmentString is an end segment (either the first or last).
func noding_NodingIntersectionFinder_isEndSegment(segStr Noding_SegmentString, index int) bool {
	if index == 0 {
		return true
	}
	if index >= segStr.Size()-2 {
		return true
	}
	return false
}

// IsDone reports whether the client of this class needs to continue testing all
// intersections in an arrangement.
func (nif *Noding_NodingIntersectionFinder) IsDone() bool {
	if nif.findAllIntersections {
		return false
	}
	return nif.interiorIntersection != nil
}
