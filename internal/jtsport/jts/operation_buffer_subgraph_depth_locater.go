package jts

import "sort"

// operationBuffer_SubgraphDepthLocater locates a subgraph inside a set of subgraphs,
// in order to determine the outside depth of the subgraph.
// The input subgraphs are assumed to have had depths
// already calculated for their edges.
type operationBuffer_SubgraphDepthLocater struct {
	subgraphs []*OperationBuffer_BufferSubgraph
	seg       *Geom_LineSegment
}

// operationBuffer_newSubgraphDepthLocater creates a new SubgraphDepthLocater.
func operationBuffer_newSubgraphDepthLocater(subgraphs []*OperationBuffer_BufferSubgraph) *operationBuffer_SubgraphDepthLocater {
	return &operationBuffer_SubgraphDepthLocater{
		subgraphs: subgraphs,
		seg:       Geom_NewLineSegment(),
	}
}

// GetDepth returns the depth at the given coordinate.
func (sdl *operationBuffer_SubgraphDepthLocater) GetDepth(p *Geom_Coordinate) int {
	stabbedSegments := sdl.findStabbedSegments(p)
	// if no segments on stabbing line subgraph must be outside all others.
	if len(stabbedSegments) == 0 {
		return 0
	}
	sort.Slice(stabbedSegments, func(i, j int) bool {
		return stabbedSegments[i].compareTo(stabbedSegments[j]) < 0
	})
	ds := stabbedSegments[0]
	return ds.leftDepth
}

// findStabbedSegments finds all non-horizontal segments intersecting the stabbing line.
// The stabbing line is the ray to the right of stabbingRayLeftPt.
func (sdl *operationBuffer_SubgraphDepthLocater) findStabbedSegments(stabbingRayLeftPt *Geom_Coordinate) []*operationBuffer_DepthSegment {
	stabbedSegments := make([]*operationBuffer_DepthSegment, 0)
	for _, bsg := range sdl.subgraphs {
		// optimization - don't bother checking subgraphs which the ray does not intersect
		env := bsg.GetEnvelope()
		if stabbingRayLeftPt.Y < env.GetMinY() || stabbingRayLeftPt.Y > env.GetMaxY() {
			continue
		}

		sdl.findStabbedSegmentsInList(stabbingRayLeftPt, bsg.GetDirectedEdges(), &stabbedSegments)
	}
	return stabbedSegments
}

// findStabbedSegmentsInList finds all non-horizontal segments intersecting the stabbing line
// in the list of dirEdges.
// The stabbing line is the ray to the right of stabbingRayLeftPt.
func (sdl *operationBuffer_SubgraphDepthLocater) findStabbedSegmentsInList(stabbingRayLeftPt *Geom_Coordinate, dirEdges []*Geomgraph_DirectedEdge, stabbedSegments *[]*operationBuffer_DepthSegment) {
	// Check all forward DirectedEdges only. This is still general,
	// because each Edge has a forward DirectedEdge.
	for _, de := range dirEdges {
		if !de.IsForward() {
			continue
		}
		sdl.findStabbedSegmentsInEdge(stabbingRayLeftPt, de, stabbedSegments)
	}
}

// findStabbedSegmentsInEdge finds all non-horizontal segments intersecting the stabbing line
// in the input dirEdge.
// The stabbing line is the ray to the right of stabbingRayLeftPt.
func (sdl *operationBuffer_SubgraphDepthLocater) findStabbedSegmentsInEdge(stabbingRayLeftPt *Geom_Coordinate, dirEdge *Geomgraph_DirectedEdge, stabbedSegments *[]*operationBuffer_DepthSegment) {
	pts := dirEdge.GetEdge().GetCoordinates()
	for i := 0; i < len(pts)-1; i++ {
		sdl.seg.P0 = pts[i]
		sdl.seg.P1 = pts[i+1]
		// ensure segment always points upwards
		if sdl.seg.P0.Y > sdl.seg.P1.Y {
			sdl.seg.Reverse()
		}

		// skip segment if it is left of the stabbing line
		maxx := sdl.seg.P0.X
		if sdl.seg.P1.X > maxx {
			maxx = sdl.seg.P1.X
		}
		if maxx < stabbingRayLeftPt.X {
			continue
		}

		// skip horizontal segments (there will be a non-horizontal one carrying the same depth info
		if sdl.seg.IsHorizontal() {
			continue
		}

		// skip if segment is above or below stabbing line
		if stabbingRayLeftPt.Y < sdl.seg.P0.Y || stabbingRayLeftPt.Y > sdl.seg.P1.Y {
			continue
		}

		// skip if stabbing ray is right of the segment
		if Algorithm_Orientation_Index(sdl.seg.P0, sdl.seg.P1, stabbingRayLeftPt) == Algorithm_Orientation_Right {
			continue
		}

		// stabbing line cuts this segment, so record it
		depth := dirEdge.GetDepth(Geom_Position_Left)
		// if segment direction was flipped, use RHS depth instead
		if !sdl.seg.P0.Equals(pts[i]) {
			depth = dirEdge.GetDepth(Geom_Position_Right)
		}
		ds := operationBuffer_newDepthSegment(sdl.seg, depth)
		*stabbedSegments = append(*stabbedSegments, ds)
	}
}

