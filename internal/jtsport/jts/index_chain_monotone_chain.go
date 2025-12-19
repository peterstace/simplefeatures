package jts

import "math"

// IndexChain_MonotoneChain represents a monotone chain, which is a way of
// partitioning the segments of a linestring to allow for fast searching of
// intersections. They have the following properties:
//
//  1. The segments within a monotone chain never intersect each other.
//  2. The envelope of any contiguous subset of the segments in a monotone chain
//     is equal to the envelope of the endpoints of the subset.
//
// Property 1 means that there is no need to test pairs of segments from within
// the same monotone chain for intersection.
//
// Property 2 allows an efficient binary search to be used to find the
// intersection points of two monotone chains. For many types of real-world
// data, these properties eliminate a large number of segment comparisons,
// producing substantial speed gains.
//
// One of the goals of this implementation of MonotoneChains is to be as space
// and time efficient as possible. One design choice that aids this is that a
// MonotoneChain is based on a subarray of a list of points. This means that new
// arrays of points (potentially very large) do not have to be allocated.
//
// MonotoneChains support the following kinds of queries:
//   - Envelope select: determine all the segments in the chain which intersect a
//     given envelope
//   - Overlap: determine all the pairs of segments in two chains whose envelopes
//     overlap
//
// This implementation of MonotoneChains uses the concept of internal iterators
// (MonotoneChainSelectAction and MonotoneChainOverlapAction) to return the
// results for queries. This has time and space advantages, since it is not
// necessary to build lists of instantiated objects to represent the segments
// returned by the query. Queries made in this manner are thread-safe.
//
// MonotoneChains support being assigned an integer id value to provide a total
// ordering for a set of chains. This can be used during some kinds of
// processing to avoid redundant comparisons (i.e. by comparing only chains
// where the first id is less than the second).
//
// MonotoneChains support using a tolerance distance for overlap tests. This
// allows reporting overlap in situations where intersection snapping is being
// used. If this is used the chain envelope must be computed providing an
// expansion distance using GetEnvelopeWithExpansion.
type IndexChain_MonotoneChain struct {
	pts     []*Geom_Coordinate
	start   int
	end     int
	env     *Geom_Envelope
	context any // User-defined information.
	id      int // Useful for optimizing chain comparisons.
}

// IndexChain_NewMonotoneChain creates a new MonotoneChain based on the given
// array of points.
func IndexChain_NewMonotoneChain(pts []*Geom_Coordinate, start, end int, context any) *IndexChain_MonotoneChain {
	return &IndexChain_MonotoneChain{
		pts:     pts,
		start:   start,
		end:     end,
		context: context,
	}
}

// SetId sets the id of this chain. Useful for assigning an ordering to a set of
// chains, which can be used to avoid redundant processing.
func (mc *IndexChain_MonotoneChain) SetId(id int) {
	mc.id = id
}

// SetOverlapDistance sets the overlap distance used in overlap tests with other
// chains.
func (mc *IndexChain_MonotoneChain) SetOverlapDistance(distance float64) {
	// This is a no-op in the Java implementation (the field is commented out).
}

// GetId gets the id of this chain.
func (mc *IndexChain_MonotoneChain) GetId() int {
	return mc.id
}

// GetContext gets the user-defined context data value.
func (mc *IndexChain_MonotoneChain) GetContext() any {
	return mc.context
}

// GetEnvelope gets the envelope of the chain.
func (mc *IndexChain_MonotoneChain) GetEnvelope() *Geom_Envelope {
	return mc.GetEnvelopeWithExpansion(0.0)
}

// GetEnvelopeWithExpansion gets the envelope for this chain, expanded by a
// given distance.
func (mc *IndexChain_MonotoneChain) GetEnvelopeWithExpansion(expansionDistance float64) *Geom_Envelope {
	if mc.env == nil {
		// The monotonicity property allows fast envelope determination.
		p0 := mc.pts[mc.start]
		p1 := mc.pts[mc.end]
		mc.env = Geom_NewEnvelopeFromCoordinates(p0, p1)
		if expansionDistance > 0.0 {
			mc.env.ExpandBy(expansionDistance)
		}
	}
	return mc.env
}

// GetStartIndex gets the index of the start of the monotone chain in the
// underlying array of points.
func (mc *IndexChain_MonotoneChain) GetStartIndex() int {
	return mc.start
}

// GetEndIndex gets the index of the end of the monotone chain in the underlying
// array of points.
func (mc *IndexChain_MonotoneChain) GetEndIndex() int {
	return mc.end
}

// GetLineSegment gets the line segment starting at index.
func (mc *IndexChain_MonotoneChain) GetLineSegment(index int, ls *Geom_LineSegment) {
	ls.P0 = mc.pts[index]
	ls.P1 = mc.pts[index+1]
}

// GetCoordinates returns the subsequence of coordinates forming this chain.
// Allocates a new slice to hold the Coordinates.
func (mc *IndexChain_MonotoneChain) GetCoordinates() []*Geom_Coordinate {
	coord := make([]*Geom_Coordinate, mc.end-mc.start+1)
	index := 0
	for i := mc.start; i <= mc.end; i++ {
		coord[index] = mc.pts[i]
		index++
	}
	return coord
}

// Select determines all the line segments in the chain whose envelopes overlap
// the searchEnvelope, and processes them.
//
// The monotone chain search algorithm attempts to optimize performance by not
// calling the select action on chain segments which it can determine are not in
// the search envelope. However, it *may* call the select action on segments
// which do not intersect the search envelope. This saves on the overhead of
// checking envelope intersection each time, since clients may be able to do
// this more efficiently.
func (mc *IndexChain_MonotoneChain) Select(searchEnv *Geom_Envelope, mcs *IndexChain_MonotoneChainSelectAction) {
	mc.computeSelect(searchEnv, mc.start, mc.end, mcs)
}

