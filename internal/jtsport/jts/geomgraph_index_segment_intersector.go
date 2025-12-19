package jts

import "github.com/peterstace/simplefeatures/internal/jtsport/java"

// GeomgraphIndex_SegmentIntersector_IsAdjacentSegments returns true if the two
// segment indices are adjacent (differ by 1).
func GeomgraphIndex_SegmentIntersector_IsAdjacentSegments(i1, i2 int) bool {
	diff := i1 - i2
	if diff < 0 {
		diff = -diff
	}
	return diff == 1
}

// GeomgraphIndex_SegmentIntersector computes the intersection of line segments,
// and adds the intersection to the edges containing the segments.
type GeomgraphIndex_SegmentIntersector struct {
	child java.Polymorphic

	// These variables keep track of what types of intersections were
	// found during ALL edges that have been intersected.
	hasIntersection   bool
	hasProper         bool
	hasProperInterior bool

	// The proper intersection point found.
	properIntersectionPoint *Geom_Coordinate

	li                 *Algorithm_LineIntersector
	includeProper      bool
	recordIsolated     bool
	isSelfIntersection bool
	numIntersections   int

	// NumTests is for testing only.
	NumTests int

	bdyNodes [][]*Geomgraph_Node
}

// GetChild returns the immediate child in the type hierarchy chain.
func (si *GeomgraphIndex_SegmentIntersector) GetChild() java.Polymorphic {
	return si.child
}

// GetParent returns the immediate parent in the type hierarchy chain.
func (si *GeomgraphIndex_SegmentIntersector) GetParent() java.Polymorphic {
	return nil
}

// GeomgraphIndex_NewSegmentIntersector creates a new SegmentIntersector.
func GeomgraphIndex_NewSegmentIntersector(li *Algorithm_LineIntersector, includeProper, recordIsolated bool) *GeomgraphIndex_SegmentIntersector {
	return &GeomgraphIndex_SegmentIntersector{
		li:             li,
		includeProper:  includeProper,
		recordIsolated: recordIsolated,
	}
}

// SetBoundaryNodes sets the boundary nodes for both geometries.
func (si *GeomgraphIndex_SegmentIntersector) SetBoundaryNodes(bdyNodes0, bdyNodes1 []*Geomgraph_Node) {
	si.bdyNodes = make([][]*Geomgraph_Node, 2)
	si.bdyNodes[0] = bdyNodes0
	si.bdyNodes[1] = bdyNodes1
}

// IsDone returns whether processing is complete.
func (si *GeomgraphIndex_SegmentIntersector) IsDone() bool {
	return false
}

// GetProperIntersectionPoint returns the proper intersection point, or nil if
// none was found.
func (si *GeomgraphIndex_SegmentIntersector) GetProperIntersectionPoint() *Geom_Coordinate {
	return si.properIntersectionPoint
}

// HasIntersection returns true if any intersection was found.
func (si *GeomgraphIndex_SegmentIntersector) HasIntersection() bool {
	return si.hasIntersection
}

// HasProperIntersection returns true if a proper intersection was found.
// A proper intersection is an intersection which is interior to at least two
// line segments. Note that a proper intersection is not necessarily in the
// interior of the entire Geometry, since another edge may have an endpoint
// equal to the intersection, which according to SFS semantics can result in
// the point being on the Boundary of the Geometry.
func (si *GeomgraphIndex_SegmentIntersector) HasProperIntersection() bool {
	return si.hasProper
}

// HasProperInteriorIntersection returns true if a proper interior intersection
// was found. A proper interior intersection is a proper intersection which is
// not contained in the set of boundary nodes set for this SegmentIntersector.
func (si *GeomgraphIndex_SegmentIntersector) HasProperInteriorIntersection() bool {
	return si.hasProperInterior
}

