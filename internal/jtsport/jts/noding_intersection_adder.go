package jts

var _ Noding_SegmentIntersector = (*Noding_IntersectionAdder)(nil)

// Noding_IntersectionAdder computes the possible intersections between two line
// segments in NodedSegmentStrings and adds them to each string using
// AddIntersections.
type Noding_IntersectionAdder struct {
	// These variables keep track of what types of intersections were found
	// during ALL edges that have been intersected.
	hasIntersection      bool
	hasProper            bool
	hasProperInterior    bool
	hasInterior          bool
	properIntersectionPt *Geom_Coordinate
	li                   *Algorithm_LineIntersector
	isSelfIntersection   bool

	NumIntersections         int
	NumInteriorIntersections int
	NumProperIntersections   int
	NumTests                 int
}

// IsNoding_SegmentIntersector is a marker method for interface identification.
func (ia *Noding_IntersectionAdder) IsNoding_SegmentIntersector() {}

// Noding_IntersectionAdder_IsAdjacentSegments returns true if the segment
// indices are adjacent.
func Noding_IntersectionAdder_IsAdjacentSegments(i1, i2 int) bool {
	diff := i1 - i2
	if diff < 0 {
		diff = -diff
	}
	return diff == 1
}

// Noding_NewIntersectionAdder creates a new IntersectionAdder with the given
// LineIntersector.
func Noding_NewIntersectionAdder(li *Algorithm_LineIntersector) *Noding_IntersectionAdder {
	return &Noding_IntersectionAdder{
		li: li,
	}
}

// GetLineIntersector returns the LineIntersector used by this IntersectionAdder.
func (ia *Noding_IntersectionAdder) GetLineIntersector() *Algorithm_LineIntersector {
	return ia.li
}

// GetProperIntersectionPoint returns the proper intersection point, or nil if
// none was found.
func (ia *Noding_IntersectionAdder) GetProperIntersectionPoint() *Geom_Coordinate {
	return ia.properIntersectionPt
}

// HasIntersection returns true if any intersection was found.
func (ia *Noding_IntersectionAdder) HasIntersection() bool {
	return ia.hasIntersection
}

// HasProperIntersection returns true if a proper intersection was found.
// A proper intersection is an intersection which is interior to at least two
// line segments. Note that a proper intersection is not necessarily in the
// interior of the entire Geometry, since another edge may have an endpoint
// equal to the intersection, which according to SFS semantics can result in
// the point being on the Boundary of the Geometry.
func (ia *Noding_IntersectionAdder) HasProperIntersection() bool {
	return ia.hasProper
}

// HasProperInteriorIntersection returns true if a proper interior intersection
// was found. A proper interior intersection is a proper intersection which is
// not contained in the set of boundary nodes set for this SegmentIntersector.
func (ia *Noding_IntersectionAdder) HasProperInteriorIntersection() bool {
	return ia.hasProperInterior
}

// HasInteriorIntersection returns true if an interior intersection was found.
// An interior intersection is an intersection which is in the interior of some
// segment.
func (ia *Noding_IntersectionAdder) HasInteriorIntersection() bool {
	return ia.hasInterior
}

// isTrivialIntersection tests whether an intersection is trivial.
// A trivial intersection is an apparent self-intersection which in fact is
// simply the point shared by adjacent line segments. Note that closed edges
// require a special check for the point shared by the beginning and end
// segments.
func (ia *Noding_IntersectionAdder) isTrivialIntersection(
	e0 Noding_SegmentString, segIndex0 int,
	e1 Noding_SegmentString, segIndex1 int,
) bool {
	if e0 == e1 {
		if ia.li.GetIntersectionNum() == 1 {
			if Noding_IntersectionAdder_IsAdjacentSegments(segIndex0, segIndex1) {
				return true
			}
			if e0.IsClosed() {
				maxSegIndex := e0.Size() - 1
				if (segIndex0 == 0 && segIndex1 == maxSegIndex) ||
					(segIndex1 == 0 && segIndex0 == maxSegIndex) {
					return true
				}
			}
		}
	}
	return false
}

// ProcessIntersections is called by clients of the SegmentIntersector class to
// process intersections for two segments of the SegmentStrings being
// intersected. Note that some clients (such as MonotoneChains) may optimize
// away this call for segment pairs which they have determined do not intersect
// (e.g. by a disjoint envelope test).
func (ia *Noding_IntersectionAdder) ProcessIntersections(
	e0 Noding_SegmentString, segIndex0 int,
	e1 Noding_SegmentString, segIndex1 int,
) {
	if e0 == e1 && segIndex0 == segIndex1 {
		return
	}
	ia.NumTests++
	p00 := e0.GetCoordinate(segIndex0)
	p01 := e0.GetCoordinate(segIndex0 + 1)
	p10 := e1.GetCoordinate(segIndex1)
	p11 := e1.GetCoordinate(segIndex1 + 1)

	ia.li.ComputeIntersection(p00, p01, p10, p11)
	if ia.li.HasIntersection() {
		ia.NumIntersections++
		if ia.li.IsInteriorIntersection() {
			ia.NumInteriorIntersections++
			ia.hasInterior = true
		}
		// If the segments are adjacent they have at least one trivial
		// intersection, the shared endpoint. Don't bother adding it if it is
		// the only intersection.
		if !ia.isTrivialIntersection(e0, segIndex0, e1, segIndex1) {
			ia.hasIntersection = true
			nss0 := e0.(*Noding_NodedSegmentString)
			nss1 := e1.(*Noding_NodedSegmentString)
			nss0.AddIntersections(ia.li, segIndex0, 0)
			nss1.AddIntersections(ia.li, segIndex1, 1)
			if ia.li.IsProper() {
				ia.NumProperIntersections++
				ia.hasProper = true
				ia.hasProperInterior = true
			}
		}
	}
}

// IsDone always returns false since all intersections should be processed.
func (ia *Noding_IntersectionAdder) IsDone() bool {
	return false
}
