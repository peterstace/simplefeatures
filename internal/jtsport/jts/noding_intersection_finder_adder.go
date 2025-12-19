package jts

var _ Noding_SegmentIntersector = (*Noding_IntersectionFinderAdder)(nil)

// Noding_IntersectionFinderAdder finds interior intersections between line
// segments in NodedSegmentStrings, and adds them as nodes using AddIntersections.
//
// This class is used primarily for Snap-Rounding. For general-purpose noding,
// use IntersectionAdder.
//
// Deprecated: see InteriorIntersectionFinderAdder.
type Noding_IntersectionFinderAdder struct {
	li                    *Algorithm_LineIntersector
	interiorIntersections []*Geom_Coordinate
}

// IsNoding_SegmentIntersector is a marker method for interface identification.
func (ifa *Noding_IntersectionFinderAdder) IsNoding_SegmentIntersector() {}

// Noding_NewIntersectionFinderAdder creates an intersection finder which finds
// all proper intersections.
func Noding_NewIntersectionFinderAdder(li *Algorithm_LineIntersector) *Noding_IntersectionFinderAdder {
	return &Noding_IntersectionFinderAdder{
		li:                    li,
		interiorIntersections: make([]*Geom_Coordinate, 0),
	}
}

// GetInteriorIntersections returns the list of interior intersections found.
func (ifa *Noding_IntersectionFinderAdder) GetInteriorIntersections() []*Geom_Coordinate {
	return ifa.interiorIntersections
}

// ProcessIntersections is called by clients of the SegmentIntersector class to
// process intersections for two segments of the SegmentStrings being
// intersected. Note that some clients (such as MonotoneChains) may optimize
// away this call for segment pairs which they have determined do not intersect
// (e.g. by a disjoint envelope test).
func (ifa *Noding_IntersectionFinderAdder) ProcessIntersections(
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

	ifa.li.ComputeIntersection(p00, p01, p10, p11)

	if ifa.li.HasIntersection() {
		if ifa.li.IsInteriorIntersection() {
			for intIndex := 0; intIndex < ifa.li.GetIntersectionNum(); intIndex++ {
				ifa.interiorIntersections = append(ifa.interiorIntersections, ifa.li.GetIntersection(intIndex))
			}
			nss0 := e0.(*Noding_NodedSegmentString)
			nss1 := e1.(*Noding_NodedSegmentString)
			nss0.AddIntersections(ifa.li, segIndex0, 0)
			nss1.AddIntersections(ifa.li, segIndex1, 1)
		}
	}
}

// IsDone always returns false since all intersections should be processed.
func (ifa *Noding_IntersectionFinderAdder) IsDone() bool {
	return false
}
