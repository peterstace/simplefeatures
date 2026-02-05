package jts

var _ Noding_SegmentIntersector = (*Noding_SegmentIntersectionDetector)(nil)

// Noding_SegmentIntersectionDetector detects and records an intersection
// between two SegmentStrings, if one exists. Only a single intersection is
// recorded. This strategy can be configured to search for proper
// intersections. In this case, the presence of any kind of intersection will
// still be recorded, but searching will continue until either a proper
// intersection has been found or no intersections are detected.
type Noding_SegmentIntersectionDetector struct {
	li           *Algorithm_LineIntersector
	findProper   bool
	findAllTypes bool

	hasIntersection          bool
	hasProperIntersection    bool
	hasNonProperIntersection bool

	intPt       *Geom_Coordinate
	intSegments []*Geom_Coordinate
}

// IsNoding_SegmentIntersector is a marker method for interface identification.
func (sid *Noding_SegmentIntersectionDetector) IsNoding_SegmentIntersector() {}

// Noding_NewSegmentIntersectionDetector creates an intersection finder using a
// RobustLineIntersector.
func Noding_NewSegmentIntersectionDetector() *Noding_SegmentIntersectionDetector {
	return Noding_NewSegmentIntersectionDetectorWithLI(Algorithm_NewRobustLineIntersector().Algorithm_LineIntersector)
}

// Noding_NewSegmentIntersectionDetectorWithLI creates an intersection finder
// using a given LineIntersector.
func Noding_NewSegmentIntersectionDetectorWithLI(li *Algorithm_LineIntersector) *Noding_SegmentIntersectionDetector {
	return &Noding_SegmentIntersectionDetector{
		li: li,
	}
}

// SetFindProper sets whether processing must continue until a proper
// intersection is found.
func (sid *Noding_SegmentIntersectionDetector) SetFindProper(findProper bool) {
	sid.findProper = findProper
}

// SetFindAllIntersectionTypes sets whether processing can terminate once any
// intersection is found.
func (sid *Noding_SegmentIntersectionDetector) SetFindAllIntersectionTypes(findAllTypes bool) {
	sid.findAllTypes = findAllTypes
}

// HasIntersection tests whether an intersection was found.
func (sid *Noding_SegmentIntersectionDetector) HasIntersection() bool {
	return sid.hasIntersection
}

// HasProperIntersection tests whether a proper intersection was found.
func (sid *Noding_SegmentIntersectionDetector) HasProperIntersection() bool {
	return sid.hasProperIntersection
}

// HasNonProperIntersection tests whether a non-proper intersection was found.
func (sid *Noding_SegmentIntersectionDetector) HasNonProperIntersection() bool {
	return sid.hasNonProperIntersection
}

// GetIntersection gets the computed location of the intersection. Due to
// round-off, the location may not be exact.
func (sid *Noding_SegmentIntersectionDetector) GetIntersection() *Geom_Coordinate {
	return sid.intPt
}

// GetIntersectionSegments gets the endpoints of the intersecting segments.
func (sid *Noding_SegmentIntersectionDetector) GetIntersectionSegments() []*Geom_Coordinate {
	return sid.intSegments
}

// ProcessIntersections is called by clients of the SegmentIntersector class to
// process intersections for two segments of the SegmentStrings being
// intersected. Note that some clients (such as MonotoneChains) may optimize
// away this call for segment pairs which they have determined do not intersect
// (e.g. by a disjoint envelope test).
func (sid *Noding_SegmentIntersectionDetector) ProcessIntersections(
	e0 Noding_SegmentString, segIndex0 int,
	e1 Noding_SegmentString, segIndex1 int,
) {
	// don't bother intersecting a segment with itself
	if e0 == e1 && segIndex0 == segIndex1 {
		return
	}

	p00 := e0.GetCoordinate(segIndex0)
	p01 := e0.GetCoordinate(segIndex0 + 1)
	p10 := e1.GetCoordinate(segIndex1)
	p11 := e1.GetCoordinate(segIndex1 + 1)

	sid.li.ComputeIntersection(p00, p01, p10, p11)

	if sid.li.HasIntersection() {

		// record intersection info
		sid.hasIntersection = true

		isProper := sid.li.IsProper()
		if isProper {
			sid.hasProperIntersection = true
		}
		if !isProper {
			sid.hasNonProperIntersection = true
		}

		// If this is the kind of intersection we are searching for
		// OR no location has yet been recorded
		// save the location data
		saveLocation := true
		if sid.findProper && !isProper {
			saveLocation = false
		}

		if sid.intPt == nil || saveLocation {

			// record intersection location (approximate)
			sid.intPt = sid.li.GetIntersection(0)

			// record intersecting segments
			sid.intSegments = make([]*Geom_Coordinate, 4)
			sid.intSegments[0] = p00
			sid.intSegments[1] = p01
			sid.intSegments[2] = p10
			sid.intSegments[3] = p11
		}
	}
}

// IsDone tests whether processing can terminate, because all required
// information has been obtained (e.g. an intersection of the desired type has
// been detected).
func (sid *Noding_SegmentIntersectionDetector) IsDone() bool {
	// If finding all types, we can stop
	// when both possible types have been found.
	if sid.findAllTypes {
		return sid.hasProperIntersection && sid.hasNonProperIntersection
	}

	// If searching for a proper intersection, only stop if one is found
	if sid.findProper {
		return sid.hasProperIntersection
	}
	return sid.hasIntersection
}
