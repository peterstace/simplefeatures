package jts

// Noding_NodingValidator validates that a collection of SegmentStrings is
// correctly noded. Throws an appropriate exception if a noding error is found.
type Noding_NodingValidator struct {
	li         *Algorithm_LineIntersector
	segStrings []Noding_SegmentString
}

// Noding_NewNodingValidator creates a new NodingValidator.
func Noding_NewNodingValidator(segStrings []Noding_SegmentString) *Noding_NodingValidator {
	rli := Algorithm_NewRobustLineIntersector()
	return &Noding_NodingValidator{
		li:         rli.Algorithm_LineIntersector,
		segStrings: segStrings,
	}
}

// CheckValid checks whether the supplied segment strings are correctly noded.
// Panics with a RuntimeException if a noding error is found.
func (nv *Noding_NodingValidator) CheckValid() {
	nv.checkEndPtVertexIntersections()
	nv.checkInteriorIntersections()
	nv.checkCollapses()
}

// checkCollapses checks if a segment string contains a segment pattern a-b-a
// (which implies a self-intersection).
func (nv *Noding_NodingValidator) checkCollapses() {
	for _, ss := range nv.segStrings {
		nv.checkCollapsesForSegmentString(ss)
	}
}

func (nv *Noding_NodingValidator) checkCollapsesForSegmentString(ss Noding_SegmentString) {
	pts := ss.GetCoordinates()
	for i := 0; i < len(pts)-2; i++ {
		nv.checkCollapse(pts[i], pts[i+1], pts[i+2])
	}
}

func (nv *Noding_NodingValidator) checkCollapse(p0, p1, p2 *Geom_Coordinate) {
	if p0.Equals(p2) {
		panic("found non-noded collapse at " + p0.String() + "-" + p1.String() + "-" + p2.String())
	}
}

// checkInteriorIntersections checks all pairs of segments for intersections
// at an interior point of a segment.
func (nv *Noding_NodingValidator) checkInteriorIntersections() {
	for _, ss0 := range nv.segStrings {
		for _, ss1 := range nv.segStrings {
			nv.checkInteriorIntersectionsBetween(ss0, ss1)
		}
	}
}

func (nv *Noding_NodingValidator) checkInteriorIntersectionsBetween(ss0, ss1 Noding_SegmentString) {
	pts0 := ss0.GetCoordinates()
	pts1 := ss1.GetCoordinates()
	for i0 := 0; i0 < len(pts0)-1; i0++ {
		for i1 := 0; i1 < len(pts1)-1; i1++ {
			nv.checkInteriorIntersection(ss0, i0, ss1, i1)
		}
	}
}

func (nv *Noding_NodingValidator) checkInteriorIntersection(e0 Noding_SegmentString, segIndex0 int, e1 Noding_SegmentString, segIndex1 int) {
	if e0 == e1 && segIndex0 == segIndex1 {
		return
	}
	p00 := e0.GetCoordinate(segIndex0)
	p01 := e0.GetCoordinate(segIndex0 + 1)
	p10 := e1.GetCoordinate(segIndex1)
	p11 := e1.GetCoordinate(segIndex1 + 1)

	nv.li.ComputeIntersection(p00, p01, p10, p11)
	if nv.li.HasIntersection() {
		if nv.li.IsProper() ||
			nv.hasInteriorIntersection(nv.li, p00, p01) ||
			nv.hasInteriorIntersection(nv.li, p10, p11) {
			panic("found non-noded intersection at " +
				p00.String() + "-" + p01.String() +
				" and " +
				p10.String() + "-" + p11.String())
		}
	}
}

// hasInteriorIntersection returns true if there is an intersection point
// which is not an endpoint of the segment p0-p1.
func (nv *Noding_NodingValidator) hasInteriorIntersection(li *Algorithm_LineIntersector, p0, p1 *Geom_Coordinate) bool {
	for i := 0; i < li.GetIntersectionNum(); i++ {
		intPt := li.GetIntersection(i)
		if !intPt.Equals(p0) && !intPt.Equals(p1) {
			return true
		}
	}
	return false
}

// checkEndPtVertexIntersections checks for intersections between an endpoint
// of a segment string and an interior vertex of another segment string.
func (nv *Noding_NodingValidator) checkEndPtVertexIntersections() {
	for _, ss := range nv.segStrings {
		pts := ss.GetCoordinates()
		nv.checkEndPtVertexIntersection(pts[0], nv.segStrings)
		nv.checkEndPtVertexIntersection(pts[len(pts)-1], nv.segStrings)
	}
}

func (nv *Noding_NodingValidator) checkEndPtVertexIntersection(testPt *Geom_Coordinate, segStrings []Noding_SegmentString) {
	for _, ss := range segStrings {
		pts := ss.GetCoordinates()
		for j := 1; j < len(pts)-1; j++ {
			if pts[j].Equals(testPt) {
				panic("found endpt/interior pt intersection at index " + string(rune('0'+j)) + " :pt " + testPt.String())
			}
		}
	}
}