// isTrivialIntersection checks if an intersection is trivial.
// A trivial intersection is an apparent self-intersection which in fact is
// simply the point shared by adjacent line segments. Note that closed edges
// require a special check for the point shared by the beginning and end
// segments.
func (si *GeomgraphIndex_SegmentIntersector) isTrivialIntersection(e0 *Geomgraph_Edge, segIndex0 int, e1 *Geomgraph_Edge, segIndex1 int) bool {
	if e0 == e1 {
		if si.li.GetIntersectionNum() == 1 {
			if GeomgraphIndex_SegmentIntersector_IsAdjacentSegments(segIndex0, segIndex1) {
				return true
			}
			if e0.IsClosed() {
				maxSegIndex := e0.GetNumPoints() - 1
				if (segIndex0 == 0 && segIndex1 == maxSegIndex) ||
					(segIndex1 == 0 && segIndex0 == maxSegIndex) {
					return true
				}
			}
		}
	}
	return false
}

// AddIntersections is called by clients of the EdgeIntersector class to test
// for and add intersections for two segments of the edges being intersected.
// Note that clients (such as MonotoneChainEdges) may choose not to intersect
// certain pairs of segments for efficiency reasons.
func (si *GeomgraphIndex_SegmentIntersector) AddIntersections(e0 *Geomgraph_Edge, segIndex0 int, e1 *Geomgraph_Edge, segIndex1 int) {
	if e0 == e1 && segIndex0 == segIndex1 {
		return
	}
	si.NumTests++
	p00 := e0.GetCoordinateAtIndex(segIndex0)
	p01 := e0.GetCoordinateAtIndex(segIndex0 + 1)
	p10 := e1.GetCoordinateAtIndex(segIndex1)
	p11 := e1.GetCoordinateAtIndex(segIndex1 + 1)

	si.li.ComputeIntersection(p00, p01, p10, p11)

	// Always record any non-proper intersections.
	// If includeProper is true, record any proper intersections as well.
	if si.li.HasIntersection() {
		if si.recordIsolated {
			e0.SetIsolated(false)
			e1.SetIsolated(false)
		}
		si.numIntersections++
		// If the segments are adjacent they have at least one trivial
		// intersection, the shared endpoint. Don't bother adding it if it is
		// the only intersection.
		if !si.isTrivialIntersection(e0, segIndex0, e1, segIndex1) {
			si.hasIntersection = true
			// In certain cases two line segments test as having a proper
			// intersection via the robust orientation check, but due to
			// roundoff the computed intersection point is equal to an
			// endpoint. If the endpoint is a boundary point the computed point
			// must be included as a node. If it is not a boundary point the
			// intersection is recorded as properInterior by logic below.
			isBoundaryPt := si.isBoundaryPoint(si.li, si.bdyNodes)
			isNotProper := !si.li.IsProper() || isBoundaryPt
			if si.includeProper || isNotProper {
				e0.AddIntersections(si.li, segIndex0, 0)
				e1.AddIntersections(si.li, segIndex1, 1)
			}
			if si.li.IsProper() {
				si.properIntersectionPoint = si.li.GetIntersection(0).Copy()
				si.hasProper = true
				if !isBoundaryPt {
					si.hasProperInterior = true
				}
			}
		}
	}
}

func (si *GeomgraphIndex_SegmentIntersector) isBoundaryPoint(li *Algorithm_LineIntersector, bdyNodes [][]*Geomgraph_Node) bool {
	if bdyNodes == nil {
		return false
	}
	if si.isBoundaryPointInternal(li, bdyNodes[0]) {
		return true
	}
	if si.isBoundaryPointInternal(li, bdyNodes[1]) {
		return true
	}
	return false
}

func (si *GeomgraphIndex_SegmentIntersector) isBoundaryPointInternal(li *Algorithm_LineIntersector, bdyNodes []*Geomgraph_Node) bool {
	for _, node := range bdyNodes {
		pt := node.GetCoordinate()
		if li.IsIntersection(pt) {
			return true
		}
	}
	return false
}