func (mc *IndexChain_MonotoneChain) computeSelect(searchEnv *Geom_Envelope, start0, end0 int, mcs *IndexChain_MonotoneChainSelectAction) {
	p0 := mc.pts[start0]
	p1 := mc.pts[end0]

	// Terminating condition for the recursion.
	if end0-start0 == 1 {
		mcs.Select(mc, start0)
		return
	}
	// Nothing to do if the envelopes don't overlap.
	if !searchEnv.IntersectsCoordinates(p0, p1) {
		return
	}

	// The chains overlap, so split each in half and iterate (binary search).
	mid := (start0 + end0) / 2

	// Assert: mid != start or end (since we checked above for end - start <= 1).
	// Check terminating conditions before recursing.
	if start0 < mid {
		mc.computeSelect(searchEnv, start0, mid, mcs)
	}
	if mid < end0 {
		mc.computeSelect(searchEnv, mid, end0, mcs)
	}
}

// ComputeOverlaps determines the line segments in two chains which may overlap,
// and passes them to an overlap action.
//
// The monotone chain search algorithm attempts to optimize performance by not
// calling the overlap action on chain segments which it can determine do not
// overlap. However, it *may* call the overlap action on segments which do not
// actually interact. This saves on the overhead of checking intersection each
// time, since clients may be able to do this more efficiently.
func (mc *IndexChain_MonotoneChain) ComputeOverlaps(other *IndexChain_MonotoneChain, mco *IndexChain_MonotoneChainOverlapAction) {
	mc.computeOverlaps(mc.start, mc.end, other, other.start, other.end, 0.0, mco)
}

// ComputeOverlapsWithTolerance determines the line segments in two chains which
// may overlap, using an overlap distance tolerance, and passes them to an
// overlap action.
func (mc *IndexChain_MonotoneChain) ComputeOverlapsWithTolerance(other *IndexChain_MonotoneChain, overlapTolerance float64, mco *IndexChain_MonotoneChainOverlapAction) {
	mc.computeOverlaps(mc.start, mc.end, other, other.start, other.end, overlapTolerance, mco)
}

// computeOverlaps uses an efficient mutual binary search strategy to determine
// which pairs of chain segments may overlap, and calls the given overlap action
// on them.
func (mc *IndexChain_MonotoneChain) computeOverlaps(start0, end0 int, other *IndexChain_MonotoneChain, start1, end1 int, overlapTolerance float64, mco *IndexChain_MonotoneChainOverlapAction) {
	// Terminating condition for the recursion.
	if end0-start0 == 1 && end1-start1 == 1 {
		mco.Overlap(mc, start0, other, start1)
		return
	}
	// Nothing to do if the envelopes of these subchains don't overlap.
	if !mc.overlaps(start0, end0, other, start1, end1, overlapTolerance) {
		return
	}

	// The chains overlap, so split each in half and iterate (binary search).
	mid0 := (start0 + end0) / 2
	mid1 := (start1 + end1) / 2

	// Assert: mid != start or end (since we checked above for end - start <= 1).
	// Check terminating conditions before recursing.
	if start0 < mid0 {
		if start1 < mid1 {
			mc.computeOverlaps(start0, mid0, other, start1, mid1, overlapTolerance, mco)
		}
		if mid1 < end1 {
			mc.computeOverlaps(start0, mid0, other, mid1, end1, overlapTolerance, mco)
		}
	}
	if mid0 < end0 {
		if start1 < mid1 {
			mc.computeOverlaps(mid0, end0, other, start1, mid1, overlapTolerance, mco)
		}
		if mid1 < end1 {
			mc.computeOverlaps(mid0, end0, other, mid1, end1, overlapTolerance, mco)
		}
	}
}

// overlaps tests whether the envelope of a section of the chain overlaps
// (intersects) the envelope of a section of another target chain. This test is
// efficient due to the monotonicity property of the sections (i.e. the
// envelopes can be determined from the section endpoints rather than a full
// scan).
func (mc *IndexChain_MonotoneChain) overlaps(start0, end0 int, other *IndexChain_MonotoneChain, start1, end1 int, overlapTolerance float64) bool {
	if overlapTolerance > 0.0 {
		return mc.overlapsWithTolerance(mc.pts[start0], mc.pts[end0], other.pts[start1], other.pts[end1], overlapTolerance)
	}
	return Geom_Envelope_IntersectsEnvelopeEnvelope(mc.pts[start0], mc.pts[end0], other.pts[start1], other.pts[end1])
}

func (mc *IndexChain_MonotoneChain) overlapsWithTolerance(p1, p2, q1, q2 *Geom_Coordinate, overlapTolerance float64) bool {
	minq := math.Min(q1.X, q2.X)
	maxq := math.Max(q1.X, q2.X)
	minp := math.Min(p1.X, p2.X)
	maxp := math.Max(p1.X, p2.X)

	if minp > maxq+overlapTolerance {
		return false
	}
	if maxp < minq-overlapTolerance {
		return false
	}

	minq = math.Min(q1.Y, q2.Y)
	maxq = math.Max(q1.Y, q2.Y)
	minp = math.Min(p1.Y, p2.Y)
	maxp = math.Max(p1.Y, p2.Y)

	if minp > maxq+overlapTolerance {
		return false
	}
	if maxp < minq-overlapTolerance {
		return false
	}
	return true
}
