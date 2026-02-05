package jts

// Noding_FastNodingValidator validates that a collection of SegmentStrings is
// correctly noded. Indexing is used to improve performance. By default
// validation stops after a single non-noded intersection is detected.
// Alternatively, it can be requested to detect all intersections by using
// SetFindAllIntersections.
//
// The validator does not check for topology collapse situations (e.g. where two
// segment strings are fully co-incident).
//
// The validator checks for the following situations which indicate incorrect
// noding:
//   - Proper intersections between segments (i.e. the intersection is interior
//     to both segments)
//   - Intersections at an interior vertex (i.e. with an endpoint or another
//     interior vertex)
//
// The client may either test the IsValid() condition, or request that a
// suitable TopologyException be thrown.
type Noding_FastNodingValidator struct {
	li *Algorithm_LineIntersector

	segStrings           []Noding_SegmentString
	findAllIntersections bool
	segInt               *Noding_NodingIntersectionFinder
	isValid              bool
}

// Noding_FastNodingValidator_ComputeIntersections gets a list of all
// intersections found. Intersections are represented as Coordinates. List is
// empty if none were found.
func Noding_FastNodingValidator_ComputeIntersections(segStrings []Noding_SegmentString) []*Geom_Coordinate {
	nv := Noding_NewFastNodingValidator(segStrings)
	nv.SetFindAllIntersections(true)
	nv.IsValid()
	return nv.GetIntersections()
}

// Noding_NewFastNodingValidator creates a new noding validator for a given set
// of linework.
func Noding_NewFastNodingValidator(segStrings []Noding_SegmentString) *Noding_FastNodingValidator {
	return &Noding_FastNodingValidator{
		li:         Algorithm_NewRobustLineIntersector().Algorithm_LineIntersector,
		segStrings: segStrings,
		isValid:    true,
	}
}

// SetFindAllIntersections sets whether all intersections should be found.
func (fnv *Noding_FastNodingValidator) SetFindAllIntersections(findAllIntersections bool) {
	fnv.findAllIntersections = findAllIntersections
}

// GetIntersections gets a list of all intersections found. Intersections are
// represented as Coordinates. List is empty if none were found.
func (fnv *Noding_FastNodingValidator) GetIntersections() []*Geom_Coordinate {
	return fnv.segInt.GetIntersections()
}

// IsValid checks for an intersection and reports if one is found.
func (fnv *Noding_FastNodingValidator) IsValid() bool {
	fnv.execute()
	return fnv.isValid
}

// GetErrorMessage returns an error message indicating the segments containing
// the intersection.
func (fnv *Noding_FastNodingValidator) GetErrorMessage() string {
	if fnv.isValid {
		return "no intersections found"
	}

	intSegs := fnv.segInt.GetIntersectionSegments()
	return "found non-noded intersection between " +
		Io_WKTWriter_ToLineStringFromTwoCoords(intSegs[0], intSegs[1]) +
		" and " +
		Io_WKTWriter_ToLineStringFromTwoCoords(intSegs[2], intSegs[3])
}

// CheckValid checks for an intersection and panics with a TopologyException if
// one is found.
func (fnv *Noding_FastNodingValidator) CheckValid() {
	fnv.execute()
	if !fnv.isValid {
		panic(Geom_NewTopologyExceptionWithCoordinate(fnv.GetErrorMessage(), fnv.segInt.GetIntersection()))
	}
}

func (fnv *Noding_FastNodingValidator) execute() {
	if fnv.segInt != nil {
		return
	}
	fnv.checkInteriorIntersections()
}

func (fnv *Noding_FastNodingValidator) checkInteriorIntersections() {
	// MD - It may even be reliable to simply check whether
	// end segments (of SegmentStrings) have an interior intersection,
	// since noding should have split any true interior intersections already.
	fnv.isValid = true
	fnv.segInt = Noding_NewNodingIntersectionFinder(fnv.li)
	fnv.segInt.SetFindAllIntersections(fnv.findAllIntersections)
	noder := Noding_NewMCIndexNoder()
	noder.SetSegmentIntersector(fnv.segInt)
	noder.ComputeNodes(fnv.segStrings)
	if fnv.segInt.HasIntersection() {
		fnv.isValid = false
		return
	}
}
