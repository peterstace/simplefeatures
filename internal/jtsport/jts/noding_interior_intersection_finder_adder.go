package jts

var _ Noding_SegmentIntersector = (*Noding_InteriorIntersectionFinderAdder)(nil)

// Noding_InteriorIntersectionFinderAdder finds interior intersections between
// line segments in NodedSegmentStrings, and adds them as nodes using
// AddIntersections.
//
// This class is used primarily for Snap-Rounding. For general-purpose noding,
// use IntersectionAdder.
type Noding_InteriorIntersectionFinderAdder struct {
	li                    *Algorithm_LineIntersector
	interiorIntersections []*Geom_Coordinate
}

// IsNoding_SegmentIntersector is a marker method for interface identification.
func (iifa *Noding_InteriorIntersectionFinderAdder) IsNoding_SegmentIntersector() {}

// Noding_NewInteriorIntersectionFinderAdder creates an intersection finder
// which finds all proper intersections.
func Noding_NewInteriorIntersectionFinderAdder(li *Algorithm_LineIntersector) *Noding_InteriorIntersectionFinderAdder {
	return &Noding_InteriorIntersectionFinderAdder{
		li:                    li,
		interiorIntersections: make([]*Geom_Coordinate, 0),
	}
}

// GetInteriorIntersections returns the list of interior intersections found.
func (iifa *Noding_InteriorIntersectionFinderAdder) GetInteriorIntersections() []*Geom_Coordinate {
	return iifa.interiorIntersections
}

// ProcessIntersections is called by clients of the SegmentIntersector class to
// process intersections for two segments of the SegmentStrings being
// intersected.
func (iifa *Noding_InteriorIntersectionFinderAdder) ProcessIntersections(
	e0 Noding_SegmentString, segIndex0 int,
	e1 Noding_SegmentString, segIndex1 int,
) {
	if e0 == e1 && segIndex0 == segIndex1 {
		return
	}

	p00 := e0.GetCoordinate(segIndex0)
	p01 := e0.GetCoordinate(segIndex0 + 1)
	p10 := e1.GetCoordinate(segIndex1)
	p11 := e1.GetCoordinate(segIndex1 + 1)

	iifa.li.ComputeIntersection(p00, p01, p10, p11)

	if iifa.li.HasIntersection() {
		if iifa.li.IsInteriorIntersection() {
			for intIndex := 0; intIndex < iifa.li.GetIntersectionNum(); intIndex++ {
				iifa.interiorIntersections = append(iifa.interiorIntersections, iifa.li.GetIntersection(intIndex))
			}
			nss0 := e0.(*Noding_NodedSegmentString)
			nss1 := e1.(*Noding_NodedSegmentString)
			nss0.AddIntersections(iifa.li, segIndex0, 0)
			nss1.AddIntersections(iifa.li, segIndex1, 1)
		}
	}
}

// IsDone always returns false since all intersections should be processed.
func (iifa *Noding_InteriorIntersectionFinderAdder) IsDone() bool {
	return false
}