// operationBuffer_DepthSegment is a segment from a directed edge which has been assigned a depth value
// for its sides.
type operationBuffer_DepthSegment struct {
	upwardSeg *Geom_LineSegment
	leftDepth int
}

// operationBuffer_newDepthSegment creates a new DepthSegment.
func operationBuffer_newDepthSegment(seg *Geom_LineSegment, depth int) *operationBuffer_DepthSegment {
	// Assert: input seg is upward (p0.y <= p1.y)
	return &operationBuffer_DepthSegment{
		upwardSeg: Geom_NewLineSegmentFromLineSegment(seg),
		leftDepth: depth,
	}
}

// isUpward tests if the segment points upward.
func (ds *operationBuffer_DepthSegment) isUpward() bool {
	return ds.upwardSeg.P0.Y <= ds.upwardSeg.P1.Y
}

// compareTo is a comparison operation which orders segments left to right.
//
// The definition of the ordering is:
//   - -1 : if DS1.seg is left of or below DS2.seg (DS1 < DS2)
//   - 1 : if DS1.seg is right of or above DS2.seg (DS1 > DS2)
//   - 0 : if the segments are identical
func (ds *operationBuffer_DepthSegment) compareTo(other *operationBuffer_DepthSegment) int {
	// If segment envelopes do not overlap, then
	// can use standard segment lexicographic ordering.
	if ds.upwardSeg.MinX() >= other.upwardSeg.MaxX() ||
		ds.upwardSeg.MaxX() <= other.upwardSeg.MinX() ||
		ds.upwardSeg.MinY() >= other.upwardSeg.MaxY() ||
		ds.upwardSeg.MaxY() <= other.upwardSeg.MinY() {
		return ds.upwardSeg.CompareTo(other.upwardSeg)
	}

	// Otherwise if envelopes overlap, use relative segment orientation.
	//
	// Collinear segments should be evaluated by previous logic
	orientIndex := ds.upwardSeg.OrientationIndexSegment(other.upwardSeg)
	if orientIndex != 0 {
		return orientIndex
	}

	// If comparison between this and other is indeterminate,
	// try the opposite call order.
	// The sign of the result needs to be flipped.
	orientIndex = -1 * other.upwardSeg.OrientationIndexSegment(ds.upwardSeg)
	if orientIndex != 0 {
		return orientIndex
	}

	// If segment envelopes overlap and they are collinear,
	// since segments do not cross they must be equal.
	// assert: segments are equal
	return 0
}

// oldCompareTo is dead code in Java but included for 1-1 correspondence.
func (ds *operationBuffer_DepthSegment) oldCompareTo(other *operationBuffer_DepthSegment) int {
	// fast check if segments are trivially ordered along X
	if ds.upwardSeg.MinX() > other.upwardSeg.MaxX() {
		return 1
	}
	if ds.upwardSeg.MaxX() < other.upwardSeg.MinX() {
		return -1
	}

	// try and compute a determinate orientation for the segments.
	// Test returns 1 if other is left of this (i.e. this > other)
	orientIndex := ds.upwardSeg.OrientationIndexSegment(other.upwardSeg)
	if orientIndex != 0 {
		return orientIndex
	}

	// If comparison between this and other is indeterminate,
	// try the opposite call order.
	// The sign of the result needs to be flipped.
	orientIndex = -1 * other.upwardSeg.OrientationIndexSegment(ds.upwardSeg)
	if orientIndex != 0 {
		return orientIndex
	}

	// otherwise, use standard lexicographic segment ordering
	return ds.upwardSeg.CompareTo(other.upwardSeg)
}

// String returns a string representation of the depth segment.
func (ds *operationBuffer_DepthSegment) String() string {
	return ds.upwardSeg.String()
}
