package jts

// Noding_FastSegmentSetIntersectionFinder finds if two sets of SegmentStrings
// intersect. Uses indexing for fast performance and to optimize repeated tests
// against a target set of lines. Short-circuited to return as soon an
// intersection is found.
//
// Immutable and thread-safe.
type Noding_FastSegmentSetIntersectionFinder struct {
	segSetMutInt Noding_SegmentSetMutualIntersector
}

// Noding_NewFastSegmentSetIntersectionFinder creates an intersection finder
// against a given set of segment strings.
func Noding_NewFastSegmentSetIntersectionFinder(baseSegStrings []Noding_SegmentString) *Noding_FastSegmentSetIntersectionFinder {
	return &Noding_FastSegmentSetIntersectionFinder{
		segSetMutInt: Noding_NewMCIndexSegmentSetMutualIntersector(baseSegStrings),
	}
}

// GetSegmentSetIntersector gets the segment set intersector used by this class.
// This allows other uses of the same underlying indexed structure.
func (f *Noding_FastSegmentSetIntersectionFinder) GetSegmentSetIntersector() Noding_SegmentSetMutualIntersector {
	return f.segSetMutInt
}

// Intersects tests for intersections with a given set of target
// SegmentStrings.
func (f *Noding_FastSegmentSetIntersectionFinder) Intersects(segStrings []Noding_SegmentString) bool {
	intFinder := Noding_NewSegmentIntersectionDetector()
	return f.IntersectsWithDetector(segStrings, intFinder)
}

// IntersectsWithDetector tests for intersections with a given set of target
// SegmentStrings using a given SegmentIntersectionDetector.
func (f *Noding_FastSegmentSetIntersectionFinder) IntersectsWithDetector(segStrings []Noding_SegmentString, intDetector *Noding_SegmentIntersectionDetector) bool {
	f.segSetMutInt.Process(segStrings, intDetector)
	return intDetector.HasIntersection()
}
